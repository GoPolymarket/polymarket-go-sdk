package bot

import "github.com/shopspring/decimal"

// Opportunity is a tradable candidate ranked by the analyzer.
type Opportunity struct {
	MarketID      string
	Question      string
	TokenID       string
	Outcome       string
	Bid           decimal.Decimal
	Ask           decimal.Decimal
	Mid           decimal.Decimal
	Spread        decimal.Decimal
	SpreadBps     decimal.Decimal
	BidDepth      decimal.Decimal
	AskDepth      decimal.Decimal
	Imbalance     decimal.Decimal
	SignalScore   decimal.Decimal
	Recommended   string
	ConfidenceBps decimal.Decimal
}

// TradePlan is a fully-specified order instruction.
type TradePlan struct {
	TokenID           string
	Side              string
	AmountUSDC        decimal.Decimal
	MaxAcceptedPrice  decimal.Decimal
	ExpectedMid       decimal.Decimal
	MaxSlippageBps    decimal.Decimal
	Reason            string
	OpportunitySource Opportunity
}
