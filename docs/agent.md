# Agent guide — rentals-search-mcp

How an LLM / LOCAL_AGENT should use this MCP. Copy the **TOOLS.md sketch**
into a host persona file (e.g. ai-gantry `persona/TOOLS.md`), or point agents
here.

Host server id: **`rentals`** → tools appear as `rentals__{tool}`.
Tool names in this binary do **not** start with `rentals_`.

## Quota (read this first)

RentCast free tier ≈ **50 requests / calendar month** (~**1–2 per day**).
Every successful `listings_search`, `listings_get`, `rent_estimate_get`, and
`markets_get` burns **one**. There is no free “preview.”

| Free (0 RentCast) | Burns 1 request each |
| --- | --- |
| `areas_resolve` | `listings_search` |
| `link_format` | `listings_get` |
| `account_get` | `rent_estimate_get` |
| | `markets_get` |

**Local usage counter** (on every search response + `account_get`):

- `requests_used` / `requests_left` / `requests_per_month` (default 50)
- `period` = `YYYY-MM` (local calendar month)
- `period_resets` = **1st of next month** (local time) — yes, the local counter
  resets on the first of the month
- Not the official RentCast dashboard (billing cycle may differ) — still treat
  `requests_left` as a hard budget before calling paid tools

**Before any burning call:** skim `usage` from the last response, or call
`rentals__account_get` (free). If `requests_left` is low, prefer `link_format`
or stop and tell the human.

## Thrifty search (one API call, not N)

**Bad (burns N):** search Capitol Hill, then Ballard, then Fremont separately.

**Good (burns 1):**

```text
# Multi-neighborhood — ONE call
rentals__listings_search(
  neighborhood="Ballard,Fremont,Wallingford",
  bedrooms="1:2", price_max=2800, new_this_week=true, limit=10
)

# Or resolve free, then one multi-zip search
rentals__areas_resolve(list_all=true)   # FREE
rentals__listings_search(
  city="Seattle", state="WA",
  zip_codes="98107,98117,98103",
  bedrooms="1", price_max=2500, limit=10
)
```

`zip_codes` / comma-separated `neighborhood` → **one** RentCast request, then
local zip filter. Always pass tight filters (`bedrooms`, `price_max`,
`property_type`, `new_this_week`) so the page is useful without offset paging.

Avoid: paging with `offset` unless the human asks for more; avoid
`markets_get` / `rent_estimate_get` / many `listings_get` on a whim.

## Contract

```text
[account_get?] → [areas_resolve FREE] → ONE listings_search → [listings_get] → handoff
                                         ↘ markets / rent_estimate only if needed
```

1. **Budget check** — notice `usage` / `account_get` (free).
2. **Where + budget** — one `listings_search` with city+state, `zip_codes`, or
   `neighborhood` (comma-separated OK), plus beds / rent / type.
3. **Fresh** — `new_this_week` or `days_old_max` when they want recent listings.
4. **Detail** — `listings_get` only after the human picks an id.
5. **Handoff** — `listing_url` / agent/office contact. Never apply or message
   landlords. `listing_url` is agent/office site when present, else Google
   `{address} rental` (no Zillow/Realtor deep-link ids from RentCast).

## Hard rules

- Residential long-term only — **not** retail / office / commercial.
- Soft prefs `pets_wanted` / `parking_wanted` / `laundry_wanted` are **not**
  RentCast filters — confirm on the listing page.
- Be thrifty: **one combined search > many area searches**.

## TOOLS.md sketch (paste into persona)

Routing row:

```text
| Apartments / houses / rentals | `rentals__…` | applying, messaging landlords, or retail/commercial leases |
```

Section:

```text
## Rentals (apartments / houses)

QUOTA: RentCast free ≈ 50 req/month (~1–2/day). Check rentals__account_get or
usage on responses BEFORE searching. period_resets = 1st of next month (local counter).
FREE: areas_resolve, link_format, account_get. EACH search/get/estimate/markets burns 1.

When friends ask for apartments or rental houses:
1. Optional rentals__account_get if usage unknown
2. Optional rentals__areas_resolve (FREE) for Seattle neighborhood names
3. ONE rentals__listings_search — combine areas:
   neighborhood="Ballard,Fremont" OR zip_codes="98107,98103" with city+state
   + bedrooms + price_max (+ new_this_week). NEVER one call per neighborhood.
4. rentals__listings_get only for a chosen id
5. Hand off listing_url / contact — never apply or message landlords

If requests_left is low → rentals__link_format (FREE) or stop and say so.
Do NOT use rentals for retail/office/commercial. Needs RENTCAST_API_KEY.
```

See also [ai-gantry.md](ai-gantry.md) for `mcp.toml` / Docker env wiring.
