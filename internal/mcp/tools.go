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
	ZipCodes      string  `json:"zip_codes,omitempty" jsonschema:"Comma/pipe-separated zips for multi-zip search (client filter when city+state set)"`
	Neighborhood  string  `json:"neighborhood,omitempty" jsonschema:"Local preset e.g. Capitol Hill / Ballard (Seattle); expands to lat/lng or zips"`
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
	DaysOld       string  `json:"days_old,omitempty" jsonschema:"Max listing age in days (RentCast daysOld); e.g. 7 or *:7"`
	DaysOldMax    int     `json:"days_old_max,omitempty" jsonschema:"Max days on market (shorthand for days_old)"`
	NewThisWeek   bool    `json:"new_this_week,omitempty" jsonschema:"Only listings ≤7 days old (sets days_old_max=7)"`
	PetsWanted    bool    `json:"pets_wanted,omitempty" jsonschema:"Soft preference only — RentCast cannot filter pets; verify on listing_url"`
	ParkingWanted bool    `json:"parking_wanted,omitempty" jsonschema:"Soft preference only — RentCast cannot filter parking; verify on listing_url"`
	LaundryWanted bool    `json:"laundry_wanted,omitempty" jsonschema:"Soft preference only — RentCast cannot filter laundry; verify on listing_url"`
	Status        string  `json:"status,omitempty" jsonschema:"Active (default) or Inactive"`
	Limit         int     `json:"limit,omitempty" jsonschema:"Page size (default 10, max 50)"`
	Offset        int     `json:"offset,omitempty" jsonschema:"Pagination offset"`
}

type areasResolveInput struct {
	Neighborhood string `json:"neighborhood,omitempty" jsonschema:"Neighborhood name or alias (e.g. Capitol Hill, u district)"`
	City         string `json:"city,omitempty" jsonschema:"Optional city filter (e.g. Seattle)"`
	State        string `json:"state,omitempty" jsonschema:"Optional 2-letter state filter"`
	ListAll      bool   `json:"list_all,omitempty" jsonschema:"List all known presets (optionally filtered by city/state)"`
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
	City          string `json:"city,omitempty" jsonschema:"City name"`
	State         string `json:"state,omitempty" jsonschema:"2-letter state"`
	ZipCode       string `json:"zip_code,omitempty" jsonschema:"Zip code"`
	Neighborhood  string `json:"neighborhood,omitempty" jsonschema:"Optional neighborhood for search term"`
	Bedrooms      string `json:"bedrooms,omitempty" jsonschema:"Optional bedrooms"`
	PriceMax      int    `json:"price_max,omitempty" jsonschema:"Optional max monthly rent"`
	PropertyType  string `json:"property_type,omitempty" jsonschema:"apartment|house|…"`
	PetsWanted    bool   `json:"pets_wanted,omitempty" jsonschema:"Hint for public search URL only"`
	ParkingWanted bool   `json:"parking_wanted,omitempty" jsonschema:"Hint noted in response; not all sites support URL filters"`
	LaundryWanted bool   `json:"laundry_wanted,omitempty" jsonschema:"Hint noted in response; not all sites support URL filters"`
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
		ZipCodes:      in.ZipCodes,
		Neighborhood:  in.Neighborhood,
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
		DaysOld:       in.DaysOld,
		DaysOldMax:    in.DaysOldMax,
		NewThisWeek:   in.NewThisWeek,
		PetsWanted:    in.PetsWanted,
		ParkingWanted: in.ParkingWanted,
		LaundryWanted: in.LaundryWanted,
		Status:        in.Status,
		Limit:         in.Limit,
		Offset:        in.Offset,
	})
	if err != nil {
		return errResult(err.Error()), nil, nil
	}
	return jsonResult(res)
}

func (s *Server) areasResolve(_ context.Context, _ *sdkmcp.CallToolRequest, in areasResolveInput) (*sdkmcp.CallToolResult, any, error) {
	res, err := rentcast.ResolveAreas(rentcast.AreaResolveRequest{
		Neighborhood: in.Neighborhood,
		City:         in.City,
		State:        in.State,
		ListAll:      in.ListAll,
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
	parts := make([]string, 0, 4)
	if n := strings.TrimSpace(in.Neighborhood); n != "" {
		parts = append(parts, n)
	}
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
		return errResult("provide city+state and/or zip_code (neighborhood optional)"), nil, nil
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
	if in.PetsWanted {
		u += "&pets=true"
	}
	_ = in.PropertyType
	note := "Fallback public search URL only. Prefer listings_search when API quota allows."
	if in.PetsWanted || in.ParkingWanted || in.LaundryWanted {
		note += " Amenity URL hints are best-effort; confirm on the site. RentCast cannot filter pets/parking/laundry."
	}
	return jsonResult(map[string]any{
		"search_url":     u,
		"location":       loc,
		"pets_wanted":    in.PetsWanted,
		"parking_wanted": in.ParkingWanted,
		"laundry_wanted": in.LaundryWanted,
		"note":           note,
	})
}

func (s *Server) accountGet(_ context.Context, _ *sdkmcp.CallToolRequest, _ accountGetInput) (*sdkmcp.CallToolResult, any, error) {
	out := map[string]any{
		"provider":       "RentCast",
		"dashboard_url":  "https://app.rentcast.io/",
		"free_tier_note": "Developer plan includes about 50 API requests per month.",
		"quota_api":      false,
		"note":           "RentCast has no public account API. usage below is a local counter of successful calls from this binary (calendar month). Verify against the dashboard. link_format / areas_resolve / account_get do not burn RentCast requests.",
	}
	if s.Client != nil {
		if u := s.Client.AccountUsage(); u != nil {
			out["usage"] = u
		}
	}
	return jsonResult(out)
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
