package heartbeat

import (
	"context"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
)

type Client interface {
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (HeartbeatResponse, error)
}

type clientImpl struct {
	httpClient *transport.Client
}

func NewClient(httpClient *transport.Client) Client {
	return &clientImpl{httpClient: httpClient}
}

func (c *clientImpl) Heartbeat(ctx context.Context, req *HeartbeatRequest) (HeartbeatResponse, error) {
	var resp HeartbeatResponse
	var body interface{}
	if req != nil && req.HeartbeatID != "" {
		body = map[string]string{"heartbeat_id": req.HeartbeatID}
	} else {
		body = map[string]interface{}{"heartbeat_id": nil}
	}
	err := c.httpClient.Post(ctx, "/v1/heartbeats", body, &resp)
	return resp, err
}
