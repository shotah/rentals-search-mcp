<p align="center">
  <img src="docs/assets/banner.svg" alt="rentals-search-mcp — search · recommend · handoff" width="100%">
</p>

# rentals-search-mcp

<p align="center">
  <a href="https://github.com/shotah/rentals-search-mcp/actions/workflows/ci.yml"><img src="https://github.com/shotah/rentals-search-mcp/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/shotah/rentals-search-mcp/actions/workflows/release.yml"><img src="https://github.com/shotah/rentals-search-mcp/actions/workflows/release.yml/badge.svg" alt="Release"></a>
  <a href="https://github.com/shotah/rentals-search-mcp/actions/workflows/ci.yml"><img src="https://github.com/shotah/rentals-search-mcp/raw/gh-pages/badges/coverage.svg" alt="Coverage"></a>
  <a href="https://pkg.go.dev/github.com/shotah/rentals-search-mcp"><img src="https://pkg.go.dev/badge/github.com/shotah/rentals-search-mcp.svg" alt="Go Reference"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/shotah/rentals-search-mcp" alt="Go version">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/shotah/rentals-search-mcp" alt="License"></a>
</p>

Static Go [MCP](https://modelcontextprotocol.io) for **long-term residential rental
search** (apartments, houses, condos, townhomes) via the
[RentCast API](https://developers.rentcast.io/reference/introduction).

Built for [ai-gantry](https://github.com/shotah/ai-gantry) / LOCAL_AGENT: lean
tool surface, stdio only, `CGO_ENABLED=0` binary that fits distroless.
**Search + recommend + listing handoff — never applies, contacts landlords, or
signs leases.**

Naming follows
[ai-gantry MCP tool naming](https://github.com/shotah/ai-gantry/blob/main/docs/mcp-naming.md):
host server id `rentals` → tools like `rentals__listings_search` (tool names do
**not** repeat the server id).

> **Publish from a personal GitHub account only** (never a work org). CI /
> GoReleaser / badges are scaffolded and light up after the personal remote
> exists — see [TODO.md](TODO.md).

## What this is (and is not)

| In scope | Out of scope |
| --- | --- |
| Long-term **for-rent** residential listings (US) | Short-term / vacation stays (Airbnb, Google Hotels) |
| Apartments, houses, condos, townhomes, multi-family | Retail / office / industrial **commercial** leases |
| Rent estimates + zip market stats | Applying, touring, or signing leases |
| Listing URL / contact handoff | Purchasing or escrow |

Commercial / tiny retail spaces (e.g. a small Seattle grocery mart) need a
**different** data source and almost certainly a **different MCP package** —
see [docs/commercial-spaces.md](docs/commercial-spaces.md).

## Tools (MVP)

| Tool | Description |
| --- | --- |
| `listings_search` | City/zip/radius search + filters (beds, baths, rent, property type) |
| `listings_get` | One listing by RentCast listing id |
| `rent_estimate_get` | Long-term rent AVM for an address (+ comps when available) |
| `markets_get` | Aggregate rent / listing stats for a US zip code |
| `link_format` | Fallback public search URL (no API call) |
| `account_get` | Soft usage note — RentCast has no public quota API; point at dashboard |

Host names: `rentals__listings_search`, `rentals__listings_get`,
`rentals__rent_estimate_get`, `rentals__markets_get`, `rentals__link_format`,
`rentals__account_get`.

### Agent contract

```text
listings_search → [listings_get] → present listing url / contact
                 ↘ markets_get / rent_estimate_get for context
```

1. **Where + budget:** `listings_search` with `city`+`state` or `zip_code`
   (or lat/lng + `radius`), plus `bedrooms`, `bathrooms`, `price` range,
   `property_type`.
2. **Detail:** `listings_get` when the human picks a candidate id.
3. **Context (optional):** `markets_get` for zip averages; `rent_estimate_get`
   when comparing a specific address to “fair rent”.
4. **Handoff:** give the human the listing URL / contact fields. Do **not**
   impersonate the renter or submit applications.

## Setup

1. Create a [RentCast](https://www.rentcast.io/api) account (Developer plan =
   **50 free API requests / month**).
2. Export the key:

```bash
export RENTCAST_API_KEY=...
# optional (tests / proxies):
# export RENTCAST_BASE_URL=https://api.rentcast.io/v1
```

## Development

```bash
make help                 # list targets
make tools                # install goimports-reviser + golangci-lint v2
make install-hooks        # git pre-commit: autofix + lint + test + coverage ≥70%
make check                # same checks as the pre-commit hook
make cli                  # static binary → ./bin/rentals-search-mcp
make self-test
```

Coverage **≥70%** is mandatory locally (`make coverage` / pre-commit) and in CI.

Status: **MVP client wired**. US-wide long-term residential search (Seattle is just
the first use case). Remaining publish steps are in [TODO.md](TODO.md).

## Releasing

GoReleaser publishes static binaries on `v*` tags (see `.github/workflows/release.yml`).
Requires a personal GitHub remote first.

```bash
make version                 # show VERSION + next patch (dry-run)
make release                 # patch bump, commit VERSION, tag, push
make release BUMP=minor
make release TAG=v0.1.0
```

Archive name for ai-gantry `download_url`:

```text
rentals-search-mcp_{version}_{os}_{arch}.tar.gz
```

## Run as MCP (stdio)

```bash
./bin/rentals-search-mcp
```

Logs go to **stderr** only — stdout is reserved for the MCP protocol.

Example Cursor / client config:

```json
{
  "mcpServers": {
    "rentals": {
      "command": "rentals-search-mcp",
      "env": {
        "RENTCAST_API_KEY": "…"
      }
    }
  }
}
```

## ai-gantry wiring (when published)

```toml
[[server]]
name = "rentals"
command = "rentals-search-mcp"
download_tag = "latest"
download_url = "https://github.com/shotah/rentals-search-mcp/releases/download/{tag}/rentals-search-mcp_{version}_{os}_{arch}.tar.gz"
```

Do **not** publish this to a work GitHub org. Use a personal account / org
(module path assumes `shotah`).

More consumer notes: [docs/ai-gantry.md](docs/ai-gantry.md).

## License

MIT
