---
PLAN: "refactor!: layout uses view's contracts, LocalTools on crud.Module, layout/app owns Browser and Demo, the chassis draws the entry screens"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

# Plan: `layout` is the chassis only the app chooses

## Tools and local dependencies

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```
```bash
go get webtyp.com/view@latest webtyp.com/appkit@latest webtyp.com/crud@latest webtyp.com/rpc@latest webtyp.com/router@latest webtyp.com/auth@latest
```
`view@latest` and `appkit@latest` must already contain `view.Demo` and `appkit.Entry` (plans L1, L2); if `go doc webtyp.com/view Demo`
fails, stop and say so in the PR. Remove every `replace` and the old «TEMPORAL» comment block from `go.mod`; `go mod tidy`.
Any `webtyp.com/router/loopback` use becomes `crud.Loopback` (`grep -rn "router/loopback" .` → empty).

## Why

Decisions D15, D16, D17 of `webtyp/docs/ARCHITECTURE_MASTER_PLAN.md`. Modules stop importing `layout`: the contracts
they implement moved to `view` (plan L1); `appkit` stopped importing `layout` (plan L2). `layout` now owns the whole
page: the app's bootstrap (`Browser`, moved from `appkit`), a module's demo (`Demo`, new) and the entry screens drawn
with the app's brand.

## APIs used

| Symbol | Signature | From |
|---|---|---|
| `view.Notifier`, `view.UsesNotifier`, `view.Duration`, `view.Auto/Persistent/For` | identical to today's `layout` ones | plan L1 |
| `view.EntryScreens`, `view.Entry` | `struct{ Login, Setup Entry; IdleSeconds int }`, `struct{ Subtitle lang.Text; Form dom.Component; Message *dom.SignalString }` | plan L1 |
| `view.Demo` | `struct{ Label, Timezone, ActingAs string; Modules []crud.Module; Migrations []ddl.Step; Screens func(view.BrowserDeps) []view.Screen }` + `Validate() error` | plan L1 |
| `view.BrowserDeps` | `struct{ Caller router.Caller; Zone time.Zone }` | plan L1 |
| `view.NewRecords().Caller(c router.Caller) router.Caller` | one record store per page | `view` |
| `appkit.App`, `appkit.Brand{Name, LogoMark string}` | unchanged | `appkit` |
| `appkit.Screens(app, caller) ([]view.Screen, error)`, `appkit.Entry(app, caller) (view.EntryScreens, error)` | | plan L2 |
| `rpc.Routes(prefix string, public []model.ModuleNaming, modules ...crud.Module) ([]router.RouteInfo, error)` | | `rpc v0.2.17` |
| `crud.LoopbackActingAs(userID string, mods ...crud.Module) router.Caller` | panics on a wiring mistake | `crud v0.0.3` |
| `rpc.NewCaller(prefix string) router.Caller` | | `rpc v0.2.17` |

If a signature differs, stop and name the symbol; never adapt it by guessing.

## Explicit deletions (grep must be empty)

- from `capabilities.go`: `Duration` and methods, `Auto`, `Persistent`, `For`, `Notifier`, `UsesNotifier` →
  `grep -rn "^func Auto\|^type Notifier\|^type UsesNotifier\|^type Duration" --include=*.go .`
- `router.OperationModule` → `grep -rn "OperationModule" --include=*.go .`
- `view.ListArgs` → `grep -rn "view.ListArgs" --include=*.go .`

## Stage 1 — test doubles on `crud.ReadArgs` (was A3)

Replace `view.ListArgs` with `crud.ReadArgs` (import `webtyp.com/crud`) in: `crudview/actions_test.go`
(`(*countingLister).List`), `crudview/bulk_test.go` (`(*noSaveBackend).List`), `crudview/consumer_test.go`
(`(*listSaveDeleteBackend).List`), `crudview/paging_test.go` (`pagedLister.lastArgs`, `(*pagedLister).List`),
`crudview/readonly_test.go` (`(*listOnly).List`), `crudview/record_identity_test.go` (`(*memoryBackend).List`,
`(*laterBackend).List`), `crudview/shared_records_test.go` (`(*sharedBackend).List`), `tests/menu_test.go`
(`(*localSavingLister).List`).

## Stage 2 — contracts from `view`

Delete the six symbols above from `capabilities.go`. Every use in this repo becomes `view.<Name>`: `factory.go`
(`UseNotifier(n view.Notifier)` ×2, and in `describedModule`/`describedGroup` of `capabilities.go`),
`platformd/platformd.go:968` (`Notify(t msgtype.Type, msg string, d view.Duration)`), `platformd/capabilities.go`,
`crudview/crudview.go`, and the tests `platformd/platformd_test.go`, `crudview/crudview_test.go`. Find every site with
`grep -rn "Notifier\|Duration\|Auto()\|Persistent()\|For(" --include=*.go .` and fix each. `Badged`, `Described`, `Step`,
`AlertedModule`, `ModuleAlert`, `GroupModule`, `UIModule` stay here.

## Stage 3 — `LocalTools` on `crud.Module` (`app/menu.go`)

```go
// LocalTools builds the tools of a page with no server — a module's demo.
//
//	tools, err := app.LocalTools([]crud.Module{mod}, ddl.Step{Name: "shop", Run: migrate.Migrate})
func LocalTools(modules []crud.Module, steps ...ddl.Step) (Tools, error) {
	var t Tools
	if len(modules) > 0 {
		routes, err := rpc.Routes(rpc.DefaultPrefix, nil, modules...)
		if err != nil {
			return Tools{}, err
		}
		t.API = apiexplorer.FromTable(router.Table(routes, nil))
	}
	// steps part unchanged
}
```
`tests/menu_test.go:387-396`: `fakeOpMod` implements `crud.Module` (`ModelName() string`, `MountOperations(crud.Registry)`
registering one `reg.Read(...)`), call `app.LocalTools([]crud.Module{fakeOpMod{}}, fakeStep)`.

## Stage 4 — `app/browser.go` (moved from `appkit`, `//go:build wasm`)

Copy `browser.go` from `appkit` at its last published tag (`git -C ../appkit show v0.2.27:browser.go`) into
`app/browser.go`, package `app`, then change exactly:

1. `func Browser(app appkit.App, tools Tools)` — the doc example becomes `func main() { layoutapp.Browser(config.App, devtools.Menu()) }`.
2. `appmenu.BuildMenu` / `appmenu.Tools` → `BuildMenu` / `Tools` (same package now).
3. Remove the entry loop (old lines 86-95) and the local `Screens` function (old lines 178-196): use
   `entry, err := appkit.Entry(app, caller)` and `screens, errScreens := appkit.Screens(app, view.NewRecords().Caller(caller))`;
   an `err` from `appkit.Entry` is logged with `app.Log` and treated as «no entry screens».
4. The session keeps `entry view.EntryScreens`; the two `Show` blocks draw the entry screens with the brand:

```go
func entryScreen(b appkit.Brand, title lang.Text, e view.Entry) *Element {
	if e.Form == nil {
		return Div()
	}
	l := &login.Login{Title: title, Subtitle: e.Subtitle, Form: e.Form, Message: e.Message, LogoMark: b.LogoMark}
	return l.Render()
}
// login view:  entryScreen(s.brand, lang.Text(s.brand.Name), s.entry.Login)
// setup view:  entryScreen(s.brand, lang.Text(s.brand.Name), s.entry.Setup)
```

`emptyArgs`, `profileIdentity`, `reload`, `logout` move unchanged. New imports: `webtyp.com/appkit`, `webtyp.com/auth`,
`webtyp.com/components/usermenu`, `webtyp.com/fetch`, `webtyp.com/layout/login`, `webtyp.com/layout/platformd`.

## Stage 5 — `app/demo.go` (new, `//go:build wasm`)

```go
// Demo draws a module's demo with the chassis: the module's operations through a loopback signed in as
// Demo.ActingAs, its screens, and the developer tools from its operations and migrations. It never returns.
// The webtyp daemon generates the only caller:  func main() { layoutapp.Demo(seed.Demo) }
func Demo(build func() (view.Demo, error)) {
	d, err := build()
	if err == nil {
		err = d.Validate()
	}
	if err != nil {
		Append("body", P().Text(err.Error())) // a broken demo says why on the page, never a blank screen
		select {}
	}
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err) // never fails in the browser
	}
	zone, err := time.LoadZone(d.Timezone)
	if err != nil {
		Append("body", P().Text(appkit.ErrNoTimezone.Error()))
		select {}
	}
	caller := view.NewRecords().Caller(crud.LoopbackActingAs(d.ActingAs, d.Modules...))
	screens := d.Screens(view.BrowserDeps{Caller: caller, Zone: zone})
	tools, err := LocalTools(d.Modules, d.Migrations...)
	if err == nil {
		var menu Menu
		menu, err = BuildMenu(screens, ids, func(model.Resource) bool { return true }, tools)
		if err == nil {
			p := &platformd.Platform{AppName: d.Label + " — demo", Modules: menu.Modules, CanView: menu.CanView,
				User: demoUser{id: d.ActingAs}}
			p.Init(nil)
			Append("body", p)
			select {}
		}
	}
	Append("body", P().Text(err.Error()))
	select {}
}

type demoUser struct{ id string }

func (u demoUser) UserName() string    { return u.id }
func (u demoUser) UserAvatar() string  { return "" }
func (u demoUser) UserRoles() []string { return nil }
```
Use the same `dom`/`html` dot-import style as `app/browser.go`. If `Append` is named differently in `dom v0.13.29`, use
the one `appkit/browser.go` used (`Render("app", …)` targets `#app`; the demo page has only `body`).

## Stage 6 — tests (`tests/`, rows in `docs/TESTS.md`, `-tags wasm` run by `gotest` in headless Chrome)

| Test | File | Setup → call → expected |
|---|---|---|
| `TestLocalTools_CrudModule` | `tests/menu_test.go` | `LocalTools([]crud.Module{fakeOpMod{}}, fakeStep)` → `t.API != nil`, `t.Database != nil` |
| `TestBrowser_ShapeOfApp` | `tests/browser_test.go` | moved from `appkit/tests/browser_test.go` (`TestApp_BrowserShape`) |
| `TestDemo_InvalidShowsReason` | `tests/demo_test.go` | `Demo(func() (view.Demo, error) { return view.Demo{}, nil })` in a goroutine; the body text contains the text of `view.ErrDemoNoLabel` |

## Stage 7 — translations and documentation

`langc sync .`, fill every empty value (zero reuse warnings). `README.md`: «Run the app's browser →
`layoutapp.Browser(config.App, devtools.Menu())` (generated)», «A module's demo → the module returns `view.Demo` from
`seed.Demo`; the daemon calls `layoutapp.Demo`». `docs/ARCHITECTURE.md`: `layout` is imported only by an app's generated
browser main, by `webtyp/devtools` and by the daemon's generated demo main — never by a module.

## Delivery check

```bash
gotest
langc sync .
grep -rn "router/loopback\|OperationModule\|view.ListArgs" --include=*.go .   # empty
```

A PR that does not compile, or whose tests could not run, is a failed delivery.

| Stage | Files |
|---|---|
| 1 | `crudview/*_test.go`, `tests/menu_test.go`, `go.mod` |
| 2 | `capabilities.go`, `factory.go`, `platformd/*.go`, `crudview/crudview.go`, their tests |
| 3 | `app/menu.go`, `tests/menu_test.go` |
| 4 | `app/browser.go` |
| 5 | `app/demo.go` |
| 6 | `tests/browser_test.go`, `tests/demo_test.go`, `docs/TESTS.md` |
| 7 | `lang.json`, `README.md`, `docs/ARCHITECTURE.md` |
