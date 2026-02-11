package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/bot"
)

func main() {
	var (
		execute = flag.Bool("execute", false, "Enable live order submission (default false)")
		yes     = flag.Bool("yes", false, "Skip interactive confirmation when --execute is set")
		topN    = flag.Int("top", 5, "Top opportunities to print")
	)
	flag.Parse()

	cfg := bot.DefaultConfig()
	cfg.TopN = *topN
	cfg.AllowExecution = *execute
	cfg.DryRun = !*execute
	cfg.RequireExplicitConfirm = !*yes

	pk := strings.TrimSpace(os.Getenv("POLYMARKET_PK"))
	if pk == "" {
		log.Fatal("missing POLYMARKET_PK")
	}

	signer, err := auth.NewPrivateKeySigner(pk, 137)
	if err != nil {
		log.Fatalf("create signer failed: %v", err)
	}

	apiKey := &auth.APIKey{
		Key:        strings.TrimSpace(os.Getenv("POLYMARKET_API_KEY")),
		Secret:     strings.TrimSpace(os.Getenv("POLYMARKET_API_SECRET")),
		Passphrase: strings.TrimSpace(os.Getenv("POLYMARKET_API_PASSPHRASE")),
	}
	if apiKey.Key == "" || apiKey.Secret == "" || apiKey.Passphrase == "" {
		log.Fatal("missing POLYMARKET_API_KEY / POLYMARKET_API_SECRET / POLYMARKET_API_PASSPHRASE")
	}

	client := polymarket.NewClient().CLOB.WithAuth(signer, apiKey)
	engine, err := bot.NewEngine(client, signer, cfg)
	if err != nil {
		log.Fatalf("create engine failed: %v", err)
	}

	ctx := context.Background()
	opps, err := engine.ScanOpportunities(ctx)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}
	if len(opps) == 0 {
		fmt.Println("No opportunities matched filters.")
		return
	}

	fmt.Printf("Found %d opportunities\n", len(opps))
	for i, op := range opps {
		fmt.Printf("[%d] %s | %s (%s) | bid=%s ask=%s spread=%s bps imbalance=%s rec=%s conf=%s bps\n",
			i,
			trim(op.Question, 90),
			op.Outcome,
			op.TokenID,
			op.Bid.StringFixed(4),
			op.Ask.StringFixed(4),
			op.SpreadBps.StringFixed(2),
			op.Imbalance.StringFixed(4),
			op.Recommended,
			op.ConfidenceBps.StringFixed(2),
		)
	}

	sort.Slice(opps, func(i, j int) bool { return opps[i].SignalScore.GreaterThan(opps[j].SignalScore) })
	best := opps[0]
	plan, err := engine.BuildTradePlan(best)
	if err != nil {
		log.Fatalf("plan failed: %v", err)
	}

	risk, err := engine.EvaluateRisk(ctx)
	if err != nil {
		log.Fatalf("risk evaluation failed: %v", err)
	}
	if err := engine.ValidatePlanAgainstRisk(plan, risk); err != nil {
		log.Fatalf("plan blocked by risk: %v", err)
	}

	fmt.Printf("\nPlanned order: side=%s token=%s amount=%s USDC price_guard=%s reason=%s\n",
		plan.Side,
		plan.TokenID,
		plan.AmountUSDC.StringFixed(2),
		plan.MaxAcceptedPrice.StringFixed(4),
		plan.Reason,
	)

	if cfg.DryRun || !cfg.AllowExecution {
		fmt.Println("Dry run mode: no order submitted. Use --execute to place a live order.")
		return
	}
	if cfg.RequireExplicitConfirm {
		if !confirm("Submit live order now? [y/N]: ") {
			fmt.Println("Canceled.")
			return
		}
	}

	resp, err := engine.ExecutePlan(ctx, plan)
	if err != nil {
		log.Fatalf("execute failed: %v", err)
	}
	fmt.Printf("Order submitted: id=%s status=%s\n", resp.ID, resp.Status)
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
