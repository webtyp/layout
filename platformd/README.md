# webtyp/layout/platformd

Shell layout with hash-based routing, nav rail, header, and notifications.

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

## Fields

| Field | Type | Purpose |
|---|---|---|
| `AppName` | `string` | Drawer title on a phone when there is no `Brand`. Optional. |
| `Brand` | `Brand` | Mark + name at the header's leading slot; it is the "go home" control. Optional. |
| `User` | `Identity` | The logged-in identity. Required — header and drawer are built from it. |
| `UserActions` | `func() Component` | Factory for the header-right slot (theme toggle, etc.). One instance per surface. Optional. |
| `Modules` | `[]UIModule` | Registered modules, in nav-rail order. |
| `CanView` | `func(resource string) bool` | Filters which modules the shell presents. `nil` shows all. |
| `DefaultID` | `string` | `ModelName()` of the module shown initially. Empty = the first viewable one. |
| `IdleTimeout` | `int` | Seconds without activity inside the platform before `OnIdle` fires once. `0` (default) disables the idle lock. |
| `OnIdle` | `func()` | Runs when `IdleTimeout` elapses. Required when `IdleTimeout > 0` — `Init` panics otherwise. |

### The idle lock

With `IdleTimeout > 0`, the shell arms a timer on its root and rearms it on
every presence signal: pointer entering the shell, click or tap, key press,
focus moving inside, and document scroll (the capture-phase listener `Init`
already wires). When the period elapses with none of those, `OnIdle` fires
**once** — the platform does not re-arm itself; the next activity starts a
fresh period. Detecting idleness is the shell's job; the consequence
(logout, reload) is the app's.

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
