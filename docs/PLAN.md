---
PLAN: "fix: CrudView.Reload has no way for a host to observe a load failure, unlike Save/Delete"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# PLAN — `CrudView`: add `OnLoadError`, the missing sibling of `OnSaved`/`OnDeleted`

You are an external agent with **zero prior context** about this project. Everything you need is
in this file. Read `crudview/AGENTS.md`'s "Component Contract" and "Documentation First" sections
first for this package's conventions, then this file fully before writing code.

## 0. Prerequisite — run this first

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

All tests run with `gotest`, never `go test` directly.

## 1. The bug, proven by a failing test already in this repo

`crudview/reload_error_test.go` (already committed on this branch) reproduces a real bug found
while manually testing a downstream app: a list screen showed "0 / 0" even though the server held
real rows for the current tenant, and there was **no way at all**, anywhere in the app, to learn
that the load had actually failed underneath (the real cause was a separate bug in
`webtyp.com/mcp`, already fixed in that repo's own `docs/PLAN.md` — but this package's silence is
what let it go completely unnoticed).

Run the test now and confirm it is RED — it does not even compile yet, which is the expected
starting point (see §3, "why this test doesn't compile today"):

```bash
gotest
# crudview/reload_error_test.go:41:3: unknown field OnLoadError in struct literal of type CrudView
```

The acceptance criterion for this plan: **`gotest` must go fully green**, including this test,
with no other test regressing.

## 2. Root cause

`CrudView.Init` (`crudview.go`) performs the component's first load like this:

```go
if v.Presenter != nil {
	v.Reload(func(err error) {
		if err != nil {
			Log(err.Error())
		}
	})
}
```

`Log` writes to the browser's JavaScript console only. No end user ever opens devtools, so a
Reload failure is, in practice, **silent**. Compare this to how the exact same kind of
asynchronous failure is already handled elsewhere on this same type:

```go
OnSaved   func(err error)                // saveAction → reportSave calls this on failure
OnDeleted func(ids []string, err error)  // deleteAction calls this on failure
```

Both `Save` and `Delete` already give a host app an exported hook to observe and display an
error. `Reload` — the one operation that runs automatically, unconditionally, on every mount — has
no equivalent. This is the asymmetry this plan closes.

## 3. Design gate (required — this changes public API)

**1. Prior art.** Within this very type, `OnSaved`/`OnDeleted` already establish the exact
convention this plan extends: an exported `OnX func(..., err error)` field a host sets to observe
one specific async operation's outcome, invoked from inside the method that performs it. More
broadly, "an `onError` callback for an async data operation" is a universal, well-worn shape
(`net/http.Server.ErrorLog`, every JS data-fetching library's `onError`) — nothing novel is being
introduced, only a documented gap in this type's own existing family of hooks is being closed.

**2. Novice-name test.** `OnLoadError` read aloud: "the function called when loading fails" —
understood without documentation, and it sits next to `OnSaved`/`OnDeleted` in the struct so the
naming family is self-evident by proximity. No abbreviation, no boolean parameter.

**3. Complexity ledger.**
```
Concepts the developer must learn   +1  (closes a gap in an existing family — OnSaved/OnDeleted
                                          already taught the shape; this is not a new concept)
Files touched to close this gap     +0  (crudview.go only, no new file)
Lines at the call site              +0 required (nil is safe, matching every other On* field)
                                     +1 optional (a host that wants to show the error)
Ways to observe a load failure      0 → 1  (never ends positive — there was no way before)
```

**4. Where it belongs.** On `CrudView` itself, in `crudview.go`, declared next to `OnSaved` /
`OnDeleted` / `OnAfterReload` — same struct, same file, same responsibility this type already
owns (reporting the outcome of its own async operations). Not a new package, not a new type.

**5. What this deletes.** The ad-hoc, unobservable inline closure inside `Init()` — `func(err
error) { if err != nil { Log(err.Error()) } }` — is replaced by moving that responsibility into
`Reload()` itself (see §4), so every caller of `Reload()`, not just `Init()`'s own first call,
gets the same reporting behavior for free. `Log` stays (it is the harness's "loud dev-mode
diagnostic" trail, principle 5) — `OnLoadError` adds the host-visible layer on top of it, it does
not replace it.

## 4. The fix

In `crudview.go`:

1. Add the field, next to `OnAfterReload`:
   ```go
   // OnLoadError reports a Reload failure — the List-side sibling of OnSaved/
   // OnDeleted. nil (the default) disables it silently, same as every other
   // On* hook here; Reload always Logs the error regardless, so this is an
   // ADDITIONAL host-visible signal, not a replacement for the dev-console
   // trail.
   OnLoadError func(err error)
   ```
2. Move the error-reporting into `Reload` itself, so `Init()` no longer needs its own bespoke
   closure:
   ```go
   func (v *CrudView) Reload(done func(error)) {
   	if done == nil {
   		done = func(error) {}
   	}
   	if v.Presenter == nil {
   		done(nil)
   		return
   	}
   	v.Presenter.Reload(func(err error) {
   		if err != nil {
   			Log(err.Error())
   			if v.OnLoadError != nil {
   				v.OnLoadError(err)
   			}
   			done(err)
   			return
   		}
   		v.filter()
   		if v.OnAfterReload != nil && v.list != nil {
   			v.OnAfterReload(v.list)
   		}
   		done(nil)
   	})
   }
   ```
3. Simplify `Init()`'s own call now that `Reload` reports on its own:
   ```go
   if v.Presenter != nil {
   	v.Reload(nil)
   }
   ```
   (`Reload`'s own `if done == nil { done = func(error) {} }` already makes `nil` safe here,
   exactly like every other internal call site in this file already does.)

### What NOT to do

- **Do not add a default visible UI** (a toast, a banner) inside `crudview` itself. This package
  is host-agnostic — deciding HOW to show an error is the composing app's job, exactly like it
  already is for `OnSaved`/`OnDeleted`. This plan only makes the failure *observable*; it does not
  design the error UI.
- **Do not remove `Log(err.Error())`.** It stays as the unconditional dev-diagnostic trail
  (harness principle 5); `OnLoadError` is additive.
- **Do not touch `webtyp.com/view`'s `core.Reload`/`callerLister.list`.** The bug that actually
  caused the browser-visible symptom this test's doc-comment describes was in `webtyp.com/mcp`
  (nil args encoding as JSON `null`) — already fixed in that repo, see its own `docs/PLAN.md`.
  This plan is scoped only to making a Reload failure observable, regardless of its cause.

## 5. Verification

```bash
gotest
# vet ✅, race ✅, tests ✅, wasm ✅ — TestCrudView_Reload_ReportsErrorViaOnLoadError now
# PASSES (and compiles), no other test in crudview/ or elsewhere changes behavior.
```

## Stages

| # | Stage | File(s) | Acceptance |
|---|---|---|---|
| 1 | Add `OnLoadError func(err error)` field | `crudview.go` | Declared next to `OnAfterReload`, doc comment as above |
| 2 | Wire it inside `Reload()` | `crudview.go` | Called on error, after `Log`, before `done(err)` |
| 3 | Simplify `Init()`'s call site | `crudview.go` | `v.Reload(nil)`, bespoke closure removed |
| 4 | Verify | — | `gotest` green, `TestCrudView_Reload_ReportsErrorViaOnLoadError` passes |
