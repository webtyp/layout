---
PLAN: "feat: layout/apiexplorer — read /_routes, show every endpoint with its access and roles, and call it"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Prerequisite:** `webtyp.com/router` **v0.4.0** published (it adds `router.RouteTable`, the
> decodable reading shape of `/_routes`). `go get webtyp.com/router@v0.4.0`. If that version
> does not exist, stop and report it: do not redeclare the `/_routes` shape here.

# Plan — `layout/apiexplorer`: see and try every endpoint of a running server

## 0. Context (read first)

Every webtyp server can serve its route table at `GET /_routes`
(`router.MountIntrospection`): method, path, access (`public` / `authenticated` / `guarded`),
the permission a guarded route requires (`resource` + CRUD `action` letters), **which roles hold
that permission**, the description, and the argument schema the route declared. Domain
operations are now ordinary routes (`POST /api/<module>/<op>`, `webtyp.com/rpc`), so this table
is the whole API.

What is missing is a screen to **read** it and **try** it, like FastAPI `/docs`, Django REST
Framework's browsable API or `rails routes`. This plan builds it as a layout of this repository,
so an app shows it as one more module of its shell (`platformd.UIModule`), behind a permission,
in development and in production alike.

**Why in the browser and in `layout`:** the route table is already JSON; rendering it server
side would put the UI kit in a Cloudflare Worker (1 MB limit) or in `server/httpd` (absent from
edge deployments). `layout` already depends on `dom`, `html` and `components`; this package adds
`webtyp.com/fetch` (zero-dependency HTTP client) to call endpoints.

This replaces an earlier draft written for a separate `tinywasm/apiexplorer` repository, which
is discarded.

## Design gate

### 1. Prior art
- **FastAPI `/docs` and springdoc (Swagger UI over OpenAPI)**: a list of endpoints, each
  expandable into a form built from its schema, with "Execute". Our interaction model.
- **Django REST Framework browsable API**: rendered for the logged-in user, using the session
  cookie; permissions apply to the explorer exactly as to any client. Our auth model.
- **`rails routes` / `php artisan route:list`**: a flat, sortable table of every route. Our
  table.
None of them shows **which roles hold each permission** or flags permissions **no role holds**;
that diagnosis is why this explorer exists.

### 2. Novice-name test
`apiexplorer.New(apiexplorer.Config{})` — "an API explorer". It is a `platformd.UIModule`, so
`Modules: []platformd.UIModule{..., explorer}` puts it in the menu.

### 3. Complexity ledger
```
Concepts the developer must learn   +1 (apiexplorer.New)
Files they must touch to do X        +1 line in the app's module list
Lines at the call site               +2
Ways to do the same thing            0
```

### 4. Where it belongs
A screen made of existing components: a layout, like `crudview` and `chatview`. The data shape
belongs to `router` (`RouteTable`); this package only renders and calls.

### 5. What this deletes
The discarded `apiexplorer` draft (in `webtyp/app-releases`, removed by the maintainer).

## 1. Target API (package `webtyp.com/layout/apiexplorer`)

```go
// Config configures an explorer. The zero value works: same origin, router.IntrospectionPath.
type Config struct {
    // RoutesURL is where the table is fetched from; "" = router.IntrospectionPath, same origin.
    RoutesURL string
}

// New returns the explorer as a platformd.UIModule (ModelName "api_explorer", Label from the
// layout's lang.json, an icon from the existing icon set). Nothing is fetched until Activate.
func New(cfg Config) (*Explorer, error)

// Explorer implements platformd.UIModule.
type Explorer struct{ /* unexported */ }
```

Follow the structure of `chatview` (`chatview/chatview.go`, `css.go`, tests) and the rules in
this repo's `AGENTS.md`: components implement only `Render()` (+ optional `Init`), state in
signals, value-embedded `dom.Element`, styles through the `widget/style` DSL (no hex literals:
an orphan row is a semantic `danger` state), labels in `lang.json`.

## 2. Behaviour (normative) — the minimum to start

### 2.1 Table
On `Activate`, `GET RoutesURL` with `webtyp.com/fetch` (the browser sends the session cookie on
same-origin requests) and decode into `router.RouteTable` with `webtyp.com/json`.
- Fetch or decode failure, or non-2xx → render the status and message in place. Never an empty
  table (it would read as "this server has no routes"). A 403 says: "you need permission to view
  the API".
- One row per route. Columns, in this order: **method · path · access · permission · roles**.
  - permission = `resource:action` for guarded routes (`patient:cu`), empty otherwise.
  - roles = comma-separated; `PolicyKnown == false` → `—` with a tooltip "the server did not
    describe its policy" (never an empty list).
  - `Orphan()` rows are styled as danger and **sorted first**; `public` rows carry a visible mark.
- A text filter over method + path (use the existing search component of `webtyp/components`
  that `crudview` uses).

### 2.2 Trying a route
Clicking a row expands it (only one expanded at a time):
- **Path parameters**: every `{name}` in the path becomes its own labelled input; the request
  path is rebuilt from them; an empty value disables "Send" (never `/api/sites//x`).
- **Body** (methods with a body): one labelled input per `ArgRecord` (kind decides the input:
  text, number, checkbox); `HasArgs == false` → a plain textarea for raw JSON. Encode the
  inputs into a JSON object with `webtyp.com/json`.
- **Send** button: `fetch` with the route's method, `Content-Type: application/json` for bodies.
  Render status, then body. A `403` is annotated: "requires `<resource:action>`, held by:
  `<roles>`" (or "nobody").
- **Nothing is sent without pressing Send.** Methods other than GET ask for a confirmation
  click ("This will run `POST /api/...` for real") before sending.
- Never persist bodies, responses or credentials (no localStorage).

## 3. Stages

| Stage | Files | Content |
|---|---|---|
| 1 | `go.mod` | `go get webtyp.com/router@v0.4.0 webtyp.com/fetch` |
| 2 | `apiexplorer/explorer.go` | `Config`, `New`, `Explorer` (UIModule), fetch + decode, table |
| 3 | `apiexplorer/try.go` | path builder, body form, send, 403 annotation |
| 4 | `apiexplorer/css.go` (`//go:build !wasm`) | styles via the DSL |
| 5 | `lang.json` | labels (Spanish and English, like the existing entries) |
| 6 | `tests/` | §4 |
| 7 | `apiexplorer/README.md`, `README.md` index, `docs/ARCHITECTURE.md` | what it is, how an app adds it, the loud warning below |

**README warning (verbatim intent):** never expose `/_routes` or this screen publicly in
production — the permission map of a service is a map of what to attack. Mount
`MountIntrospection(...).Requires("api_explorer", model.Read)` and show the module only to roles
that hold that permission.

## 4. Tests (`tests/`)

Pure logic must be testable without a browser — keep it out of `//go:build wasm` files:
1. Path builder: `/api/sites/{site}/assets/{key}` + `{site: a, key: b}` → `/api/sites/a/assets/b`;
   an empty value → error, not `/api/sites//assets/b`.
2. Sorting: every `Orphan()` route before every other route; stable otherwise (path order).
3. Body encoding from `ArgRecord`s: text, number and bool inputs produce the right JSON types.
4. 403 annotation text with roles, with no roles ("nobody"), with `PolicyKnown == false`.
5. WASM view test (pattern of `chatview_wasm_test.go`): a fake `/_routes` response (mock the
   fetch layer the way other wasm tests in this repo do) renders one row per route, orphans first,
   and expanding a row does **not** issue any request until Send is clicked.

## 5. Code rules (non-negotiable)
- WASM code: `webtyp.com/fmt`, `webtyp.com/json`; no `fmt`, `strings`, `errors`, `encoding/json`,
  `reflect` from the standard library; no `map`.
- Value embedding of `dom.Element`; no `ssr.go` / `front.go`; CSS only in `css.go`.
- No exported symbol beyond §1. Tests in `tests/`; never export for a test.

## 6. Acceptance criteria
- `gotest ./...` green (stdlib and wasm).
- `grep -rn "map\[\|\*dom.Element" apiexplorer/` → empty.
- No request is issued by rendering or expanding (asserted by test 5).
