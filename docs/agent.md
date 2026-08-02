# Agent guide — rentals-search-mcp

How an LLM / LOCAL_AGENT should use this MCP. Copy the **TOOLS.md sketch**
into a host persona file (e.g. ai-gantry `persona/TOOLS.md`), or point agents
here.

Host server id: **`rentals`** → tools appear as `rentals__{tool}`.
Tool names in this binary do **not** start with `rentals_`.

## Contract

```text
[areas_resolve] → listings_search → [listings_get] → hand off listing_url / contact
                              ↘ markets_get / rent_estimate_get for context
```

1. **Where + budget** — `listings_search` with `city`+`state`, `zip_code` /
   `zip_codes`, or `neighborhood` (Seattle presets), plus beds / rent / type.
2. **Fresh** — `new_this_week` or `days_old_max` when they want recent listings.
3. **Detail** — `listings_get` after the human picks an id.
4. **Context (optional)** — `markets_get` (zip averages), `rent_estimate_get`
   (“is this rent fair?”).
5. **Handoff** — give `listing_url` / agent/office contact. **Never** apply,
   message landlords, or schedule tours for the human.

## Tool cheat sheet

| Tool | Burns RentCast? | When |
| --- | --- | --- |
| `areas_resolve` | No | Neighborhood name → zips / lat/lng (Seattle first) |
| `listings_search` | Yes | Main search |
| `listings_get` | Yes | One listing by id |
| `rent_estimate_get` | Yes | Fair-rent AVM for an address |
| `markets_get` | Yes | Zip-level rent stats |
| `link_format` | No | Public search URL fallback when quota is tight |
| `account_get` | No | Local usage estimate + dashboard link |

## Hard rules

- Residential long-term only (apartments, houses, condos, townhomes).
  **Not** retail / office / commercial leases.
- Soft prefs `pets_wanted` / `parking_wanted` / `laundry_wanted` are **not**
  RentCast filters — confirm on the listing page.
- Free tier ≈ **50 requests/month**. Prefer one tight `listings_search` over
  paging. Local `usage` on responses / `account_get` is approximate; dashboard
  is source of truth.

## TOOLS.md sketch (paste into persona)

Routing row:

```text
| Apartments / houses / rentals | `rentals__…` | applying, messaging landlords, or retail/commercial leases |
```

Section:

```text
## Rentals (apartments / houses)

When friends ask for apartments or rental houses:
1. Optional rentals__areas_resolve for a Seattle neighborhood name
2. rentals__listings_search with city+state, zip, or neighborhood; bedrooms; price_max; optional new_this_week
3. rentals__listings_get for a chosen id
4. Optional: rentals__markets_get / rentals__rent_estimate_get for context
5. Hand off listing_url / contact — never apply or message landlords for the human

Do NOT use rentals tools for retail/office/commercial leases.
Quota ~50 free RentCast req/month — prefer tight searches; rentals__account_get
shows a local usage estimate (dashboard is source of truth).
Needs RENTCAST_API_KEY.
```

See also [ai-gantry.md](ai-gantry.md) for `mcp.toml` / Docker env wiring.
