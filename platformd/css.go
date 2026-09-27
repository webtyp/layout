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
		// On a phone the header is gone and the hamburger (inside msg-stack,
		// Docked top-end) FLOATS over the module content — it does not reserve
		// space. What stows it is SCROLL and nothing else: down stows it, up
		// brings it back, and the top of a scroller always has it (see
		// onScroll in platformd.go). That is the gesture every mobile app
		// trains its users on, so it needs no explaining.
		//
		// It deliberately does NOT yield to what the user is doing inside the
		// panel. Two :has()-from-root rules used to hide it — one on
		// :focus-within, one on any [data-open="true"] control in the panel —
		// and both were wrong for the same reason: navigation is not a
		// distraction from the task, it is part of it. A user halfway through
		// an edit still needs to look something up in another module, and a
		// rule that keeps the menu gone "through the whole select→edit flow"
		// makes the app a dead end until they cancel. The focus rule was worse
		// still: a pointer click leaves focus parked inside the panel, so on a
		// desktop browser (and in a device emulator) the button vanished on
		// arrival and only came back by clicking outside the page entirely —
		// a real phone never reproduced it, because iOS does not focus on tap.
		//
		// The cost accepted here is that the floating button can overlap
		// whatever a module puts in its top-end corner. That is the normal
		// bargain for floating chrome, and scrolling clears it.
		// A slim bar, not a banner: Space1 padding and the IconMd avatar keep
		// it near 40px — the header frames the stage, it does not compete
		// with it.
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
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
		Part(widget.Part("msg"), style.KeepSize(), style.Glyph(style.Primary)).
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
		// screen; showing it again in the hover-revealed rail panel would say
		// the app's own name twice at once.
		OnlyOn(css.Mobile, widget.Part("drawer-brand"),
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
		// A tap target, not decoration: the same padded row a nav-link gets,
		// so it reads as the first item of the list it leads rather than a
		// caption above it.
		On(css.Mobile, widget.Part("brand"),
			style.Pad(style.Space2),
			style.Width(style.Full),
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
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
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
		// ── mobile-only chrome ────────────────────────────────────────────────
		// No header on a phone: the module brings its own title and the chrome
		// floats over the content. The button pins to the screen so it is
		// reachable from either page of a swipe strip.
		On(css.Mobile, widget.Part("header"),
			style.Hide(),
		).
		// AccentInverse, not Primary: the hamburger is the one control that is
		// always "current" -- it is how the user reaches every module, so it
		// wears the same amber-with-white-icon the rail's current nav-link and
		// crudview's open action already use, instead of blending into the
		// plain blue every other control defaults to.
		//
		// The position comes from the msg-stack wrapper it rides inside, not
		// from this rule: two docked boxes in the same corner would need an
		// offset calculation to stay apart. PushEnd pushes it to the end
		// edge — the same corner it always had — while the toast block
		// below widens with its messages.
		//
		// Exact 50px square from closed scales only: FontSize(TextLg) +
		// IconBox(IconLg) = 2.5em of 1.25rem — the same pair that boxes
		// every square control, replacing the old content-plus-Pad sizing
		// (40px glyph + Space2 = 56, matching nothing). The 40px nav glyphs
		// ride centered inside with room to spare.
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.AccentInverse),
			style.FontSize(style.TextLg),
			style.IconBox(style.IconLg),
			style.Round(style.RadiusSm),
			style.Raise(style.Floating),
			style.CenterContent(),
			style.PushEnd(),
			// Se guarda mientras el usuario baja. Es un estado, no una clase: lo
			// escribe Go y lo lee la hoja, y el atributo sale del propio State para
			// que marcado y selector no puedan discrepar.
			style.RevealedBy(widget.Open),
		).
		OnlyOn(css.Mobile, widget.Part("nav-overlay"),
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.RevealedBy(widget.Open),
		).
		// The toast stack is the hamburger's phone home. It docks the button
		// exactly where the button used to sit alone (top-end, Space4 from the
		// edge), then lets the toasts hang below it. One wrapper positioned
		// once beats two floating pieces that would each need an offset
		// calculation to stay apart — the plan's "justo debajo del botón"
		// falls out of the Stack gap, no arithmetic.
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
		//
		// FloatMiddle, not a top Docked: the hamburger rides the vertical
		// middle of the end edge — reachable mid-scroll, clear of both the
		// header zone and the footer chrome — with the same Space4 off the
		// edge the footer keeps off the frame. Toasts hang below it down
		// the Stack gap, as before.
		OnlyOn(css.Mobile, widget.Part("msg-stack"),
			style.Stack(style.Space2),
			style.KeepSize(),
			style.FloatMiddle(style.SideEnd, style.Space4),
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
			style.As(style.Primary),
			style.Round(style.RadiusSm),
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
		)
}

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
