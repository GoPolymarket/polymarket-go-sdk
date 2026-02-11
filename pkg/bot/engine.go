package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/shopspring/decimal"
)

// Engine contains end-to-end scan/analyze/plan/execute flow.
type Engine struct {
	client clob.Client
	signer auth.Signer
	cfg    Config
}

func NewEngine(client clob.Client, signer auth.Signer, cfg Config) (*Engine, error) {
	cfg = cfg.MergeEnv()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Engine{client: client, signer: signer, cfg: cfg}, nil
}

func (e *Engine) ScanOpportunities(ctx context.Context) ([]Opportunity, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	marketsResp, err := e.client.Markets(ctx, &clobtypes.MarketsRequest{Active: boolPtr(true), Limit: e.cfg.ScanLimit})
	if err != nil {
		return nil, fmt.Errorf("markets query failed: %w", err)
	}

	opps := make([]Opportunity, 0, len(marketsResp.Data)*2)
	for _, m := range marketsResp.Data {
		for _, t := range m.Tokens {
			o, err := e.analyzeToken(ctx, m, t)
			if err != nil {
				continue
			}
			if o.SpreadBps.LessThan(e.cfg.MinSpreadBps) {
				continue
			}
			if o.BidDepth.LessThan(e.cfg.MinBookDepthShares) || o.AskDepth.LessThan(e.cfg.MinBookDepthShares) {
				continue
			}
			if o.Imbalance.Abs().LessThan(e.cfg.MinImbalance) {
				continue
			}
			opps = append(opps, o)
		}
	}

	sort.Slice(opps, func(i, j int) bool {
		return opps[i].SignalScore.GreaterThan(opps[j].SignalScore)
	})

	if len(opps) > e.cfg.TopN {
		return opps[:e.cfg.TopN], nil
	}
	return opps, nil
}

func (e *Engine) BuildTradePlan(op Opportunity) (*TradePlan, error) {
	if op.ConfidenceBps.LessThan(e.cfg.MinConfidenceBps) {
		return nil, fmt.Errorf("confidence too low: %s bps", op.ConfidenceBps.StringFixed(2))
	}

	amount := decimal.Min(e.cfg.DefaultAmountUSDC, e.cfg.MaxPerTradeUSDC)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("invalid amount after risk caps")
	}

	priceGuard := slippageGuardPrice(op, e.cfg.MaxSlippageBps)

	plan := &TradePlan{
		TokenID:           op.TokenID,
		Side:              strings.ToUpper(op.Recommended),
		AmountUSDC:        amount,
		MaxAcceptedPrice:  priceGuard,
		ExpectedMid:       op.Mid,
		MaxSlippageBps:    e.cfg.MaxSlippageBps,
		Reason:            fmt.Sprintf("imbalance=%s spread=%sbps", op.Imbalance.StringFixed(4), op.SpreadBps.StringFixed(2)),
		OpportunitySource: op,
	}
	return plan, nil
}

func (e *Engine) ExecutePlan(ctx context.Context, plan *TradePlan) (clobtypes.OrderResponse, error) {
	if plan == nil {
		return clobtypes.OrderResponse{}, fmt.Errorf("plan is required")
	}
	if e.cfg.DryRun || !e.cfg.AllowExecution {
		return clobtypes.OrderResponse{}, fmt.Errorf("execution disabled (dry-run=%t allow-execution=%t)", e.cfg.DryRun, e.cfg.AllowExecution)
	}

	ctx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	builder := clob.NewOrderBuilder(e.client, e.signer).
		TokenID(plan.TokenID).
		Side(plan.Side).
		AmountUSDC(plan.AmountUSDC.InexactFloat64()).
		PriceDec(plan.MaxAcceptedPrice).
		OrderType(clobtypes.OrderTypeFAK)

	signable, err := builder.BuildMarketWithContext(ctx)
	if err != nil {
		return clobtypes.OrderResponse{}, fmt.Errorf("build market order failed: %w", err)
	}

	return e.client.CreateOrderFromSignable(ctx, signable)
}

func boolPtr(v bool) *bool { return &v }

func slippageGuardPrice(op Opportunity, maxSlippageBps decimal.Decimal) decimal.Decimal {
	delta := op.Mid.Mul(maxSlippageBps).Div(decimal.NewFromInt(10000))
	if strings.EqualFold(op.Recommended, "BUY") {
		return op.Mid.Add(delta)
	}
	guard := op.Mid.Sub(delta)
	if guard.LessThanOrEqual(decimal.Zero) {
		return decimal.RequireFromString("0.01")
	}
	return guard
}
