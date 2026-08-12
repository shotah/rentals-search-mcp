package mcp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shotah/rentals-search-mcp/internal/rentcast"
)

// Host server id for ai-gantry is "rentals" → tools appear as rentals__{tool}.
// Tool names must NOT start with "rentals_" (no double prefix).
const ServerName = "rentals-search-mcp"

// ServerVersion is set at build time via ldflags.
var ServerVersion = "dev"

// RentalsAPI is the RentCast surface used by tools (mockable in tests).
type RentalsAPI interface {
	SearchListings(ctx context.Context, req rentcast.ListingsSearchRequest) (*rentcast.ListingsSearchResult, error)
	GetListing(ctx context.Context, id string) (*rentcast.ListingGetResult, error)
	RentEstimate(ctx context.Context, req rentcast.RentEstimateRequest) (*rentcast.RentEstimateResult, error)
	MarketStats(ctx context.Context, zipCode string) (*rentcast.MarketStatsResult, error)
	AccountUsage() *rentcast.Usage
}

// Server is the stdio MCP rental search surface.
type Server struct {
	Log    *log.Logger
	Client RentalsAPI
}

// New creates an MCP server. Client may be nil until Run (loaded from env).
func New(client RentalsAPI) *Server {
	return &Server{
		Log:    log.New(os.Stderr, "rentals-search-mcp: ", log.LstdFlags|log.Lmsgprefix),
		Client: client,
	}
}

// RegisteredToolNames returns tool names in registration order (for tests).
func RegisteredToolNames() []string {
	return []string{
		"listings_search",
		"listings_get",
		"rent_estimate_get",
		"markets_get",
		"areas_resolve",
		"link_format",
		"account_get",
	}
}

// newMCPServer builds the protocol server with tools registered.
func (s *Server) newMCPServer() *sdkmcp.Server {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "listings_search",
		Description: "Search long-term residential rentals (burns 1 RentCast request). Free tier ≈50/month — use sparingly. " +
			"THRIFTY: pass multiple neighborhoods as comma-separated neighborhood=Ballard,Fremont OR zip_codes=98107,98103 " +
			"with city+state for ONE call — never one search per area. " +
			"Filter tightly (bedrooms, price_max, new_this_week). Soft pets/parking/laundry prefs are not API filters. " +
			"Returns listing_url / contact handoff. Does not apply or contact landlords.",
	}, s.listingsSearch)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "listings_get",
		Description: "Get one listing by id (burns 1 RentCast request). Only after the human picks a candidate from search. " +
			"Do not prefetch many ids. Does not apply or contact landlords.",
	}, s.listingsGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "rent_estimate_get",
		Description: "Fair-rent AVM for one address (burns 1 RentCast request). Use sparingly when asking if a listed rent is reasonable. " +
			"Not for commercial spaces.",
	}, s.rentEstimateGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "markets_get",
		Description: "Zip rental market aggregates (burns 1 RentCast request). Optional context — skip if quota is tight; " +
			"prefer a tight listings_search instead.",
	}, s.marketsGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "areas_resolve",
		Description: "FREE (no RentCast call). Resolve Seattle neighborhood name(s) to zips/lat/lng. " +
			"Use list_all=true once, then ONE listings_search with neighborhood=A,B,C or zip_codes=… — not N searches.",
	}, s.areasResolve)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "link_format",
		Description: "FREE public rental search URL fallback (no RentCast call). Prefer when usage.requests_left is low.",
	}, s.linkFormat)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "account_get",
		Description: "FREE local usage counter (requests_used / requests_left / period_resets on the 1st). " +
			"Check before burning RentCast calls. Free tier ≈50/month (~1–2/day). Does not call the network.",
	}, s.accountGet)

	return server
}

// Run starts the MCP server over stdio. Logs go to stderr only.
func (s *Server) Run(ctx context.Context) error {
	if s.Client == nil {
		c, err := rentcast.NewFromEnv()
		if err != nil {
			return err
		}
		s.Client = c
	}
	return s.serve(ctx, &sdkmcp.StdioTransport{})
}

func (s *Server) serve(ctx context.Context, transport sdkmcp.Transport) error {
	server := s.newMCPServer()
	s.Log.Printf("starting stdio MCP (%s %s) tools=%s",
		ServerName, ServerVersion, strings.Join(RegisteredToolNames(), ","))
	return server.Run(ctx, transport)
}

// SelfTest validates tool registration and a dry (stub) client path — no API key.
func SelfTest() error {
	names := RegisteredToolNames()
	if len(names) == 0 {
		return errors.New("no tools registered")
	}
	for _, n := range names {
		if strings.HasPrefix(n, "rentals_") {
			return fmt.Errorf("tool %q must not start with rentals_ (double prefix)", n)
		}
		parts := strings.Split(n, "_")
		if len(parts) < 2 {
			return fmt.Errorf("tool %q must be service_verb…", n)
		}
	}

	stub := rentcast.NewStub()
	s := New(stub)
	_ = s.newMCPServer()

	ctx := context.Background()
	if _, err := stub.SearchListings(ctx, rentcast.ListingsSearchRequest{
		City: "Seattle", State: "WA", Limit: 5,
	}); err != nil {
		return fmt.Errorf("stub SearchListings: %w", err)
	}
	if _, err := stub.GetListing(ctx, "self-test-id"); err != nil {
		return fmt.Errorf("stub GetListing: %w", err)
	}
	if _, err := stub.RentEstimate(ctx, rentcast.RentEstimateRequest{
		Address: "1 Main St, Seattle, WA 98101",
	}); err != nil {
		return fmt.Errorf("stub RentEstimate: %w", err)
	}
	if _, err := stub.MarketStats(ctx, "98101"); err != nil {
		return fmt.Errorf("stub MarketStats: %w", err)
	}
	if pt := rentcast.NormalizePropertyType("apartment"); pt != "Apartment" {
		return fmt.Errorf("NormalizePropertyType: got %q", pt)
	}
	if _, err := rentcast.ResolveAreas(rentcast.AreaResolveRequest{Neighborhood: "Capitol Hill"}); err != nil {
		return fmt.Errorf("ResolveAreas: %w", err)
	}
	return nil
}
