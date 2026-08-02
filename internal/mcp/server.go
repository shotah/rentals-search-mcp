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
	GetListing(ctx context.Context, id string) (*rentcast.Listing, error)
	RentEstimate(ctx context.Context, req rentcast.RentEstimateRequest) (*rentcast.RentEstimateResult, error)
	MarketStats(ctx context.Context, zipCode string) (*rentcast.MarketStatsResult, error)
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
		Description: "Search long-term residential rental listings (apartments, houses, condos, townhomes) " +
			"by city/state, zip, or lat/lng+radius. Filter by bedrooms, bathrooms, rent, square footage, " +
			"and property_type. Returns ranked listing summaries with listing_url / contact handoff fields. " +
			"Does not apply or contact landlords. Not for retail/office/commercial leases.",
	}, s.listingsSearch)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "listings_get",
		Description: "Get one long-term rental listing by id from listings_search. " +
			"Use after the human picks a candidate. Does not apply or contact landlords.",
	}, s.listingsGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "rent_estimate_get",
		Description: "Estimate fair long-term rent for a US residential address (AVM) with comparable " +
			"properties when available. Use when asking whether a listed rent is reasonable. " +
			"Not for commercial retail spaces.",
	}, s.rentEstimateGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "markets_get",
		Description: "Get aggregate rental market statistics for a US zip code " +
			"(averages, listing trends). Useful context before or after listings_search.",
	}, s.marketsGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "link_format",
		Description: "Build a public rental SEARCH URL (fallback) from city/state/zip and filters. " +
			"No API call and no application. Prefer listings_search when quota allows.",
	}, s.linkFormat)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name: "account_get",
		Description: "Remind the agent about RentCast quota (≈50 free requests/month on Developer) " +
			"and point at the dashboard. RentCast has no public usage JSON like SerpAPI — " +
			"this does not call the network.",
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
	return nil
}
