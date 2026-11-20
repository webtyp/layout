//go:build !wasm

package login

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the login screen's visual contract using the style DSL.
func (l *Login) RenderCSS() *css.Stylesheet {
	return style.For(l).
		// Calm neutral surface backdrop (Panel): in ui-consistency Primary is
		// reserved for the brand mark and the primary action button ("Entrar"),
		// never a full-viewport saturated gradient behind the form.
		Root(
			style.Cover(),
			style.CenterContent(),
			style.As(style.Panel),
			style.Pad(style.Space4),
			style.Backdrop(style.Viewport),
		).
		// Elevated white card (Field = ColorBackground + ColorOnSurface +
		// 1px ColorOutline) over the Panel backdrop, with Popover shadow and
		// generous Space8 inset so the credentials form reads as a clean,
		// focused card.
		Part(PartCard,
			style.Stack(style.Space6),
			style.Width(style.Compact),
			style.As(style.Field),
			style.Round(style.RadiusLg),
			style.Raise(style.Popover),
			style.Pad(style.Space8),
		).
		Part(PartHeader,
			style.Stack(style.Space2),
			style.CenterContent(),
		).
		// PartMark is the <img> element itself. KeepSize() prevents flexbox from
		// shrinking the image along the column's main axis (which previously
		// cropped the top and bottom of the crest when MediaBox emitted
		// display:flex + overflow:hidden directly on the <img>).
		Part(PartMark,
			style.Width(style.Third),
			style.KeepSize(),
		).
		Part(PartTitle,
			style.FontSize(style.TextXl),
			style.FontWeight(style.WeightBold),
		).
		Part(PartSubtitle,
			style.FontSize(style.TextSm),
			style.As(style.Subtle),
		).
		// DangerWash, no Danger: un error de login es una nota dentro de la
		// tarjeta, no un bloque de alarma que compita con el formulario que
		// la persona tiene que volver a usar.
		Part(PartMessage,
			style.As(style.DangerWash),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
			style.FontSize(style.TextSm),
		).
		// Secondary sign-in modes inside the same card: a top hairline separates
		// each additional mode from the primary form above it.
		Part(PartEntry,
			style.Stack(style.Space4),
			style.PadEdge(style.EdgeTop, style.Space4),
		).
		Part(PartDivider,
			style.FontSize(style.TextSm),
			style.As(style.Subtle),
			style.CenterContent(),
		).
		Stylesheet()
}
