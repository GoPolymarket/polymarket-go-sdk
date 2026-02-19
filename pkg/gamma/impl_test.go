package gamma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

type staticDoer struct {
	responses map[string]string
}

func (d *staticDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	payload, ok := d.responses[key]
	if !ok {
		return nil, fmt.Errorf("unexpected request %q", key)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func TestGammaMethods(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/status":             `"OK"`,
			"/teams":              `[]`,
			"/sports":             `[]`,
			"/sports/market-types": `{"sports":[]}`,
			"/tags":               `[]`,
			"/tags/1":             `{"id":"1","name":"tag1"}`,
			"/tags/slug/tag-slug": `{"id":"1","name":"tag1"}`,
			"/tags/1/related-tags": `[]`,
			"/tags/slug/tag-slug/related-tags": `[]`,
			"/tags/1/related-tags/tags": `[]`,
			"/events":             `[]`,
			"/events/1":           `{"id":"1","title":"event1"}`,
			"/events/slug/slug1":  `{"id":"1","title":"event1"}`,
			"/events/1/tags":      `[]`,
			"/markets":            `[]`,
			"/markets/1":          `{"id":"1","question":"market1"}`,
			"/markets/slug/slug1": `{"id":"1","question":"market1"}`,
			"/markets/1/tags":     `[]`,
			"/series":             `[]`,
			"/series/1":           `{"id":"1"}`,
			"/comments":           `[]`,
			"/comments/1":         `[]`,
			"/comments/user_address/0x123": `[]`,
			"/public-profile?address=0x123": `{"id":"1"}`,
			"/public-search?q=test": `{"events":[],"markets":[]}`,
		},
	}
	client := NewClient(transport.NewClient(doer, BaseURL))
	ctx := context.Background()

	t.Run("Status", func(t *testing.T) {
		resp, err := client.Status(ctx)
		if err != nil || string(resp) != "OK" {
			t.Errorf("Status failed: %v", err)
		}
	})

	t.Run("Sports", func(t *testing.T) {
		_, err := client.Sports(ctx)
		if err != nil {
			t.Errorf("Sports failed: %v", err)
		}
	})

	t.Run("SportsMarketTypes", func(t *testing.T) {
		_, err := client.SportsMarketTypes(ctx)
		if err != nil {
			t.Errorf("SportsMarketTypes failed: %v", err)
		}
	})

	t.Run("Tags", func(t *testing.T) {
		_, err := client.Tags(ctx, nil)
		if err != nil {
			t.Errorf("Tags failed: %v", err)
		}
	})

	t.Run("TagByID", func(t *testing.T) {
		resp, err := client.TagByID(ctx, &TagByIDRequest{ID: "1"})
		if err != nil || resp.ID != "1" {
			t.Errorf("TagByID failed: %v", err)
		}
	})

	t.Run("TagBySlug", func(t *testing.T) {
		resp, err := client.TagBySlug(ctx, &TagBySlugRequest{Slug: "tag-slug"})
		if err != nil || resp.ID != "1" {
			t.Errorf("TagBySlug failed: %v", err)
		}
	})

	t.Run("RelatedTags", func(t *testing.T) {
		_, _ = client.RelatedTagsByID(ctx, &RelatedTagsByIDRequest{ID: "1"})
		_, _ = client.RelatedTagsBySlug(ctx, &RelatedTagsBySlugRequest{Slug: "tag-slug"})
		_, _ = client.TagsRelatedToTagByID(ctx, &RelatedTagsByIDRequest{ID: "1"})
		_, _ = client.TagsRelatedToTagBySlug(ctx, &RelatedTagsBySlugRequest{Slug: "tag-slug"})
	})

	t.Run("Events", func(t *testing.T) {
		_, err := client.Events(ctx, nil)
		if err != nil {
			t.Errorf("Events failed: %v", err)
		}
		_, _ = client.EventByID(ctx, &EventByIDRequest{ID: "1"})
		_, _ = client.EventBySlug(ctx, &EventBySlugRequest{Slug: "slug1"})
		_, _ = client.EventTags(ctx, &EventTagsRequest{ID: "1"})
	})

	t.Run("Markets", func(t *testing.T) {
		_, err := client.Markets(ctx, nil)
		if err != nil {
			t.Errorf("Markets failed: %v", err)
		}
		_, _ = client.MarketByID(ctx, &MarketByIDRequest{ID: "1"})
		_, _ = client.MarketBySlug(ctx, &MarketBySlugRequest{Slug: "slug1"})
		_, _ = client.MarketTags(ctx, &MarketTagsRequest{ID: "1"})
	})

	t.Run("Series", func(t *testing.T) {
		_, _ = client.Series(ctx, nil)
		_, _ = client.SeriesByID(ctx, &SeriesByIDRequest{ID: "1"})
	})

	t.Run("Comments", func(t *testing.T) {
		_, _ = client.Comments(ctx, nil)
		_, _ = client.CommentByID(ctx, &CommentByIDRequest{ID: "1"})
		_, _ = client.CommentsByUserAddress(ctx, &CommentsByUserAddressRequest{UserAddress: "0x123"})
	})

	t.Run("PublicProfile", func(t *testing.T) {
		_, err := client.PublicProfile(ctx, &PublicProfileRequest{Address: "0x123"})
		if err != nil {
			t.Errorf("PublicProfile failed: %v", err)
		}
	})

	t.Run("PublicSearch", func(t *testing.T) {
		_, _ = client.PublicSearch(ctx, &PublicSearchRequest{Query: "test"})
	})

	t.Run("PaginationAll", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{
				"/markets?limit=1&offset=0": `[{"id":"1"}]`,
				"/markets?limit=1&offset=1": `[]`,
				"/events?limit=1&offset=0":  `[{"id":"1"}]`,
				"/events?limit=1&offset=1":  `[]`,
			},
		}
		client := NewClient(transport.NewClient(doer, BaseURL))
		limit := 1
		
		markets, _ := client.MarketsAll(ctx, &MarketsRequest{Limit: &limit})
		if len(markets) != 1 {
			t.Errorf("expected 1 market")
		}

		events, _ := client.EventsAll(ctx, &EventsRequest{Limit: &limit})
		if len(events) != 1 {
			t.Errorf("expected 1 event")
		}
	})

	t.Run("LegacyAliases", func(t *testing.T) {
		_, _ = client.GetMarkets(ctx, nil)
		_, _ = client.GetMarket(ctx, "1")
		_, _ = client.GetEvents(ctx, nil)
		_, _ = client.GetEvent(ctx, "1")
	})
}

func TestMarket_NegRiskFields(t *testing.T) {
	raw := `{
		"id": "m1",
		"question": "Will X happen?",
		"conditionId": "0xcond",
		"negRisk": true,
		"negRiskMarketId": "0xneg",
		"enableOrderBook": true,
		"questionId": "0xq",
		"volume24hr": "1000000",
		"spread": "0.02",
		"bestBid": "0.48",
		"bestAsk": "0.52",
		"lastTradePrice": "0.50",
		"commentCount": 42,
		"cyom": false
	}`

	var m Market
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !m.NegRisk {
		t.Error("expected NegRisk=true")
	}
	if m.NegRiskMarketID != "0xneg" {
		t.Errorf("NegRiskMarketID = %s, want 0xneg", m.NegRiskMarketID)
	}
	if !m.EnableOrderBook {
		t.Error("expected EnableOrderBook=true")
	}
	if m.Volume24hr != "1000000" {
		t.Errorf("Volume24hr = %s, want 1000000", m.Volume24hr)
	}
	if m.BestBid != "0.48" {
		t.Errorf("BestBid = %s, want 0.48", m.BestBid)
	}
	if m.CommentCount != 42 {
		t.Errorf("CommentCount = %d, want 42", m.CommentCount)
	}
}

func TestEvent_NegRiskFields(t *testing.T) {
	raw := `{
		"id": "e1",
		"title": "Election 2024",
		"negRisk": true,
		"enableNegRisk": true,
		"negRiskAugmented": false,
		"commentCount": 100,
		"competitionState": "active",
		"cyom": true
	}`

	var e Event
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !e.NegRisk {
		t.Error("expected NegRisk=true")
	}
	if !e.EnableNegRisk {
		t.Error("expected EnableNegRisk=true")
	}
	if e.CommentCount != 100 {
		t.Errorf("CommentCount = %d, want 100", e.CommentCount)
	}
	if e.CompetitionState != "active" {
		t.Errorf("CompetitionState = %s, want active", e.CompetitionState)
	}
}
