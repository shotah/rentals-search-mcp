package rentcast

import "context"

// Stub is an in-memory client for self-test / unit tests (no network).
type Stub struct {
	Usage *UsageTracker
}

// NewStub returns a non-nil RentalsAPI that returns canned empty results.
func NewStub() *Stub {
	return &Stub{Usage: NewUsageTrackerForTest("", defaultMonthlyQuota)}
}

func (s Stub) SearchListings(_ context.Context, req ListingsSearchRequest) (*ListingsSearchResult, error) {
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
		Usage:    s.Usage.Snapshot(),
	}, nil
}

func (s Stub) GetListing(_ context.Context, id string) (*ListingGetResult, error) {
	return &ListingGetResult{
		Listing: Listing{ID: id, Status: "stub"},
		Usage:   s.Usage.Snapshot(),
	}, nil
}

func (s Stub) RentEstimate(_ context.Context, req RentEstimateRequest) (*RentEstimateResult, error) {
	return &RentEstimateResult{Address: req.Address, Note: "stub client", Usage: s.Usage.Snapshot()}, nil
}

func (s Stub) MarketStats(_ context.Context, zipCode string) (*MarketStatsResult, error) {
	return &MarketStatsResult{ZipCode: zipCode, Note: "stub client", Usage: s.Usage.Snapshot()}, nil
}

func (s Stub) AccountUsage() *Usage {
	return s.Usage.Snapshot()
}
