# layout
<img src="docs/img/badges.svg">

Predefined UI layouts for webtyp modules

## Docs

Permanent documentation only — plans and stages live outside the README and are ephemeral.

- [Architecture](docs/ARCHITECTURE.md) — package layout, platform lifecycle, signal fields, activation and notification flows
- [Translation Dictionary](docs/DICTIONARY.md) — how consumers provide a translation dictionary for layout's UI chrome
- [Refactor arquitectónico](docs/ARQ_REFACTOR.md) — por qué `crudview` y `rightpanel` deben fusionarse, evaluado contra el construction harness
- [Reverse Engineering](docs/IE_LAYOUT.md) — Reference from pa100t shell
- [Reverse Engineering Module](docs/IE_MODULE_CONTENT.md) — Reference from pa100t crud module
- [Bug Analysis](docs/BUG_DOM.md) — resolved panic: dom `Show` on second modal open (fixed in dom v0.13.0 / components v0.4.1)
- [`api-design` skill](https://github.com/webtyp/devskills/blob/main/skills/api-design/SKILL.md) — ecosystem principles: typed, explicit APIs that fail at compile time

## Theming — surface vs tint vs inherited ink

A consumer sets the palette in one place, `config/css.go`, and may express the
primary colour either way — layouts must render correctly under both:

- **solid** — `css.Set(css.ColorPrimary, "#…")` rewrites `--color-primary`;
  every consumer (fills, icon tints, text, borders) repaints.
- **gradient** — `css.SetGradient(css.ColorPrimary, "135deg", from, to)` adds a
  companion `--color-primary-image` painted as `background-image` on **every
  element styled `style.As(style.Primary)`**. It is a *surface-only* mechanism
  by construction; each element re-origins the gradient over its own box.

So an element must pick the right primitive:

| Intent | Primitive | Emits |
|---|---|---|
| A filled surface — button, chip, card, panel, a band with its own colour | `style.As(X)` | background + text + border + radius default. Reads the gradient. |
| A tinted icon / text on the element's *own* neutral background | `style.Glyph(X)` | `color: var(--color-<x>)` (X's **base** colour). No background. |
| Text sitting on an ancestor already `style.As(X)` of the same family | *nothing* | `color` inherits `--color-on-<x>` (the reset sets no `color` on `h1..h6`). |

`style.As(X)` used only to colour text drags in a background, a border and a
radius it never wanted — invisible under a solid theme, a detached rectangle
under a gradient one. Icons are painted through `currentColor`, which is always
a solid colour: a gradient never reaches an icon, by design.

## Packages

### platformd

The platform shell, providing the header, navigation rail, and module hosting.
The stage is a slide-deck of layers (not a scroller): the active module slides
in left→right, and a swipe inside a module can never chain onto the stage. On a
phone the hamburger carries the active module's icon and stows while you scroll
down.

- `NewUIModule(id, label, iconID, view)`: helper to create modules.
- `CanView`: function field to gate module access.
- `Platform.Brand` (`BrandName()`/`BrandMark()`): optional leading header slot; empty mark falls back to the shell's default glyph.

The reference demo — `devices`, `medicalhistory` and `about` modules on this
shell — lives in its own public repo, [github.com/webtyp/app-demo](https://github.com/webtyp/app-demo),
so the demo's ORM/storage/input dependencies do not weigh on this module's
`go.mod`. Run it from that repo with the webtyp dev server. Each module owns
its view, its data and its icon; the chassis ships only its own chrome glyphs.
Adding a module to an application is a package plus one line in `p.Modules` —
no `if` in the shell.

### crudview

A CRUD controller for `rightpanel` (form left, list right) that replicates the Pa100T experience. Renders no frame of its own — it builds a `rightpanel.RightPanel`, fills its slots, and owns only the state machine.

- **Preconfigure, don't assemble**: The composition root should use the high-level constructor `crudview.New(crudview.Config)` to wire the entire form↔list↔transport cycle once per module.
- **Presenter-Based**: Takes a `view.Presenter` that handles list, selection, saving, and deleting.

```go
view, err := crudview.New(crudview.Config{
    ParentID:  "my-module",
    Presenter: myPresenter,
})
```

### login

The pre-authentication screen: an elevated card centered on a full-bleed brand
backdrop (`As(Primary)`, so each app brings its own `--color-primary`), with an
optional corner mark pinned to the viewport independently of the card's height.

It is served pre-rendered from the server using `RenderHTML()`, and the supplied `Form`
component should be configured in SSR mode (e.g. `SetSSR(true)`) to support functioning
without JavaScript.

It owns none of the form's fields or validation. `Form` is built by the
composition root exactly like `platformd.Platform.Modules` are, so the package
never assumes a shape for what it centers — only that it renders.

```go
(&login.Login{
    Title:    "Demo CMS",
    Subtitle: "Ingrese sus credenciales para continuar", // optional
    Form:     myLoginForm.Render(),
    LogoMark: crestDataURI, // optional — a URL/data-URI, not an svg.Icon
}).Render()
```

`LogoMark` mirrors `platformd.Brand.BrandMark`'s contract: a crest or seal is a
full-color image, not a `currentColor` glyph this package could recolor.

### landing

The public website layout: several URLs, each with its own metadata, built from
typed data instead of session state. The other four packages are application
shells — one screen, logged in, internal navigation; this one is a site.

Sections are composed in call order and each one names its navigation anchor
explicitly with `.At(id)` — an id derived from the title breaks silently the day
someone fixes an accent in it.

```go
site := landing.New(brand,
    landing.InfoBar(contact),
    landing.Header(menu...),
    landing.Hero(headline, tagline, cta, slides...).At("home"),
    landing.Split("Our Story", photo, paragraphs...).At("about"),
    landing.Cards("Services", cards...).At("services"),
    landing.Stats(figures...).At("commitment"),
    landing.Hours("Contact", contact, schedules...).At("contact"),
    landing.MapEmbed("Location", mapURL).At("location"),
    landing.Footer(menu...),
).WithSEO(homeDoc).WithSubPages(detailPages...)

pages := site.RenderPages() // implements html.PagesProvider: home + one page per item
```

`Split`, `MapEmbed`, `Footer` and `Hours` are horizontal bands with no state, so
they live here; everything with behaviour is consumed from `webtyp/components`
(`infobar`, `sitenav`, `herobanner`, `statgrid`, `contentcard`), never
reimplemented. Images in `Split` and `Cards` accept base image paths without variant
suffixes (e.g., `/img/foto.jpg`, not `/img/foto.M.jpg`) and emit responsive `srcset`
variants (`.S`, `.M`, `.L`) with responsive `sizes` matching layout grid breakpoints.
Detail pages keep the site chrome and must not repeat another
page's `Title` or `Description` — `RenderPages()` panics on a duplicate, because
that is how per-page ranking is lost.
