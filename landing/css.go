//go:build !wasm

package landing

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for landing.
func (p *Page) RenderSheet() *style.Sheet {
	return style.For(p).
		Root(
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.As(style.Secondary),
		).
		Part(partHeader,
			style.Stack(style.SpaceNone),
			style.KeepSize(),
		).
		// The band: one padded, stacked strip per section. Everything below it
		// arranges content inside that strip and never re-pads it.
		//
		// Anchor(): a section can hold an arbitrary component — herobanner,
		// among others — that uses Backdrop(Parent) to go full-bleed. Backdrop
		// resolves against the nearest POSITIONED ancestor; without Anchor()
		// here, that search finds nothing between the section and the
		// viewport, so the "full-bleed hero" covers the ENTIRE page — header
		// included — instead of just its own section. landing is the one
		// composing arbitrary parts into a section, so it owns establishing
		// that positioning context; the embedded component can't know it.
		Part(partSection,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
			style.Anchor(),
		).
		Part(partFooter,
			style.Stack(style.Space2),
			style.Pad(style.Space6),
			style.As(style.Primary),
		).
		// Two children (text + image): two columns at most.
		Part(partSplit,
			style.Grid(2, style.ColumnMedium, style.Space4),
		).
		Part(partCards,
			style.Stack(style.Space4),
		).
		Part(partForm,
			style.Stack(style.Space4),
		).
		Part(partHours,
			style.Stack(style.Space4),
		).
		Part(partMap,
			style.Stack(style.Space4),
		).
		Part(partTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(partSubtitle,
			style.FontSize(style.TextLg),
			style.Glyph(style.Subtle),
		).
		Part(partBody,
			style.Stack(style.Space2),
		).
		Part(partMedia,
			style.MediaBox(style.Aspect16x9),
		).
		// Four: what fits a 1440px viewport at ColumnMedium, kept as the cap.
		Part(partGrid,
			style.Grid(4, style.ColumnMedium, style.Space4),
		).
		Part(partBadge,
			style.FontSize(style.TextSm),
			style.Glyph(style.Subtle),
		).
		Part(partBrand,
			style.FontWeight(style.WeightBold),
		)
}

// RenderCSS implements visual contract for landing layout.
func (p *Page) RenderCSS() *css.Stylesheet {
	return css.NewStylesheet(
		css.Raw(p.RenderSheet().Stylesheet().String()),
		css.Raw(sectionHeightFix),
	)
}

// sectionHeightFix: a section that directly wraps a full-bleed component
// (herobanner, via style.Backdrop(Parent) on its root) has that component
// removed from normal flow — position:absolute, sized by min-height alone.
// With nothing else in flow, the SECTION collapses to its own padding, and
// the visually 75vh-tall herobanner overflows past that collapsed box,
// painting over whatever comes after it in the document (the next section,
// or — before Part(partSection, ...Anchor()) existed — the page header,
// since Backdrop(Parent) had no positioned ancestor to stop at).
//
// :has() scopes this to exactly the sections that need it — Cards, Split,
// Stats and friends size normally from their own in-flow content and are
// untouched. The 75vh figure matches herobanner's own min-height, wherever
// a consumer declares it — this just gives the SECTION the same floor so
// document flow matches what's visually painted, instead of collapsing to
// its own padding while the hero overflows past it uncounted.
const sectionHeightFix = `.landing__section:has(> .herobanner){min-height:75vh}`
