//go:build !wasm

package login

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the login screen's visual contract using the style DSL.
func (l *Login) RenderCSS() *css.Stylesheet {
	return style.For(l).
		Root(
			style.Cover(),
			style.CenterContent(),
			// As(Primary), not Page: this is the one screen with no
			// authenticated chrome around it to carry the brand, so the
			// backdrop itself does — ColorPrimary's own default gradient
			// (see webtyp.com/css's brandRoot/ColorPrimaryGradient), softened
			// by Veil()'s translucent ColorSurface wash + blur below so it
			// reads as ambient color, not a flat saturated fill the earlier
			// version of this file specifically avoided. Veil's wash paints
			// AFTER this and wins the background-color property, but never
			// touches background-image, so the gradient still shows through
			// it, blurred.
			style.As(style.Primary),
			style.Pad(style.Space4),
			style.Backdrop(style.Viewport),
			style.Veil(),
		).
		// Inset, not Panel or Page: fieldset's input is a framed Panel now, so
		// its own hairline is what separates it from whatever it sits on —
		// contrast against the card behind it no longer has to do that work.
		// Inset stays because it is the surface crudview already puts its own
		// field stack on, so a login form and a module form come out of the
		// same box.
		//
		// Space6 twice over: the gap between header and form is the same air
		// as the card's own inset, which is what keeps a card this small from
		// reading as cramped. Compact caps it at a single column of controls
		// — Readable's 65ch is a measure for prose and made the card wider
		// than most of the screens it fronts.
		//
		// No Backdrop/Veil here: those take the element out of flow
		// (position: absolute; inset: 0), which is what a full-coverage
		// scrim needs but breaks two-axis centering for a Width-constrained
		// card — Root's own CenterContent() already centers this card on
		// both axes as long as it stays a normal flex child. A translucent
		// card look, if wanted later, needs Root's veil color mixed into
		// this Part's own background instead of Backdrop(Parent).
		Part(PartCard,
			style.Stack(style.Space6),
			style.Width(style.Compact),
			style.As(style.Inset),
			style.Round(style.RadiusLg),
			style.Raise(style.Floating),
			style.Pad(style.Space6),
		).
		// Space1, against the card's Space6: title and subtitle are one block
		// that happens to be set in two sizes, and spacing them like siblings
		// of the form would read as three unrelated things stacked up.
		Part(PartHeader,
			style.Stack(style.Space3),
			style.CenterContent(),
		).
		// Logo sits inside the header, above the title, centered and constrained.
		// Width Third (~33% of the Compact card) keeps the falcon crest readable
		// without covering the form. Stack Space3 separates logo → title → subtitle.
		Part(PartMark,
			style.MediaBox(style.AspectSquare),
			style.Width(style.Third),
		).
		Part(PartTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(PartSubtitle,
			style.FontSize(style.TextBase),
			style.Glyph(style.Primary),
		).
		Stylesheet()
}
