---
PLAN: "feat: schemaview — the «Database» screen of the Developer tools group"
TAG: v0.3.81
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 3069625319767653291
PR: https://github.com/webtyp/layout/pull/52
REVIEWED_COMMIT: eb052d10a132d358a6b6145e53224f7650815bb8
---

> Read AGENTS.md at the repo root before any change.

<!-- Parked: rename to docs/PLAN.md and dispatch only after PLAN_DEVTOOLS_MENU (v0.3.80) is closed and published. -->

# Plan — `webtyp.com/layout`: `schemaview`

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/erd@v0.0.2 webtyp.com/icons@v0.0.38
```

## Why

An app with a database serves its diagram at `GET /_schema` (`webtyp.com/appkit`, permission `db_schema:read`):
one SVG with every table grouped by module, plus the list of model errors. This screen shows it as the second
child of the «Developer tools» group, after the API explorer. Each child keeps its own permission; the group
shows only what the user may see. An app without a database never grants `db_schema`, so the link never shows.

## Ecosystem search

`layout/apiexplorer` is the model: same shape (fetch on `Activate`, search bar, error line, list). `erd`
(v0.0.2) owns the path, the permission and the decodable `Report`. `icons/database` (v0.0.38) is the glyph. New:
package `schemaview` (one screen, like `apiexplorer`).

## APIs used

| Symbol | Signature | Tag | File |
|---|---|---|---|
| `erd.SchemaPath` | `const SchemaPath = "/_schema"` | erd v0.0.2 | `erd.go` |
| `erd.ResourceSchema` | `const ResourceSchema model.Resource = "db_schema"` | erd v0.0.2 | `erd.go` |
| `erd.IDPrefix` | `const IDPrefix = "erd-table-"` — each table in the SVG is `<g id="erd-table-<name>">` | erd v0.0.2 | `draw.go` |
| `erd.Report` | `type Report struct { SVG string; Findings []Finding }` (implements `model.Decodable` by pointer) | erd v0.0.2 | `erd.go` |
| `erd.Finding` | `type Finding struct { Rule Rule; Table, Column, Target string }` | erd v0.0.2 | `erd.go` |
| `erd.Rule` | `func (r Rule) Level() Level` (`erd.LevelError` / `erd.LevelWarning`); `func (r Rule) Text() lang.Text` | erd v0.0.2 | `erd.go` |
| `database.Ref`, `database.Def` | `const Ref = svg.Icon("database")`; `func Def() sprite.Definition` (svg.go, `!wasm`) | icons v0.0.38 | `database/` |
| `dom.Get` | `func Get(id string) (Reference, bool)`; `Reference.ScrollIntoView()` | as in `go.mod` | `dom.go`, `reference.go` |
| `dom.Trust` / `(*Element).Raw` | `func Trust(html string) TrustedHTML`; `func (b *Element) Raw(h TrustedHTML) *Element` | as in `go.mod` | `trusted.go`, `element.go` |

The rest (`fetch.Get`, `json.Decode`, `searchbar`, `NewString`, `NewNodes`, `widget.Name`) is used exactly as in
`apiexplorer/explorer.go`: read that file first and mirror it. If a signature in the published tag differs from
this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — `schemaview/schemaview.go` (new)

```go
// Package schemaview is the «Database» screen: the app's database diagram (GET erd.SchemaPath) and the list of
// model errors, each one jumping to its table.
package schemaview

import (
	"webtyp.com/components/searchbar"
	"webtyp.com/erd"
	"webtyp.com/fetch"
	"webtyp.com/icons/database"
	"webtyp.com/json"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
)

// View is a module of the platform shell: proven at compile time.
var _ layout.UIModule = (*View)(nil)

const NameSchemaView widget.Name = "schemaview"

var (
	clsRoot     = NameSchemaView.Root()
	clsHeader   = NameSchemaView.Class("header")
	clsError    = NameSchemaView.Class("error")
	clsFindings = NameSchemaView.Class("findings")
	clsFinding  = NameSchemaView.Class("finding")
	clsBad      = NameSchemaView.Class("finding-error")
	clsCanvas   = NameSchemaView.Class("canvas")
)

type View struct {
	Element
	errorMsg *SignalString
	search   *SignalString
	filter   *searchbar.SearchBar
	report   *erd.Report
	svgNodes *SignalNodes
	rows     *SignalNodes
}

func New() *View {
	return &View{errorMsg: NewString(""), search: NewString(""), svgNodes: NewNodes(), rows: NewNodes()}
}

func (v *View) ModelName() string        { return string(erd.ResourceSchema) }
func (v *View) Label() string            { return lang.Translate("Database").String() }
func (v *View) Icon() svg.Icon           { return database.Ref }
func (v *View) View() Component          { return v }
func (v *View) WidgetName() widget.Name  { return NameSchemaView }
func (v *View) WidgetKind() widget.Kind  { return widget.Region }

func (v *View) Init(ctx Ctx) {
	v.filter = &searchbar.SearchBar{Placeholder: lang.Text("Search tables")}
	v.filter.OnFilterChange(func(term string) {
		v.search.Set(term)
		v.updateRows()
		v.jumpTo(term)
	})
}

// Activate fetches the diagram each time the screen opens (the schema changes when the app is redeployed).
func (v *View) Activate() {
	v.errorMsg.Set("")
	fetch.Get(erd.SchemaPath).Send(func(resp *fetch.Response, err error) {
		if err != nil {
			v.errorMsg.Set(err.Error())
			return
		}
		if resp.Status == 403 {
			v.errorMsg.Set(lang.Translate("no access to", "database diagram").String())
			return
		}
		if resp.Status != 200 {
			v.errorMsg.Set(lang.Translate("could not load", "database diagram").String() + " " + Sprintf("%d", resp.Status))
			return
		}
		var rep erd.Report
		if err := json.Decode(resp.Body(), &rep); err != nil {
			v.errorMsg.Set(lang.Translate("could not load", "database diagram").String() + ": " + err.Error())
			return
		}
		v.report = &rep
		// The SVG is produced by this app's own server (erd), never by a user: trusted markup.
		v.svgNodes.Set([]*Element{Div().Raw(Trust(rep.SVG))})
		v.updateRows()
	})
}

// updateRows lists the findings whose table or column contains the search term; errors first (erd sorts them).
func (v *View) updateRows() {
	if v.report == nil {
		return
	}
	term := Convert(v.search.Get()).ToLower().String()
	var rows []*Element
	for _, f := range v.report.Findings {
		where := f.Table
		if f.Column != "" {
			where += "." + f.Column
		}
		if term != "" && !Contains(Convert(where).ToLower().String(), term) {
			continue
		}
		text := lang.Translate(f.Rule.Text()).String()
		if f.Target != "" {
			text += " (" + f.Target + ")"
		}
		row := Div().Set(clsFinding.AsAttr())
		if f.Rule.Level() == erd.LevelError {
			row.Set(clsBad.AsAttr())
		}
		table := f.Table
		row.Child(Span().Text(where), Span().Text(text)).OnClick(func(Event) { v.jumpTo(table) })
		rows = append(rows, row)
	}
	v.rows.Set(rows)
}

// jumpTo scrolls the diagram to the table named exactly name; nothing happens when there is none.
func (v *View) jumpTo(name string) {
	if ref, ok := Get(erd.IDPrefix + name); ok {
		ref.ScrollIntoView()
	}
}

func (v *View) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())
	header := Div().Set(clsHeader.AsAttr())
	if v.filter != nil {
		header.Child(v.filter.Render())
	}
	root.Child(header)
	root.Child(Div().Set(clsError.AsAttr()).BindText(v.errorMsg))
	root.Child(Div().Set(clsFindings.AsAttr()).BindChildren(v.rows))
	root.Child(Div().Set(clsCanvas.AsAttr()).BindChildren(v.svgNodes))
	return root
}
```

If `Set` on an `Element` replaces the class instead of adding it (check `apiexplorer`'s rows: they combine
`clsRow` with state classes), use the same mechanism `apiexplorer` uses for `clsOrphan`.

## Stage 2 — `schemaview/css.go` and `schemaview/svg.go` (new, `//go:build !wasm`, asset producers)

`css.go` — mirror `apiexplorer/css.go`:

```go
func (v *View) RenderSheet() *style.Sheet {
	return style.For(v).
		Root(style.Fill(), style.Stack(style.Space3), style.Pad(style.Space3)).
		Part("header", style.Stack(style.Space2), style.KeepSize()).
		Part("error", style.As(style.Danger)).
		Part("findings", style.Stack(style.Space1), style.KeepSize()).
		Part("finding", style.Row(style.Space3), style.Pad(style.Space2), style.As(style.Page), style.Round(style.RadiusMd), style.Interactive(style.Page)).
		Part("finding-error", style.As(style.DangerWash)).
		Part("canvas", style.Scroll(), style.As(style.Panel), style.Round(style.RadiusMd))
}

func (v *View) RenderCSS() *css.Stylesheet { return v.RenderSheet().Stylesheet() }
```

The canvas must scroll both ways (the SVG is wider than the stage). If `style.Scroll()` scrolls only vertically,
that is a missing recipe: stop and say so in the PR (rule of the skill widget-styling); do not hand-write CSS.

`svg.go`:

```go
// IconSvg ships the screen's nav glyph (database) to the page sprite.
func (v *View) IconSvg() *sprite.Sprite { return sprite.NewSprite(database.Def()) }
```

## Stage 3 — `app/menu.go`: second child of the group

In `BuildMenu`, after `explorer, err := apiexplorer.New(...)` and its error check:

```go
	schema := schemaview.New()
```

The group's children become `[]layout.UIModule{layout.Describe(explorer, helpAPIExplorer), layout.Describe(schema, helpSchemaView)}`,
with `const helpSchemaView lang.Text = "Every table of this app, its references and the model errors to fix."` next to
`helpAPIExplorer`, and one more permission line:
`perms = append(perms, screenPerm{id: schema.ModelName(), permission: erd.ResourceSchema})`.
Update the `BuildMenu` doc comment: «… inside the «Developer tools» group, with the database diagram, visible to
whoever may read erd.ResourceSchema».

## Stage 4 — translations

`langc sync .`, then fill every empty value. New keys and their Spanish (fill the other 7 languages too):
`"Database"` → «Base de datos»; `"Search tables"` → «Buscar tablas»; `"database diagram"` → «diagrama de la base de
datos»; `"no access to"` → «sin acceso al»; `"could not load"` → «no se pudo cargar el»;
`"Every table of this app, its references and the model errors to fix."` → «Todas las tablas de esta aplicación,
sus referencias y los errores del modelo por corregir.». `langc sync .` must end with zero untranslated keys and zero
reuse warnings.

## Stage 5 — tests

- `tests/menu_test.go`: the group (`"devtools"`) has two children, `string(router.ResourceIntrospection)` then
  `string(erd.ResourceSchema)`; with `allows` true only for `erd.ResourceSchema`, `m.CanView("db_schema")` is true and
  `m.CanView(string(router.ResourceIntrospection))` false; with `allows` false for both, both are false.
- `schemaview/schemaview_wasm_test.go` (`//go:build wasm`, root package like `apiexplorer/explorer_wasm_test.go`,
  same mount setup): set `v.report = &erd.Report{SVG: "<svg><g id=\"erd-table-user\"></g></svg>", Findings:
  []erd.Finding{{Rule: erd.RuleMissingTable, Table: "user", Column: "tenant_id", Target: "tenant"}, {Rule:
  erd.RuleIsolatedTable, Table: "setting"}}}`, call `v.svgNodes.Set(...)` as `Activate` does and `v.updateRows()`,
  render, then: two `.schemaview__finding` rows; one of them has `schemaview__finding-error`; after
  `v.search.Set("tenant")` + `v.updateRows()` only one row remains; `document.getElementById("erd-table-user")`
  exists.

## Stage 6 — docs

- `README.md`: a `### schemaview` section after `### apiexplorer`: what it shows, `GET /_schema`, permission
  `db_schema:read`, never public, «served by `webtyp.com/appkit` when the app passes `Migrations`».
- `docs/ARCHITECTURE.md`: in the tree add `schemaview/    # the database diagram: reads /_schema (erd.Report)`, and
  in the `BuildMenu` paragraph say the group holds the explorer and the database diagram.

## Delivery

- `gotest` green (also `-tags browser` and the wasm pass). Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`,
  `GOFLAGS=-mod=mod`. A PR that does not compile, or whose tests could not run, is a failed delivery.
- `langc sync .` → no untranslated key, no reuse warning.
- `git diff --stat main` lists: `go.mod`, `go.sum`, `schemaview/schemaview.go`, `schemaview/css.go`,
  `schemaview/svg.go`, `schemaview/schemaview_wasm_test.go`, `app/menu.go`, `lang.json`, `tests/menu_test.go`,
  `README.md`, `docs/ARCHITECTURE.md`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `schemaview/schemaview.go` | the screen |
| 2 | `schemaview/css.go`, `schemaview/svg.go` | layout and glyph |
| 3 | `app/menu.go` | second child of «Developer tools», `db_schema` |
| 4 | `lang.json` | keys translated |
| 5 | `tests/menu_test.go`, `schemaview/schemaview_wasm_test.go` | permissions per child; rows and jump |
| 6 | `README.md`, `docs/ARCHITECTURE.md` | documented |
