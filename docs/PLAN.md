---
PLAN: "feat(platformd): a module explains itself — description + \"?\" button, welcome card with steps, alert strip"
EXECUTOR: jules
REVIEWER: none
---

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

In the `headerTitle` construction (`platformd.go:~631-634`), after the `H1`, conditionally add a button:

```go
headerTitle := Div().Set(clsHeaderTitleText.AsAttr()). // existing wrapper, name may differ slightly — read the current code
	Child(H1().Set(clsHeaderTitleText.AsAttr()).BindTextFunc(p.activeLabel))
if d, ok := p.activeDescribed(); ok {
	_ = d
	headerTitle.Child(
		Button().Set(clsHelpButton.AsAttr()). // new class — add its recipe in css.go, follow style.Button's own recipe, never hand-composed (widget-styling: "a missing recipe is a defect in widget/style" — if style has no fitting recipe, use style.Button(Subtle) or whatever this repo's css.go already uses for a small icon-only control, and say in the PR which you picked and why)
			Attr("aria-label", "Help").
			OnClick(func(e Event) { e.StopPropagation(); p.openWelcome.Set(true) }).
			Child(Text("?")))
}
```

Add `p.openWelcome = NewBool(false)` next to `p.active`/`p.navStowed`'s own initialization (same
`BindState`-style reactive field this file already uses throughout — grep `NewBool`/`NewString` for the
exact constructor name and follow it).

The welcome card: a `Div` bound to render only `if d, ok := p.activeDescribed(); ok && p.openWelcome.Get()`,
drawn as a dismissible panel over or beside the module's content (pick the placement that doesn't fight the
phone-strip rule in `ui-consistency` — a module's content strip on phone is full-width, so the welcome card
docks the same way a toast does: `style.DockedClear`, see `ui-consistency` → "Messages and the phone strip").
Contents: `d.Description()` as body text, then — if `d.Steps()` is non-empty — a vertical list, each `Step`
rendered with a ✓ (done) or an empty circle (not done), text struck through or dimmed when `Done`. A "Got
it"/dismiss control sets `p.openWelcome.Set(false)`.

**First-visit auto-open (D3):** the welcome card opens automatically the first time a person visits a module
that implements `Described`, remembered per person in the browser (`localStorage`, keyed by module id —
read `webtyp/dom`'s existing storage helper if one exists in this repo already; if none does, use the
browser's `localStorage` directly behind the existing `wasm`/`!wasm` split this file already has elsewhere,
and no-op on the server side). Do not block render on it; treat a storage read failure as "not seen yet"
(never crash on a blocked/private-mode storage).

## Stage 3 — the alert strip

A `Div`, placed above the module's content area (same header region a `platformd.Auto()` toast already
docks to — reuse that placement, do not invent a second one), bound to render only when
`activeAlerted()` (same walk-pattern helper as `activeDescribed`) returns a non-empty `UnseenAlerts()`. Each
row: `Subject`, a "Ver" button calling `OnView`, a "Marcar visto" button calling `OnSeen` then removing that
row from the rendered list (the chassis does not re-fetch — the module that produced `ModuleAlert` owns
re-rendering `UnseenAlerts()` on its next read, typically after its own `OnSeen` closure updates its state
and asks the chassis to redraw, following whatever reactive-redraw mechanism `BindChildren` already uses
elsewhere in this file).

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
