# webtyp/layout/platformd

Shell layout with hash-based routing, nav rail, header, and notifications.

The header is the primary surface at every width and carries the page's single
`<h1>`: the module's name (inside a `GroupModule`, the active child's label), large,
beside the app's mark; centred and smaller on a phone. The mark opens a panel with
the app's name and `Platform.AppVersion` (optional). Modules draw no title of their
own.

## Usage

A module is anything satisfying `platformd.UIModule` — `ModelName()` is its id
(and its hash route), the other three feed the rail and the stage:

    type Home struct{}

    func (Home) ModelName() string  { return "home" }
    func (Home) Label() string      { return "Home" }
    func (Home) Icon() svg.Icon     { return "icon-home" }
    func (Home) View() dom.Component { return &MyView{} }

    p := &platformd.Platform{
        AppName:     "My App",
        User:        currentUser,
        DefaultID:   "home",
        IdleTimeout: 30 * 60,
        OnIdle:      logoutAndReload,
        Modules:     []platformd.UIModule{Home{}},
    }
    p.Init(ctx)
    dom.Append("body", p)

## Module capabilities

Modules can optionally implement capabilities detected by `platformd`:

- **`Badged`**: a count the nav rail draws over the module's entry (`Badge() *countbadge.CountBadge`).
- **`UsesNotifier`**: the chassis hands itself (`Notifier`) over once in `Init` before any module renders (`UseNotifier(n platformd.Notifier)`).

### Help and alerts

Two capabilities exist for explaining and warning users:
- **`Described`**: a module can implement `Description() lang.Text` and optionally `Steps() []Step`. When the active module is `Described` and the description text is not empty, a «?» help button appears beside the module title. Tapping it reveals a panel with the description text. When the module also has `Steps()`, they are rendered in the panel. The panel automatically opens as a welcome card on first visit (remembered in browser localStorage). Alternatively, use `Describe(module, "Text")` to add this capability functionally.
- **`AlertedModule`**: a module can implement `UnseenAlerts() []ModuleAlert`. When the active module has unseen alerts, an alert strip appears above its content on the stage. The strip lists the alerts by subject, with optional "View" and "Mark as seen" actions.

```go
type chatModule struct {
	platformd.UIModule
	badge *countbadge.CountBadge
	notify platformd.Notifier
}

func (m *chatModule) Badge() *countbadge.CountBadge { return m.badge }
func (m *chatModule) UseNotifier(n platformd.Notifier) { m.notify = n }
func (m *chatModule) Description() lang.Text { return lang.Text("Chat interface.") }
func (m *chatModule) Steps() []platformd.Step { return nil }
```

## Fields

| Field | Type | Purpose |
|---|---|---|
| `AppName` | `string` | Drawer title on a phone when there is no `Brand`. Optional. |
| `Brand` | `Brand` | Mark + name at the header's leading slot; it is the "go home" control. Optional. |
| `User` | `Identity` | The logged-in identity. Required — header and drawer are built from it. |
| `UserActions` | `func() Component` | Factory for the app's actions in the user menu (sign-out, site picker). The theme control is built in. One instance per surface. Optional. |
| `Status` | `func() Component` | Factory for what the app says about itself now (connection state, a new version waiting), at the header right before the user menu. Optional. |
| `Modules` | `[]UIModule` | Registered modules, in nav-rail order. |
| `CanView` | `func(resource string) bool` | Filters which modules the shell presents. `nil` shows all. |
| `DefaultID` | `string` | `ModelName()` of the module shown initially. Empty = the first viewable one. |
| `IdleTimeout` | `int` | Seconds without activity anywhere on the page before `OnIdle` fires once. `0` (default) disables the idle lock. |
| `OnIdle` | `func()` | Runs when `IdleTimeout` elapses. Required when `IdleTimeout > 0` — `Init` panics otherwise. |

### The idle lock

With `IdleTimeout > 0`, `Init` registers `webtyp.com/dom.OnUserActivity` and
rearms a timer on every pulse it reports — pointer movement/press, keypress,
wheel, or scroll, anywhere on the page (`dom` listens on the document, not on
the shell's own subtree — see that function's doc comment for why an
element-level listener cannot do this reliably). When the period elapses with
no pulse, `OnIdle` fires **once** — the platform does not re-arm itself; the
next activity starts a fresh period. Detecting idleness is the shell's job;
the consequence (logout, reload) is the app's.

## Icons

Icons are SVG sprite references. Register them in your consumer's `svg.go`:

    //go:build !wasm
    package main

    import "webtyp.com/svg"
    // or consume Platform's built-in icons (icon-home, icon-products, icon-info)

The platform registers `icon-home`, `icon-products`, `icon-info` by default via `Platform.IconSvg()`.

## CSS Tokens

See `platformd/tokens.go` for the full list of CSS custom properties.
All visual customization should go through these tokens.
