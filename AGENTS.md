# Agent Guide — `webtyp/layout`

Constraints for agents working on layout (e.g. `platformd`). Read this before any change.

---

## Hot reload — do NOT compile manually

The WebTyp dev server has hot reload: every source change (Go, CSS in `css.go`,
SSR assets) is recompiled and re-served automatically. Do **not** run `go build`,
`GOOS=js GOARCH=wasm go build`, or re-run `start_development` to "pick up" a change,
and do not poll the wasm endpoint waiting for a rebuild. Just edit the file, then
look at the running app (reload the browser / screenshot). The only reason to build
by hand is a one-off compile check; the running app never needs it.

---

## Construction Harness — the WebTyp way (read first)

The typed, explicit code **is** the harness. Whoever writes against this library is
often an agent that does **not** know it; they must produce correct code guided only
by the signatures, and the compiler must **reject** whatever is wrong. Correctness
lives in the compiler and the signatures, not in a manual you must remember. A
harness moves correctness to the compiler; a manual moves it to the reader — the
first is orders of magnitude more reliable for someone with no context.

Every public API must hold to these principles:

1. **Typed over `any`.** No generic holes (`func(...any)`, `interface{}`) in the
   API — intent-typed methods, like the `webtyp/json` writer (`String`, `Int`,
   `Bool`, `Object`, `Array`). `any` is allowed **only** at the I/O edge, never in
   the data. **Reuse already-declared types** (e.g. `fmt.KeyValue`) instead of
   inventing new ones. Generics with an `any` constraint are the same hole in
   disguise: a signature that does not name its real type is not self-describing.
2. **Explicit over implicit.** The name declares the intent; reading the call is
   enough to know what it does, without opening the implementation.
3. **Illegal states unrepresentable.** If something must not happen, it must not be
   writable. One intent = one path, typed to demand exactly what it needs.
4. **One way to do each thing.** A single construction pattern, with no alternatives
   that force a choice or a trip to the docs.
5. **Minimal public surface.** Export exactly what the author uses; internal
   machinery stays unexported — you cannot misuse what you cannot see.
6. **Fail at compile time, not at runtime.** Order of preference to catch an error:
   compile error → noisy dev-mode diagnostic → (never) silent failure.
7. **Self-describing signatures.** Autocomplete must be enough to build. If using
   the API needs a long document, the API is incomplete.

**Litmus test:** if an agent with no context produces correct code guided only by
autocomplete and a few-line example, the harness is closed. If it needs a manual to
avoid mistakes, something is still untyped.

---

## Component Contract — ONE way (signals)

Layout components implement **only** `Render() *dom.Element` (+ optional `Init(ctx dom.Ctx)`).
There is **NO** `OnMount`/`OnUpdate`/`OnUnmount` and **NO** manual `Update()` (unexported in `dom`).

- One-time setup (hash listener, initial route activation) goes in `Init(ctx)`. The framework runs
  it ONCE — do **not** hand-roll a `mounted bool` guard.
- State the UI shows lives in **typed signals**: e.g. `menuOpen *SignalBool`, `active *SignalString`,
  notifications as a `SignalNodes` rendered with `container.BindChildren(s)`.
- Event handlers only mutate signals (e.g. `menuOpen.Toggle()`); the bound DOM patches
  surgically. **Never** re-render the whole platform, never a Virtual DOM.
- For state changed outside a handler (timers, programmatic notifications): `Set` the signal — it
  patches directly, even from a goroutine. Register teardown with `ctx.OnCleanup(fn)`.

## No Generics

Zero generic functions (follow `webtyp/fmt` codec rule "cero any, cero map"). Concrete typed
signals only: `SignalString`/`SignalBool`/`SignalNodes`, `DeriveString`/`DeriveBool`, `Bind*`,
`Show`. Never `Signal[T]`.

## Minimal Public Surface

Export only the platform's user-facing API (`Platform`, module registration, `Notify`). Unexport
internal helpers and anything only this package uses. Struct fields stay unexported.

## WASM / TinyGo

- Reactive code in `//go:build wasm`; `!wasm` stubs where called from tag-less code.
- No Go stdlib: use `webtyp.com/fmt`. DOM only via `webtyp.com/dom`, never
  `syscall/js`. `switch` not `map`. No `defer/recover`. Embed `dom.Element` by value.
- **No `encoding/json`:** Direct use of the standard library `encoding/json` is prohibited in WASM paths. Use `webtyp.com/json` instead.

## Translatable messages — `lang.Text` + `lang.json`

Framework-authored UI chrome (dialog titles, button labels, empty states,
aria-labels) is English and goes through `webtyp.com/lang` (never
`webtyp.com/fmt/lang`, removed):

- **Fixed UI text is `lang.Text`, data stays `string`.** A component field that
  holds fixed text is typed `lang.Text` and translated in `Render` with
  `lang.Translate(x).String()`. Pass English literals to such fields; never
  pre-translate them.
- **Each argument is one key, exactly as written.** The comma decides word or
  phrase: `lang.Translate("Delete", "%s?", "This", "action", ...)` makes word
  keys; `lang.Translate("No messages yet")` one phrase key. Format verbs pass
  through.
- **Translations are data.** This library's are in `lang.json` at the repo root
  (library format: no `default`, positional lists in `languages` order). Keep it
  in step with `langc sync .`. There is no Go dictionary API.

## Testing

```bash
go install webtyp.com/devflow/cmd/gotest@latest
gotest
```

- `gotest`, never `go test`. Stdlib assertions only. Dual WASM/stdlib; real DOM in WASM tests.
- Cover frequent use cases: menu toggle patches only the nav class; toasts (keyed list) insert/remove
  a single row; routing activates once; `Init` runs once (no `mounted` guard). Publish with `gopush 'message'`.

## UIModule Contract

The `UIModule` interface is the ONLY way to provide a view to the platform chassis.

---

## Build tags belong to the consumer

The isomorphic Go ecosystem means files compile to both backend and browser. We use build tags to control what goes into the WASM binary.

| File tag | Target | Purpose |
|---|---|---|
| (untagged) | Both | Shared logic, constants, `svg.Icon` references. Ships to WASM. |
| `//go:build wasm` | Browser | Browser-only logic, `syscall/js` interactions (rare, via `dom`). |
| `//go:build !wasm` | Backend/SSR | SSR asset collection, `sprite.Define`. NEVER ships to WASM. |

> [!IMPORTANT]
> `webtyp/svg` is split by PACKAGE, not by build tag inside the library.
> - `webtyp.com/svg`: Shared reference (safe for untagged code).
> - `webtyp.com/svg/sprite`: Backend-only definition.
>
> `svg/sprite` compiles for WASM too, so forgetting the `!wasm` tag on your
> `svg.go` does NOT fail the build — it silently ships every path `d` string
> into the browser bundle. Only the dependency-graph check catches it:
>
> ```bash
> GOOS=js GOARCH=wasm go list -deps ./platformd | grep webtyp/svg/sprite  # MUST be empty
> GOOS=js GOARCH=wasm go list -deps ./crudview | grep webtyp/svg/sprite   # MUST be empty
> ```

## SVG icons — name is shared, drawing is backend-only

1. **Declare** the reference in an **untagged** file using `svg.Icon("name")`.
2. **Define** the drawing in a `//go:build !wasm` file (typically `svg.go`) using `sprite.Define(iconConst, "viewBox", sprite.Path("d..."))`.
3. **Render** using `icon.Render("class")`. This is the ONLY supported render path.

Do NOT hand-build `<svg><use href="#id"/></svg>` blocks using `svg.Svg()` or `svg.Use()`.

**A glyph shared with a component comes from `webtyp/icons`.** `crudview`'s
footer trash/pencil are the same marks `targetlist`/`targetdate` draw on their
rows — so all of them import `webtyp.com/icons/trash` and
`.../pencil` and take `Ref` + `Def()`. Never re-define a shared glyph's
geometry here, never import it sideways from `components`. `plus`/`undo` (the
toggle button) come from `webtyp/icons` too. A glyph private to `crudview`
still lives in `crudview/svg.go`.

**The mark's *skin* is `listselect.Apply` — assembled, not re-declared.** The
check box / glyph reveal / colour rules live once in `components/listselect`;
`crudview` styles only its own footer buttons to match (`As(Danger)` for the
delete commit, so it wears the same white glyph as the checked rows).

## crudview footer — capability-gated affordances

The three footer actions are gated at **render time** by type assertion on the
presenter, each on the capability it needs:

| Affordance | Capability | Absent → |
|---|---|---|
| create (`+` state of the toggle) | `view.Saver` | toggle still renders (it is also `↺` for leaving selection mode) but its `+` state is `disabled` and `toggleAction` no-ops |
| `🗑` (`cv-cruddelete`) | `view.Deleter` | button not rendered |
| `✏` (`cv-crudedit`, bulk field patch) | `view.Updater` | button not rendered |

Editing ONE loaded record is **not** a footer affordance and not gated on
`Updater` — it rides `Saver.Save` via `autoSaveAction`. `✏` is the *bulk* entry
point only. See `docs/BULK_ACTIONS_MASTER_PLAN.md` §4.

## RevealedBy + an icon child → the part needs a flow

A footer part that carries `style.RevealedBy(widget.Open)` **and** wraps an
icon (it is a `<button>` that is a flex parent, not the icon itself) must also
carry a flow — `style.Row(style.SpaceNone)`. Without it the `@layer states`
reveal restores `display:block` (a flow-less part), `CenterContent()` goes
inert, and the glyph strands at the leading edge. `action-delete` /
`action-edit` learned this; `action` (never hidden) does not need it. Same fix
and reason as `components/listselect/css.go`.

## SSR asset provider names are matched by regex — the name must be EXACT

`webtyp/ssr` collects a package's SSR output by scanning `css.go`/`js.go`/`svg.go`/`html.go`
for functions whose names match **exactly**: `RenderCSS`, `RootCSS`, `RenderHTML`, `RenderJS`,
`IconSvg`. A CSS builder named anything else (e.g. `GenerateCSS`) is **silently never emitted** —
the component renders with **zero styling** and nothing fails at build time.

Two rules keep a package detectable:

- **Exact name.** The CSS entry point MUST be `RenderCSS` (not `GenerateCSS`, `Styles`, …). Because
  the dot-imported `css` package already exports a free `RenderCSS`, declare it as a **method**
  (`func (v *CrudView) RenderCSS() *Stylesheet`) to avoid the name collision, matching
  `platformd`/`rightpanel`.
- **One receiver per package.** `ssr` requires all providers in a package to share ONE receiver
  (or all be free functions). So `RenderCSS` and `IconSvg` in the same package must both be methods
  on the same type — never mix a method with a free function, or receiver detection produces code
  that calls a method that doesn't exist.

Symptom to recognize: a component renders unstyled while its icons appear giant/black (icons emitted
via `IconSvg`, CSS not emitted because of a wrong name).

---

## Documentation First

Update docs **before** code and before `gopush`: `docs/ARCHITECTURE.md` (platform lifecycle:
`Init(ctx)` once, signals, no `Update()`), update any lifecycle diagram (`flowchart TD`, no
`subgraph`, `<br/>` for breaks), and re-index `README.md` so every `docs/` file is linked.
