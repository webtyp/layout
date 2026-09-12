# webtyp/layout/platformd

Shell layout with hash-based routing, nav rail, header, and notifications.

## Usage

    p := &platformd.Platform{
        AppName:     "My App",
        IdleTimeout: 30 * 60,
        OnIdle:      logoutAndReload,
        Modules: []platformd.Module{
            {
                ID:      "home",
                Label:   "Home",
                Default: true,
                Icon:    svg.Icon("icon-home", "pd-nav-icon"),
                View:    &MyView{},
            },
        },
    }
    dom.Append("body", p)

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
