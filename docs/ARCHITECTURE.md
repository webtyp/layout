# webtyp/layout Architecture

## Package Layout

    webtyp/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    │   └── modules/    # Demo modules, one package each (devices, medicalhistory, about):
    │                   # view + data + icon, owned by the module, not the chassis
    ├── rightpanel/     # THE module skeleton: frame, two columns, aside bands,
    │                   # mobile master-detail strip. Owns every layout primitive.
    ├── crudview/       # CRUD controller: state machine + orchestration.
    │                   # Renders NO frame — composes rightpanel.
    ├── login/          # Pre-authentication screen: card on a brand backdrop.
    └── landing/        # Public multi-page website: typed sections in call
                        # order, explicit anchors, RenderPages() per URL.

## Who owns what

The module frame has exactly one owner: `rightpanel`. It owns every layout
primitive — the split, the two columns, the aside bands, the mobile
master-detail strip — and every module in this repository composes it.
`crudview` renders NO frame: it is a CRUD controller that builds a
`rightpanel.RightPanel`, fills its slots (Title, Article, Aside, AsideControls,
AsideFooter) and keeps only the state machine. `platformd` is the shell
(routing, chrome) and is a third thing.

`landing` sits outside that trio: it is not an application shell but a public
site, so it owns page-level concerns the others never have — one document per
URL, per-page SEO metadata, and navigation by anchor. A section there holds a
*builder*, not an element: the header and footer are rendered once per emitted
page, and a `dom.Element` has exactly one parent, so a shared instance would
panic on the second page. Components (`infobar`, `sitenav`, `herobanner`,
`statgrid`, `contentcard`) are always embedded through `Render()` — a component
handed to `Child()` as itself serializes from its zero embedded `Element` and
disappears as an empty `<></>`.

The reference demo (the `about`, `devices` and `medicalhistory` modules) lives
in its own repo — [github.com/webtyp/app-demo](https://github.com/webtyp/app-demo) —
so the demo's ORM/storage/input dependencies never enter layout's `go.mod`.
Each module owns its view, its data and its icon (declared shared in the
untagged file, drawn in a `//go:build !wasm` `svg.go` via `sprite.Define`). The
chassis only ships its own chrome glyphs
(`IconUser`, `IconBrand`, the menu button): content icons are a
module decision, and the rail renders whatever `Module.Icon()` returns.

Three tests in `conformance_test.go` make the split un-reintroducible:

- `TestOnlyOneOwnerOfTheGrid` fails if a second package emits
  `Split`/`MasterDetail` — a second skeleton is how the duplication started.
- `TestNoLocallyDeclaredSeams` rejects a locally-declared `Filterable`-style
  interface that forks an upstream contract.
- `TestEverySlotIsRendered` fails when a slot exists on `RightPanel` but is
  never wired into `Render()`.

See [ARQ_REFACTOR.md](ARQ_REFACTOR.md) for why the two skeletons diverged and
what the construction harness has to say about it.

## Dependencies

    webtyp/layout → webtyp/html (element builders)
    webtyp/layout → webtyp/svg  (Icon helper, *Sprite)
    webtyp/layout → webtyp/dom  (Component, Event, lifecycle)
    webtyp/layout → webtyp/css  (Stylesheet, Token)
	webtyp/layout → webtyp/router (Caller interface)

---

## Platform Lifecycle (`platformd`)

`Platform` implements `Init(ctx dom.Ctx)` + `Render() *dom.Element`. The framework calls `Init` exactly once when the component is first mounted — no `mounted` guard is needed.

The application chassis is built exclusively from the `Cover`, `Sidebar` and
`SlideDeck` layouts using the style DSL. Sizing, grid flow, mobile reflow, and
drawer navigation are defined by widget-level tokens and layout models. Route
selection and visibility are carried exclusively by reactive state attributes
(`data-current`, `data-open`) tied to `widget.Current` and `widget.Open` states,
completely eliminating legacy CSS class toggles.

`Cover` fixes the shell to the viewport height, so the frame itself never
scrolls; the active module panel carries `Scroll` and is the only element that
does. Any part rendering a bare `<svg>` declares `IconBox` — without a box an
svg falls back to 300×150 and breaks the layout. Both require `widget` v0.4.4 or
later.

### The stage is a SlideDeck, not a scroller

The module stage is **not** a horizontal scroller: panels are absolute layers
parked at `translateX(-100%)`, and the panel carrying `widget.Current` enters
sliding left→right (`SlideDeck`), each panel being an absolutely-positioned
containing block for its own content (a module's floating chrome — crudview's
Fab — resolves against ITS panel). Because the stage never scrolls, a swipe
inside a module's `MasterDetail` strip (the only horizontal snap scroller on the
page) can never chain onto it and drag the app to another section. `Activate`
writes `data-current`; the CSS transition does the rest — no `ScrollIntoView`.

```flowchart TD
A[New Platform] --> B[Render: build DOM tree with signal bindings]
B --> C[Init: create signals, register OnHashChange + OnScrollCapture]
C --> D{hash present?}
D -- yes --> E[Activate hash module]
D -- no --> F[Activate DefaultID / first module]
E & F --> G[Runtime — signals drive all UI updates]
G --> H[user clicks hamburger] --> I[menuOpen.Toggle <br/> BindAttrBool data-open patches UI]
G --> J[Notify called with a Duration] --> K[toastNodes builds desktop + mobile copies]
K --> L[notifications.Set / notificationsMobile.Set <br/> each BindChildren inserts its row]
L --> M{expires?} --> N[time.AfterFunc → dismiss]
L --> O{tap / hover / focus} --> P[dismiss, or pause/resume countdown]
G --> Q[hash changes] --> R[active.Set <br/> DeriveBool patches panel and link data-current]
G --> S[user scrolls a module container] --> T[onScroll: navStowed.Set <br/> hamburger data-open hides/reveals it]
```

### Signal fields on `Platform`

| Field | Type | Bound to |
|---|---|---|
| `active` | `*SignalString` | nav link `BindAttrBool("data-current", DeriveBool(...))`, panel `BindAttrBool("data-current", DeriveBool(...))`, header `BindText(DeriveString(...))` |
| `menuOpen` | `*SignalBool` | nav overlay `BindAttrBool("data-open", ...)`, navigation rail `BindAttrBool("data-open", ...)` |
| `notifications` | `*SignalNodes` | header `msgSlot` container via `BindChildren` |
| `notificationsMobile` | `*SignalNodes` | mobile `msg-slot-mobile` container via `BindChildren` |
| `navIcon` | `*SignalNodes` | hamburger button via `BindChildren` — the active module's glyph |
| `navStowed` | `*SignalBool` | hamburger `BindStateFunc(widget.Open, !Get())` — stowed while scrolling down |

### Activation flow (`Activate`)

`Activate(moduleID)` is the single write point for routing:
1. Early-return if `active == moduleID && !menuOpen` (no-op).
2. Check `p.CanView(moduleID)` if provided; fallback to default if denied.
3. `p.active.Set(moduleID)` — signals patch nav/panel state attributes reactively.
4. `p.menuOpen.Set(false)` — closes the mobile overlay.
5. `p.navIcon.Set(activeIcon().Render(...))` — the hamburger carries the new
   module's glyph on a phone, where no header names the section.
6. `p.lastScrollTop = 0; p.navStowed.Set(false)` — the section refresh resets the
   chrome: the new module starts from the top with the button in reach.
7. `SetHash("#" + moduleID)` — keeps the URL in sync.

### Scroll chrome (`onScroll`)

`Init` registers `OnScrollCapture` (dom's capture-phase document listener): the
`scroll` event does not bubble, and the vertical scrollers live inside other
packages' containers. `onScroll` drives the `navStowed` signal — scrolling down
past an 8px threshold stows the hamburger (`display: none` via
`RevealedBy(widget.Open)` with a negated binding); scrolling up or returning to
the top brings it back. It must be visible at rest: a module that fits without
scrolling would otherwise leave the menu unreachable.

### Notification flow (`Notify` / `dismiss`)

- `Notify(t, msg, d Duration)` appends to `rawNotifications` and calls
  `notifications.Set(toastNodes(""))` and
  `notificationsMobile.Set(toastNodes("-m"))`: every toast renders into BOTH
  slots — the header's `msg-slot` on wide screens and the mobile stack under
  the hamburger, because on a phone the header is `display:none` and a fixed
  descendant of a hidden ancestor is never painted. One `*Element` cannot have
  two parents, so the two copies carry distinct id/key suffixes.
- The duration is a typed decision, not a number: `Auto()` sizes it to the
  message (`clamp(2000ms, 1200ms + words×350ms, 8000ms)`), `Persistent()`
  keeps the toast until dismissed (the right call for errors, WCAG 2.2.1),
  `For(ms)` pins an exact window. Auto-dismiss runs on `time.AfterFunc`,
  which is why `p.mu` guards `rawNotifications`.
- A11y: `role="status"` (polite) for info/success, `role="alert"` (assertive)
  for warning/error. Tapping a toast dismisses it; hovering or focusing pauses
  the countdown (`pauseToast`/`resumeToast`) with the deadline fixed across the
  pause.
- `dismiss`: removes from `rawNotifications` (stopping a pending timer on the
  manual path), re-renders both slots → each `BindChildren` removes one keyed
  row; untouched rows keep DOM identity.

No `Update()` is called anywhere — all UI changes go through signal `Set`.

---

## Header chrome (`platformd`, v0.2.0 pass)

The header is a three-slot frame: **brand** at the leading edge, **messages**
centred (`CenterContent` on `msg-slot`), **user menu** at the trailing edge.
Parts touching the application frame — header, nav, menu, msg, and the
`rightpanel` root/header/title, the `crudview` root — are squared with
`EdgeToEdge()`; interior elements keep their radius.

- **`Platform.Brand`** (`Brand` interface: `BrandName()`, `BrandMark()`) fills the
  leading slot. An empty mark falls back to the shell's `IconBrand` glyph at the
  same box as the avatar (`IconBox(IconLg)` + `Round(RadiusFull)` +
  `HideOverflow`); a `nil` Brand renders no slot at all. `AppName` is NOT a
  fallback: it titles the phone drawer, Brand lives in the desktop header, and
  only one of the two surfaces exists at a time.
- **Notifications** are plain text in severity colour: `Glyph(Subtle|Success|
  Accent|Danger)` emits `color` + `fill: currentColor` with no background box.
  On a phone there is no header to tint, so the same toast becomes a slab —
  `As(Inset)` + radius + `Raise(Floating)` — in the `msg-stack` under the
  hamburger, the severity colours still carried by the glyph variants.
- **Navigation semantics**: the current route renders `As(Accent)` (amber —
  "where I am"), hover renders `As(Inset)` (tonal). The light selection tint
  (`Highlight`) is deliberately not used in the nav; selected rows in consumer
  lists re-point at `Accent` for the same reason.
- **On a phone the module header is gone** (`rightpanel` hides it): the only
  chrome naming the section is the hamburger, which renders the active module's
  icon (`navIcon`) instead of a fixed hamburger glyph — "you are here, and from
  here you change". The `iconMenu` glyph stays as fallback for a module that
  declares `Icon() == ""`.

---

## CRUD Layout (`crudview`)

Standard two-column layout: Form (left, 66vw) and List (right, 29vw).

### Structure

```flowchart TD
    CV[CrudView] --> L[Left Column: 66vw]
    CV --> R[Right Column: 29vw]
    L --> T[Title Band]
    T --> H[Title: h1]
    T --> CTX[Context slot: dom.Component]
    L --> F[Form: dom.Component]
    L --> B[CRUD Bar: Buttons]
    R --> LIST[List: SignalNodes]
    R --> S[Search: local filter]
```

### The `Context` slot — scope, not search

`crudview.Config.Context dom.Component` is a control rendered in the **title band**
(below the `h1`), for a context/scope selector — "which professional / which área"
(its ancestor is the legacy Pa100T professional dropdown that re-scoped the whole
screen). It rides `rightpanel.RightPanel.HeadControls`, the band that already
existed for that exact purpose.

Deliberately, `Context` is **not auto-wired to the list filter** even when it
satisfies `widget.Filterable`: `Filter` (e.g. a calendar) and `Context` (a doctor)
write into the same single-term `Presenter.Filter(term)`, so together they could
not represent "doctor X + day Y". Changing the context **re-scopes the data** — the
module that composes the `Context` calls its presenter + `CrudView.Reload()` —
while `Filter` remains the term filter. `crudview` only renders the slot; wiring
is the consumer's, by design.

### Data Flow (`view.Presenter`)

`crudview` is a pure renderer: it never talks to a `router.Caller` and holds no data of its own.
A domain module builds a `view.Presenter` (via `view.New(caller, record, listOp, newList, project,
opts...)`, importing `view`+`model`+`router`, never `layout`) and hands it to `crudview.New(Config{
Presenter: p})`.

1. `Init` (or a save/delete callback) calls `Presenter.Reload()`, which invokes the list op and
   decodes into `Presenter.Items()` synchronously.
2. `filter()` reads straight from `Presenter.Items()` (no local `allItems` copy).
3. `items` (SignalNodes) is updated, triggering a DOM patch.

Save/Delete follow the same shape: `crudview` syncs form values into `Presenter.Record()`, then
calls `Presenter.Save(record)` / `Presenter.Delete(id)` — both synchronous, both returning `error`
directly.

### Signal Fields

| Field | Type | Role |
|---|---|---|
| `items` | `*SignalNodes` | Holds the rendered card elements for the list. |
| `selected` | `*SignalString` | Holds the ID of the currently selected item. |
| `search` | `*SignalString` | Holds the current search term; triggers `filter()`. |
| `canSave` | `*SignalBool` | Controls the enabled state of the Save button. |
| `canDelete` | `*SignalBool` | Controls the enabled state of the Delete button. |

### El pegamento vive aquí

The high-level pattern for constructing a CRUD view is `crudview.New(Config)`. This constructor is the single place where the standard CRUD view↔form↔`Presenter` loop is wired:

- A module builds a `view.Presenter` (owns list/select/save/delete against a `router.Caller`) and passes it once via `Config{ParentID, Presenter}`.
- All callbacks (`OnSelect`, `OnNew`, `OnSave`, `OnDelete`, `OnCancel`) are automatically wired.
- Form inputs are automatically populated on selection using `form.LoadValues` with records returned by `Presenter.Select(id)`.
- Saves are validated and synced via `form.SyncValues` before shipping to `Presenter.Save`.
- `OnSave`/`OnDelete` are only wired when `Presenter.CanSave()`/`CanDelete()` are true.
- Empty search string placeholders default to `"Search…"`, but can be customized via `Presenter.SearchPlaceholder()`.

#### Principle: Standard-shaped tests

As a policy established in this layer (C4), no public high-level API should be published without a consumer-shaped test (like `crudview/consumer_test.go`) validating the entire integration of forms, models, and transport logic within this library.
