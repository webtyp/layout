---
PLAN: "feat: promote UIModule contracts to root layout package and add layout/app.BuildMenu"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 10323412434008975531
PR: https://github.com/webtyp/layout/pull/48
REVIEWED_COMMIT: 27256583f66ffd2a9dc8e815d612dd514e4abf81
---
> Read AGENTS.md at the repo root before any change.

# PLAN — `webtyp/layout`: root module contracts (`UIModule`, `GroupModule`, `Described`, `Notifier`) and `layout/app.BuildMenu`

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/view@v0.6.45 webtyp.com/lang@v0.1.10 webtyp.com/router@v0.4.8 webtyp.com/unixid@v0.3.1
```

## Ecosystem search

- `webtyp.com/layout` (`layout.go`) currently declares only `Module` (`ModelName() string`), while `UIModule`, `GroupModule`, `Described`, `Step`, `Badged`, `AlertedModule`, `ModuleAlert`, `Notifier`, `UsesNotifier`, `Duration`, `NewUIModule`, `NewGroupUIModule` and `Describe` live inside `webtyp.com/layout/platformd`. Because of that, `crudview` (`crudview/crudview.go:16`) and `apiexplorer` (`apiexplorer/explorer.go:15`) had to import `webtyp.com/layout/platformd` just to satisfy `UsesNotifier` and `UIModule`.
- `webtyp.com/view` (`v0.6.45`) declares `view.Screen`, `view.Lookup`, and `Screen.Validate()`.
- **Decision**: **extend** `webtyp.com/layout` by promoting `UIModule`, `GroupModule`, `Described`, `Step`, `Badged`, `AlertedModule`, `ModuleAlert`, `Notifier`, `UsesNotifier`, `Duration`, `Auto`, `Persistent`, `For`, `NewUIModule`, `NewGroupUIModule`, and `Describe` to the root `layout` package (keeping aliases in `layout/platformd`), decouple `crudview` and `apiexplorer` from `platformd`, and add subpackage `webtyp.com/layout/app` (`app/menu.go`) with `Menu` and `BuildMenu` against `layout.UIModule` (never importing `platformd`).

## Design gate

1. **Prior art**:
   - Flutter (`Scaffold` vs `RouterConfig`/`Widget` contracts), React Admin (`<Admin layout={MyLayout}>` where resources/screens are layout-agnostic), Django Admin (`AdminSite` vs swappable templates). In all three, the screen/menu contract is separate from the concrete shell layout so the application can swap or customize the shell (`platformd` or a custom layout) without rewriting screen assembly.
2. **Novice-name test**:
   - `layout.UIModule`, `layout.GroupModule`, `layout.Described`, `layout.NewUIModule`, `layout.NewGroupUIModule`, `layout.Describe` in `webtyp.com/layout`.
   - `app.Menu` and `app.BuildMenu(screens, ids, allows)` in `webtyp.com/layout/app`.
3. **Complexity ledger**:
   ```
   Concepts the developer must learn   +1 (app.BuildMenu) / -1 (per-module manual crudview + onFirstActivation + fillOptions boilerplate)
   Files they must touch to do X       +1 (config/browser.go) / -14 (ui/browser.go in each domain module)
   Lines at the call site              +3 / -45 per module
   Ways to do the same thing           +1 / -1 (replaces manual UIModule construction in modules)
   ```
4. **Where does it belong**:
   - Root `webtyp.com/layout` owns the chassis-neutral contracts (`UIModule`, `GroupModule`, `Described`, `Notifier`, etc.).
   - `webtyp.com/layout/platformd` is one concrete chassis implementation and re-exports type aliases (`type UIModule = layout.UIModule`, etc.) so all existing consumers compile unchanged.
   - `webtyp.com/layout/app` joins `view.Screen` with `layout.UIModule` and `layout/crudview` without importing `layout/platformd`, and keeps `webtyp/appkit`'s `go.mod` free of UI dependencies.
5. **What does this change delete?**:
   - Removes the direct `"webtyp.com/layout/platformd"` import from `crudview/crudview.go`, `crudview/crudview_test.go`, and `apiexplorer/explorer.go`, and removes the duplicate struct definitions from `platformd/capabilities.go` and `platformd/factory.go` (replaced by aliases to `layout`).

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `view.Screen` | `type Screen struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Lookups []Lookup; Help lang.Text; Children []Screen }` | `v0.6.45` | `webtyp.com/view/screen.go` |
| `(view.Screen).Validate` | `func (s Screen) Validate() error` | `v0.6.45` | `webtyp.com/view/screen.go` |
| `view.Lookup` | `type Lookup struct { Field string; Source Presenter }` | `v0.6.45` | `webtyp.com/view/screen.go` |
| `view.Presenter` | `type Presenter interface { Record() model.Model; NewRecord() model.Model; Items() []Item; SearchPlaceholder() lang.Text; Description() lang.Text; Reload(done func(err error)) }` | `v0.6.45` | `webtyp.com/view/view.go` |
| `crudview.New` | `func New(cfg Config) (*CrudView, error)` | local | `webtyp.com/layout/crudview/crud.go` |
| `(crudview.CrudView).Activate` | `func (v *CrudView) Activate()` | local | `webtyp.com/layout/crudview/crudview.go` |
| `(crudview.CrudView).SetOptions` | `func (v *CrudView) SetOptions(fieldName string, opts ...fmt.KeyValue)` | local | `webtyp.com/layout/crudview/crudview.go` |
| `lang.Translate` | `func Translate(key Text, args ...any) TranslatedText` | `v0.1.10` | `webtyp.com/lang` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

---

## Stage 1 — Promote module contracts to root `package layout` and alias in `platformd`

1. Update `layout.go` at the repo root to define `Module` and `UIModule`:

```go
// Package layout provides pre-built, consistent, reusable UI layout skeletons and
// chassis contracts for webtyp modules. A layout defines WHERE elements are placed
// (structure, grid, spacing) but NOT what those elements contain — that is the
// consumer's responsibility.
package layout

import (
	"webtyp.com/dom"
	"webtyp.com/svg"
)

// Module is the identity contract for the ecosystem.
// Any component that can be identified by a slug (for routing or keys)
// implements this interface.
type Module interface {
	ModelName() string
}

// UIModule is a module that provides its UI to any layout chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	Module               // identity: ModelName() → used as ID
	Label() string       // text in the nav rail
	Icon() svg.Icon      // chassis renders via the sprite
	View() dom.Component // module content
	Activate()           // called by chassis when the module is activated (on-demand loading)
}
```

2. Create `capabilities.go` at the repo root (`package layout`):

```go
package layout

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/fmt"
	"webtyp.com/lang"
	"webtyp.com/msgtype"
)

// Duration is how long a notification stays before dismissing itself.
type Duration struct {
	millis func(msg string) int // nil = persistent
	auto   bool                 // built by Auto(): the caller left the length to the chassis
}

// IsAuto reports whether the duration was created with Auto().
func (d Duration) IsAuto() bool { return d.auto }

// HasTimer reports whether the duration auto-dismisses after a positive number of milliseconds.
func (d Duration) HasTimer() bool { return d.millis != nil }

// Millis computes the duration in milliseconds for msg, or 0 if persistent.
func (d Duration) Millis(msg string) int {
	if d.millis == nil {
		return 0
	}
	return d.millis(msg)
}

// Auto sizes the duration to the message (~350ms per word plus 2.5s, floored at 4s and capped at 10s).
func Auto() Duration {
	return Duration{auto: true, millis: func(msg string) int {
		words := fmt.Count(msg, " ") + 1
		ms := 2500 + words*350
		if ms < 4000 {
			return 4000
		}
		if ms > 10000 {
			return 10000
		}
		return ms
	}}
}

// Persistent keeps the notification until the user dismisses it.
func Persistent() Duration { return Duration{} }

// For pins an exact duration in milliseconds.
func For(ms int) Duration {
	return Duration{millis: func(string) int { return ms }}
}

// Badged is an optional capability of a UIModule: a count the nav rail draws
// over the module's entry.
type Badged interface {
	Badge() *countbadge.CountBadge
}

// Described is an optional capability of a UIModule: what it is, for someone who has never used it, plus
// optional first-steps guidance.
type Described interface {
	Description() lang.Text
	Steps() []Step
}

// Step is one setup prerequisite shown with a checkmark once Done is true.
type Step struct {
	Label lang.Text
	Done  bool
}

// AlertedModule is an optional capability of a UIModule: it has alerts a person has not seen yet.
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}

// ModuleAlert is one row an alert strip draws.
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}

// Notifier raises a notification in the chassis.
type Notifier interface {
	Notify(t msgtype.Type, msg string, d Duration)
}

// UsesNotifier is an optional capability of a UIModule that raises notifications.
type UsesNotifier interface {
	UseNotifier(n Notifier)
}

// GroupModule is an optional capability of a UIModule indicating that it groups
// nested child modules.
type GroupModule interface {
	UIModule
	Children() []UIModule
}

// Describe returns m with the Described capability, preserving UsesNotifier, Badged, and GroupModule.
func Describe(m UIModule, text lang.Text, steps ...Step) UIModule {
	if gm, ok := m.(GroupModule); ok {
		return &describedGroup{
			GroupModule: gm,
			text:        text,
			steps:       steps,
		}
	}
	return &describedModule{
		UIModule: m,
		text:     text,
		steps:    steps,
	}
}

type describedModule struct {
	UIModule
	text  lang.Text
	steps []Step
}

func (m *describedModule) Description() lang.Text { return m.text }
func (m *describedModule) Steps() []Step          { return m.steps }

func (m *describedModule) UseNotifier(n Notifier) {
	if u, ok := m.UIModule.(UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *describedModule) Badge() *countbadge.CountBadge {
	if b, ok := m.UIModule.(Badged); ok {
		return b.Badge()
	}
	return nil
}

type describedGroup struct {
	GroupModule
	text  lang.Text
	steps []Step
}

func (m *describedGroup) Description() lang.Text { return m.text }
func (m *describedGroup) Steps() []Step          { return m.steps }

func (m *describedGroup) UseNotifier(n Notifier) {
	if u, ok := m.GroupModule.(UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *describedGroup) Badge() *countbadge.CountBadge {
	if b, ok := m.GroupModule.(Badged); ok {
		return b.Badge()
	}
	return nil
}
```

3. Create `factory.go` at the repo root (`package layout`):

```go
package layout

import (
	"webtyp.com/dom"
	"webtyp.com/svg"
)

// NewUIModule returns an implementation of UIModule with an optional onActivate hook.
func NewUIModule(id, label string, icon svg.Icon, view dom.Component, onActivate ...func()) UIModule {
	var act func()
	if len(onActivate) > 0 {
		act = onActivate[0]
	}
	return &uiModule{
		id:         id,
		label:      label,
		icon:       icon,
		view:       view,
		onActivate: act,
	}
}

type uiModule struct {
	id         string
	label      string
	icon       svg.Icon
	view       dom.Component
	onActivate func()
}

func (m *uiModule) ModelName() string   { return m.id }
func (m *uiModule) Label() string       { return m.label }
func (m *uiModule) Icon() svg.Icon      { return m.icon }
func (m *uiModule) View() dom.Component { return m.view }
func (m *uiModule) Activate() {
	if m.onActivate != nil {
		m.onActivate()
	} else if act, ok := m.view.(interface{ Activate() }); ok {
		act.Activate()
	}
}

func (m *uiModule) UseNotifier(n Notifier) {
	if u, ok := m.view.(UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

var _ UIModule = (*uiModule)(nil)
var _ UsesNotifier = (*uiModule)(nil)

// NewGroupUIModule returns an implementation of GroupModule that wraps child UIModules.
func NewGroupUIModule(id, label string, icon svg.Icon, children []UIModule) UIModule {
	return &uiGroupModule{
		id:       id,
		label:    label,
		icon:     icon,
		children: children,
	}
}

type uiGroupModule struct {
	id       string
	label    string
	icon     svg.Icon
	children []UIModule
}

func (m *uiGroupModule) ModelName() string    { return m.id }
func (m *uiGroupModule) Label() string        { return m.label }
func (m *uiGroupModule) Icon() svg.Icon       { return m.icon }
func (m *uiGroupModule) View() dom.Component  { return nil }
func (m *uiGroupModule) Children() []UIModule { return m.children }
func (m *uiGroupModule) Activate() {
	if len(m.children) > 0 {
		m.children[0].Activate()
	}
}

func (m *uiGroupModule) UseNotifier(n Notifier) {
	for _, ch := range m.children {
		if u, ok := ch.(UsesNotifier); ok {
			u.UseNotifier(n)
		}
	}
}

var _ GroupModule = (*uiGroupModule)(nil)
var _ UsesNotifier = (*uiGroupModule)(nil)
```

4. Update `platformd/capabilities.go`, `platformd/factory.go`, `platformd/platformd.go`, and `platformd/platformd_test.go` to alias the root `layout` types and functions instead of duplicating them:
   - In `platformd/capabilities.go`:
     ```go
     package platformd

     import "webtyp.com/layout"

     type (
     	Duration      = layout.Duration
     	Badged        = layout.Badged
     	Described     = layout.Described
     	Step          = layout.Step
     	AlertedModule = layout.AlertedModule
     	ModuleAlert   = layout.ModuleAlert
     	Notifier      = layout.Notifier
     	UsesNotifier  = layout.UsesNotifier
     	GroupModule   = layout.GroupModule
     )

     var (
     	Auto       = layout.Auto
     	Persistent = layout.Persistent
     	For        = layout.For
     	Describe   = layout.Describe
     )

     var _ Notifier = (*Platform)(nil)
     ```
   - In `platformd/factory.go`:
     ```go
     package platformd

     import "webtyp.com/layout"

     var (
     	NewUIModule      = layout.NewUIModule
     	NewGroupUIModule = layout.NewGroupUIModule
     )
     ```
   - In `platformd/platformd.go`:
     - Replace `type UIModule interface { ... }` (lines 96–102) with `type UIModule = layout.UIModule`.
     - Remove the duplicate `Duration`, `Auto`, `Persistent`, `For` definitions (lines 242–278, now aliased in `platformd/capabilities.go`).
     - In `(p *Platform) Notify` (lines 990–1014), use `d.IsAuto()`, `d.HasTimer()`, and `d.Millis(msg)` instead of unexported fields `d.auto` and `d.millis`.
   - In `platformd/platformd_test.go`:
     - On line 554 (`autoMillis`), replace `return Auto().millis(msg)` with `return Auto().Millis(msg)`.

5. Decouple `crudview` and `apiexplorer` from `platformd`:
   - In `crudview/crudview.go`: replace import `"webtyp.com/layout/platformd"` with `"webtyp.com/layout"`, and replace `platformd.Notifier`, `platformd.UsesNotifier`, `platformd.Auto()` with `layout.Notifier`, `layout.UsesNotifier`, `layout.Auto()`.
   - In `crudview/crudview_test.go`: replace import `"webtyp.com/layout/platformd"` with `"webtyp.com/layout"`, and `platformd.Duration` with `layout.Duration`.
   - In `apiexplorer/explorer.go`: replace import `"webtyp.com/layout/platformd"` with `"webtyp.com/layout"`, and `var _ platformd.UIModule = (*Explorer)(nil)` with `var _ layout.UIModule = (*Explorer)(nil)`.
   - Verification: `grep -rn "layout/platformd" crudview/ apiexplorer/` must return empty.

---

## Stage 2 — `app/menu.go`: `BuildMenu` against `layout.UIModule` (never importing `platformd`)

Create `app/menu.go` (`package app`, **no build tag** so `gopls` and native tests analyze it directly). Reference for leaf construction and lookup loading: `_temp/device_manager/ui/browser.go` (`onFirstActivation`, `fillOptions`).

```go
package app

import (
	"webtyp.com/fmt"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/svg"
	"webtyp.com/view"
)

// Menu is what layout/app hands to any layout chassis (such as platformd or a custom layout):
// the modules and the visibility rule.
type Menu struct {
	Modules []layout.UIModule
	CanView func(id string) bool
}

type screenPerm struct {
	id         string
	permission model.Resource
}

// BuildMenu validates every screen and builds its layout.UIModule tree. allows answers whether the
// signed-in user holds a permission (in an app: func(r model.Resource) bool { return profile.Allows(string(r)) }).
func BuildMenu(screens []view.Screen, ids model.IDGenerator, allows func(model.Resource) bool) (Menu, error) {
	for _, s := range screens {
		if err := s.Validate(); err != nil {
			return Menu{}, err
		}
	}

	var perms []screenPerm
	mods := make([]layout.UIModule, 0, len(screens))
	for _, s := range screens {
		m, err := buildScreen(s, ids, &perms)
		if err != nil {
			return Menu{}, err
		}
		mods = append(mods, m)
	}

	canView := func(id string) bool {
		if allows == nil {
			return false
		}
		for _, p := range perms {
			if p.id == id {
				return allows(p.permission)
			}
		}
		return false
	}

	return Menu{
		Modules: mods,
		CanView: canView,
	}, nil
}

func buildScreen(s view.Screen, ids model.IDGenerator, perms *[]screenPerm) (layout.UIModule, error) {
	if len(s.Children) > 0 {
		children := make([]layout.UIModule, 0, len(s.Children))
		for _, c := range s.Children {
			ch, err := buildScreen(c, ids, perms)
			if err != nil {
				return nil, err
			}
			children = append(children, ch)
		}
		grp := layout.NewGroupUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), children)
		return layout.Describe(grp, s.Help), nil
	}

	v, err := crudview.New(crudview.Config{
		ParentID:  s.ID,
		Presenter: s.Presenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}

	lookups := s.Lookups
	activate := onFirstActivation(v, func() {
		for _, l := range lookups {
			fillOptions(v, l.Field, l.Source)
		}
	})

	*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
	m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), v, activate)
	if desc := s.Presenter.Description(); desc != "" {
		m = layout.Describe(m, desc)
	}
	return m, nil
}

func onFirstActivation(v *crudview.CrudView, load func()) func() {
	loaded := false
	return func() {
		if !loaded {
			loaded = true
			load()
		}
		v.Activate()
	}
}

func fillOptions(v *crudview.CrudView, field string, source view.Presenter) {
	source.Reload(func(err error) {
		if err != nil {
			return
		}
		items := source.Items()
		opts := make([]fmt.KeyValue, 0, len(items))
		for _, it := range items {
			opts = append(opts, fmt.KeyValue{Key: it.ID, Value: it.Label})
		}
		v.SetOptions(field, opts...)
	})
}
```

Verification: `grep -rn "layout/platformd" app/` must return empty.

---

## Stage 3 — Menu tests (`tests/menu_test.go`)

Create `tests/menu_test.go` (`package tests`, no build tag so it runs in both native and wasm passes). Use `view.New(&conformance.FakeLister{}, &conformance.MockRecord{})` (`webtyp.com/view/conformance`) as presenters and a simple `model.IDGenerator` (or `unixid`) as `ids`.

| Test | Setup | Expect |
|---|---|---|
| `TestBuildMenu_InvalidScreenIsAnError` | a leaf without `Permission` | error containing `needs a Permission` |
| `TestBuildMenu_LeafVisibleByItsPermission` | leaf `x` with `Permission: "service"`; `allows` true only for `"service"` | `CanView("x")` true; with `allows` always false → false |
| `TestBuildMenu_GroupHasNoOwnPermission` | group `g` with leaf `g.a` | `CanView("g")` false; one module returned, it implements `layout.GroupModule` with 1 child |
| `TestBuildMenu_DescriptionFromPresenter` | leaf whose presenter uses `view.WithDescription("Help")` | module implements `layout.Described`, `Description() == "Help"` |

---

## Stage 4 — Docs and cleanup

- `README.md`: document the root `layout` contracts (`UIModule`, `GroupModule`, `Described`, `Notifier`) and the `webtyp.com/layout/app` subpackage (`BuildMenu` turning `[]view.Screen` into layout-agnostic `[]layout.UIModule` + `CanView`, usable with `platformd` or any custom layout).
- Delete `_temp/`. Check: `test ! -e _temp`.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
grep -rn "layout/platformd" app/ crudview/ apiexplorer/ && exit 1 || true
test ! -e _temp
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission or report progress: when the check passes, commit, push and open
  the PR. Stuck for 30 minutes: open the PR anyway listing what fails.
- Commit after every stage (`git add -A` then `git commit -m "stage N: …"`); never `git reset`,
  `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `layout.go`, `capabilities.go`, `factory.go`, `platformd/capabilities.go`, `platformd/factory.go`, `platformd/platformd.go`, `platformd/platformd_test.go`, `crudview/crudview.go`, `crudview/crudview_test.go`, `apiexplorer/explorer.go` | root `layout` owns contracts; `crudview` and `apiexplorer` no longer import `platformd` |
| 2 | `app/menu.go` | `BuildMenu` builds leaves, groups and `CanView` against `layout.UIModule` without importing `platformd` |
| 3 | `tests/menu_test.go` | four tests pass in `gotest` |
| 4 | `README.md`; delete `_temp/` | `test ! -e _temp` |
