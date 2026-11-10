---
PLAN: "refactor!: layout implements view.Layout and no longer imports appkit (D50); crudview runs list-wide actions with router.Operation; the demo loopback takes the (absent) publisher"
EXECUTOR: antigravity
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage and stop. Never ask whether to
> proceed. This plan runs **locally**: do not tag, push or publish anything; the owner reviews first.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Goal

Two changes in one delivery: **(a)** the type follows `router.Operation` (stages 1–3, below); **(b)** the layout stops importing `appkit` (decision D50, stage 5): an app's `main` picks the layout and hands it to `appkit.Browser(app, layout)`; `layout/app` offers `New(tools) view.Layout`, whose `Show(view.Session)` draws what `appkit` built. `appkit` and `layout` never import each other; the contract between them is `view`.

`view.Action.Operation`, `view.Actioner.Run` and `view.ActionRunner.RunAction` now carry a `router.Operation` (view plan). `layout/crudview` draws the action buttons and runs them with a string; `layout/app/demo.go` builds the in-memory
loopback with the old signature. This plan follows the type: no behaviour changes, no new screen.

## Ecosystem search

Searched `actionEnabled`, `runAction`, `OnAction`, `actionRunning`, `LoopbackActingAs` over `webtyp/*` and `veltylabs/*`: the action band lives only in `layout/crudview/crudview.go`; `layout/app/demo.go` is the only caller of the loopback in this repo. **Extend** `layout`; no new library.

## Design gate

1. **Prior art**: none new (type propagation of the `router`/`view` plans).
2. **Novice-name test**: no new name; `OnAction(op router.Operation, err error)` tells which operation ran.
3. **Ledger**: concepts 0; operation strings in logic −6.
4. **Where**: `crudview` is the list screen; the identity type is `router.Operation`.
5. **Deletes**: string comparisons of operations; `actionNamePrefix+a.Operation` concatenation of a string field.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

Dependencies (all published, no `replace`): `go get webtyp.com/model@v0.2.18 webtyp.com/router@v0.4.21 webtyp.com/crud@v0.0.10 webtyp.com/view@v0.6.71 webtyp.com/mcp@v0.3.24 webtyp.com/rpc@v0.2.21 webtyp.com/fetch@v0.1.30 && go mod tidy`; afterwards `webtyp.com/appkit` must leave `go.mod` (stage 5). Prerequisite: `view` v0.6.71 (published), `rpc` and `fetch` published.

## APIs used

| symbol | exact signature | file |
|---|---|---|
| `router.Operation` | `struct{ Module, Entity string; Action model.Action }`; `String() string`; comparable | `router/op.go` (router plan) |
| `view.Action` | `struct{ Operation router.Operation; Label, Confirm lang.Text; Args func(records []model.Model) model.Encodable }` | `view/action.go` (view plan) |
| `view.Actioner` | `Actions() []Action`; `Run(op router.Operation, done func(error))` | `view/action.go` (view plan) |
| `crud.LoopbackActingAs` | `func LoopbackActingAs(pub events.Publisher, userID string, mods ...Module) router.Caller` | `crud/loopback.go` (crud plan) |
| `fetch.Get/Post/Put/Patch/Delete` | `func X(endpoint any) *Request`; `.ContentTypeJSON()`, `.Body([]byte)`, `.Send(cb func(*Response, error))` | `fetch/fetch.go` (fetch plan adds `Patch`) |
| `view.Session` | `struct{ Brand Brand; Caller router.Caller; Zone time.Zone; Screens []Screen; Entry EntryScreens; Log func(...any) }` | `view/session.go` |
| `view.Layout` | `interface{ Show(Session) }` | `view/session.go` |
| `router.QueryEscape` | `func QueryEscape(s string) string` | `router/query_codec.go` (router plan) |
| `signal string` of `actionRunning` | `Get() string`, `Set(string)` — read `crudview.go:215-216` for its exact type | `crudview.go` |

If a signature differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Test fixtures against `crud` v0.0.10 (a module fixture that implements `crud.Module` or `crud.Registry`)

`crud` v0.0.8, v0.0.9 and v0.0.10 changed the shapes below; every fixture and fake in this repo's tests follows them. The reference is `crud`'s own `tests/stored_entity_test.go` (`testModule`, `declareRules`): it is in the module cache at `webtyp.com/crud@v0.0.10/tests/` (`go list -m -f '{{.Dir}}' webtyp.com/crud`).

| Old | New (`crud` v0.0.10) |
|---|---|
| `crud.Caller` (hook and read arguments) | `crud.CurrentUser` (same field `UserID`) |
| `reg.Read(&Rec{}, nil, func(c crud.Caller, args crud.ReadArgs, f model.Model) (model.ModelSlice, error) {…})` | `reg.AddReadOnlyEntity(&Rec{}, nil, func(c crud.CurrentUser, args crud.ReadArgs, f model.Model) (model.ModelSlice, error) {…})` (same parameters, new name) |
| `reg.Table(crud.Table{…})` | `reg.AddStoredEntity(crud.StoredEntity{…})`; its hooks return `(crud.RuleAnswer, error)`: `return crud.ApplyAndPublish, nil` (an error: `return 0, err`); `Rules: model.Rules{…}` is required (`HaveNoBusinessRule` where a hook is nil, `AddMyBusinessRule` where it is set) |
| a `crud.Module` has `ModelName()` and `MountOperations(reg crud.Registry)` | it also has `Definition() model.ModuleDefinition`, listing **exactly** the entities it mounts, or `crud.Mount` refuses it (`ErrGeneratedOutOfDate`) |
| `crud.Op(&Rec{}, model.Read)` (a string) | `router.Operation{Module: "<ModelName()>", Entity: "<entity ModelName()>", Action: model.Read}` |
| `crud.Mount(reg, public, m)` | `crud.Mount(reg, public, nil, m)` (the new third argument is the `events.Publisher`; `nil` = nobody listens) |
| `crud.Loopback(m...)`, `crud.LoopbackActingAs(id, m...)` | `crud.Loopback(nil, m...)`, `crud.LoopbackActingAs(nil, id, m...)` |

`Definition()` for the two usual fixtures:

```go
// a read-only entity mounted with AddReadOnlyEntity (a Wire model, no table)
func (m myModule) Definition() model.ModuleDefinition {
	return model.ModuleDefinition{Name: m.ModelName(), Entities: []model.Definition{{
		Name: (&Rec{}).ModelName(), Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule}}}}
}

// a stored entity mounted with AddStoredEntity: copy what the entity says about itself
func (m myModule) Definition() model.ModuleDefinition {
	return model.ModuleDefinition{Name: m.ModelName(), Entities: []model.Definition{{
		Name: m.entity.Model.ModelName(), Allow: m.entity.Model.(model.Allower).Allow(), Rules: m.entity.Rules}}}
}
```

There is no `model.BaseData` and no `crud.Table`/`crud.Op` any more: if a name is not in `crud` v0.0.10, `go doc webtyp.com/crud` lists what is.

## Stage 1 — tests first (red)

| Where | Old | New |
|---|---|---|
| `crudview/actions_test.go:53,83,105,133`, `actions_wasm_test.go:41`, `consumer_stylesheet_test.go:231` | `view.Action{Op: "apply", …}`, `{Op: "resend", …}` | `view.Action{Operation: router.Operation{Module: "m", Entity: "apply", Action: model.Create}, …}` (distinct `Entity` per old name: `apply`, `resend`) |
| the same files, assertions and drivers that compare or click by op string (`ClickAction("apply")`, `ActionEnabled("apply")`, the button `name` `cv-action-apply`) | string | the `router.Operation`; the button name is `actionNamePrefix + op.String()` → `cv-action-m.apply.create` |
| `tests/menu_test.go:383` | `reg.Read(&testSpecialty{}, nil, func(c crud.Caller, …)` | `func(c crud.CurrentUser, …)`; its module fixture implements `Definition()` (`Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule}` entry for the wire entity) |
| every `crud.Loopback(mods…)` / `crud.LoopbackActingAs(user, mods…)` in tests | old signature | `crud.Loopback(nil, mods…)` / `crud.LoopbackActingAs(nil, user, mods…)` |
| `OnAction` hooks in tests | `func(op string, err error)` | `func(op router.Operation, err error)` |

New tests (`apiexplorer/try_test.go`):

| Test | setup → call → expected |
|---|---|
| `TestBuildQuery_EscapesAndSkipsEmpty` | args `username` (string) `"jules ñ"`, `age` (int) `"42"`, `active` (bool) `"true"`, `note` (string) `""` → `"username=jules%20%C3%B1&age=42&active=true"` (the empty `note` is left out) |
| `TestBuildQuery_NoArgsIsEmpty` | no args → `""` |

New test (`crudview/actions_test.go`):

| Test | setup → call → expected |
|---|---|
| `TestActions_ButtonNameIsTheOperationText` | a fake lister with one action `m.apply.create` → the rendered button has `name="cv-action-m.apply.create"` |
| `TestActions_RunPassesTheTypedOperation` | click it → the fake lister's `RunAction` received exactly the `router.Operation` of the action |
| `TestActions_RunningDisablesOnlyThatAction` | two actions; while the first runs, the first button is disabled and the second is enabled |

Run `gotest`: it does not compile. Commit: `test: typed actions in crudview (red)`.

## Stage 2 — the code

`crudview/crudview.go` (imports: add `webtyp.com/router`):

| Line | Old | New |
|---|---|---|
| 346 | `func (v *CrudView) actionEnabled(op string) bool { … v.actionRunning.Get() != op }` | `func (v *CrudView) actionEnabled(op router.Operation) bool { … v.actionRunning.Get() != op.String() }` |
| 353, 363 | `v.actionEnabled(a.Operation)`, `v.runAction(a.Operation)` | unchanged text (they now pass a `router.Operation`) |
| 369-374 | `op := v.pendingAction.Operation; … if op != "" { v.runAction(op) }` | `op := v.pendingAction.Operation; … if op.Valid() { v.runAction(op) }` |
| 380 | `func (v *CrudView) runAction(op string)` | `func (v *CrudView) runAction(op router.Operation)` |
| 384 | `v.actionRunning.Set(op)` | `v.actionRunning.Set(op.String())` |
| 392 | `v.reportAction(op, err)` / `func (v *CrudView) reportAction(op string, err error)` | `op router.Operation`; `OnAction` field type `func(op router.Operation, err error)` (find the field with `grep -n "OnAction" crudview/*.go`) |
| 419 | `Attr("name", actionNamePrefix+a.Operation)` | `Attr("name", actionNamePrefix+a.Operation.String())` |

Update the doc comment at `crudview.go:46`: the button name is `cv-action-<module>.<entity>.<verb>`.
`app/demo.go:38`: `crud.LoopbackActingAs(nil, d.ActingAs, d.Modules...)` — the demo announces nothing outside the page. `app/browser.go` changes in stage 5.
`gotest` passes. Commit: `refactor!: crudview runs typed actions`.

## Stage 3 — the API explorer sends the REST request

`apiexplorer/try.go` (the "try it" button) sends every route as `POST`/`PUT`/`DELETE` or, for anything else, as a `GET`; `PATCH` (an update) would be sent as a `GET`, and the arguments of a read — which now travel in the query — would be sent as a JSON body.
Replace the request construction (`try.go:131-148`) by a small function and add `buildQuery`:

```go
// requestFor builds the request a route needs: the method is the route's method, the arguments travel in the query for a GET
// and in a JSON body for every other method.
func requestFor(method, path string) *fetch.Request {
	switch method {
	case "POST":
		return fetch.Post(path)
	case "PUT":
		return fetch.Put(path)
	case "PATCH":
		return fetch.Patch(path)
	case "DELETE":
		return fetch.Delete(path)
	}
	return fetch.Get(path)
}

// buildQuery writes the arguments the person filled in as "name=value&…", leaving out the empty ones.
func buildQuery(args []router.ArgRecord, signals []*SignalString) string {
	query := ""
	for i, arg := range args {
		val := signals[i].Get()
		if val == "" {
			continue
		}
		if query != "" {
			query += "&"
		}
		query += router.QueryEscape(arg.Name) + "=" + router.QueryEscape(val)
	}
	return query
}
```

and in the click handler:

```go
		hasBody := r.Method != "GET" && r.Method != "HEAD"
		if !hasBody && r.HasArgs {
			if q := buildQuery(argRecords, argSignals); q != "" {
				finalPath += "?" + q
			}
		}
		req := requestFor(r.Method, finalPath)

		if hasBody {
			req.ContentTypeJSON()
			if !r.HasArgs {
				req.Body([]byte(rawBodySig.Get()))
			} else {
				req.Body([]byte(buildJSON(argRecords, argSignals)))
			}
		}
```

(The existing declaration `var req *fetch.Request` and the old `switch` are deleted.) Known limit, not changed here: for `POST`/`PATCH`/`DELETE` the explorer builds a flat JSON object from the advertised argument fields, while `crud` expects the wrappers `{records}`, `{ids, fields, record}`, `{ids}`;
"try it" on a write route therefore needs the raw body (`!r.HasArgs`). Do not widen this plan to fix it.
`requestFor` is glue over `fetch` (the method strings are proved by the `fetch` plan's tests); `buildQuery` is the logic and has its tests above. `gotest` passes. Commit: `feat: the API explorer sends REST requests`.

## Stage 5 — `layout` implements `view.Layout` and stops importing `appkit` (D50)

Tests first (in `tests/browser_test.go`, `//go:build wasm`; replace the old `appkit.App` fixture): `TestLayout_IsAViewLayout` (`var _ view.Layout = app.New(tools)` compiles), `TestShow_DrawsTheSessionItIsGiven` (`app.New(tools).Show(view.Session{Brand: view.Brand{Name: "acme"}, Caller: <fake Caller that answers `me` with a profile>, Zone: zone, Screens: [one screen], Entry: view.EntryScreens{}})` → the chassis renders with `AppName == "acme"` and the screen in the menu), `TestShow_AnonymousSessionShowsTheEntry` (the fake caller answers `me` with an error and `Entry.Login` set → the login form is shown). Run `gotest`: it does not compile (red). Commit: `test: layout is a view.Layout (red)`.

Code (`app/browser.go`, `app/demo.go`):

1. Delete `func Browser(app appkit.App, tools Tools)` and the imports of `appkit` and `rpc`. Add

```go
// New is the layout an app hands to appkit.Browser: it draws the Session that appkit built.
//
//	func main() { appkit.Browser(config.App, layoutapp.New(devtools.Menu())) }
func New(tools Tools) view.Layout { return layout{tools: tools} }

type layout struct{ tools Tools }

// Show asks the server who the user is and shows the first-admin setup, the access screen or the chassis with the screens the user may open.
func (l layout) Show(s view.Session) {
	st := buildSession(s, l.tools)
	st.platform.Init(nil)
	Render("app", st)
}
```

2. `buildSession(s view.Session, tools Tools) *session`: the body is today's, with `caller := s.Caller`, `entry := s.Entry`, `screens := s.Screens` (appkit already built them with the shared `view.NewRecords()` page), `app.Log` → `s.Log` (guard `s.Log != nil`), `app.Brand` → `s.Brand`. Delete the two `appkit.EntryScreens` / `appkit.Screens` calls and their error branches. `entryScreen(b view.Brand, …)` and `session.brand view.Brand`.
3. `opMe` becomes a typed operation: `var opMe = router.Operation{Module: "authority", Entity: "me", Action: model.Read}` and `caller.Call(opMe, …)`. (`authority.me` is the operation of the `auth` module; if its plan names it otherwise, this variable follows it: one place.)
4. `app/demo.go`: delete the `appkit` import; the broken-zone branch shows `err.Error()` of `time.LoadZone`.
5. `go mod tidy`: `webtyp.com/appkit` leaves `go.mod`. `grep -rn "webtyp.com/appkit" --include=*.go --include=go.mod .` → empty.
6. `README.md`/`docs/ARCHITECTURE.md`: `layoutapp.Browser(config.App, tools)` becomes `appkit.Browser(config.App, layoutapp.New(tools))`; `web/browser/main.go` is generated by the `app` library (its plan follows this change).

`gotest` passes. Commit: `refactor!: layout is a view.Layout; no appkit import (D50)`.

## Stage 6 — documentation

- `README.md` (≤ 150 lines, four ordered sections) and `docs/ARCHITECTURE.md` (four sections): any example with `view.Action{Op: "…"}` or `OnAction` uses the typed form; the button-name rule.
- Do not link `docs/PLAN.md` from permanent docs. `docs/TESTS.md` (if present): one row per new test.

Commit: `docs: typed actions`.

## Delivery check

1. `gotest` (builds, vets, runs; a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `grep -rn "op string\|Operation: \"" --include=*.go crudview app` → empty; `grep -rn "webtyp.com/appkit" --include=*.go --include=go.mod .` → empty.
3. `git diff --stat main` lists only `crudview/crudview.go`, `app/demo.go`, `apiexplorer/try.go`, `apiexplorer/try_test.go`, `crudview/*_test.go`, `tests/menu_test.go`, `app/browser.go`, `go.mod`, `go.sum`, `README.md`, `docs/*`.

| Stage | Files | Commit |
|---|---|---|
| 1 | tests (including `apiexplorer/try_test.go`) | `test: … (red)` |
| 2 | `crudview/crudview.go`, `app/demo.go` | `refactor!: …` |
| 3 | `apiexplorer/try.go`, `apiexplorer/try_test.go` | `feat: …` |
| 5 | `tests/browser_test.go` (red), then `app/browser.go`, `app/demo.go`, `go.mod`, `go.sum` | `test: … (red)` then `refactor!: …` |
| 6 | `README.md`, `docs/*` | `docs: …` |
