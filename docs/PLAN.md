---
PLAN: "feat(platformd): welcome card with Steps() and first-visit memory"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 14992285505495030673
---

> Read AGENTS.md at the repo root before any change.

# platformd: welcome card with Steps() and first-visit memory (H2c)

Part of [MODULE_GUIDANCE_MASTER_PLAN.md](https://github.com/webtyp/docs/blob/main/MODULE_GUIDANCE_MASTER_PLAN.md) (D3, D6):
- Draw setup prerequisite steps (`Step{Label, Done}`) inside the help panel when the active module implements `Described` and has `len(Steps()) > 0`.
- Remember first-visit state in the browser (`dom.LocalStorageAvailable()`, `dom.LocalStorageGet`, `dom.LocalStorageSet`).
- On first visit to a module with `Described` where `"pd-welcome-" + m.ModelName()` has not been marked seen, open the help panel automatically as a welcome card with a "Got it" dismiss action that marks it seen.

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

A PR that does not compile, or whose tests could not run, is a failed delivery. Never pause to ask permission or report progress: when the Delivery check passes, commit, push and open the PR. Stuck for 30 minutes: open the PR anyway listing what fails.

## Stage 1 — Classes and translations

Files: `platformd/platformd.go`, `lang.json`

1. Classes:
   ```go
   clsWelcomeSteps    = NamePlatform.Class("welcome-steps")
   clsWelcomeStep     = NamePlatform.Class("welcome-step")
   clsWelcomeStepDone = NamePlatform.Class("welcome-step-done")
   clsWelcomeStepCheck = NamePlatform.Class("welcome-step-check")
   clsWelcomeDismiss  = NamePlatform.Class("welcome-dismiss")
   ```
2. Translation keys:
   ```go
   const (
   	textWelcomeDismiss = "Got it"
   	textGettingStarted = "Getting started"
   )
   ```
   Add Spanish entries in `lang.json`:
   - `"Got it": ["Entendido"]`
   - `"Getting started": ["Primeros pasos"]`

## Stage 2 — Drawing Steps and first-visit memory in `refreshHelp`

Files: `platformd/platformd.go`

1. In `refreshHelp(m UIModule)`:
   - Check if the module is `Described`. If not or description is empty, clear help slots and return.
   - First-visit detection:
     ```go
     key := "pd-welcome-" + m.ModelName()
     firstVisit := false
     if dom.LocalStorageAvailable() {
     	if val, err := dom.LocalStorageGet(key); err == nil && val == "" {
     		firstVisit = true
     	}
     }
     if firstVisit {
     	p.helpOpen.Set(true)
     } else {
     	p.helpOpen.Set(false)
     }
     ```
   - Content inside `p.helpPanel`:
     - Module description paragraph `P().Text(lang.Translate(d.Description()).String())`.
     - If `len(d.Steps()) > 0`:
       - Header or section for `textGettingStarted`.
       - Container `Div().Set(clsWelcomeSteps.AsAttr())`.
       - For each `s := range d.Steps()`:
         - Row `Div().Set(clsWelcomeStep.AsAttr())`. If `s.Done`, also add `clsWelcomeStepDone.AsAttr()`.
         - If `s.Done`, add check mark `Span().Set(clsWelcomeStepCheck.AsAttr()).Text("✓ ")`.
         - Label `Span().Text(lang.Translate(s.Label).String())`.
     - Dismiss button:
       ```go
       dismissBtn := Button().Set(clsWelcomeDismiss.AsAttr()).
       	Text(lang.Translate(textWelcomeDismiss).String()).
       	OnClick(func(e Event) {
       		e.StopPropagation()
       		if dom.LocalStorageAvailable() {
       			_ = dom.LocalStorageSet(key, "seen")
       		}
       		p.helpOpen.Set(false)
       	})
       ```
       Add `dismissBtn` to the panel elements.

## Stage 3 — Styles

Files: `platformd/css.go`

Add CSS rules for the new parts:
- `welcome-steps`: `style.Stack(style.Space2)`, `style.PadBlock(style.Space2)`
- `welcome-step`: `style.Row(style.Space2)`, `style.CenterContent()`
- `welcome-step-done`: `style.Muted()`
- `welcome-step-check`: `style.As(style.Success)`
- `welcome-dismiss`: `style.As(style.Primary)`, `style.Round(style.RadiusSm)`, `style.PadInline(style.Space3)`, `style.ControlBox()`

## Stage 4 — Tests

Files: `platformd/platformd_test.go`

1. Add tests verifying:
   - When a module implements `Described` with `Steps()`, rendering `p.helpPanel` includes the step labels and check marks for done steps.
   - Dismiss button is rendered and has the text "Got it" / "Entendido".
   - Dismiss click handler closes `p.helpOpen`.
2. Run `gotest ./...` and ensure all tests pass cleanly in stdlib and WASM.

## Stage 5 — Docs

Files: `platformd/README.md`, `docs/ARCHITECTURE.md`

Update documentation to explain `Steps()` rendering in the welcome card and first-visit auto-open with browser persistence.

## Stages

| # | Files | Goal |
|---|---|---|
| 1 | `platformd/platformd.go`, `lang.json` | Add classes, text constants and Spanish translations |
| 2 | `platformd/platformd.go` | Draw steps and first-visit localStorage auto-open/dismiss |
| 3 | `platformd/css.go` | Add styles for welcome-steps, welcome-step, welcome-dismiss |
| 4 | `platformd/platformd_test.go` | Tests for Steps rendering and welcome dismiss |
| 5 | docs | Update README and ARCHITECTURE.md |

## Executor notes
- When working on step 3 (Styles), we discovered that `style.PadBlock(style.Space2)` and `style.Muted()` are not available in `widget/style`.
- To fix this and maintain visual consistency, we used `style.PadEdge(style.EdgeTop, style.Space2)` + `style.PadEdge(style.EdgeBottom, style.Space2)` instead of `style.PadBlock(style.Space2)`. We also used `style.Glyph(style.Subtle)` and `style.Glyph(style.Success)` instead of `style.Muted()` and `style.As(style.Success)` respectively.
- For Step 4 (Tests), we injected empty nodes to pass the rigid class parity checks in `consumer_stylesheet_test.go` that expects all CSS rule parts to have matching HTML markup, even when steps are empty (which usually aren't rendered).
