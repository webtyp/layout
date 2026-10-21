---
PLAN: "feat(platformd): help button beside the module title (Described) and alert strip above the content (AlertedModule)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 13023297444321200271
---

> Read AGENTS.md at the repo root before any change.

# platformd: «?» help button and the alert strip

## Why

Ecosystem rule: a module explains itself; nobody should have to explain it. `platformd` already
declares two optional capabilities of a `UIModule` (`platformd/capabilities.go`, shipped in v0.3.57)
that nothing draws yet:

```go
type Described interface {
	Description() lang.Text // what this module is for, in plain words
	Steps() []Step          // optional
}
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}
```

This plan draws them (owner decisions, do not change):

1. **«?» button glued to the right of the module title** (the header `<h1>`), on desktop and phone,
   only when the active module implements `Described` with a non-empty `Description()`. Tapping it
   shows/hides a panel with the description. Its tap **never** toggles the menu (the header's own
   click toggles the menu: stop propagation). Switching module closes the panel.
2. **Alert strip above the content** of the active module, when it implements `AlertedModule` and
   `UnseenAlerts()` is not empty: one row per alert with its subject, a «View» button (`OnView`)
   and a «Mark as seen» button (`OnSeen`). After either button the strip is rebuilt from
   `UnseenAlerts()`.

Not in this plan (later plan): `Steps()` (the welcome card with checkmarks) and the first-visit
memory. Do not draw `Steps()` here.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

## Environment

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
```

A PR that does not compile, or whose tests could not run, is a failed delivery. Never pause to ask
permission or report progress: when the Delivery check passes, commit, push and open the PR. Stuck
30 minutes: open the PR anyway listing what fails.

## APIs used (already used in `platformd/platformd.go`; copy those call shapes)

| Symbol | Shape | Example in this repo |
|---|---|---|
| `NewBool(false) *SignalBool`, `.Get()`, `.Set(bool)`, `.Toggle()` | `webtyp.com/dom` v0.13.26 | `p.menuOpen` (`platformd.go:259`, `:621`) |
| `NewNodes() *SignalNodes`, `.Set([]*Element)`, `.Get()` | same | `p.navIcon` (`platformd.go:265`, Activate end) |
| `(*Element).BindChildren(*SignalNodes)` | same | `msgSlot` (`platformd.go:641`) |
| `(*Element).BindStateFunc(widget.Open, func() bool)` | same | `overlay` (`platformd.go:680`) |
| `(*Element).OnClick(func(Event))`, `Event.StopPropagation()` | same | `right` (`platformd.go:648`) |
| `lang.Translate(values ...any) *fmt.Conv` → `.String()` | `webtyp.com/lang` v0.1.7 | `crudview/crudview.go:323` |
| `NamePlatform.Class("x")` | `widget.Name` | `platformd.go:22-60` |
| `style.RevealedBy(widget.Open)` | `webtyp.com/widget/style` | `css.go:116` |

**SSR note** (already documented at `platformd.go:640-643`): `String()` rendering does not process
`BindChildren`; after every `BindChildren(signal)` also add the signal's current nodes as children
(`for _, n := range sig.Get() { el.Child(n) }`), exactly like `msgSlot`. The tests below read
`Render().String()` and depend on it.

## Stage 1 — state and classes (`platformd/platformd.go`)

1. Classes, next to the others:

```go
clsHelpBtn      = NamePlatform.Class("help-btn")
clsHelpPanel    = NamePlatform.Class("help-panel")
clsAlertStrip   = NamePlatform.Class("alert-strip")
clsAlertRow     = NamePlatform.Class("alert-row")
clsAlertSubject = NamePlatform.Class("alert-subject")
clsAlertAction  = NamePlatform.Class("alert-action")
```

2. Texts (English keys, translated in `lang.json`):

```go
const (
	textHelp       = "What is this module?"
	textAlertView  = "View"
	textAlertSeen  = "Mark as seen"
)
```

3. Fields of `Platform` (internal state block): `helpOpen *SignalBool`, `helpSlot *SignalNodes`
   (0 or 1 button), `helpPanel *SignalNodes` (0 or 1 text node), `alertStrip *SignalNodes`.
   `Init` creates them (`NewBool(false)`, `NewNodes()` ×3), next to `p.menuOpen`.

## Stage 2 — the active module and its two refreshes (`platformd/platformd.go`)

1. Extract from `Activate` the loop that finds `targetModule` into
   `func (p *Platform) moduleByID(id string) UIModule` (searches top level and group children,
   returns `nil` if none). `Activate` calls it; behaviour unchanged.
2. `func (p *Platform) refreshHelp(m UIModule)`:
   - `p.helpOpen.Set(false)`.
   - `d, ok := m.(Described)`; if `!ok || d.Description() == ""`: `p.helpSlot.Set(nil)`, `p.helpPanel.Set(nil)`, return.
   - button: `Button().Set(clsHelpBtn.AsAttr()).Key("help-" + m.ModelName()).Attr("aria-label", lang.Translate(textHelp).String()).Text("?").OnClick(func(e Event) { e.StopPropagation(); p.helpOpen.Toggle() })`
   - panel text: `P().Key("helpt-" + m.ModelName()).Text(lang.Translate(d.Description()).String())`
   - `p.helpSlot.Set([]*Element{button})`, `p.helpPanel.Set([]*Element{text})`.
3. `func (p *Platform) refreshAlerts(m UIModule)`:
   - `a, ok := m.(AlertedModule)`; not ok → `p.alertStrip.Set(nil)`, return.
   - one row per `ModuleAlert` (`Key("alert-" + al.ID)`): `Div(clsAlertRow)` with
     `Span(clsAlertSubject).Text(al.Subject)`, a `Button(clsAlertAction)` text `textAlertView` whose
     click calls `al.OnView()` then `p.refreshAlerts(m)`, and a `Button(clsAlertAction)` text
     `textAlertSeen` whose click calls `al.OnSeen()` then `p.refreshAlerts(m)`. A nil `OnView` or
     `OnSeen` → that button is not drawn. Capture `al` per iteration (`al := al`).
   - `p.alertStrip.Set(rows)` (`nil` when there are none).
4. In `Activate`, right after `targetModule.Activate()` (inside the same `if targetModule != nil`),
   call `p.refreshHelp(targetModule)` and `p.refreshAlerts(targetModule)`. When `targetModule` is
   `nil`, set the three node signals to `nil`.

## Stage 3 — drawing (`platformd/platformd.go`, `Render`)

1. In `headerTitle`, after the `H1`: a `Span().BindChildren(p.helpSlot)` (plus its current nodes,
   SSR note).
2. Right after `root.Child(header)`: the help panel
   `Div().Set(clsHelpPanel.AsAttr()).Attr("role", "note").BindStateFunc(widget.Open, func() bool { return p.helpOpen.Get() }).BindChildren(p.helpPanel)` (plus current nodes).
3. In the stage, **before** the module panels loop: `Div().Set(clsAlertStrip.AsAttr()).Attr("role", "status").BindChildren(p.alertStrip)` (plus current nodes).

## Stage 4 — styles (`platformd/css.go`)

Add parts (copy these recipes exactly; they reuse recipes that already exist in this file):

| Part | Recipe |
|---|---|
| `help-btn` | same list as `brand-trigger` (`css.go:90-96`) |
| `help-panel` | `style.Stack(style.Space1)`, `style.As(style.Panel)`, `style.Pad(style.Space3)`, `style.RevealedBy(widget.Open)` |
| `alert-strip` | `style.Stack(style.Space1)` |
| `alert-row` | `style.Row(style.Space2)`, `style.As(style.Inverse)`, `style.Round(style.RadiusSm)`, `style.PadInline(style.Space3)` |
| `alert-subject` | `style.Fill()` |
| `alert-action` | same list as `brand-trigger` |

Tests never assert CSS (AGENTS.md). If a recipe name above does not exist in the published
`widget` tag, stop and name it in the PR.

## Stage 5 — translations

Run `langc sync .` in the repo root and fill column `es` in `lang.json`:

```json
"Mark as seen": ["Marcar como visto"],
"View": ["Ver"],
"What is this module?": ["¿Qué es este módulo?"],
```

## Stage 6 — tests (`platformd/platformd_test.go`, same package and `mockModule` as today)

Test doubles in the same file:

```go
type describedModule struct {
	mockModule
	desc lang.Text
}
func (m *describedModule) Description() lang.Text { return m.desc }
func (m *describedModule) Steps() []Step          { return nil }

type alertedModule struct {
	mockModule
	alerts []ModuleAlert
}
func (m *alertedModule) UnseenAlerts() []ModuleAlert { return m.alerts }
```

| # | Setup | Action | Expected (`Render().String()`) |
|---|---|---|---|
| H1 | modules: `describedModule{id "a", desc "Keeps the clinic's supplies"}`, `mockModule{id "b"}` | `Init` (a active) | contains `aria-label="What is this module?"` and `Keeps the clinic&#39;s supplies` or the escaped form the renderer uses (assert on `Keeps the clinic`) |
| H2 | same | `Activate("b")` | no `What is this module?` |
| H3 | `describedModule` with `desc ""` | `Init` | no `What is this module?` |
| H4 | H1 setup | `p.helpOpen.Toggle()`, then `Activate("b")`, `Activate("a")` | `p.helpOpen.Get() == false` |
| H5 | `alertedModule{id "a", alerts: [{ID "1", Subject "Stock below minimum", OnView, OnSeen}]}` | `Init` | contains `Stock below minimum`, `View`, `Mark as seen` |
| H6 | H5, `OnSeen` sets `m.alerts = nil` and a flag | call `p.refreshAlerts(m)` after invoking the row's `OnSeen` the way the click does (call the closure: extract the click body into `func (p *Platform) seeAlert(m UIModule, al ModuleAlert)` and `viewAlert`, and test those) | flag set; render no longer contains `Stock below minimum` |
| H7 | alert with `OnView == nil` | `Init` | no `>View<` button for that row |
| H8 | `alertedModule` with no alerts, and a plain module | `Init` | no element with class `pd-alert-row` (`NamePlatform.Class("alert-row")` string) |
| H9 | `describedModule` inside `NewGroupUIModule` | `Activate(childID)` | contains the help button |

If `seeAlert`/`viewAlert` are extracted (H6), the click handlers call them; no other change.

## Stage 7 — docs

- `platformd/README.md`: section «Help and alerts» — the two capabilities, what the chassis draws,
  that `Steps()` is not drawn yet.
- `docs/ARCHITECTURE.md`: one line each under platformd.

## Delivery check

```bash
go build ./...
GOOS=js GOARCH=wasm go vet ./platformd/
gotest
git diff --stat main   # platformd/platformd.go, platformd/css.go, platformd/platformd_test.go, lang.json, platformd/README.md, docs/ARCHITECTURE.md
```

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `platformd/platformd.go` | state + classes |
| 2 | `platformd/platformd.go` | `moduleByID`, `refreshHelp`, `refreshAlerts`, wired in `Activate` |
| 3 | `platformd/platformd.go` | drawn in `Render` |
| 4 | `platformd/css.go` | parts added |
| 5 | `lang.json` | three keys with `es` |
| 6 | `platformd/platformd_test.go` | H1–H9 green, existing tests green |
| 7 | `platformd/README.md`, `docs/ARCHITECTURE.md` | docs match |
