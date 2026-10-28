---
PLAN: "feat: API explorer inside a «Developer tools» menu group"
TAG: v0.3.80
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 9772909075004117390
---

> Read AGENTS.md at the repo root before any change.

<!-- Parked: rename to docs/PLAN.md and dispatch only after the v0.3.79 plan (model.Item) is closed and published. -->

# Plan — `webtyp.com/layout`: «Developer tools» group

## Why

`BuildMenu` appends the API explorer as a loose leaf at the end of the menu. A second developer screen (the
database diagram) is coming, so the explorer moves into a group «Developer tools» that keeps the current
`terminal` glyph; the explorer itself gets its own glyph, `code` (`</>`). Each child keeps its own permission:
a group has none, it shows only the children the user may see, and it is hidden when the user may see none
(`platformd/platformd.go` `isViewable`, and the `continue` on non-viewable children in `Render`). This plan adds
the test that pins that rule for a group with one allowed and one forbidden child.

## Stage 1 — dependency

`go get webtyp.com/icons@v0.0.38` (adds the `code` glyph), `go mod tidy`.

## Stage 2 — `apiexplorer/`

- `explorer.go`: import `webtyp.com/icons/code` instead of `webtyp.com/icons/terminal`;
  `func (e *Explorer) Icon() svg.Icon { return code.Ref }`.
- `svg.go`, replace the body and comment:

```go
// IconSvg ships the explorer's nav glyph (code) and the glyph of the «Developer tools» group
// layout/app.BuildMenu wraps it in (terminal), so both menu entries point at symbols that exist.
func (e *Explorer) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(code.Def(), terminal.Def())
}
```

## Stage 3 — `app/menu.go`

Add, next to `helpAPIExplorer`:

```go
// idDevTools is the menu id of the «Developer tools» group. A group has no permission of its own.
const idDevTools = "devtools"

const (
	labelDevTools lang.Text = "Developer tools"
	helpDevTools  lang.Text = "Tools to inspect how this app is built."
)
```

In `BuildMenu`, replace `mods = append(mods, layout.Describe(explorer, helpAPIExplorer))` with:

```go
	devtools := layout.NewGroupUIModule(idDevTools, lang.Translate(labelDevTools).String(), terminal.Ref,
		[]layout.UIModule{layout.Describe(explorer, helpAPIExplorer)})
	mods = append(mods, layout.Describe(devtools, helpDevTools))
```

(import `webtyp.com/icons/terminal`). The `perms` line for the explorer stays exactly as it is. Change the doc
comment line `// The API explorer is always added last, …` to:
`// The API explorer is always added last, inside the «Developer tools» group, visible to whoever may read router.ResourceIntrospection; the app never adds it.`

## Stage 4 — `lang.json`

Add two keys in alphabetical position (order of `languages`: es, zh, hi, ar, pt, fr, de, ru):

```json
"Developer tools": ["Herramientas de desarrollo", "开发者工具", "डेवलपर टूल", "أدوات المطور", "Ferramentas de desenvolvimento", "Outils de développement", "Entwicklerwerkzeuge", "Инструменты разработчика"],
"Tools to inspect how this app is built.": ["Herramientas para revisar cómo está construida esta aplicación.", "用于检查此应用构建方式的工具。", "यह ऐप कैसे बना है, इसकी जाँच के लिए टूल।", "أدوات لفحص كيفية بناء هذا التطبيق.", "Ferramentas para inspecionar como este aplicativo é construído.", "Outils pour inspecter la construction de cette application.", "Werkzeuge, um zu prüfen, wie diese App aufgebaut ist.", "Инструменты для проверки того, как устроено это приложение."],
```

## Stage 5 — tests

- `tests/menu_test.go`, the test that checks `last.ModelName() != string(router.ResourceIntrospection)`: the last
  module is now the group. Assert `last.ModelName() == "devtools"`, that it satisfies `layout.GroupModule`, and that
  its only child's `ModelName()` is `string(router.ResourceIntrospection)`. The two `CanView` assertions stay.
  `TestBuildMenu_DescriptionFromPresenter` still expects 2 modules (1 screen + the group): fix its message to say so.
- `platformd/platformd_test.go`, new test next to `TestPlatform_GroupModuleAccordion`:

```go
// A group shows only the children the user may see, and is hidden when the user may see none.
func TestPlatform_GroupShowsOnlyAllowedChildren(t *testing.T) {
	allowed := &mockModule{id: "dev.routes", label: "Routes", icon: svg.Icon("code")}
	denied := &mockModule{id: "dev.schema", label: "Schema", icon: svg.Icon("database")}
	group := NewGroupUIModule("dev", "Developer tools", svg.Icon("terminal"), []UIModule{allowed, denied})

	p := &Platform{Modules: []UIModule{group}, CanView: func(id string) bool { return id == "dev.routes" }}
	p.Init(NilCtx())
	html := p.Render().String()
	if !contains(html, "data-id='dev.routes'") {
		t.Error("allowed child must be rendered")
	}
	if contains(html, "data-id='dev.schema'") {
		t.Error("forbidden child must not be rendered, neither as link nor as panel")
	}
	p.Activate("dev.schema")
	if p.active.Get() == "dev.schema" {
		t.Error("a forbidden child must not be activatable by id")
	}

	none := &Platform{Modules: []UIModule{group}, CanView: func(string) bool { return false }}
	none.Init(NilCtx())
	if contains(none.Render().String(), "data-id='dev'") {
		t.Error("a group with no visible child must not be rendered")
	}
}
```

If this test fails, that is a real security defect: fix `platformd.go` so the rule holds, and say so in the PR.

## Stage 6 — docs

- `docs/ARCHITECTURE.md` line «`BuildMenu` adds the API explorer last under `router.ResourceIntrospection`» →
  «`BuildMenu` adds the API explorer last, inside the «Developer tools» group (`terminal` glyph), under
  `router.ResourceIntrospection`».
- `README.md` `### apiexplorer` section: add one sentence: «In the menu it lives inside the «Developer tools» group.»

## Delivery

- `gotest` green (also `-tags browser` tests). Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`,
  `GOFLAGS=-mod=mod`.
- `grep -rn "icons/terminal" --include=*.go .` → only `apiexplorer/svg.go` and `app/menu.go`.
- `git diff --stat main` lists exactly: `go.mod`, `go.sum`, `apiexplorer/explorer.go`, `apiexplorer/svg.go`,
  `app/menu.go`, `lang.json`, `tests/menu_test.go`, `platformd/platformd_test.go` (+ `platformd/platformd.go` only
  if stage 5 found the defect), `docs/ARCHITECTURE.md`, `README.md`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `go.mod`, `go.sum` | `icons` v0.0.38 |
| 2 | `apiexplorer/explorer.go`, `apiexplorer/svg.go` | `code` glyph; sprite ships `code` + `terminal` |
| 3 | `app/menu.go` | explorer inside the «Developer tools» group |
| 4 | `lang.json` | two keys |
| 5 | `tests/menu_test.go`, `platformd/platformd_test.go` | group last; group shows only allowed children |
| 6 | `docs/ARCHITECTURE.md`, `README.md` | group mentioned |
