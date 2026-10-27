---
PLAN: "feat: menu draws Component screens, ships the API explorer by itself and is closed by default"
TAG: v0.3.77
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 12902786279715637843
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp.com/layout`: the menu draws every kind of screen, with its permission, closed by default

## Why

`layout/app.BuildMenu` turns the screens modules declare (`view.Screen`) into the chassis menu. Three gaps:

1. **Screens that are not lists.** `webtyp/view` v0.6.46 adds `Screen.Component` (a wizard, a chat). The menu must
   draw it with its permission and its «?» help like any list.
2. **The API explorer is written by every app.** Every app appends `apiexplorer.New(...)` by hand next to the
   menu, and `CanView` then hides it from everyone: `BuildMenu` only knows the permissions of the screens it built.
   The menu now adds the explorer itself, guarded by `router.ResourceIntrospection` (the same permission that
   guards `/_routes`, `webtyp/router` v0.4.9).
3. **A chassis without a visibility rule shows everything.** `platformd.Platform.CanView == nil` shows every
   module (`platformd.go:151,320`). Closed by default: `nil` shows nothing.

Also: a behaviour test that pins «edit a valid field → the list row shows the new value», with a backend that
answers later (like the network), in `crudview`.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/router@v0.4.9
go get webtyp.com/view@v0.6.46
```

`router@v0.4.9` is the first tag with `ResourceIntrospection`; `view@v0.6.46` the first with `Screen.Component`.
If either `go get` fails or the symbol is missing, stop and say so in the PR.

## Ecosystem search

- `view.Screen.Component dom.Component` and its `Validate` rules (view v0.6.46) → reuse.
- `router.ResourceIntrospection` (router v0.4.9) → reuse; deletes this repo's literal `"api_explorer"`.
- `layout.NewUIModule`, `layout.Describe` (this repo, `factory.go`, `capabilities.go`) → reuse to wrap a Component.
- `apiexplorer.New` (this repo) → reuse inside `BuildMenu`.

## Design gate

1. **Prior art.** Component screens: Django admin custom views, ActiveAdmin `register_page`, react-admin
   `<CustomRoutes>`, Filament Pages — one menu, one permission check for lists and custom pages. Explorer: Laravel
   Telescope, Spring Boot Actuator, Django `admindocs` register themselves with their own permission. Closed by
   default: Spring Security denies unless granted; Django `has_perm` is false without a grant.
2. **Novice-name test.** No new exported name: `BuildMenu` keeps its signature; `CanView`'s meaning changes for
   `nil` only.
3. **Complexity ledger.** Concepts +0; lines in every app −8 (the explorer block); ways to add the explorer −1;
   ways to declare a non-list screen −1 (no `Browser()` beside `Screens()`).
4. **Where it belongs.** Drawing screens is `layout/app`'s job; the visibility default is `platformd`'s.
5. **What it deletes.** The literal `"api_explorer"` in `apiexplorer/explorer.go:94`; the «nil = show all» branch of
   `isViewable`.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `view.Screen` | `struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Component dom.Component; Lookups []Lookup; Help lang.Text; Children []Screen }` | view v0.6.46 | `screen.go` |
| `(view.Screen).Validate` | `func (s Screen) Validate() error` — exactly one of `Presenter`/`Component` on a leaf; a Component leaf has `Help` | view v0.6.46 | `screen.go` |
| `router.ResourceIntrospection` | `const ResourceIntrospection model.Resource = "api_explorer"` | router v0.4.9 | `introspection.go` |
| `layout.NewUIModule` | `func NewUIModule(id, label string, icon svg.Icon, view dom.Component, onActivate ...func()) UIModule` | this repo | `factory.go` |
| `layout.Describe` | `func Describe(m UIModule, text lang.Text, steps ...Step) UIModule` | this repo | `capabilities.go` |
| `apiexplorer.New` | `func New(cfg Config) (*Explorer, error)`; `*Explorer` satisfies `layout.UIModule` | this repo | `apiexplorer/explorer.go` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — the explorer's identity (`apiexplorer/explorer.go`)

Line 94 becomes `func (e *Explorer) ModelName() string { return string(router.ResourceIntrospection) }`.
Check: `grep -rn '"api_explorer"' --include=*.go .` → empty.

## Stage 2 — `layout/app/menu.go`

1. Doc comment of `BuildMenu`, replace the example with
   `func(r model.Resource) bool { return profile.Can(r, model.Read) }` (the browser profile of `webtyp/auth`).
   Add one sentence: «The API explorer is always added last, visible to whoever may read
   `router.ResourceIntrospection`; the app never adds it.»
2. Add the help text constant (English key, translated in `lang.json`):

```go
const helpAPIExplorer lang.Text = "Every operation of this app: who may call it, and a form to try it."
```

3. In `buildScreen`, right after the group branch and before `crudview.New`, add the Component leaf:

```go
	if s.Component != nil {
		*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
		m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), s.Component)
		return layout.Describe(m, s.Help), nil
	}
```

4. In `BuildMenu`, after the loop that builds `mods` and before `canView`:

```go
	explorer, err := apiexplorer.New(apiexplorer.Config{})
	if err != nil {
		return Menu{}, err
	}
	mods = append(mods, layout.Describe(explorer, helpAPIExplorer))
	perms = append(perms, screenPerm{id: explorer.ModelName(), permission: router.ResourceIntrospection})
```

   Imports: `webtyp.com/layout/apiexplorer`, `webtyp.com/router`. If importing `apiexplorer` from `layout/app`
   creates an import cycle, stop and say so in the PR (it must not: `apiexplorer` does not import `layout/app`).

## Stage 3 — closed by default (`platformd/platformd.go`)

- Field comment (line ~151): `// CanView decides which modules the shell presents. nil = none (closed by default):
  pass layout/app's Menu.CanView, or your own rule.`
- `isViewable`: replace `if p.CanView == nil { return true }` with `if p.CanView == nil { return false }`.
- In `platformd/platformd_test.go`, every `&Platform{…}` literal that has `Modules` and no `CanView` gets
  `CanView: func(string) bool { return true },` so those tests keep testing what they tested.
- Also every other `Platform{` literal in this repo (`grep -rn "Platform{" --include=*.go .`), including demos under
  `web/`, gets an explicit `CanView`.

## Stage 4 — tests

**`tests/menu_test.go`** (extend):

| Test | Setup | Expected |
|---|---|---|
| `TestBuildMenu_ComponentLeaf` | one screen `{ID:"booking", Label:"Booking", Icon:"laptop", Permission:"reservation", Help:"Book.", Component: &fakeComp{}}` (`type fakeComp struct{ dom.Element }`, embedded by value) | no error; `Modules[0].ModelName() == "booking"`; `Modules[0].View()` is the same component; `CanView("booking")` true only when `allows` returns true for `"reservation"`; the module satisfies `layout.Described` with text `"Book."` |
| `TestBuildMenu_AddsAPIExplorerLast` | any valid screens; `allows` true only for `router.ResourceIntrospection` | last module's `ModelName() == string(router.ResourceIntrospection)`; `CanView(string(router.ResourceIntrospection))` true; with `allows` always false → false |
| existing tests | — | update every `len(m.Modules)` expectation by +1 (the explorer) |

**`platformd/platformd_test.go`** — add `TestPlatform_NilCanViewShowsNothing`: a `Platform` with two modules and no
`CanView`; `Render().String()` contains neither `data-id='mod1'` nor `data-id='mod2'`.

**`crudview/record_identity_test.go`** — add the D2 guard (an async backend answers later, like the network):

```go
type laterBackend struct {
	memoryBackend
	pending []func()
}

func (b *laterBackend) List(done func([]model.Model, error)) {
	b.pending = append(b.pending, func() { b.memoryBackend.List(done) })
}

func (b *laterBackend) Save(recs []model.Model, done func(error)) {
	cp := make([]model.Model, len(recs))
	for i, r := range recs {
		d := *(r.(*Device))
		cp[i] = &d
	}
	b.pending = append(b.pending, func() { b.memoryBackend.Save(cp, done) })
}

func (b *laterBackend) answer() {
	for len(b.pending) > 0 {
		f := b.pending[0]
		b.pending = b.pending[1:]
		f()
	}
}

// What is saved shows at once in its list row (skill ui-consistency), also when the
// backend answers later.
func TestRecordIdentity_EditedValueShowsInItsRow(t *testing.T) {
	b := &laterBackend{memoryBackend: memoryBackend{rows: []Device{{Id: "a1", Name: "Alpha", Ip: "10.0.0.1"}}}}
	v, err := New(Config{ParentID: "edit", Presenter: view.New(b, &Device{}), IDs: testIDs})
	if err != nil {
		t.Fatal(err)
	}
	v.Init(&fakeCtx{})
	v.Activate()
	b.answer()
	v.selectAction(view.Item{ID: "a1"})
	v.form.SetValues("name", "Alpha Edited")
	v.autoSaveAction()
	b.answer()
	items := v.list.Items()
	if len(items) != 1 || items[0].Label != "Alpha Edited" {
		t.Fatalf("row after edit = %+v, want one row labelled \"Alpha Edited\"", items)
	}
}
```

(This test passes today; it guards the behaviour while the list moves to a shared record store in a later plan.)

## Stage 5 — docs

- `docs/ARCHITECTURE.md`, section «Chassis contracts and menu assembly»: add that a leaf is a list (`Presenter`) or a
  component (`Component` + `Help`), that `BuildMenu` adds the API explorer last under
  `router.ResourceIntrospection`, and that `Platform.CanView == nil` shows nothing.
- `README.md`: wherever `CanView` or the explorer wiring is shown, show `Menu.CanView` and drop any hand-added
  explorer.

## Stage 6 — translations

`langc sync .`, then fill every empty value it reports in `lang.json` — including the 7 empty columns of
`"API Explorer"` and the new `helpAPIExplorer` key — for every language in `"languages"`. Zero untranslated keys and
zero reuse warnings.

## Delivery

- `gotest` green (it also runs the `-tags browser` tests in headless Chrome). A PR that does not compile, or whose
  tests could not run, is a failed delivery. Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`,
  `GOFLAGS=-mod=mod`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `apiexplorer/explorer.go`, `go.mod`, `go.sum` | identity from `router.ResourceIntrospection` |
| 2 | `app/menu.go` | Component leaf; explorer added last with its permission |
| 3 | `platformd/platformd.go`, `platformd/platformd_test.go`, every `Platform{` literal | `nil` CanView = nothing |
| 4 | `tests/menu_test.go`, `platformd/platformd_test.go`, `crudview/record_identity_test.go` | behaviour tests |
| 5 | `docs/ARCHITECTURE.md`, `README.md` | document |
| 6 | `lang.json` | translate |
