# Proto 3.0 evidence: menu/page provider registry — Candidates A vs B

- Date: 2026-08-23
- Branch: `feat/plugin-framework-phase3` (worktree)
- Plan ref: `.ai/runs/2026-08-23-plugin-framework-phase3.md` Phase 3.0 (3.0.1)
- Scope: falsification prototype only. Do NOT build on `registry_proto_a.go` / `registry_proto_b.go`; they are evidence artifacts.
- Working tree state: call sites are currently wired through **Candidate B** (see "Working-tree wiring" at the end). Candidate A's registry is fully registered and compiles; its call-site numbers below were measured from an actual A migration + `go build ./...` + `go test ./internal/ui/ ./internal/framework/` run.

## Sample set

| Sample | Constructor | Args |
|---|---|---|
| PlaylistDetail | `NewPlaylistDetailMenu(base baseMenu, playlistId int64)` | 1 |
| ArtistDetail | `NewArtistDetailMenu(base, artistId int64, name string)` | 2 |
| SearchResult | `NewSearchResultMenu(base, searchType SearchType)` | 1 |
| AddToUserPlaylist | `NewAddToUserPlaylistMenu(base, userId int64, song structs.Song, isAdd bool)` | 3 |
| Ranks | `NewRanksMenu(base baseMenu)` | 0 |
| Login (page) | `NewLoginPage(netease *Netease)` | 1 (page, not menu) |

## Candidate shapes

- **A (variadic)** — `MenuProviderA{Key, Build func(base baseMenu, args ...any) (Menu, error)}`; registry `map[string]*MenuProviderA`; call sites pass positional args. Each provider closure hand-writes arg count/type/order assertions.
- **B (generic)** — `menuFactory[T]{Key, Build func(base baseMenu, opts T) (Menu, error)}` stored in `map[string]any`; call sites pass a named opts struct; the only type assertion lives inside `BuildMenuB[T]` (`.(menuFactory[T])`).

## Per-candidate summary table

Measured as **net lines changed at the call site** (old lines removed vs new lines added) with `go build ./...` green and `go test ./internal/ui/ ./internal/framework/` green for both.

| Sample menu | # call sites migrated | A: net lines | B: net lines | A: type assertions at call site | B: type assertions at call site | Readability notes |
|---|---|---|---|---|---|---|
| PlaylistDetail | 5 (menu_ranks:42, menu_high_quality:39, menu_daily_recommend:43, menu_user_playlist:54, menu_search_result:82) | +20 | +20 | 0 | 0 | A: `BuildMenuA("playlist_detail", base, id)` reads like today's positional constructor. B: `BuildMenuB("playlist_detail", base, PlaylistDetailOpts{PlaylistID: id})` — named field is self-documenting but adds noise for a single arg. |
| ArtistDetail | 5 (menu_search_result:87, menu_artists_subscribe:47, menu_hot_artists:39, menu_artist_list:24, operate:423) | +19 | +19 | 0 | 0 | A: two positional args (`id, name`) fine. B: `ArtistDetailOpts{ArtistID: …, Name: …}` clearly labels which string is the name — the real win over A's `(int64, string)` ambiguity. |
| SearchResult | 2 (menu_search_type:50, operate:1010) | +8 | +8 | 0 | 0 | A terse; B forces `SearchResultOpts{SearchType: StSingleSong}` — mildly noisy for one field. |
| AddToUserPlaylist | 1 (operate:729) | +3 | +3 | 0 | 0 | A's 3 positional args (`userId, song, isAdd`) are readable but order-fragile. B's named struct (`UserID/Song/IsAdd`) is the clearest case for B. |
| Ranks | 1 (menu_main:49) | 0 | 0 | 0 | 0 | A: `mustBuildMenuA("ranks", base)` — zero overhead for no-arg. B: `mustBuildMenuB("ranks", base, NoArgMenuOpts{})` — placeholder type is pure noise at every call site. |
| Login (page) | 1 (netease:110) | +4 | +4 | **1** | **1** | Both need `loginPage.(*LoginPage)` because `Netease.login` is a concrete `*LoginPage` while `BuildPage*` returns `model.Page`. Same cost for A and B. |
| **Total** | **15 call sites** | **+54** | **+54** | 1 (login only) | 1 (login only) | Line-count parity; the differentiators are type-safety and registration ergonomics, not size. |

Notes:
- The +N is dominated by the `(Menu, error)` error contract (3-4 lines of `if err != nil { return nil }` per navigation site) — identical for A and B, orthogonal to the shape choice.
- Menu call sites need **zero** type assertions in both shapes. A's assertions live inside the provider registrations; B's single assertion is hidden in the registry. The login bootstrap is the only place both shapes force a call-site assertion (concrete field type).

## No-arg menu handling story

- **A**: natural. `BuildMenuA("ranks", base)` with no args; provider checks `len(args) != 0`. No placeholder types anywhere. A no-arg menu costs the same as a 1-arg menu.
- **B**: forced. Go has no default type parameters, so the generic signature needs an explicit opts type. Prototype uses `NoArgMenuOpts struct{}`; every call site must spell `NoArgMenuOpts{}`. Documented choice: a shared placeholder type rather than per-menu empty structs. `mustBuildMenuB("ranks", base, NoArgMenuOpts{})` is the compact form. If no-arg menus are common, this is B's weakest point.

## Login-page story

- A: `BuildPageA("login", n)` — provider asserts `args[0].(*Netease)`.
- B: `BuildPageB("login", LoginPageOpts{Netease: n})` — typed, no provider assertion.
- Both: `BuildPage*` returns `model.Page`, so the single bootstrap assignment `n.login = NewLoginPage(n)` becomes a 4-line block with `loginPage.(*LoginPage)`. The assertion is unavoidable while `Netease.login` is a concrete type (changing that field is out of prototype scope).
- `ToLoginPage` / `ToSearchPage` navigation paths were not touched — they consume the already-built `n.login`, so the prototype only covers the construction site, which is the real "registry" boundary for pages.

## MainMenu bootstrap-path note

`NewMainMenu(netease)` at `internal/commands/netease.go:72` is the app-bootstrap entry and is NOT registered as a provider in this prototype (it stays a plain constructor). What the prototype demonstrates for the bootstrap path:

- MainMenu's internal composite literal builds 14 child menus eagerly, including the sample `Ranks` at `menu_main.go:49`. In a constructor with no error channel, both candidates use a `mustBuildMenuA` / `mustBuildMenuB` helper that panics on registration errors:
  - A: `mustBuildMenuA("ranks", base)` — 1 line, identical to today.
  - B: `mustBuildMenuB("ranks", base, NoArgMenuOpts{})` — 1 line, extra placeholder.
- If `NewMainMenu` itself were registered as a provider later: A would register `Build func(base baseMenu, args ...any)` and the call site would be `mustBuildMenuA("main_menu", newBaseMenu(n))`; B would use a `MainMenuOpts` (likely `struct{ Netease *Netease }` or empty) and `BuildMenuB("main_menu", …, MainMenuOpts{…})`. Neither shape blocks the bootstrap; both would keep the same must-panic escape hatch, since a missing bootstrap provider is a fatal programmer error.
- Recommendation for Phase 3.2 regardless of A/B: treat bootstrap (MainMenu, login, search) as a separate "must" surface from navigation (error-returning `BuildMenu`), so bootstrap sites stay 1-line.

## Preliminary recommendation

**Candidate B (generic).** The measured migration cost is identical (both +54 net lines, both zero call-site assertions except the shared login bootstrap), so the decision rests on type safety and the registration boundary: B moves every arg contract into the type system once (the exported opts struct is the discoverable plugin contract and wrong field names/types fail at compile time), while A re-implements arg parsing per provider and surfaces mistakes only at runtime navigation. B's real costs — the `NoArgMenuOpts{}` placeholder for no-arg menus and slightly more verbose single-field call sites — are minor relative to A's per-provider assertion boilerplate (roughly 8-12 lines per registration in this prototype vs 1-3 for B) and its runtime-only failure mode. This is a preliminary lean, not the adjudication; the orchestrator makes the final call in 3.0.2, and the working tree can be flipped to A by replacing `BuildMenuB`/`mustBuildMenuB` calls (all 15 sites are in this commit's diff).

## Working-tree wiring

- Active call sites: Candidate B (`BuildMenuB` / `BuildPageB` / `mustBuildMenuB`).
- Both registries present and registered: `internal/ui/registry_proto_a.go`, `internal/ui/registry_proto_b.go` (each with `// PROTO` header).
- Old constructors untouched and still used by non-sample call sites (e.g. `NewAlbumDetailMenu`, `NewArtistSongMenu`, `NewDjRadioDetailMenu` were not routed through the registries — only the 5 sample menus + login page were migrated).
- Validation: `go build ./...` green; `go test ./internal/ui/ ./internal/framework/` green (both candidates measured). `gofmt -l` clean for all new/changed files except pre-existing drift in `netease.go` (changelog block) and `operate.go` (trailing-whitespace line) that predates this prototype and was left untouched.
