---
PLAN: "feat(layout)!: module title moves to the platformd header; rightpanel loses its title band; crudview Context → Scope in the list header"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — the module title belongs to the shell

## Tools — install before anything else

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Problem

On desktop each module draws its own title band inside the content (`rightpanel`'s `<h1>` "Catálogo"
in a gradient band, plus `HeadControls` beside it), while the `platformd` header shows the active
module's name **only on phones**. Modules that do not use `rightpanel` have no title at all, so screens
look different from each other and the band wastes the top of every screen. Decided (master plan
`UI_CONSISTENCY`, D6, P12–P15):

- **Phone: unchanged.**
- **Desktop:** the header shows, at its leading edge, `[mark] [app name] / [active module name]`;
  the whole header bar uses the primary surface (as it already does on phones), text and icons white
  (on-primary). The module name is the page's single `<h1>`.
- For a module inside a `GroupModule`, the name shown is the **active child's** label (what
  `activeLabel()` already returns).
- `rightpanel` has no title band any more; the article starts at the top.
- A control that **filters the list** (which professional's records are shown) lives in the list's
  header, above the search box — crudview's `Config.Context` becomes `Config.Scope` and moves there.
  A control that is a **value of the record** is a form field (consumers move it in their own plans).

## Design gate

1. **Prior art.** Android/Material `TopAppBar` and Flutter `Scaffold.appBar` (the screen title is the
   shell's; the screen supplies only the body), iOS `UINavigationBar` (`navigationItem.title` drawn by
   the navigation controller), Polaris `Page title` rendered by the frame. All put the title in the
   chrome so every screen has it in the same place.
2. **Novice-name.** `Config.Scope` / `CrudView.Scope`: "the scope of the list" — what the selector
   does. `Context` described the legacy screen it mirrored, not what it does. No other new name.
3. **Ledger.**
   ```
   Concepts                 −3  (RightPanel.Title, RightPanel.Head, RightPanel.HeadControls)
   Lines at call site       −1  (no Title per screen)
   Ways to show a title     2 → 1 (header only)
   ```
4. **Where it belongs.** Screen title: `platformd` (it knows the active module). List scope:
   `crudview`, beside the filter it sits with. `rightpanel` stays a two-column skeleton only.
5. **What it deletes.** `RightPanel.Title`, `RightPanel.Head`, `RightPanel.HeadControls`, the parts
   `title-row`, `title`, `controls` and the `header` element of `rightpanel`; `CrudView.Context` and
   `Config.Context` (renamed, not kept as alias); `CrudView.Title`'s use as a heading (see Stage 2).

## Stage 1 — `rightpanel`

1. `rightpanel.go`: delete the fields `Title`, `Head`, `HeadControls`; delete the header row built in
   `Render` (`header`, `titleRow`, the `H1`, the `HeadControls` div) and the part constants/classes
   `partTitleRow`, `partTitle`, `partHeadControls`, `clsHeader`, `clsTitleRow`, `clsTitle`,
   `clsHeadControls`. `main` now holds only the article. Update the usage comment.
2. `css.go`: delete the rules for those parts and any comment that refers to them.
3. Drop the `webtyp.com/lang` import if unused.

## Stage 2 — `crudview`

1. Rename `Config.Context` → `Config.Scope` (`crud.go`) and `CrudView.Context` → `CrudView.Scope`
   (`crudview.go`). Doc comment: "Scope is a selector that decides which records the list shows (e.g.
   which professional). It is rendered in the list header above Filter. Changing it re-scopes the data:
   the composing module reloads its presenter; crudview only renders it. A value that belongs to the
   record is a form field, never a Scope."
2. Where `crudview.go` builds `rightpanel.RightPanel` (today `Title: v.Title, … HeadControls:
   v.Context`): remove `Title` and `HeadControls`; set `AsideControls` to a stack containing `v.Scope`
   (when non-nil) then `v.Filter` (when non-nil); one of them alone is passed as is, both nil → nil.
   The stack is a `Div` with a crudview part `scope-controls` styled `Stack(Space2)` in `css.go`.
3. `CrudView.Title` and `Config`'s title from `Presenter.Title()` stay for non-heading uses (confirm
   dialog text, aria labels) if any exist; if after step 2 nothing reads `Title`, delete it.

## Stage 3 — `platformd` header

`platformd.go`:

1. The header title element becomes an `H1` (today `Span` with `clsHeaderTitleText`) bound to
   `p.activeLabel` — the page's single `<h1>`.
2. Order inside the header: brand slot, a separator `Span` with part `header-sep` text `/`, the title,
   then msg-slot, right, hamburger (unchanged).

`css.go`:

1. Header: `As(style.Primary)` at every width (move it out of the mobile-only rule; the mobile rule
   keeps `AppBar(...)`).
2. `header-title` and `header-title-text`: today `OnlyOn(css.Mobile, …)`. Desktop: shown, left-aligned
   right after the separator, `FontSize(TextBase)`, `FontWeight(WeightBold)`, `KeepSize()`. Mobile: keep
   the current rules exactly (centred, the brand name hidden).
3. `header-sep`: shown only on desktop (`OnlyOn(css.Desktop, …)` or the equivalent `Hide()` on
   mobile used elsewhere in this file), `Glyph(...)` inheriting the on-primary text.
4. Brand mark, brand name, user menu trigger and status in the header read on the primary surface
   (on-primary text/glyph). Use the surface's own on-colour; never a hand-written colour.

## Stage 4 — tests: behaviour only, never the look

- Delete the tests of `Title`/`Head`/`HeadControls` and every assertion on the removed parts' styles
  (`consumer_stylesheet_test.go` entries included).
- `crudview`: rename `context_test.go` → `scope_test.go`; keep the behaviour it asserts (the scope
  control is never wired to the list filter).
- `platformd`: the page has exactly one `h1` and it holds the active module's label, including the
  active child of a `GroupModule` (an accessibility contract: one `h1` per page).
- No test asserts colours, surfaces or CSS declarations.

## Stage 5 — docs

`rightpanel/README.md`, `crudview` docs, `platformd/README.md`, `docs/ARCHITECTURE.md`: remove
`Title`/`Head`/`HeadControls`/`Context`; document `Scope` with the record-value vs list-scope rule;
document the desktop header `[mark] [app] / [module]`.

## Acceptance

- `grep -rn "HeadControls\|\.Context\b\|Config.Context\|partTitleRow" --include=*.go .` → empty.
- `gotest ./...` green (stdlib + wasm).

| Stage | Files |
|---|---|
| 1 | `rightpanel/rightpanel.go`, `rightpanel/css.go` |
| 2 | `crudview/crud.go`, `crudview/crudview.go`, `crudview/css.go` |
| 3 | `platformd/platformd.go`, `platformd/css.go` |
| 4 | `rightpanel/*_test.go`, `crudview/scope_test.go` (from `context_test.go`), `platformd/platformd_test.go`, `*/consumer_stylesheet_test.go` (removed parts) |
| 5 | `rightpanel/README.md`, `platformd/README.md`, `docs/ARCHITECTURE.md`, crudview docs |
