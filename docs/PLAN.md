---
PLAN: "feat!: developer tools only when there is an API or a database; a demo shows them from its own operations and migrations"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Developer tools from a source

## Context

`BuildMenu` always adds the «Developer tools» group with the API explorer and the database diagram
(`app/menu.go:120-131`), and both screens always read from a server (`apiexplorer`: `router.IntrospectionPath`;
`schemaview`: `erd.SchemaPath`). Two problems: a module's **demo** runs in the browser with no server, so both screens
are empty (404); and an app with **no API and no database** (fully offline) still shows two empty tools. Decided:

- each screen reads from a **source**: the server (as today) or data built in process;
- the tools are a **required** parameter of `BuildMenu`: a nil tool is left out, and with none the whole group is
  left out — every app and demo states what it has, and the compiler makes it.
- a demo builds its sources with the same pieces the server uses: `rpc.Routes` + `router.Table` for the routes,
  `erd.ReadSteps` for the diagram.

In a demo there is no server to send a request to, so the explorer's «try» form is shown only with the server source.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/router@v0.4.17
go get webtyp.com/rpc@v0.2.15
go get webtyp.com/erd@v0.0.14
go get webtyp.com/ddl@v0.0.36
go mod tidy
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.

## Ecosystem search

`router.Table` (v0.4.17), `rpc.Routes` (v0.2.15) and `erd.ReadSteps` (v0.0.14) were built for this — **reuse**. No
other package builds a route table or a diagram. `apiexplorer`, `schemaview` and `app` are **extended**;
`apiexplorer.Config.RoutesURL` is **deleted** (only `app/menu.go` built a `Config`, empty).

## Design gate

1. **Prior art.** Django Debug Toolbar and Spring Boot Actuator show their panels only when the feature behind them is
   present (unverified: pages not read in this session). Storybook-style demos render a component with in-process data
   instead of a backend (unverified). Same here: a tool appears only with a source, and a demo gives in-process data.
2. **Novice names.** `app.Tools{API, Database}`, `apiexplorer.FromServer()`, `apiexplorer.FromTable(t)`,
   `schemaview.FromServer()`, `schemaview.FromReport(r)`, `app.LocalTools(modules, steps...)`.
3. **Complexity ledger.** `BuildMenu` +1 required parameter · ways to configure the explorer: 2 named constructors
   instead of a struct with a URL · an app with neither API nor database: −1 empty group.
4. **Where it belongs.** `layout` (the screens and the menu). Building the data stays in `router`, `rpc`, `erd`.
5. **Deletes.** `apiexplorer.Config.RoutesURL` and `Explorer.routesURL`; the unconditional devtools group.

## APIs used

| Symbol | Exact signature | Tag |
|---|---|---|
| `router.Table` | `func Table(routes []RouteInfo, policy model.PolicyDescriber) RouteTable` | `webtyp.com/router` v0.4.17 |
| `rpc.Routes` | `func Routes(prefix string, modules ...router.OperationModule) ([]router.RouteInfo, error)`; `rpc.DefaultPrefix = "/api"` | `webtyp.com/rpc` v0.2.15 |
| `erd.ReadSteps` | `func ReadSteps(steps ...ddl.Step) (Diagram, error)`; `(Diagram).Report() Report` | `webtyp.com/erd` v0.0.14 |
| `ddl.Step` | `struct { Name string; Run func(conn Execer, compiler Compiler) error }` | `webtyp.com/ddl` v0.0.36 |
| `layout.NewGroupUIModule`, `layout.Describe` | as used in `app/menu.go` today | this repo |

If a signature differs, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — tests first

Red after this stage.

- `tests/menu_test.go`: every existing `app.BuildMenu(screens, ids, allows)` call gains a fourth argument. Where the
  test expects the developer tools (`TestBuildMenu_AddsDevToolsLast`), pass
  `app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()}`; everywhere else pass `app.Tools{}`
  and adjust only the module counts that included the group. Add:

| Test | Call | Expected |
|---|---|---|
| `TestBuildMenu_NoTools_NoGroup` | `BuildMenu(screens, ids, allowAll, app.Tools{})` | no module with id `"devtools"` |
| `TestBuildMenu_OnlyAPI` | `Tools{API: apiexplorer.FromServer()}` | the `"devtools"` group has exactly one child, the explorer |
| `TestBuildMenu_OnlyDatabase` | `Tools{Database: schemaview.FromServer()}` | one child, the diagram |
| `TestLocalTools` (table) | `app.LocalTools(mods, steps...)` with: a fake `router.OperationModule` declaring two operations and one `ddl.Step` creating one table; no modules and no steps; a step whose `Run` returns `errors.New("boom")` | `API != nil && Database != nil`; both nil; error containing the step name and `boom` |

  Find group children the same way `TestBuildMenu_AddsDevToolsLast` already does.
- `apiexplorer/explorer_test.go` (internal, already justified): replace `New(Config{})` with
  `New(FromTable(router.RouteTable{}))`; add `TestNewWithoutSource` (`New(nil)` → error text `errNoSource`),
  `TestFromTableCannotTry` (`FromTable(t).canTry == false`, `FromServer().canTry == true`) and `TestFromTableLoads`
  (`FromTable(t).load` calls `done` with `t` and `nil`).
- `schemaview/source_test.go` (new, internal `package schemaview`, first line comment: `// Root-level test (justified): the source's load function is unexported.`):
  `New(nil)` → error text `errNoSource`; `FromReport(r).load` calls `done` with `r` and `nil`.

Commit after this stage.

## Stage 2 — `apiexplorer`

New file `apiexplorer/source.go`:

```go
package apiexplorer

import (
	"webtyp.com/fetch"
	"webtyp.com/json"
	"webtyp.com/lang"
	"webtyp.com/router"

	. "webtyp.com/fmt"
)

const errNoSource = "apiexplorer: no source: use FromServer or FromTable"

// Config says where the explorer reads its route table and whether a route can
// be tried. Build it with FromServer or FromTable.
type Config struct {
	load   func(done func(router.RouteTable, error))
	canTry bool
}

// FromServer reads the table the app's server serves at router.IntrospectionPath;
// each route can be tried against that server.
// Example: app.Tools{API: apiexplorer.FromServer()}
func FromServer() *Config {
	return &Config{load: fetchRoutes(router.IntrospectionPath), canTry: true}
}

// FromTable shows a table built in process — a module's demo, which has no
// server, so routes cannot be tried.
// Example: apiexplorer.FromTable(router.Table(routes, nil))
func FromTable(t router.RouteTable) *Config {
	return &Config{load: func(done func(router.RouteTable, error)) { done(t, nil) }}
}

func fetchRoutes(url string) func(done func(router.RouteTable, error)) {
	return func(done func(router.RouteTable, error)) {
		fetch.Get(url).Send(func(resp *fetch.Response, err error) {
			if err != nil {
				done(router.RouteTable{}, err)
				return
			}
			if resp.Status == 403 {
				done(router.RouteTable{}, Err(lang.Translate("you need permission to view the API").String()))
				return
			}
			if resp.Status != 200 {
				done(router.RouteTable{}, Err(lang.Translate("Error").String()+" "+Sprintf("%d", resp.Status)))
				return
			}
			var table router.RouteTable
			if err := json.Decode(resp.Body(), &table); err != nil {
				done(router.RouteTable{}, Err(lang.Translate("Failed to decode routes").String()+": "+err.Error()))
				return
			}
			done(table, nil)
		})
	}
}
```

In `explorer.go`:
- `Config` (the old struct with `RoutesURL`) and the `routesURL` field are deleted.
- `New(cfg *Config) (*Explorer, error)`: `if cfg == nil { return nil, Err(errNoSource) }`; store `cfg`; keep the rest.
- `load()` becomes:

```go
func (e *Explorer) load() {
	e.cfg.load(func(table router.RouteTable, err error) {
		if err != nil {
			e.errorMsg.Set(err.Error())
			return
		}
		e.tableData = &table
		e.updateRows()
	})
}
```

- Where the row gets its try panel (`explorer.go:251-252`), add it only when `e.cfg.canTry`.
- Remove imports that become unused (`fetch`, `json` may move entirely to `source.go`).

## Stage 3 — `schemaview`

New file `schemaview/source.go`:

```go
package schemaview

import (
	"webtyp.com/erd"
	"webtyp.com/fetch"
	"webtyp.com/json"
	"webtyp.com/lang"

	. "webtyp.com/fmt"
)

const errNoSource = "schemaview: no source: use FromServer or FromReport"

// Source says where the diagram comes from. Build it with FromServer or FromReport.
type Source struct {
	load func(done func(erd.Report, error))
}

// FromServer reads the diagram the app's server serves at erd.SchemaPath.
// Example: app.Tools{Database: schemaview.FromServer()}
func FromServer() *Source { return &Source{load: fetchReport(erd.SchemaPath)} }

// FromReport shows a diagram built in process (a module's demo).
// Example: schemaview.FromReport(diagram.Report())
func FromReport(r erd.Report) *Source {
	return &Source{load: func(done func(erd.Report, error)) { done(r, nil) }}
}

func fetchReport(url string) func(done func(erd.Report, error)) {
	return func(done func(erd.Report, error)) {
		fetch.Get(url).Send(func(resp *fetch.Response, err error) {
			if err != nil {
				done(erd.Report{}, err)
				return
			}
			if resp.Status == 403 {
				done(erd.Report{}, Err(lang.Translate("no access to", "database diagram").String()))
				return
			}
			if resp.Status != 200 {
				done(erd.Report{}, Err(lang.Translate("Could not", "load", "database diagram").String()+" "+Sprintf("%d", resp.Status)))
				return
			}
			var rep erd.Report
			if err := json.Decode(resp.Body(), &rep); err != nil {
				done(erd.Report{}, Err(lang.Translate("Could not", "load", "database diagram").String()+": "+err.Error()))
				return
			}
			done(rep, nil)
		})
	}
}
```

In `schemaview.go`: `View` gains `src *Source`; `New(src *Source) (*View, error)` returns `Err(errNoSource)` for nil;
`Activate` calls `v.src.load(func(rep erd.Report, err error) { … })` with exactly the current post-processing (the
module-label translation of the SVG, `svgNodes`, `updateRows`) and sets `errorMsg` from `err`.

## Stage 4 — `app/menu.go`

Add:

```go
// Tools are the developer screens under «Developer tools». A nil field leaves
// that screen out; Tools{} leaves the whole group out — an app with no API and
// no database shows no developer tools.
// Example (an app with a server): app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()}
type Tools struct {
	API      *apiexplorer.Config
	Database *schemaview.Source
}

// LocalTools builds the tools of a page with no server — a module's demo — with
// the pieces the server uses: the routes its operations declare (rpc.Routes)
// and its migration steps (erd.ReadSteps). No modules leaves the explorer out;
// no steps leaves the diagram out.
// Example: tools, err := app.LocalTools([]router.OperationModule{mod}, ddl.Step{Name: "shop", Run: migrate.Migrate})
func LocalTools(modules []router.OperationModule, steps ...ddl.Step) (Tools, error) {
	var t Tools
	if len(modules) > 0 {
		routes, err := rpc.Routes(rpc.DefaultPrefix, modules...)
		if err != nil {
			return Tools{}, err
		}
		t.API = apiexplorer.FromTable(router.Table(routes, nil))
	}
	if len(steps) > 0 {
		d, err := erd.ReadSteps(steps...)
		if err != nil {
			return Tools{}, err
		}
		t.Database = schemaview.FromReport(d.Report())
	}
	return t, nil
}
```

`BuildMenu` gets a fourth parameter `tools Tools` (update its doc comment: the group appears only with at least one
tool). Replace the block from `explorer, err := apiexplorer.New(apiexplorer.Config{})` to the `perms = append(...)` of
the two tools with:

```go
	var devs []layout.UIModule
	if tools.API != nil {
		explorer, err := apiexplorer.New(tools.API)
		if err != nil {
			return Menu{}, err
		}
		devs = append(devs, layout.Describe(explorer, helpAPIExplorer))
		perms = append(perms, screenPerm{id: explorer.ModelName(), permission: router.ResourceIntrospection})
	}
	if tools.Database != nil {
		schema, err := schemaview.New(tools.Database)
		if err != nil {
			return Menu{}, err
		}
		devs = append(devs, layout.Describe(schema, helpSchemaView))
		perms = append(perms, screenPerm{id: schema.ModelName(), permission: erd.ResourceSchema})
	}
	if len(devs) > 0 {
		devtools := layout.NewGroupUIModule(idDevTools, lang.Translate(labelDevTools).String(), terminal.Ref, devs)
		mods = append(mods, layout.Describe(devtools, helpDevTools))
	}
```

Run `gotest`: green. Commit after stages 2–4 (one commit per stage).

## Stage 5 — translations and docs

- `langc sync .`; fill every empty value (zero untranslated, zero reuse warnings). The texts moved, not new ones,
  except the two `errNoSource` messages if `langc` collects them.
- `README.md` (and `apiexplorer/README.md`): how an app with a server declares its tools, how a demo uses
  `app.LocalTools`, and that no tools means no group.

Commit after this stage.

## Delivery check

1. `gotest` passes (native and wasm). A PR that does not compile, or whose tests could not run, is a failed delivery.
2. `grep -rn "RoutesURL\|routesURL" --include=*.go .` → empty; `grep -n "apiexplorer.Config{}" app/menu.go` → empty.
3. `langc sync .` → zero untranslated, zero reuse warnings.
4. `git diff --stat main` lists only: `apiexplorer/source.go`, `apiexplorer/explorer.go`, `apiexplorer/explorer_test.go`,
   `schemaview/source.go`, `schemaview/schemaview.go`, `schemaview/source_test.go`, `app/menu.go`, `tests/menu_test.go`,
   `lang.json` (if changed), `README.md`, `apiexplorer/README.md`, `go.mod`, `go.sum`.

## Delivery rules

- Commit after every stage. Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
- Stuck for 30 minutes: open the PR anyway, listing what fails.
- The apps and demos that call `BuildMenu` stop compiling until they pass their tools: that is intended; they are
  updated by separate plans. Do not touch other repositories.

## Stages

| # | Files | Result |
|---|---|---|
| 1 | `tests/menu_test.go`, `apiexplorer/explorer_test.go`, `schemaview/source_test.go` | Tools present/absent, local tools, no source; red |
| 2 | `apiexplorer/source.go`, `apiexplorer/explorer.go` | Explorer reads from a source; try only with a server |
| 3 | `schemaview/source.go`, `schemaview/schemaview.go` | Diagram reads from a source |
| 4 | `app/menu.go` | `Tools`, `LocalTools`, `BuildMenu(..., tools)`; green |
| 5 | `lang.json`, `README.md`, `apiexplorer/README.md` | Translated and documented |
