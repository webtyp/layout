---
PLAN: "refactor(layout): rightpanel and platformd stop naming element ids"
TAG: v0.2.26
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 12483450681348040283
PR: https://github.com/webtyp/layout/pull/35
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **`webtyp.com/dom v0.13.12` and `webtyp.com/components v0.6.20` are published
> and already required by this repo's `go.mod`.** `dom` is what introduces
> `(*Element).Ref()` and makes a keyed element addressable; every stage below
> depends on it. Do not downgrade either.
>
> **Already done, do not redo:** `crudview`'s WASM row lookups were migrated to
> the `data-row` breadcrumb in v0.2.25 (`findRow` in `bulk_wasm_test.go`). Leave
> them alone.

# PLAN — `layout` stops naming element ids

## Why

`dom` now offers the typed contract that was missing: an author keeps the
`*Element` they built, marks it with `Key()`, and asks it for its live node with
`Ref()`. No global name is invented, so two instances of one component on the
same page cannot collide.

```go
// before — a name the author made up, global, collides at the second instance
el := Div().ID(r.panelID())
strip, ok := Get(r.panelID())

// after — a handle, scoped to the instance that built it
el := Div().Key("strip")
strip, ok := el.Ref()
```

`layout` has two offenders: `rightpanel` (which even **exports** its id scheme)
and `platformd`.

**`ID()` is not forbidden** and this plan adds no check. It stays valid for
elements the *application or chassis* declares. What changes is that these
components stop using it.

## Rules for this plan

- **Never invent a replacement upstream symbol.** If a site needs something
  `dom` does not expose (e.g. an `aria-describedby` equivalent of
  `(*Element).For`), **stop and report it** in the PR description — do not
  declare a local helper. Per `CONSTRUCTION_HARNESS.md`: *"A missing contract at
  a boundary is a defect in the library, not in the consumer."*
- **No `map`** (TinyGo binary budget) and **no standard library** — use
  `webtyp/fmt`.
- Keep every `Key()` that exists: `BindChildren` reconciles on it.
- `gotest ./...` must stay green, `wasm ✅` included.

## Stage 1 — `rightpanel`: the id scheme becomes element handles

`rightpanel/rightpanel.go` builds three elements from a private `panelID()` and
exports two of the names:

```go
101:	wrapper := Div().Set(clsWrapper.AsAttr()).ID(r.panelID())
104:	main   := Section().Set(clsMain.AsAttr()).ID(r.MainPanelID())
134:	aside  := Aside().Set(clsAside.AsAttr()).ID(r.AsidePanelID())
155:	func (r *RightPanel) MainPanelID() string  { return r.panelID() + ".main" }
156:	func (r *RightPanel) AsidePanelID() string { return r.panelID() + ".aside" }
165:	func (r *RightPanel) ShowMain()  { r.showPanel(r.MainPanelID()) }
166:	func (r *RightPanel) ShowAside() { r.showPanel(r.AsidePanelID()) }
169:	strip, ok := Get(r.panelID())
```

Two rightpanels on one page write the same three ids.

**Do this:**

1. Keep the three elements on the struct as fields — `wrapper`, `main`, `aside`
   (`*Element`) — assigned where they are built today, and give each a `Key`:
   `Key("strip")`, `Key("main")`, `Key("aside")`. `Key` is instance-local, so two
   rightpanels do not collide.
2. `showPanel` takes the target `*Element` instead of an id string, and resolves
   the scroll strip through the field:

```go
func (r *RightPanel) ShowMain()  { r.showPanel(r.main) }
func (r *RightPanel) ShowAside() { r.showPanel(r.aside) }

func (r *RightPanel) showPanel(target *Element) {
	strip, ok := r.wrapper.Ref()
	if !ok || target == nil {
		return
	}
	…                       // keep the existing body, but resolve the target
	                        // with target.Ref() instead of Get(id)
}
```
3. **Delete `MainPanelID()`, `AsidePanelID()` and `panelID()`.** A repo-wide grep
   (`grep -rn 'MainPanelID\|AsidePanelID' .`) shows their only callers are
   `ShowMain`/`ShowAside` inside this same file — no consumer outside `layout`
   uses them, so removing them breaks nothing. If the grep disagrees when you run
   it, **stop and report** rather than keeping the exported names.

`crudview` snapshots assert a `.aside'` substring for the scroll-snap target
(`crudview/crudview_test.go`, *"the aside carries the scroll-snap target id"*).
That assertion is about a **dom-assigned** id now: update it to select on the
`aside` part class instead of the id suffix.

## Stage 2 — `platformd`: four sites, three different cases

`platformd/platformd.go`:

| Line | Element | Case |
|---|---|---|
| 599 | `Div().Set(clsMsg…).ID(id+suffix).Key(id+suffix)` | **duplicate** — `ID` and `Key` carry the same string. Delete the `ID(id+suffix).` line; the `Key` already makes it addressable |
| 474 | `Section().Set(clsPanel…).ID(id).Attr("data-id", id)` | replace `ID(id)` with `Key(id)`; `data-id` already exists for selection. Migrate any `Get(id)` for these panels to the element's `Ref()` |
| 390 | `Div().Set(clsMsgSlot…).ID("pd-msg-slot")` | chassis singleton |
| 436 | `…ID("pd-hamburger-btn")` | chassis singleton |

For lines 390 and 436: `platformd` **is** the application chassis, and there is
exactly one of each per page — these are the case `ID()` legitimately exists for.
**Leave them, and add a one-line comment on each** saying so, e.g.:

```go
	// Chassis singleton: one per page by construction, so a stable global id is
	// the intended use of ID() (see webtyp/dom, "Element ids are owned by dom").
```

Before leaving them, confirm with `grep -rn 'pd-msg-slot\|pd-hamburger-btn' .`
that nothing outside `platformd` depends on the literal — if something does, note
it in the PR description.

## Stage 3 — if an id is genuinely required by HTML semantics

`<label for>`, `aria-describedby`, `aria-labelledby` and `<datalist>` need a real
id. Use `(*Element).For(other)` where the relation is label→control. **If the
relation is an `aria-*` one and `dom` exposes no equivalent of `For`, do not
build one here** — record it in the PR description as a `dom` gap.

## Tests

- Existing tests keep passing; update only selectors that referenced a deleted
  id (prefer part classes or `data-*` attributes).
- **New WASM test `TestTwoRightPanels_DoNotCollide`**: mount two `RightPanel`
  instances as siblings in ONE render; assert no panic, that both wrappers exist
  with **different** ids, and that `ShowAside()` on A does not move B's strip.
  This is the proof the collision is gone.

## Acceptance criteria

- `grep -rn '\.ID(\|^\s*ID(' --include=*.go . | grep -v _test | grep -v '/web/'`
  → only the two commented chassis singletons in `platformd`.
- `grep -rn 'MainPanelID\|AsidePanelID\|panelID' --include=*.go .` → **empty**.
- `gotest ./...` green, `wasm ✅` and `race ✅` included.
- `gofmt -l .` → empty; `go vet ./...` clean.

## Out of scope

- `webtyp/dom` — upstream, already published.
- `webtyp/components`, `webtyp/form` — their own plans.
- Any check or panic forbidding `ID()`. That is a later `dom` plan, only after
  every consumer has migrated.

## Stages

| # | Files | Change |
|---|---|---|
| 1 | `rightpanel/rightpanel.go`, `crudview/crudview_test.go` | element handles + `Ref()`; delete the exported id scheme |
| 2 | `platformd/platformd.go` | delete the duplicate `ID`, `Key` for panels, document the two chassis singletons |
| 3 | — | report any missing `aria-*` contract upstream instead of patching |
