package data

import "context"

// Client defines the Data API interface.
type Client interface {
	Health(ctx context.Context) (string, error)
	Positions(ctx context.Context, req *PositionsRequest) (PositionsResponse, error)
	Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error)
	Activity(ctx context.Context, req *ActivityRequest) (ActivityResponse, error)
	Holders(ctx context.Context, req *HoldersRequest) (HoldersResponse, error)
	Value(ctx context.Context, req *ValueRequest) (ValueResponse, error)
	ClosedPositions(ctx context.Context, req *ClosedPositionsRequest) (ClosedPositionsResponse, error)
	Traded(ctx context.Context, req *TradedRequest) (TradedResponse, error)
	OpenInterest(ctx context.Context, req *OpenInterestRequest) (OpenInterestResponse, error)
	LiveVolume(ctx context.Context, req *LiveVolumeRequest) (LiveVolumeResponse, error)
	Leaderboard(ctx context.Context, req *LeaderboardRequest) (LeaderboardResponse, error)
	BuildersLeaderboard(ctx context.Context, req *BuildersLeaderboardRequest) (BuildersLeaderboardResponse, error)
	BuildersVolume(ctx context.Context, req *BuildersVolumeRequest) (BuildersVolumeResponse, error)

	// V2 endpoints
	AccountingSnapshot(ctx context.Context) ([]byte, error)
	MarketPositions(ctx context.Context, req *MarketPositionsRequest) (MarketPositionsResponse, error)
	ActivityCombos(ctx context.Context, req *ActivityCombosRequest) (ActivityCombosResponse, error)
	PositionsCombos(ctx context.Context, req *PositionsCombosRequest) (PositionsCombosResponse, error)
}
