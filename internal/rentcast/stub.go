package rentcast

import "context"

// Stub is an in-memory client for self-test / unit tests (no network).
type Stub struct{}

// NewStub returns a non-nil RentalsAPI that returns canned empty results.
func NewStub() *Stub { return &Stub{} }

func (Stub) SearchListings(_ context.Context, req ListingsSearchRequest) (*ListingsSearchResult, error) {
	expanded, notes, err := ExpandSearchRequest(req)
	if err != nil {
		return nil, err
	}
	req = expanded
	if err := validateSearchRequest(req); err != nil {
		return nil, err
	}
	limit, offset := normalizePage(req.Limit, req.Offset)
	_, query := searchParams(req, limit, offset)
	return &ListingsSearchResult{
		Listings: []Listing{},
		Count:    0,
		Total:    0,
		Limit:    limit,
		Offset:   offset,
		Summary:  summarizeListings(nil, 0, limit, offset),
		Query:    query,
		Note:     joinNotes("stub client", notes),
	}, nil
}

func (Stub) GetListing(_ context.Context, id string) (*Listing, error) {
	return &Listing{ID: id, Status: "stub"}, nil
}

func (Stub) RentEstimate(_ context.Context, req RentEstimateRequest) (*RentEstimateResult, error) {
	return &RentEstimateResult{Address: req.Address, Note: "stub client"}, nil
}

func (Stub) MarketStats(_ context.Context, zipCode string) (*MarketStatsResult, error) {
	return &MarketStatsResult{ZipCode: zipCode, Note: "stub client"}, nil
}
