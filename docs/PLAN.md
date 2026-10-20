---
PLAN: "feat(platformd): a module explains itself — description + help button, welcome card with steps, alert strip"
EXECUTOR: jules
REVIEWER: none
---

> **Note to the next executor (redispatch after a failed session):** a previous attempt failed with no
> pushed work. The plan below was tightened after reading `AGENTS.md`'s "Component Contract" section
> closely: this repo has **no conditional mounting** — visibility is always a `*SignalBool` driving
> `BindStateFunc(widget.Open, ...)` on an always-present element, never a Go-level `if` around a `Div`. Every
> new piece below follows that exact construction, modeled on `menuOpen`/`brandOpen`/`msgSlot`, which already
> exist in `platformd.go` — read those three before writing anything new.

> Read AGENTS.md at the repo root before any change. This plan is complete: do not guess; if something does
> not fit, stop and ask in the PR. Design history: `webtyp/docs/MODULE_GUIDANCE_MASTER_PLAN.md` (D1 closes
> P1, D2 closes P2, D3 closes P3, D6 closes P6, D19 closes P16) and `devskills/skills/ui-consistency/SKILL.md`
> ("A module explains itself").

## Goal

Four additive pieces in `platformd`, all optional capabilities a `UIModule` may implement — exactly the
pattern `Badged` (`platformd/capabilities.go`) already establishes: the chassis checks with a type
assertion and only draws when present, so a module that implements nothing behaves exactly as today.

1. **`Described`**: a module declares what it is and how to start, shown as a dismissible welcome card and
   reachable again from a "?" button beside the module's title.
2. **The "?" button**: next to the `<h1>` title in the header (`headerTitle`, `platformd.go:~631`), on
   desktop and phone; tapping it opens/reopens the welcome card without opening the nav menu (its own
   `OnClick` must call `e.StopPropagation()` so it does not bubble into the header's own toggle).
3. **A generic steps-with-checkmarks component**, inside the welcome card, for a module with setup
   prerequisites ("1. Add a room ✓ 2. Add a category").
4. **An alert strip**: when a module has unseen alerts (per `MODULE_GUIDANCE_MASTER_PLAN.md`'s wave — the
   alert data itself comes later, from `alert_manager`), the chassis draws a strip above the module's content
   listing them, each with "Ver"/"Marcar visto". This plan builds the **strip and its optional capability**
   only — wiring real alert data through it is `alert_manager`'s own plan (H6), dispatched later.

## Git rules (strict)

- Work on the branch codejob created. Never `git reset`/`checkout -- <file>`/`clean`/`stash`. Commit often.
- **Never pause to ask permission or to report progress.** Delivery check passes → commit, push, open the PR.
  Still failing after 30 minutes → open the PR anyway, list exactly what fails.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 1 — `platformd/capabilities.go`: the two new capabilities

```go
// Described is an optional capability of a UIModule: what it is, for someone who has never used it, plus
// optional first-steps guidance. The chassis draws it as a dismissible welcome card and reopens it from the
// "?" button beside the module's title.
type Described interface {
	Description() lang.Text // what this module is for, in plain words — never the internal structure
	Steps() []Step           // optional; nil/empty means no step list in the card
}

// Step is one setup prerequisite shown with a checkmark once Done is true.
type Step struct {
	Label lang.Text
	Done  bool
}

// AlertedModule is an optional capability of a UIModule: it has alerts a person has not seen yet. The
// chassis draws them as a strip above the module's content. ModuleAlert and this interface are deliberately
// minimal — alert_manager (not this repo) owns storage, "seen" state and delivery; this is only the shape
// the chassis needs to draw the strip.
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}

// ModuleAlert is one row of the strip.
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}
```

Add `lang` to this file's imports if not already present.

## Stage 2 — `platformd.go`: the "?" button and the welcome card

Add an `activeDescribed() (Described, bool)` helper right after `activeLabel` (same walk-the-Modules-and-
Groups pattern `activeIcon`/`activeLabel` already use — do not refactor those two into it, just copy the
walk).

This repo's reactive model (`AGENTS.md` → "Component Contract") has **no conditional mounting**: an
element is always in the tree and a `*SignalBool` drives its CSS-visible state via
`BindStateFunc(widget.Open, func() bool { return someSignal.Get() })`, exactly like `p.menuOpen` already
toggles the nav drawer (`platformd.go:~620,680,732`) and `p.brandOpen` toggles the brand panel
(`platformd.go:~549`). The welcome card and the "?" button follow that **same** construction, not a Go-level
`if`.

Add fields next to `menuOpen`/`brandOpen`'s own declaration and initialization (`platformd.go:~160`,
`~259-265`): `welcomeOpen *SignalBool` / `p.welcomeOpen = NewBool(false)`.

In the `headerTitle` construction (`platformd.go:631-634`, exact current code):

```go
headerTitle := Div().Set(clsHeaderTitle.AsAttr()).
	Child(H1().Set(clsHeaderTitleText.AsAttr()).BindTextFunc(p.activeLabel))
header.Child(headerTitle)
```

Add the "?" button as a sibling of the `H1` inside `headerTitle`, always present (its own CSS hides it when
there is nothing to describe — see below), mirroring how `header.Child(msgSlot)` is unconditionally added
right after:

```go
headerTitle := Div().Set(clsHeaderTitle.AsAttr()).
	Child(H1().Set(clsHeaderTitleText.AsAttr()).BindTextFunc(p.activeLabel)).
	Child(Button().Set(clsHelpButton.AsAttr()).
		Attr("aria-label", "Help").
		OnClick(func(e Event) { e.StopPropagation(); p.welcomeOpen.Set(true) }).
		Child(Text("?")))
header.Child(headerTitle)
```

`clsHelpButton`'s `css.go` recipe hides the button entirely (`display: none`) unless the active module
implements `Described` — drive that from a bound state, not a Go conditional: add a
`BindStateFunc(widget.Current /* or a new widget.State if none fits — check widget's existing states before
adding one */, func() bool { _, ok := p.activeDescribed(); return ok })` on the button, and write the CSS
rule keyed on that state. If no existing `widget.State` fits this "is describable" concept, say so in the
PR with the two states you considered, and use the cheapest one — do not invent one without checking.

**The welcome card**, same construction as `msgSlot` (`platformd.go:~635-640`, a `Div` with a stable
`ID()` and `BindChildren` on a `SignalNodes`) crossed with `menuOpen`'s open/closed toggle
(`BindStateFunc(widget.Open, func() bool { return p.welcomeOpen.Get() })`): a single `Div`, always
mounted, placed in the header region near `msgSlot`, whose **content** (description + steps) is rebuilt
via `BindChildren` driven by a `SignalNodes` that `activeDescribed()`'s result populates whenever
`p.active` changes (hook into wherever `p.active.Set(...)` already runs — grep every call site — so the
card's content tracks the active module without a new signal subscription mechanism). Its open/closed
state is `p.welcomeOpen`, following the `menuOpen` pattern exactly, docked the same way a toast is
(`style.DockedClear`, see `ui-consistency` → "Messages and the phone strip"). A dismiss control calls
`p.welcomeOpen.Set(false)`.

Steps render as a list, each `Step` with a ✓ (done) or empty circle (not done), text dimmed when `Done`.

**First-visit auto-open (D3):** the welcome card opens automatically the first time a person visits a module
that implements `Described`, remembered per person in the browser (`localStorage`, keyed by module id —
read `webtyp/dom`'s existing storage helper if one exists in this repo already; if none does, use the
browser's `localStorage` directly behind the existing `wasm`/`!wasm` split this file already has elsewhere,
and no-op on the server side). Do not block render on it; treat a storage read failure as "not seen yet"
(never crash on a blocked/private-mode storage).

## Stage 3 — the alert strip

Same construction as `msgSlot` exactly (`platformd.go:~635-640`): a `Div` with a stable `ID()`, its rows
driven by `BindChildren` on a new `*SignalNodes` field (`p.alertRows`, initialized next to
`p.notifications`), placed in the same header region `msgSlot` already docks to. Each row: `Subject`, a
"Ver" button calling `OnView`, a "Marcar visto" button that calls `OnSeen` and then removes that row from
`p.alertRows` directly (`p.alertRows.Set(...)` with the row filtered out — the same "mutate the signal, the
bound DOM patches surgically" rule as `menuOpen.Toggle()`; never a full re-render).

Rebuild `p.alertRows` from `activeAlerted()` (same walk-pattern helper as `activeDescribed`) at the same
point(s) the welcome card's content is rebuilt — whenever `p.active` changes (same call sites you hooked in
Stage 2). An empty `UnseenAlerts()` means an empty `p.alertRows`, which renders nothing — no separate
show/hide state needed, unlike the welcome card (a strip with zero rows already takes zero visual space, so
there is no `*SignalBool` for this one).

## Stage 4 — tests

Through the exported API, `tests/` package `platformd_test`, with small fake `UIModule`s implementing
`Described`/`AlertedModule` selectively:

- A module with no `Described`/`AlertedModule` renders exactly as before (no button, no card, no strip) —
  regression guard.
- A module implementing `Described` shows the "?" button; clicking it does not toggle the nav menu
  (`e.StopPropagation()` reached).
- `Steps()` renders each step's done/not-done state correctly.
- A module implementing `AlertedModule` with a non-empty `UnseenAlerts()` shows the strip; an empty slice
  shows nothing.
- `OnView`/`OnSeen` closures are called when their button is activated.

## Stage 5 — docs

`docs/ARCHITECTURE.md`/`README.md`: document `Described`, `Step`, `AlertedModule`, `ModuleAlert` next to
`Badged`'s own documentation, with the same "optional capability, nil-safe, chassis-only-paints" framing.

## Delivery check

- `gotest` green (vet, race, tests, wasm).
- `git diff --stat main --diff-filter=D` empty.
- A module implementing neither new capability renders byte-for-byte as before (no behaviour change).

## Stages

| # | File(s) | What |
|---|---|---|
| 1 | `platformd/capabilities.go` | `Described`, `Step`, `AlertedModule`, `ModuleAlert` |
| 2 | `platformd/platformd.go`, `platformd/css.go` | "?" button, welcome card, first-visit persistence |
| 3 | `platformd/platformd.go` | alert strip |
| 4 | `tests/*` | capability-present/absent regression, button/strip behaviour |
| 5 | `docs/ARCHITECTURE.md`/`README.md` | document the new capabilities |
