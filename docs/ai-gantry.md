# ai-gantry consumer wiring

**Do this only after** the binary exists on a **personal** GitHub release
(or you bake a local binary into the image). Do not point `download_url` at a
work org.

## `mcp.toml`

```toml
# Long-term residential rentals (shotah/rentals-search-mcp) — docs/rentals.md
# Recommend-only: listings + rent estimate + market stats. No applications.
[[server]]
name = "rentals"
command = "rentals-search-mcp"
download_tag = "latest"
download_url = "https://github.com/shotah/rentals-search-mcp/releases/download/{tag}/rentals-search-mcp_{version}_{os}_{arch}.tar.gz"
```

## Env

```bash
RENTCAST_API_KEY=...
# Optional pin for Docker bake / native fetch:
# RENTALS_SEARCH_MCP_VERSION=v0.0.1
```

## Persona (`TOOLS.md`) sketch

```text
## Rentals (apartments / houses)

When friends ask for apartments or rental houses:
1. rentals__listings_search with city+state or zip, bedrooms, price_max
2. rentals__listings_get for a chosen id
3. Optional: rentals__markets_get for zip context; rentals__rent_estimate_get for “is this rent fair?”
4. Hand off listing_url / contact — never apply or message landlords for the human

Do NOT use rentals tools for retail/office/commercial leases.
```

## Host tool names

| Tool | Host |
| --- | --- |
| `listings_search` | `rentals__listings_search` |
| `listings_get` | `rentals__listings_get` |
| `rent_estimate_get` | `rentals__rent_estimate_get` |
| `markets_get` | `rentals__markets_get` |
| `link_format` | `rentals__link_format` |
| `account_get` | `rentals__account_get` |

Update [ai-gantry `docs/mcp-naming.md`](https://github.com/shotah/ai-gantry/blob/main/docs/mcp-naming.md)
host table with a `rentals` row when shipping the consumer PR.

## Cost reminder

RentCast Developer ≈ **50 requests/month**. Prefer one tight `listings_search`
over many exploratory pages. `link_format` is free (no API).
