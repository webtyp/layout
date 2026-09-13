//go:build !wasm

package crudview

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// cardInset is the gap between a card's own border and its content, on the
// axis where both panels agree: the inline sides of fields and list. The
// fields card takes a larger BLOCK inset (formGap) because the form inside it
// carries legend chips that seat half a chip-height above each input — the
// list has no such overhang, so its block inset stays cardInset too.
const cardInset = style.Space1

// formGap is the one spacing value inside the fields card: the block inset from
// the card border, AND the gap between consecutive fields (fieldset puts it on
// the <form> container). Every visible gap in the form — card edge to first
// chip, card edge to the sides and bottom of the inputs, one field to the next
// — lands on this value.
const formGap = style.Space3

// RenderSheet returns the style Sheet containing the rules for crudview.
func (v *CrudView) RenderSheet() *style.Sheet {
	return style.For(v).
		// crudview paints no frame: rightpanel owns the root, the columns and
		// the mobile strip. What remains here are the widgets this controller
		// puts INTO those slots.
		// Fill: the fields are the article's whole content now — rp__article is
		// a plain block (page surface, pad, scroll), not a flex column, so a
		// content-height child would collapse to the form's natural height and
		// leave the column floating at the top of the card.
		Part(widget.Part("fields"),
			style.As(style.Inset),
			style.Pad(formGap),
			style.PadInline(cardInset),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
		).
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Pad(cardInset),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
		). // The footer is a Row of equal buttons: delete leads, add follows
		// (see Render). Every button answers to --control-height — the same
		// token the search bar measures by — so the two ends of the column
		// agree by construction instead of by hand-tuned padding. Grow, not
		// Full: each button takes its share of the free space and yields to
		// its siblings. Glyph IconMd + Pad Space2 = 40px of content, so the
		// ControlBox floor lands the bar at exactly 50px on every
		// breakpoint — no mobile bump (the old IconLg overrides inflated the
		// bar to 64 on phones, matching nothing).
		Part(widget.Part("action"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Grow(),
			style.ControlBox(),
			style.CenterContent(),
		).
		// Same box as action, same surface: delete reads as the footer's
		// leading button, not as an alarm. No MediaBox: a square chip next
		// to a growing bar can never measure equal. Anchor: the button is
		// the positioning reference for its countbadge bubble, which rides
		// the top-end corner out of the flow — relative costs nothing visual.
		//
		// Row(SpaceNone) is not cosmetic: this part carries RevealedBy(Open),
		// so its base rule ends `display:none` and the @layer states reveal
		// re-emits only `display: <displayFor(flowType)>`. Without a flow that
		// is `block`, and CenterContent()'s flex centring goes inert on the
		// shown button — the glyph strands at the leading edge. A flow makes
		// the reveal restore `display:flex`, so CenterContent applies in every
		// mode. Same fix, same reason as components/listselect/css.go:23-31.
		// (action, above, needs no flow: it is never display:none, so its
		// CenterContent flex always stands.)
		Part(widget.Part("action-delete"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Grow(),
			style.ControlBox(),
			style.Row(style.SpaceNone),
			style.CenterContent(),
			style.Anchor(),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("action-edit"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Grow(),
			style.ControlBox(),
			style.Row(style.SpaceNone),
			style.CenterContent(),
			style.Anchor(),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("footer"),
			style.Row(style.Space1),
			style.Fill(),
			style.Width(style.Full),
		).
		// No As(): action-new is the icon INSIDE the Primary "action" button, not
		// a surface of its own. As(Primary) gave it a second Primary box —
		// background + border-radius + (under css.SetGradient) its own gradient
		// slice — a visible square over the button. It inherits the button's
		// white currentColor, exactly like action-cancel below.
		Part(widget.Part("action-new"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("action-cancel"),
			style.RevealedBy(widget.Open),
			style.IconBox(style.IconMd),
		).
		// Boxed like action-new/cancel above: a bare <svg> paints at its
		// intrinsic size and dwarfs the button. Same token, so all three
		// footer glyphs measure equal on every breakpoint.
		Part(widget.Part("action-delete-icon"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("action-edit-icon"),
			style.IconBox(style.IconMd),
		).
		// On a phone the footer travels as ONE row: rightpanel already docks
		// the aside-footer slot (see rightpanel/css.go), so docking a single
		// button here as well would rip it out of the row and strand its
		// siblings at full width. No per-button Docked, on any breakpoint:
		// every footer button keeps the same Grow box it has on desktop.
		// ControlBox's min-height stays on everywhere, so the row keeps the
		// control token as its measure instead of hugging the glyphs. The
		// glyphs stay IconMd on every breakpoint too: the old mobile IconLg
		// bump inflated the bar to 64px on phones, matching nothing.
		//
		// Exact 50px squares on mobile from the same pair as every square
		// control (FontSize TextLg + IconBox Lg = 2.5em of 1.25rem): the
		// docked slot is shrink-to-fit, so Grow has no free space to share
		// and the buttons would render content-sized (40px) — the IconBox
		// owns the 50 instead. Desktop keeps the Grow bars (static parent,
		// height already 50 via ControlBox). border-box keeps Pad inside
		// the 50; the Md glyphs ride 30px inside on the buttons' 20px type.
		On(css.Mobile, widget.Part("action"), style.FontSize(style.TextLg), style.IconBox(style.IconLg)).
		On(css.Mobile, widget.Part("action-delete"), style.FontSize(style.TextLg), style.IconBox(style.IconLg)).
		On(css.Mobile, widget.Part("action-edit"), style.FontSize(style.TextLg), style.IconBox(style.IconLg)).
		// Bare, not Inset, on a phone: the grey card + border of fields/list
		// adds nothing over the panel's own page surface, and the list is
		// where the user first wants whitespace to breathe. Bare strips
		// background and border without touching the text color — Subtle
		// would also fix the grey but would silently mute inherited text,
		// which is a different statement than "this is not a card".
		// The rows (targetlist, white Page) and the inputs (fieldset, framed
		// Panel) each declare their own surfaces, so nothing loses contrast.
		On(css.Mobile, widget.Part("fields"),
			style.As(style.Bare),
		).
		On(css.Mobile, widget.Part("list"),
			style.As(style.Bare),
		).
		// cardInset even on mobile: fields and list keep the same pad between
		// the panel border and the content on every breakpoint, because the
		// list no longer has to buy space for the ⋮ menu in the sliver
		// MasterDetail(Most) leaves visible of it — the ⋮ is unreachable from
		// the sliver now (see targetlist/css.go's PartList comment), so the
		// indent budget is simply the form's: rightpanel 4 + cardInset 4 +
		// targetlist's own 8 = 16px on both columns. The old flush was a
		// deliberate mobile exception to cardInset's promise; it is retired
		// with the premise that justified it.
		//
		// FloatingChrome(EdgeBottom, IconMd, Space4): the footer row floats over
		// content this Part contains — rightpanel docks the aside-footer slot
		// to THIS panel's corner. The icons (IconMd, matching every footer
		// icon's own size) plus the docked gap, doubled,
		// is the footprint the scroller must clear. This Part is not itself what scrolls — targetlist's
		// own Fill()+Scroll() two levels in is — but --floating-bottom is an
		// inherited custom property, so it reaches that descendant without
		// either package knowing the other's class name. Whoever changes the
		// footer icons' size or the slot's docked gap must update this call to
		// match; nothing enforces the two staying in sync beyond that they are
		// three lines apart.
		On(css.Mobile, widget.Part("list"),
			style.Pad(cardInset),
			style.FloatingChrome(style.EdgeBottom, style.IconMd, style.Space4),
		).
		// Amber while open, matching the selection language the list already
		// uses: the button IS what is currently "selected" in the sense that
		// its own panel is the one on screen. Cancelling (closing) drops it
		// back to Primary blue, since at that point nothing is active anymore.
		When(widget.Open, widget.Part("action"),
			style.As(style.AccentInverse),
		).
		// Red while deleting, matching the rows it will act on: the rule
		// selects Within the footer's Open state, which holds exactly in
		// delete mode (see Render) — normal mode stays Primary blue. A When
		// on the button itself cannot say this: the button's own Open means
		// "visible", true in normal mode too, and Disclosure admits no
		// other state. Solid Danger (not DangerWash): the commit button is
		// itself a destructive surface, so it carries the white glyph
		// (--color-on-danger) instead of the wash's dark text.
		WhenWithin(widget.Open, widget.Part("footer"), widget.Part("action-delete"),
			style.As(style.Danger),
		).
		// The delete-confirmation modal's holder must not cost the frame a flex
		// share: as a plain in-flow child of rightpanel's Split it would take a
		// third of the width (the skeleton's `.rp > *` grow beats this sheet's
		// KeepSize across sheets). Pinning it fixed to the viewport corner takes
		// it out of the flow entirely; it is 0x0 while the modal is closed, and
		// the modal's own Backdrop(Viewport) covers the screen when it opens.
		Part(widget.Part("delconfirm-mount"),
			style.Docked(style.Viewport, style.EdgeTop, style.SideStart, style.SpaceNone),
		).
		// Space2: the same step modaldialog's panel puts between its header and
		// its body, so title→question and question→buttons read as one rhythm.
		Part(widget.Part("delconfirm-body"),
			style.Stack(style.Space2),
		).
		Part(widget.Part("delconfirm-actions"),
			style.Row(style.Space1),
		).
		// Button(), not a hand-composed As+Round+Pad: these two were the last
		// place in this sheet that spelled a button out by hand, and they drifted
		// exactly the way Button's doc warns about — RadiusSm and Space1 where
		// every other button in the app uses the surface's own radius and the
		// control box, so the confirm modal read as if it came from a different
		// product. Panel also dragged an outline in (it is a panel surface, and
		// panels carry a border); Secondary is the same fill with no border,
		// which is what a button wants. Danger keeps the destructive colour and
		// now gets the press/focus treatment derived from it instead of none.
		Part(widget.Part("delconfirm-btn"),
			style.Button(style.Secondary),
		).
		Part(widget.Part("delconfirm-btn-danger"),
			style.Button(style.Danger),
		).
		When(widget.Open, widget.Part("action-new"),
			style.RevealedBy(widget.Open),
		)
}

// RenderCSS implements the visual contract for crudview using the style DSL.
func (v *CrudView) RenderCSS() *css.Stylesheet {
	return v.RenderSheet().Stylesheet()
}
