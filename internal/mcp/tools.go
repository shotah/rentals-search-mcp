package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shotah/rentals-search-mcp/internal/rentcast"
)

type listingsSearchInput struct {
	City          string  `json:"city,omitempty" jsonschema:"City name (case-sensitive upstream; e.g. Seattle)"`
	State         string  `json:"state,omitempty" jsonschema:"2-letter state (e.g. WA)"`
	ZipCode       string  `json:"zip_code,omitempty" jsonschema:"5-digit US zip code"`
	Address       string  `json:"address,omitempty" jsonschema:"Full address Street, City, State, Zip"`
	Latitude      float64 `json:"latitude,omitempty" jsonschema:"Search center latitude"`
	Longitude     float64 `json:"longitude,omitempty" jsonschema:"Search center longitude"`
	Radius        float64 `json:"radius,omitempty" jsonschema:"Radius in miles (max 100); use with lat/lng or address"`
	PropertyType  string  `json:"property_type,omitempty" jsonschema:"apartment|house|condo|townhouse|manufactured|multi_family or RentCast type"`
	Bedrooms      string  `json:"bedrooms,omitempty" jsonschema:"Bedrooms; 0=studio; ranges like 1:2 ok"`
	Bathrooms     string  `json:"bathrooms,omitempty" jsonschema:"Bathrooms; ranges ok"`
	SquareFootage string  `json:"square_footage,omitempty" jsonschema:"Living area sqft; ranges ok"`
	PriceMin      int     `json:"price_min,omitempty" jsonschema:"Minimum monthly rent USD"`
	PriceMax      int     `json:"price_max,omitempty" jsonschema:"Maximum monthly rent USD"`
	Status        string  `json:"status,omitempty" jsonschema:"Active (default) or Inactive"`
	Limit         int     `json:"limit,omitempty" jsonschema:"Page size (default 10, max 50)"`
	Offset        int     `json:"offset,omitempty" jsonschema:"Pagination offset"`
}

type listingsGetInput struct {
	ListingID string `json:"listing_id" jsonschema:"RentCast listing id from listings_search"`
}

type rentEstimateInput struct {
	Address       string `json:"address" jsonschema:"Full address Street, City, State, Zip"`
	PropertyType  string `json:"property_type,omitempty" jsonschema:"Optional property type hint"`
	Bedrooms      string `json:"bedrooms,omitempty" jsonschema:"Optional bedrooms (0=studio)"`
	Bathrooms     string `json:"bathrooms,omitempty" jsonschema:"Optional bathrooms"`
	SquareFootage string `json:"square_footage,omitempty" jsonschema:"Optional living area sqft"`
}

type marketsGetInput struct {
	ZipCode string `json:"zip_code" jsonschema:"5-digit US zip code"`
}

type linkFormatInput struct {
	City         string `json:"city,omitempty" jsonschema:"City name"`
	State        string `json:"state,omitempty" jsonschema:"2-letter state"`
	ZipCode      string `json:"zip_code,omitempty" jsonschema:"Zip code"`
	Bedrooms     string `json:"bedrooms,omitempty" jsonschema:"Optional bedrooms"`
	PriceMax     int    `json:"price_max,omitempty" jsonschema:"Optional max monthly rent"`
	PropertyType string `json:"property_type,omitempty" jsonschema:"apartment|house|…"`
}

type accountGetInput struct{}

func (s *Server) listingsSearch(ctx context.Context, _ *sdkmcp.CallToolRequest, in listingsSearchInput) (*sdkmcp.CallToolResult, any, error) {
	if s.Client == nil {
		return errResult("client not configured"), nil, nil
	}
	res, err := s.Client.SearchListings(ctx, rentcast.ListingsSearchRequest{
		City:          in.City,
		State:         in.State,
		ZipCode:       in.ZipCode,
		Address:       in.Address,
		Latitude:      in.Latitude,
		Longitude:     in.Longitude,
		Radius:        in.Radius,
		PropertyType:  in.PropertyType,
		Bedrooms:      in.Bedrooms,
		Bathrooms:     in.Bathrooms,
		SquareFootage: in.SquareFootage,
		PriceMin:      in.PriceMin,
		PriceMax:      in.PriceMax,
		Status:        in.Status,
		Limit:         in.Limit,
		Offset:        in.Offset,
	})
	if err != nil {
		return errResult(err.Error()), nil, nil
	}
	return jsonResult(res)
}

func (s *Server) listingsGet(ctx context.Context, _ *sdkmcp.CallToolRequest, in listingsGetInput) (*sdkmcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.ListingID) == "" {
		return errResult("listing_id is required"), nil, nil
	}
	if s.Client == nil {
		return errResult("client not configured"), nil, nil
	}
	res, err := s.Client.GetListing(ctx, in.ListingID)
	if err != nil {
		return errResult(err.Error()), nil, nil
	}
	return jsonResult(res)
}

func (s *Server) rentEstimateGet(ctx context.Context, _ *sdkmcp.CallToolRequest, in rentEstimateInput) (*sdkmcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.Address) == "" {
		return errResult("address is required"), nil, nil
	}
	if s.Client == nil {
		return errResult("client not configured"), nil, nil
	}
	res, err := s.Client.RentEstimate(ctx, rentcast.RentEstimateRequest{
		Address:       in.Address,
		PropertyType:  in.PropertyType,
		Bedrooms:      in.Bedrooms,
		Bathrooms:     in.Bathrooms,
		SquareFootage: in.SquareFootage,
	})
	if err != nil {
		return errResult(err.Error()), nil, nil
	}
	return jsonResult(res)
}

func (s *Server) marketsGet(ctx context.Context, _ *sdkmcp.CallToolRequest, in marketsGetInput) (*sdkmcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.ZipCode) == "" {
		return errResult("zip_code is required"), nil, nil
	}
	if s.Client == nil {
		return errResult("client not configured"), nil, nil
	}
	res, err := s.Client.MarketStats(ctx, in.ZipCode)
	if err != nil {
		return errResult(err.Error()), nil, nil
	}
	return jsonResult(res)
}

func (s *Server) linkFormat(_ context.Context, _ *sdkmcp.CallToolRequest, in linkFormatInput) (*sdkmcp.CallToolResult, any, error) {
	parts := make([]string, 0, 3)
	if c := strings.TrimSpace(in.City); c != "" {
		parts = append(parts, c)
	}
	if st := strings.TrimSpace(in.State); st != "" {
		parts = append(parts, st)
	}
	if z := strings.TrimSpace(in.ZipCode); z != "" {
		parts = append(parts, z)
	}
	loc := strings.Join(parts, " ")
	if loc == "" {
		return errResult("provide city+state and/or zip_code"), nil, nil
	}
	// Public fallback: Zillow rental search (human click-around). Not a checkout.
	u := "https://www.zillow.com/homes/for_rent/?" + url.Values{
		"searchQueryState": {fmt.Sprintf(`{"usersSearchTerm":%q}`, loc)},
	}.Encode()
	if in.Bedrooms != "" {
		u += "&beds=" + url.QueryEscape(in.Bedrooms)
	}
	if in.PriceMax > 0 {
		u += "&price_max=" + url.QueryEscape(strconv.Itoa(in.PriceMax))
	}
	_ = in.PropertyType
	return jsonResult(map[string]any{
		"search_url": u,
		"location":   loc,
		"note":       "Fallback public search URL only. Prefer listings_search when API quota allows.",
	})
}

func (s *Server) accountGet(_ context.Context, _ *sdkmcp.CallToolRequest, _ accountGetInput) (*sdkmcp.CallToolResult, any, error) {
	return jsonResult(map[string]any{
		"provider":       "RentCast",
		"dashboard_url":  "https://app.rentcast.io/",
		"free_tier_note": "Developer plan includes about 50 API requests per month.",
		"quota_api":      false,
		"note":           "RentCast has no public account/usage JSON. Check the dashboard for remaining quota. link_format does not burn requests; each other tool call does.",
	})
}

func errResult(msg string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
		IsError: true,
	}
}

func jsonResult(v any) (*sdkmcp.CallToolResult, any, error) {
	return &sdkmcp.CallToolResult{}, v, nil
}
