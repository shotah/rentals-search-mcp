package rentcast

import (
	"fmt"
	"strconv"
	"strings"
)

const amenityUnsupportedNote = "RentCast listings API does not expose pet/parking/laundry filters — verify on listing_url or with the listing agent/office."

// ExpandSearchRequest applies neighborhood presets, multi-zip parsing, and
// new-this-week / days-old helpers before the HTTP call.
func ExpandSearchRequest(req ListingsSearchRequest) (ListingsSearchRequest, []string, error) {
	notes := make([]string, 0, 4)

	zips := ParseZipList(req.ZipCodes)
	if z := strings.TrimSpace(req.ZipCode); z != "" {
		zips = mergeZips(zips, ParseZipList(z))
	}

	var err error
	req, zips, notes, err = applyNeighborhood(req, zips, notes)
	if err != nil {
		return req, nil, err
	}

	if len(zips) == 1 && strings.TrimSpace(req.ZipCode) == "" {
		req.ZipCode = zips[0]
	}
	req.ZipFilter = zips

	// Multi-zip without a location anchor: require city+state or use first zip for the API.
	if len(zips) > 1 && strings.TrimSpace(req.ZipCode) == "" &&
		(strings.TrimSpace(req.City) == "" || strings.TrimSpace(req.State) == "") &&
		req.Latitude == 0 && req.Longitude == 0 && strings.TrimSpace(req.Address) == "" {
		req.ZipCode = zips[0]
		notes = append(notes, fmt.Sprintf("multi-zip search using primary zip %s (pass city+state to search once and filter)", zips[0]))
	}

	if req.NewThisWeek && strings.TrimSpace(req.DaysOld) == "" && req.DaysOldMax <= 0 {
		req.DaysOldMax = 7
		notes = append(notes, "new_this_week → days_old_max=7")
	}
	if req.DaysOldMax > 0 && strings.TrimSpace(req.DaysOld) == "" {
		req.DaysOld = strconv.Itoa(req.DaysOldMax)
	}

	if req.PetsWanted || req.ParkingWanted || req.LaundryWanted {
		notes = append(notes, amenityUnsupportedNote)
	}

	return req, notes, nil
}

func applyNeighborhood(req ListingsSearchRequest, zips, notes []string) (ListingsSearchRequest, []string, []string, error) {
	nb := strings.TrimSpace(req.Neighborhood)
	if nb == "" {
		return req, zips, notes, nil
	}
	area := LookupNeighborhood(nb, req.City, req.State)
	if area == nil {
		return req, zips, notes, fmt.Errorf("unknown neighborhood %q — call areas_resolve (list_all=true) for Seattle presets", nb)
	}
	req.Neighborhood = area.Name
	if strings.TrimSpace(req.City) == "" {
		req.City = area.City
	}
	if strings.TrimSpace(req.State) == "" {
		req.State = area.State
	}
	zips = mergeZips(zips, area.Zips)

	useGeo := area.Latitude != 0 && area.Longitude != 0 &&
		req.Latitude == 0 && req.Longitude == 0 &&
		strings.TrimSpace(req.ZipCode) == "" &&
		len(ParseZipList(req.ZipCodes)) == 0

	switch {
	case useGeo:
		req.Latitude = area.Latitude
		req.Longitude = area.Longitude
		if req.Radius <= 0 {
			req.Radius = area.Radius
			if req.Radius <= 0 {
				req.Radius = defaultNeighborhoodRadius
			}
		}
		notes = append(notes, fmt.Sprintf("neighborhood %s → lat/lng + %.1f mi radius", area.Name, req.Radius))
	case len(area.Zips) == 1 && strings.TrimSpace(req.ZipCode) == "":
		req.ZipCode = area.Zips[0]
		notes = append(notes, fmt.Sprintf("neighborhood %s → zip %s", area.Name, area.Zips[0]))
	case len(area.Zips) > 1:
		notes = append(notes, fmt.Sprintf("neighborhood %s → filter zips %s", area.Name, strings.Join(area.Zips, ", ")))
	}
	return req, zips, notes, nil
}

func mergeZips(a, b []string) []string {
	if len(a) == 0 {
		return append([]string(nil), b...)
	}
	if len(b) == 0 {
		return append([]string(nil), a...)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, z := range a {
		if !seen[z] {
			seen[z] = true
			out = append(out, z)
		}
	}
	for _, z := range b {
		if !seen[z] {
			seen[z] = true
			out = append(out, z)
		}
	}
	return out
}

func filterListingsByZips(listings []Listing, zips []string) []Listing {
	if len(zips) == 0 {
		return listings
	}
	want := map[string]bool{}
	for _, z := range zips {
		want[z] = true
	}
	out := make([]Listing, 0, len(listings))
	for i := range listings {
		if want[strings.TrimSpace(listings[i].ZipCode)] {
			out = append(out, listings[i])
		}
	}
	return out
}

func joinNotes(base string, notes []string) string {
	if len(notes) == 0 {
		return base
	}
	if base == "" {
		return strings.Join(notes, " ")
	}
	return base + " " + strings.Join(notes, " ")
}
