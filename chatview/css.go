//go:build !wasm

package chatview

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for chatview.
func (v *ChatView) RenderSheet() *style.Sheet {
	return style.For(v).
		Root(
			style.Fill(),
			style.HideOverflow(),
		).
		Part(partHeader,
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
			style.DividerBelow(),
			style.FontWeight(style.WeightBold),
		).
		Part(partWork,
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.HideOverflow(),
		).
		Part(partThread,
			style.Grow(),
			style.Fill(),
			style.HideOverflow(),
		).
		// The list column: tab strip above a deck that takes the rest.
		Part(partAside,
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(partTabs,
			style.Row(style.Space2),
			style.KeepSize(),
			style.PadInline(style.Space2),
			style.DividerBelow(),
		).
		Part(partTab,
			style.Button(style.Subtle),
			style.Row(style.Space2),
			style.CenterContent(),
		).
		When(widget.Current, partTab,
			style.As(style.AccentInverse),
		).
		// Layers, not a strip: a horizontal scroller would chain with the one a
		// panel's own content may have.
		Part(partDeck,
			style.SlideDeck(style.MotionBase),
			style.Fill(),
		).
		Part(partPanel,
			style.Stack(style.SpaceNone),
			style.Scroll(),
		)
}

// RenderCSS implements the visual contract for chatview using the style DSL.
func (v *ChatView) RenderCSS() *css.Stylesheet {
	return v.RenderSheet().Stylesheet()
}
