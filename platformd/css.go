//go:build !wasm

package platformd

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for platformd.
func (p *Platform) RenderSheet() *style.Sheet {
	return style.For(p).
		// The outermost frame of the application: locks to the viewport, stacks
		// header over body. HideOverflow() keeps the frame itself from ever
		// scrolling — a tall module scrolls inside its own panel instead.
		Root(
			style.Cover(),
			style.HideOverflow(),
			style.As(style.Page),
		).
		// On a phone the header is an integrated App Bar: it stays in the flow,
		// sits above the module content, and holds the brand/title and hamburger.
		// What stows it is SCROLL (via style.AppBar and onScroll in platformd.go):
		// down stows the entire bar off-screen, up brings it back, and the top of
		// a scroller always has it. It leaves the content 100% free of floating
		// chrome collisions while keeping navigation accessible.
		//
		// A slim bar, not a banner: Space1 padding and the IconMd avatar keep
		// it near 40px — the header frames the stage, it does not compete
		// with it.
		// The primary surface at every width: the header is the app's own chrome
		// (the module name, the user, the status), and its text reads on-primary.
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.KeepSize(),
			style.As(style.Primary),
			style.Pad(style.Space1),
			style.EdgeToEdge(),
		).
		// The brand slot mirrors the user menu at the other end of the header:
		// the mark is the avatar's exact box (IconMd, full round, clipped), the
		// name the trigger's text treatment.
		Part(widget.Part("brand"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		Part(widget.Part("brand-mark"),
			style.IconBox(style.IconMd),
			style.Round(style.RadiusFull),
			style.HideOverflow(),
			style.KeepSize(),
		).
		Part(widget.Part("brand-name"),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
		).
		// The module's name is the header's one large thing on a desktop (the size
		// the module's own title band used to have); a phone keeps its smaller,
		// centred title.
		Part(widget.Part("header-title"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		On(css.Mobile, widget.Part("header-title"),
			style.CenterContent(),
			style.Fill(),
		).
		Part(widget.Part("header-title-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
		).
		On(css.Mobile, widget.Part("header-title-text"),
			style.FontSize(style.TextBase),
		).
		// The app's mark opens a small panel with its name and version. The panel
		// hangs from it like the user menu's, with a viewport backdrop that
		// catches the tap that closes it.
		Part(widget.Part("brand-menu"),
			style.Anchor(),
			style.KeepSize(),
		).
		// Just the mark: nothing around it. No Button() recipe — its control box
		// (the shared control height) is taller than the header bar and drew a
		// pill behind the icon; the mark sits on the header like the hamburger's
		// glyph does.
		Part(widget.Part("brand-trigger"),
			style.Row(style.SpaceNone),
			style.CenterContent(),
			style.As(style.Bare),
			style.Interactive(style.Bare),
			style.KeepSize(),
		).
		Part(widget.Part("brand-panel"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
			style.Width(style.Content),
			style.Raise(style.Floating),
			style.Flyout(style.SideStart),
		).
		Part(widget.Part("brand-panel-name"),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		Part(widget.Part("brand-panel-version"),
			style.FontSize(style.TextXs),
			style.Glyph(style.Inactive),
		).
		Part(widget.Part("brand-backdrop"),
			style.Backdrop(style.Viewport),
			style.RevealedBy(widget.Open),
		).

		// Fill() here is what pushes header-right to the far edge: it grows to
		// take the free space between the two blocks. CenterContent() makes the
		// message itself read in the middle of that space instead of hard
		// against the brand — the same block, held apart at the two edges.
		Part(widget.Part("msg-slot"),
			style.Row(style.Space1),
			style.Fill(),
			style.CenterContent(),
		).
		// No per-severity tint: every notification carries the SAME identity,
		// the system's own Primary blue — never a severity color (green/amber/
		// red) and never the plain body ink. One shared color instead of one
		// per Msg type is what closes a reported bug: the mobile rule below
		// and four variant Glyph() rules used to each declare their own
		// `color:` at equal specificity, and — because @media rules are
		// grouped and emitted after the base rules regardless of declaration
		// order — the mobile declaration always won the cascade, so a success
		// toast read green on desktop and near-black on mobile, and the
		// delete-confirmation dialog (a plain, unstyled <p>) was a third color
		// again. Three places that could each drift on their own. With one
		// color and one mechanism there is nothing left to disagree.
		//
		// Two ROLES of that one color, not two colors: Glyph(Primary) here is
		// blue TEXT with no background — a slab would compete with the
		// header's own Panel surface it sits on. Plain ColorOnSurface (body
		// ink) was tried first and read as flat against the header's grey; the
		// system's own blue reads as "this is the platform talking," not
		// ambient chrome text.
		//
		// role="status"/"alert" (see toastNodes in platformd.go) is what
		// carries severity to assistive tech; it never depended on color, so
		// dropping the four-way tint costs sighted users nothing they could
		// already rely on (a plain-text toast never had an icon to pair the
		// color with anyway — WCAG 1.4.1 wants a second channel for a
		// color-coded signal, and there wasn't one).
		Part(widget.Part("msg"),
			style.As(style.Inverse),
			style.Round(style.RadiusSm),
			style.PadInline(style.Space3),
			style.KeepSize(),
		).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		// The rail sits at the inline-end edge; the stage takes everything else.
		// Below the stage's minimum width the two reflow into one column with no
		// media query — that is Sidebar's own behaviour, not something to add.
		Part(widget.Part("body"),
			style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone),
			style.Fill(),
		).
		// Capas, no tira: los paneles se apilan y el activo entra deslizándose desde
		// el borde inicial. RevealedBy conmuta `display`, que es discreta y no puede
		// transicionar, de ahí que el movimiento venga de un transform.
		//
		// Esto NO es un scroller, y esa es la razón de ser del cambio: cuando lo era,
		// el scroll-snap horizontal de cada módulo (rightpanel MasterDetail, en móvil)
		// encadenaba con este al llegar a su extremo y un gesto dentro del contenido
		// arrastraba la aplicación al módulo siguiente. Un solo eje horizontal
		// desplazable por página.
		//
		// Todos los módulos siguen montados; el estado Current decide cuál se ve.
		Part(widget.Part("stage"),
			style.SlideDeck(style.MotionBase),
			style.Fill(),
		).
		// Scroll(), no Fill(): un módulo más alto que el viewport tiene que
		// desplazarse dentro de su propia capa. Scroll() es Fill() más overflow-y.
		//
		// Sin Anchor(): SlideDeck ya posiciona cada capa en absoluto, lo que la
		// convierte en el bloque contenedor de su contenido — que es lo que el cromo
		// flotante de un módulo necesita para resolverse contra SU panel. Añadir
		// Anchor() aquí lo ROMPE: emite position:relative en @layer widgets, que gana
		// sobre el position:absolute que el flujo emite en @layer primitives, y las
		// capas volverían al flujo apiladas una debajo de otra.
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Scroll(),
		).
		// Grow() for its min-width:0 — without it the rail is sized by whatever
		// it contains, so revealing the labels on hover widened it and pushed
		// the stage. Its width is the Sidebar's rail token and nothing else.
		// EdgeToEdge: the rail is welded to the window frame on its right and
		// bottom, where a radius leaves two background slivers.
		//
		// Primary, not Secondary: the rail IS the primary surface. An available
		// route paints nothing of its own and reads white (ColorOnPrimary,
		// inherited) directly on this fill; only the current route commits to a
		// distinct block (AccentInverse). Under css.SetGradient the whole rail
		// carries the gradient as ONE surface — no per-item box, no seams
		// between tiles. EdgeToEdge keeps it welded and square to the window
		// frame; the mobile drawer below and the hover float both restate
		// As(Primary) so the chrome reads as one colour everywhere.
		//
		// GradientAngle: the rail sits at the inline-end, where a panel frame's
		// own gradient has run toward its light end. Re-angling the rail's copy
		// to 315° (the mirror of the common 135° theme) puts the light stop at
		// its leading edge too, so the two independently-painted gradients meet
		// light-on-light and the join reads softer. Inert until an app sets a
		// gradient — see style.GradientAngle.
		Part(widget.Part("menu"),
			style.Anchor(),
			style.Stack(style.Space1),
			style.As(style.Primary),
			style.GradientAngle("315deg"),
			style.Fill(),
			style.Grow(),
			style.EdgeToEdge(),
		).
		// The drawer's copy of the identity exists only where there is no header
		// to hold one. Two visible at once is the redundancy this whole change
		// set out to remove.
		OnlyOn(css.Mobile, widget.Part("drawer-identity"),
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
		// Text with no icon beside it: it can only exist where the panel opens
		// wholesale, which is the phone drawer. No hover rule reveals it on a
		// wide screen, on purpose.
		OnlyOn(css.Mobile, widget.Part("app-name"),
			style.Row(style.SpaceNone),
			style.Pad(style.Space2),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		// The drawer's copy of the brand exists only where there is no header
		// to hold one — same reasoning as drawer-identity above. On a wide
		// screen the header's own brand, top-left, is already permanently on
		OnlyOn(css.Mobile, widget.Part("drawer-brand"),
			style.Row(style.SpaceNone),
			style.Pad(style.Space2),
			style.KeepSize(),
		).
		// In the mobile header, brand has no extra padding, keeping it exactly 24x24px
		// to match hamburger (24x24px) for symmetric centering of the title.
		On(css.Mobile, widget.Part("brand"),
			style.Pad(style.SpaceNone),
		).
		// Pad(Space1): the inset that, with navbar's own Space1 stack gap, gives
		// every nav item the SAME clearance on all four sides instead of items
		// welded to the rail edge. Mobile drops it back to flush (below) so the
		// drawer rows keep their edge-to-edge DividerBelow hairline.
		Part(widget.Part("drawer-panel"),
			style.Stack(style.Space1),
			style.Pad(style.Space1),
			style.Fill(),
		).
		// Space1 gap between items (was SpaceNone): on desktop the items read as
		// a spaced list of rounded chips. Mobile restores SpaceNone (below) so
		// rows sit flush and the DividerBelow line is their separator.
		Part(widget.Part("navbar"),
			style.Stack(style.Space1),
			style.Fill(),
		).
		On(css.Mobile, widget.Part("navbar"), style.Stack(style.SpaceNone)).
		On(css.Mobile, widget.Part("drawer-panel"), style.Pad(style.SpaceNone)).
		// No Fill: spreading the items down the rail makes every row resize the
		// moment the panel floats out and stops being full height. A rail packed
		// from the top measures the same collapsed and expanded.
		Part(widget.Part("nav-item"),
			style.Stack(style.SpaceNone),
			style.KeepSize(),
		).
		Part(widget.Part("submenu"),
			style.Stack(style.Space1),
			style.Pad(style.Space1),
			style.Round(style.RadiusSm),
			style.As(style.Secondary),
			style.Fill(),
			style.Hide(),
		).
		OnlyOn(css.Mobile, widget.Part("submenu"),
			style.Stack(style.SpaceNone),
			style.Pad(style.SpaceNone),
			style.Round(style.RadiusNone),
			style.As(style.Secondary),
			style.RevealedBy(widget.Open),
		).
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("submenu"),
			style.Stack(style.Space1),
			style.Pad(style.Space1),
			style.Round(style.RadiusSm),
			style.As(style.Secondary),
			style.Fill(),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("nav-parent"),
			style.Anchor(),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Width(style.Full),
			style.CenterContent(),
			style.Round(style.RadiusSm),
			style.Animate(style.MotionFast),
		).
		On(css.Mobile, widget.Part("nav-parent"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
			style.Round(style.RadiusNone),
			style.DividerBelow(),
			style.As(style.Primary),
		).
		When(widget.Current, widget.Part("nav-parent"),
			style.As(style.AccentInverse),
		).
		Cue(widget.Hover, widget.Part("nav-parent"),
			style.As(style.AccentHover),
		).
		Cue(widget.Focus, widget.Part("nav-parent"),
			style.As(style.AccentHover),
		).
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("nav-parent"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
			style.As(style.Primary),
		).
		WhenWithin(widget.Open, widget.Part("menu"), widget.Part("nav-parent"),
			style.As(style.Primary),
		).
		Part(widget.Part("nav-group-icon"),
			style.IconBox(style.IconLg),
		).
		On(css.Mobile, widget.Part("nav-group-icon"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("nav-active-icon"),
			style.IconBox(style.IconLg),
			style.Hide(),
		).
		Part(widget.Part("subnav-item"),
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
		Part(widget.Part("subnav-link"),
			style.Anchor(),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.PadInline(style.Space2),
			style.Width(style.Full),
			style.CenterContent(),
			style.Round(style.RadiusSm),
			style.FontSize(style.TextSm),
			style.Animate(style.MotionFast),
		).
		On(css.Mobile, widget.Part("subnav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.PadInline(style.Space4),
			style.StartContent(),
			style.Round(style.RadiusNone),
			style.FontSize(style.TextSm),
			style.DividerBelow(),
		).
		When(widget.Current, widget.Part("subnav-link"),
			style.As(style.AccentInverse),
		).
		Cue(widget.Hover, widget.Part("subnav-link"),
			style.As(style.AccentHover),
		).
		Cue(widget.Focus, widget.Part("subnav-link"),
			style.As(style.AccentHover),
		).
		Part(widget.Part("subnav-icon"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("chevron"),
			style.IconBox(style.IconSm),
			style.PushEnd(),
			style.Rotate(style.TurnNone),
			style.Animate(style.MotionFast),
			style.Hide(),
		).
		OnlyOn(css.Mobile, widget.Part("chevron"), style.Row(style.SpaceNone)).
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("chevron"), style.Row(style.SpaceNone)).
		WhenWithin(widget.Open, widget.Part("nav-parent"), widget.Part("chevron"), style.Rotate(style.TurnHalf)).
		// An available route paints NOTHING of its own: transparent box, white

		// icon and label inherited (ColorOnPrimary) from the Primary rail it
		// sits on. So the rail reads as one continuous Primary surface — under
		// css.SetGradient, one gradient — instead of a column of per-item
		// boxes with seams between them. The current route (When(widget.Current)
		// → AccentInverse, below) is the only filled block; hover/focus
		// (AccentHover) is the only other fill. Both are derived surfaces that
		// emit `background-image: none` (widget/style emit_decls), so the rail's
		// gradient cannot bleed through and overpaint their amber.
		// Round(RadiusSm): every item — available, current, hovered — carries the
		// SAME rounded shape always. Hover/focus/current change only the fill
		// COLOUR, never the geometry; the items are inset (drawer-panel Pad) so
		// a rounded corner never butts the rail's welded edge. Mobile drops the
		// radius (below): there the rows are flush with a DividerBelow seam.
		Part(widget.Part("nav-link"),
			style.Anchor(),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Width(style.Full),
			style.CenterContent(),
			style.Round(style.RadiusSm),
			style.Animate(style.MotionFast),
		).
		// A bare <svg> with no box falls back to 300x150 and wrecks the rail.
		Part(widget.Part("nav-icon"),
			style.IconBox(style.IconLg),
		).
		// The rail is icon-only: at RailNarrow a label does not fit, and forcing
		// it in is what widened the rail past its token. The mobile drawer is
		// two thirds of the viewport, so there the label rides along.
		// Row() is not decoration here: OnlyOn hides the part outside the device
		// and only a flow puts a display back on it inside, so a rule carrying
		// nothing but FontSize would stay hidden everywhere.
		// Left-aligned on a phone too: the drawer shows icon and label together,
		// which is the same row the expanded rail draws, so it aligns the same
		// way. Centring is only right for a lone icon.
		// DividerBelow: on a phone the drawer rows are flush (navbar is
		// Stack(SpaceNone)), so a hairline under each item is what separates
		// them — a cleaner, more finished read than a plain colour block.
		On(css.Mobile, widget.Part("nav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
			style.Round(style.RadiusNone),
			style.DividerBelow(),
		).
		OnlyOn(css.Mobile, widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		// The active route reads as "current", the same vocabulary the rail and
		// crudview's list rows share. It is a STATE, never a class.
		// The route you are on is the one filled block in the rail; its icon
		// rides the filled surface through currentColor. AccentInverse keeps
		// the icon white against the fully committed amber fill — the same
		// "current" language as the mobile hamburger button and crudview's
		// open action.
		//
		// No Round() here: AccentInverse's own default radius-sm is exactly what
		// the base nav-link now carries, so "current" is the SAME rounded chip
		// as every other item, only amber. The items are inset (drawer-panel
		// Pad), so a rounded corner never butts the rail's welded edge.
		When(widget.Current, widget.Part("nav-link"),
			style.As(style.AccentInverse),
		).
		// El control entero se ilumina, no solo el glifo: el hover habla del
		// mismo ámbar al que apunta. AccentHover, no AccentWash: AccentWash
		// diluye el ámbar al 15% -- casi el blanco de la página -- y un ícono
		// blanco encima deja de leerse. AccentHover lo diluye solo un 30%, lo
		// bastante fuerte para que el mismo blanco de AccentInverse siga
		// siendo legible, y lo bastante mas suave que el ámbar sólido de
		// Current para que las dos etapas se distingan por intensidad en vez
		// de por color de ícono.
		//
		// AccentHover no lleva borde, así que el ítem mide lo mismo con y sin
		// puntero encima sin necesitar el truco de outline que Inset requería
		// -- el rail mide exactamente --rail-narrow y el panel flotante se
		// dimensiona con width: max-content: un cambio de tamaño aquí saca el
		// ítem del puntero que lo activó, lo pierde, encoge, y el menú entero
		// entra en un bucle de parpadeo.
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.AccentHover),
		).
		// Paired with Hover above on purpose: a keyboard user tabbing through
		// the rail gets no :hover at all, so leaving Focus on its old default
		// would strand keyboard navigation on a different, unrelated color
		// from what a mouse hover now shows.
		Cue(widget.Focus, widget.Part("nav-link"),
			style.As(style.AccentHover),
		).
		// Hovering the rail floats the drawer-panel out over the content at label
		// width. The panel leaves the flow, so the rail's box — already pinned
		// to the Sidebar's rail token by Grow's min-width:0 — cannot change, and
		// nothing beside it moves.
		// CueWithinHover, not CueWithin: the rule lives inside
		// `@media (hover: hover)`. A touch tap fires :hover but never the
		// fine-pointer capability, so the floating panel cannot activate on a
		// phone and cannot duplicate inside the Drawer — no JS involved.
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("drawer-panel"),
			style.Docked(style.Parent, style.EdgeTop, style.SideEnd, style.SpaceNone),
			style.Width(style.Content),
			style.As(style.Primary),
			style.Raise(style.Floating),
			style.Pad(style.Space1),
			// RadiusSm, matching the nav items inside it — the float and its
			// chips share one corner radius instead of a RadiusMd shell around
			// RadiusSm children.
			style.Round(style.RadiusSm),
			style.Stack(style.Space1),
		).
		// The labels only exist while the rail is expanded — or on a phone,
		// where the drawer is two thirds of the viewport and has room for them.
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		// Left-aligned once expanded: centred is right for a lone icon in a
		// narrow rail, wrong for an icon-and-label row where the labels have to
		// start on the same line as each other.
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("nav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
		).
		CueWithinHover(widget.Hover, widget.Part("menu"), widget.Part("subnav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.PadInline(style.Space3),
			style.StartContent(),
		).

		// ── mobile-only chrome ────────────────────────────────────────────────
		// Mobile top bar: the platform header is an integrated App Bar that stays
		// in the flow, holds the brand/app-name and hamburger, and smoothly slides
		// off-screen when the user scrolls down.
		On(css.Mobile, widget.Part("header"),
			style.AppBar(widget.Open, style.MotionBase),
		).
		// Off a phone the mark and the user menu must not sit against the frame
		// edge: the phone's bar is padded by its AppBar, a desktop's is not.
		// The header keeps its 56px; its CONTENT moves. The frame below is the
		// same colour and has a gutter on top, so the band the eye sees is
		// header + gutter: the content is seated in the lower part of the header
		// (top pad = gutter, no bottom pad) so it centres in that whole band
		// instead of sitting half a gutter high.
		On(css.Tablet, widget.Part("header"),
			style.Pad(style.SpaceNone),
			style.PadEdge(style.EdgeTop, style.Space2),
			style.PadInline(style.Space3),
		).
		On(css.Desktop, widget.Part("header"),
			style.Pad(style.SpaceNone),
			style.PadEdge(style.EdgeTop, style.Space2),
			style.PadInline(style.Space3),
		).
		On(css.Mobile, widget.Part("header-right"),
			style.Hide(),
		).
		On(css.Mobile, widget.Part("msg-slot"),
			style.Hide(),
		).
		// AccentInverse, not Primary: the hamburger is the one control that is
		// always "current" -- it is how the user reaches every module, so it
		// wears the same amber-with-white-icon the rail's current nav-link and
		// crudview's open action already use. Inside the header it rides as part
		// of the App Bar with no floating shadow or fixed viewport pin.
		//
		// Exact 24px square (IconMd = 1.5em): perfectly fits inside the 48px
		// App Bar with 12px vertical clearance matching the brand mark.
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.SpaceNone),
			style.As(style.Bare),
			style.IconBox(style.IconMd),
			style.CenterContent(),
			style.KeepSize(),
		).
		On(css.Mobile, widget.Part("nav-icon"),
			style.IconBox(style.IconMd),
		).
		OnlyOn(css.Mobile, widget.Part("nav-overlay"),
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.RevealedBy(widget.Open),
		).
		// The toast stack is the hamburger's phone home. It docks the button
		// in the top-end corner (Space3 from the edges) and lets toasts hang
		// below it.
		//
		// Stack(Space2): the wrapper's own gap lands between the button and
		// the toast block; the gap BETWEEN toasts is the slot's own Stack
		// (declared below), so both seams read the same.
		//
		// No Raise here: the wrapper is transparent, and a box-shadow follows
		// the border box even when nothing is painted inside it — a shadow
		// hugging the rectangle around button+toasts would be a visible seam
		// around an invisible box. Each toast raises itself instead.
		//
		// Stacking: a Viewport pin claims the widget's layer (LayerDropdown,
		// the same --z-dropdown the drawer and the overlay ride), so the stack
		// ties with them and DOM order breaks the tie — which is why
		// msg-stack is the ROOT'S LAST CHILD in Render(): a toast must paint
		// above the drawer's overlay when both are on screen. (The --z-toast
		// layer exists for widgets whose Kind is Alert; platformd is a Menu,
		// and switching kinds to steal the layer would change its role.)
		OnlyOn(css.Mobile, widget.Part("msg-stack"),
			style.Stack(style.Space2),
			style.KeepSize(),
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space3),
			style.DockedClear(style.Space12),
		).
		OnlyOn(css.Mobile, widget.Part("msg-slot-mobile"),
			style.Stack(style.Space2),
			style.Width(style.Content),
			style.KeepSize(),
			// Width(Content) sizes the slot to its widest toast's max-content;
			// the toasts stretch to that width, so the stack reads as one
			// aligned column hugging the end edge. A long message wraps inside
			// its box instead of overflowing the screen — the DSL has no
			// max-width, so the honest behaviour is max-content: the box is as
			// wide as the widest word and the text wraps below it.
		).
		// The msg part is shared with the DESKTOP header slot (blue text on
		// the header's own Panel), so this is On, not OnlyOn: OnlyOn's
		// display:none off-device would switch the desktop toasts off too.
		//
		// On a phone the toast floats over CONTENT, not chrome — the rows,
		// cards and fields around it are all Page-white now (see targetlist
		// and crudview's own mobile rules), so blue TEXT with no fill (the
		// desktop treatment above) would read as one more line of pale ink
		// lost against that white, exactly the "se pierde" this replaced.
		// A filled Primary chip — blue fill, white text (ColorOnPrimary) —
		// is what actually separates "this is a message" from "this is
		// content": the same commit-to-color-not-tint move this chassis
		// already makes for every other floating mobile chip (the hamburger's
		// AccentInverse two rules up, crudview's own mobile action button),
		// none of which try to blend into a neutral panel. Primary's own
		// triplet carries no border, so the chip is a clean fill with no
		// outline to clash with Raise's shadow.
		On(css.Mobile, widget.Part("msg"),
			style.Raise(style.Floating),
			style.Pad(style.Space2),
		).
		// On a phone the rail stops being a column and becomes a panel that
		// slides in from the inline-end edge, gated by the same Open state as
		// the overlay. Drawer(..., MotionSlow): it parks off-screen on a
		// transform and RevealedBy transitions it in AND back out, so closing
		// is the same slide as opening — not a hard cut.
		// As(Primary), like the desktop rail: the drawer is the same chrome
		// surface. The nav rows are already flush (navbar is Stack(SpaceNone)),
		// so the per-item DividerBelow hairline is their separator; this Stack
		// only spaces brand → navbar → identity.
		On(css.Mobile, widget.Part("menu"),
			style.Drawer(style.SideEnd, style.TwoThirds, style.MotionSlow),
			style.Stack(style.Space1),
			style.As(style.Primary),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("help-btn"),
			style.Button(style.Secondary),
		).
		Part(widget.Part("help-panel"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Pad(style.Space3),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("welcome-steps"),
			style.Stack(style.Space2),
			style.PadEdge(style.EdgeTop, style.Space2),
			style.PadEdge(style.EdgeBottom, style.Space2),
		).
		Part(widget.Part("welcome-step"),
			style.Row(style.Space2),
			style.CenterContent(),
		).
		Part(widget.Part("welcome-step-done"),
			style.Glyph(style.Subtle),
		).
		Part(widget.Part("welcome-step-check"),
			style.Glyph(style.Success),
		).
		Part(widget.Part("welcome-dismiss"),
			style.As(style.Primary),
			style.Round(style.RadiusSm),
			style.PadInline(style.Space3),
			style.ControlBox(),
		).
		Part(widget.Part("alert-strip"),
			style.Stack(style.Space1),
		).
		Part(widget.Part("alert-row"),
			style.Row(style.Space2),
			style.As(style.Inverse),
			style.Round(style.RadiusSm),
			style.PadInline(style.Space3),
		).
		Part(widget.Part("alert-subject"),
			style.Fill(),
		).
		Part(widget.Part("alert-action"),
			style.Row(style.SpaceNone),
			style.CenterContent(),
			style.As(style.Bare),
			style.Interactive(style.Bare),
			style.KeepSize(),
		)
}

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
