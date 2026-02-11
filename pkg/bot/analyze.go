package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/shopspring/decimal"
)

func (e *Engine) analyzeToken(ctx context.Context, m clobtypes.Market, t clobtypes.MarketToken) (Opportunity, error) {
	book, err := e.client.OrderBook(ctx, &clobtypes.BookRequest{TokenID: t.TokenID})
	if err != nil {
		return Opportunity{}, err
	}
	if len(book.Bids) == 0 || len(book.Asks) == 0 {
		return Opportunity{}, fmt.Errorf("empty book")
	}

	bestBid, bidDepth, err := topOfBook(book.Bids)
	if err != nil {
		return Opportunity{}, err
	}
	bestAsk, askDepth, err := topOfBook(book.Asks)
	if err != nil {
		return Opportunity{}, err
	}
	if bestAsk.LessThanOrEqual(bestBid) {
		return Opportunity{}, fmt.Errorf("crossed/invalid book")
	}

	mid := bestBid.Add(bestAsk).Div(decimal.NewFromInt(2))
	spread := bestAsk.Sub(bestBid)
	spreadBps := decimal.Zero
	if !mid.IsZero() {
		spreadBps = spread.Div(mid).Mul(decimal.NewFromInt(10000))
	}

	totalDepth := bidDepth.Add(askDepth)
	imbalance := decimal.Zero
	if totalDepth.GreaterThan(decimal.Zero) {
		imbalance = bidDepth.Sub(askDepth).Div(totalDepth)
	}

	recommended := "HOLD"
	if imbalance.GreaterThan(decimal.Zero) {
		recommended = "BUY"
	} else if imbalance.LessThan(decimal.Zero) {
		recommended = "SELL"
	}

	confidenceBps := imbalance.Abs().Mul(decimal.NewFromInt(10000)).Mul(spreadBps.Div(decimal.NewFromInt(100)))
	score := confidenceBps

	return Opportunity{
		MarketID:      m.ID,
		Question:      m.Question,
		TokenID:       t.TokenID,
		Outcome:       t.Outcome,
		Bid:           bestBid,
		Ask:           bestAsk,
		Mid:           mid,
		Spread:        spread,
		SpreadBps:     spreadBps,
		BidDepth:      bidDepth,
		AskDepth:      askDepth,
		Imbalance:     imbalance,
		SignalScore:   score,
		Recommended:   strings.ToUpper(recommended),
		ConfidenceBps: confidenceBps,
	}, nil
}

func topOfBook(levels []clobtypes.PriceLevel) (price decimal.Decimal, depth decimal.Decimal, err error) {
	price, err = decimal.NewFromString(levels[0].Price)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("bad price: %w", err)
	}
	depth, err = decimal.NewFromString(levels[0].Size)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("bad size: %w", err)
	}
	return price, depth, nil
}
