//go:build !wasm

package rightpanel

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// frameGutter is Root's own rhythm: the blue frame that separates the
// two-column content from the platform chrome around it (header, menu,
// footer) on Tablet/Desktop, and the seam of the Split between article and
// aside — the file's Pad and Split calls were always meant to agree with
// each other, so they share this one constant. It is deliberately NOT gutter
// below: the outer frame and the cards living inside it are two different
// tiers, and forcing them to the same number one Space step apart is what
// had made the whole screen read as flat rather than as a frame with panels
// in it.
const frameGutter = style.Space2

// gutter is the inset every panel and card below the frame shares: article's
// and aside's own Pad, and the rhythm inside aside (header→list→footer). One
// constant instead of one-off Space values per Part is what keeps THESE from
// drifting apart from each other — it does not extend to frameGutter above,
// which is a coarser, outer-tier measurement on purpose.
const gutter = style.Space1

// RenderSheet returns the style Sheet containing the rules for rightpanel.
func (r *RightPanel) RenderSheet() *style.Sheet {
	return style.For(r).
		// Pad is what turns the primary surface into a visible frame: without it
		// the content and the aside sit flush against the panel edges and the
		// module reads as stacked rectangles instead of one view.
		// One gutter everywhere: the frame's pad is the Split's gap, so the
		// margin touching the four sides measures the same as the seam between
		// the two columns.
		Root(
			style.Split(style.SplitTwoThirds, frameGutter),
			style.Fill(),
			style.As(style.Primary),
			style.Pad(frameGutter),
			style.EdgeToEdge(),
		).
		Part(partMain,
			style.Stack(style.Space2),
			style.Fill(),
		).
		// The heading needs its own indent: it sits directly on the primary
		// surface with no card of its own to inset it. PadInline, not Pad —
		// the indent is horizontal; vertically the band answers to
		// --control-height, the same token every control measures by, so the
		// frame reads as one rhythm.
		Part(partHeader,
			style.Row(style.Space1),
			style.PadInline(style.Space2),
			style.ControlBox(),
			style.KeepSize(),
		).
		// No Center(): the header's own PadInline above is the indent, and it
		// applies to BOTH rows of the header. Center() added a second, rival
		// one — a 65ch cap with auto margins — that only engages once the row
		// is wider than the measure. So the title's alignment depended on
		// whether the panel happened to have an aside: with one, main is ~2/3
		// of the frame, narrower than 65ch, and the cap was inert; without
		// one, main takes the full width, the cap engaged, and the title
		// floated ~130px in while HeadControls right below it stayed flush at
		// the padding. One header, two width policies, and the visible result
		// changed with an unrelated structural fact.
		//
		// A title row is not prose either: the readable measure is what a LINE
		// of running text wants, and a heading beside its controls is a band.
		Part(partTitleRow,
			style.Row(style.Space2),
		).
		// The heading is TEXT on Root's Primary surface — it takes only the
		// inherited on-primary ink (Root is style.As(style.Primary), and color
		// inherits through the reset, which sets no color on h1..h6). It must
		// NOT carry style.As() of its own: As() is background + border + radius
		// as well as text, and under a gradient Primary theme (css.SetGradient)
		// that fill re-origins the gradient across the heading's own box and
		// paints a detached rectangle over the panel. FontSize is explicit
		// because the reset forces h1 font-size: inherit.
		Part(partTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(partHeadControls,
			style.Row(style.Space1),
		).
		// Secondary, not Page: aside's own backdrop is Secondary/Panel's
		// ColorSurface at every breakpoint (see the "aside" Part below), so on
		// mobile — where neither panel is overridden to Panel yet — article
		// must share that same backdrop. Otherwise fields' Inset border sits on
		// Page's white and list's identical Inset border sits on Secondary's
		// grey, and the same border color reads as two different tones purely
		// from contrast against two different backgrounds.
		// EdgeToEdge for the same reason aside needs it below: Secondary
		// defaults to RadiusSm, and on mobile article lands flush against the
		// physical screen edge — a rounded corner there meets a square screen
		// corner instead of framing anything.
		Part(partArticle,
			style.As(style.Secondary),
			style.Pad(gutter),
			style.Scroll(),
			style.Fill(),
			style.EdgeToEdge(),
		).
		// Borderless by default: on mobile the aside fills the viewport edge to
		// edge (Root's Pad drops to SpaceNone below), so Panel's border and
		// radius-md would sit flush against the physical screen edge AND stack
		// with the list's own Inset border a few pixels in — two parallel lines
		// where one is expected. Secondary carries Panel's exact background and
		// text but no border; EdgeToEdge zeroes the radius Secondary would
		// otherwise default to, so mobile gets a plain flush fill and the list
		// card is the only border on screen. Tablet and Desktop restore Panel
		// below, where Root's padding already keeps the aside off the frame's
		// edge and the border reads as a proper card.
		Part(partAside,
			style.Stack(gutter),
			style.As(style.Secondary),
			style.Pad(gutter),
			style.Fill(),
			style.EdgeToEdge(),
		).
		// The controls band carries no surface, padding or radius of its own:
		// whatever the consumer puts here — a search bar, a calendar, a select —
		// brings its own. A second frame around a control that already has one
		// reads as a box inside a box. All the band owes it is a refusal to be
		// squashed when the content below grows, which is what KeepSize buys.
		Part(partAsideHeader,
			style.KeepSize(),
		).
		Part(partAsideContent,
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(partAsideFooter,
			style.Row(style.Space1),
			style.KeepSize(),
		).
		// Floating chrome keeps one step off the frame on every edge it
		// meets: Space3 here, plus the aside's own Space1 gutter, lands the
		// footer exactly Space4 (16px) off the frame — the same Space4 the
		// hamburger's msg-stack keeps. Same criterion, same total, both
		// corners; flush (SpaceNone) is banned for floating chrome, where
		// elements get lost and tap targets die at the edge.
		On(css.Mobile, partAsideFooter, style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space3)).
		// Restores the aside's card border once there is room to show it — see
		// the "aside" Part above for why it starts borderless.
		On(css.Tablet, partAside, style.As(style.Panel)).
		On(css.Desktop, partAside, style.As(style.Panel)).
		// Article stays square on mobile for the same reason aside does: it is
		// the panel that lands flush against the screen edge once the strip
		// scrolls to it. From Tablet up it sits inset inside Root's frame same
		// as aside, so it takes the same Panel treatment aside does — Round()
		// alone would have rounded the corners but left Page's zero-width
		// border, so the two cards would still disagree on the one line that
		// actually separates them from the frame.
		On(css.Tablet, partArticle, style.As(style.Panel)).
		On(css.Desktop, partArticle, style.As(style.Panel)).
		// gutter (not SpaceNone) is kept here on purpose, on every breakpoint
		// including mobile: this is the buffer between crudview__list's own
		// Inset border and the physical screen edge. Without it the card's
		// border sits exactly flush with the aside's edge — border and edge
		// coincide, so the border reads as invisible with nothing to frame
		// against, exactly the article/fields side's gutter is what keeps ITS
		// border legible. The ⋮ menu's mobile clearance (the reason this was
		// ever tried at SpaceNone) is now reclaimed instead from targetlist's
		// own PartList padding — inside the card, not at its outer edge — see
		// components/targetlist/css.go.
		// On a phone the desktop Split becomes a horizontal scroll-snap strip:
		// the aside is what shows on arrival, and selecting an item slides the
		// main panel in from the left, leaving a sliver of the aside on the right
		// so it is obvious where you came from.
		// Pad(SpaceNone) is part of the contract, not a detail: the panels are
		// sized as a share of the scroll container, and any padding on it makes
		// each panel that much narrower than the window, so a strip of the
		// neighbour shows through at rest.
		On(css.Mobile, "",
			style.MasterDetail(style.Most),
			style.Pad(style.SpaceNone),
		).
		// En móvil no hay cabecera: el chasis lleva el nombre de la sección en su
		// botón de menú, que es el único cromo que sobrevive ahí. Una chapa flotante
		// con el título repetía ese dato encima del contenido y tapaba la barra de
		// búsqueda del aside.
		//
		// Se apaga la cabecera entera y no solo el <h1>: ControlBox le fija una
		// altura mínima de --control-height, así que ocultar únicamente el título
		// dejaría una banda vacía de 50px en lo alto de cada módulo.
		//
		// HeadControls cae con ella. Es aceptable porque hoy no lo usa nadie en el
		// repositorio salvo los tests; un control que deba sobrevivir en móvil va en
		// AsideControls, que es la banda que sí se ve ahí.
		On(css.Mobile, partHeader,
			style.Hide(),
		)
}

// RenderCSS implements visual contract for rightpanel layout.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return r.RenderSheet().Stylesheet()
}
