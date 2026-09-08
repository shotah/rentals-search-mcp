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
- `soft_cap` (default 40) / `cap_state` (`ok` | `confirm_required` | `exhausted`)
- `period` = `YYYY-MM` (local calendar month)
- `period_resets` = **1st of next month** (local time) — yes, the local counter
  resets on the first of the month
- Not the official RentCast dashboard (billing cycle may differ)

**Caps (enforced before HTTP — not honor-system):**

- **Soft (40 used):** billed tools error until you re-call with
  `confirm_spend=true`. Required on **each** remaining call, not a one-time unlock.
- **Hard (50 used):** billed tools error. Use `link_format` or wait for
  `period_resets`. `confirm_spend` does **not** bypass this. Only a human setting
  `RENTCAST_ALLOW_OVERAGE=1` in MCP env can continue (paid plan).

**Before any burning call:** skim `usage` from the last response, or call
`rentals__account_get` (free). If `cap_state` is `confirm_required` or
`exhausted`, prefer `link_format` or stop and tell the human.

## Thrifty search (one API call, not N)

**Bad (burns N):** search Capitol Hill, then Ballard, then Fremont separately.

**Good (burns 1):**

```text
# Multi-neighborhood — ONE call
rentals__listings_search(
  intent="rent",
  neighborhood="Ballard,Fremont,Wallingford",
  bedrooms="1:2", price_max=2800, new_this_week=true, limit=10
)

# Buy — same filters, purchase-price budget, house OR condo in ONE call
rentals__listings_search(
  intent="buy", city="Seattle", state="WA",
  property_type="house,condo", bedrooms="2:3", price_max=750000, limit=10
)

# Or resolve free, then one multi-zip search
rentals__areas_resolve(list_all=true)   # FREE
rentals__listings_search(
  intent="rent",
  city="Seattle", state="WA",
  zip_codes="98107,98117,98103",
  bedrooms="1", price_max=2500, limit=10
)
```

`zip_codes` / comma-separated `neighborhood` / `property_type=house,condo` →
**one** RentCast request (multi-type is an upstream OR, then optional local zip
filter). Always pass tight filters (`bedrooms`, `price_max`, `property_type`,
`new_this_week`) so the page is useful without offset paging.

Avoid: paging with `offset` unless the human asks for more; avoid
`markets_get` / `rent_estimate_get` / many `listings_get` on a whim.

## Contract

```text
ASK rent vs buy → [account_get?] → [areas_resolve FREE] → ONE listings_search(intent=…) → [listings_get] → handoff
                                                              ↘ markets / rent_estimate only if needed
```

1. **Rent or buy** — if the human has not said which, **ask before searching**.
   Then pass `intent=rent` or `intent=buy` on `listings_search`, `listings_get`,
   and `link_format`. Do not guess. Sale and rental listing ids are different
   catalogs. `price_min` / `price_max` = monthly rent when renting, purchase
   price when buying.
2. **Budget check** — notice `usage` / `account_get` (free).
3. **Where + budget** — one `listings_search` with city+state, `zip_codes`, or
   `neighborhood` (comma-separated OK), plus beds / price / type.
4. **Fresh** — `new_this_week` or `days_old_max` when they want recent listings.
5. **Detail** — `listings_get` only after the human picks an id (same `intent`).
6. **Handoff** — ALWAYS include `listing_url` for every listing you mention so
   the human can review. Never apply, make offers, or message landlords/agents.
   `listing_url` is agent/office site when present, else Google `{address} rental`
   or `{address} for sale` (no Zillow/Realtor deep-link ids from RentCast).

## Hard rules

- Residential long-term only — **not** retail / office / commercial.
- ALWAYS give the human each `listing_url` (they review; you do not apply or offer).
- Soft prefs `pets_wanted` / `parking_wanted` / `laundry_wanted` are **not**
  RentCast filters — confirm on the listing page.
- Be thrifty: **one combined search > many area or type searches**.

## TOOLS.md sketch (paste into persona)

Routing row:

```text
| Apartments / houses / rentals **or** homes for sale | `rentals__…` | applying, offers, messaging landlords/agents, or retail/commercial leases |
```

Section:

```text
## Housing (apartments / houses — rent or buy)

QUOTA: RentCast free ≈ 50 req/month (~1–2/day). HARD CAP at 50 — billed tools
are blocked; the model cannot unlock them. After soft cap (~40) re-call with
confirm_spend=true (each remaining call). Check rentals__account_get (cap_state)
or usage on responses BEFORE searching. period_resets = 1st of next month.
FREE: areas_resolve, link_format, account_get. EACH search/get/estimate/markets burns 1.

When friends ask for apartments, houses, or condos:
0. If they have not said rent vs buy, ASK. Then pass intent=rent or intent=buy
   on every listings_search / listings_get / link_format. Do NOT guess or default.
1. Optional rentals__account_get if usage unknown
2. Optional rentals__areas_resolve (FREE) for Seattle neighborhood names
3. ONE rentals__listings_search — combine areas AND types:
   intent=rent|buy + neighborhood="Ballard,Fremont" OR zip_codes="98107,98103"
   + property_type="house,condo" + bedrooms + price_max (+ new_this_week).
   price_max is monthly rent when renting, purchase price when buying.
   NEVER one call per neighborhood or property type.
4. rentals__listings_get only for a chosen id (same intent)
5. ALWAYS include each listing_url so the human can review — never apply,
   make offers, or message landlords/agents

If cap_state is confirm_required → re-call with confirm_spend=true only if this
search is worth 1 remaining request; else rentals__link_format (FREE, still needs
intent) or stop.
If cap_state is exhausted → do NOT retry billed tools. link_format or wait until
period_resets. The model cannot unlock the hard cap.
Do NOT use rentals for retail/office/commercial. Needs RENTCAST_API_KEY.
```

See also [ai-gantry.md](ai-gantry.md) for `mcp.toml` / Docker env wiring.
