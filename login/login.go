package login

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/view"
	"webtyp.com/widget"
)

// NameLogin is the widget identity; it produces the class prefix "login" and
// the part classes "login__card", "login__header", "login__title",
// "login__subtitle", "login__mark", "login__message", "login__entry",
// "login__divider".
const NameLogin = widget.Name("login")

const (
	PartCard     = widget.Part("card")     // the centered card: header + entry blocks
	PartHeader   = widget.Part("header")   // mark + title + primary subtitle, packed as one block
	PartTitle    = widget.Part("title")    // the app's own name, leading the card
	PartSubtitle = widget.Part("subtitle") // the one line of orientation under it
	PartMark     = widget.Part("mark")     // the brand mark at the top of the card
	PartMessage  = widget.Part("message")  // el resultado del último intento de envío
	PartEntry    = widget.Part("entry")    // one sign-in mode block (divider + message + form)
	PartDivider  = widget.Part("divider")  // orientation line above a secondary sign-in mode
)

var (
	clsLogin    = NameLogin.Root()
	clsCard     = NameLogin.Class(PartCard)
	clsHeader   = NameLogin.Class(PartHeader)
	clsTitle    = NameLogin.Class(PartTitle)
	clsSubtitle = NameLogin.Class(PartSubtitle)
	clsMark     = NameLogin.Class(PartMark)
	clsMessage  = NameLogin.Class(PartMessage)
	clsEntry    = NameLogin.Class(PartEntry)
	clsDivider  = NameLogin.Class(PartDivider)
)

// Login is the pre-authentication screen: an elevated card (brand mark, title,
// subtitle, and the open sign-in modes) centered on a calm neutral backdrop.
//
// It owns none of the form's fields or validation — each entry's Form is built
// by the sign-in mode (or the composition root), so this package never assumes
// a shape for what it centers, only that it Renders.
type Login struct {
	Element // value embed — NEVER *dom.Element (TinyGo heap constraint)

	// Title leads the card — the app's own name, or "Sign in" if the app
	// prefers that framing over its brand. English; translated in Render.
	// Required: the screen reads as unbranded without it.
	Title lang.Text

	// Subtitle is the single line under the title telling the user what this
	// screen wants from them ("Enter your key to continue"). When empty and
	// Entries is non-empty, Entries[0].Subtitle is used.
	Subtitle lang.Text

	// Form is a single login form, built and validated by the caller. When
	// Entries is also empty, Form + Subtitle + Message act as a single entry.
	Form Component

	// Message es el resultado del último intento de envío cuando se usa Form
	// directamente.
	Message *SignalString

	// Entries are the sign-in modes open at this door (view.Session.Entries).
	// When more than one mode is open (e.g. trusted_ip + passkey), all render
	// inside the same card: the first leads under the header, and each
	// subsequent mode renders below a subtle divider with its own Subtitle,
	// Message slot and Form.
	Entries []view.Entry

	// LogoMark is a data-URI or URL for the brand mark at the top of the card.
	// Optional — a Login with no mark simply does not render it.
	LogoMark string
}

func (l *Login) WidgetName() widget.Name { return NameLogin }

// Region: this screen is a landmark a user orients by, not a control with
// its own interaction pattern — the Form inside it is the interactive
// piece, and carries its own Kind.
func (l *Login) WidgetKind() widget.Kind { return widget.Region }

func (l *Login) resolvedEntries() []view.Entry {
	if len(l.Entries) > 0 {
		return l.Entries
	}
	if l.Form != nil || l.Message != nil || l.Subtitle != "" {
		return []view.Entry{{Subtitle: l.Subtitle, Form: l.Form, Message: l.Message}}
	}
	return nil
}

func (l *Login) Render() *Element {
	entries := l.resolvedEntries()
	leadSubtitle := l.Subtitle
	if leadSubtitle == "" && len(entries) > 0 {
		leadSubtitle = entries[0].Subtitle
	}

	header := Div().Set(clsHeader.AsAttr())
	if l.LogoMark != "" {
		header.Child(NewElement("img").NoCloseTag().Set(clsMark.AsAttr()).Attr("src", l.LogoMark).Attr("alt", ""))
	}
	header.Child(H1().Set(clsTitle.AsAttr()).Text(lang.Translate(l.Title).String()))
	if leadSubtitle != "" {
		header.Child(P().Set(clsSubtitle.AsAttr()).Text(lang.Translate(leadSubtitle).String()))
	}

	card := Div().Set(clsCard.AsAttr()).Child(header)

	for i, e := range entries {
		if i == 0 {
			if e.Message != nil {
				card.Child(messageNode(e.Message))
			}
			if e.Form != nil {
				card.Child(e.Form)
			}
			continue
		}
		block := Div().Set(clsEntry.AsAttr())
		if e.Subtitle != "" {
			block.Child(P().Set(clsDivider.AsAttr()).Text(lang.Translate(e.Subtitle).String()))
		}
		if e.Message != nil {
			block.Child(messageNode(e.Message))
		}
		if e.Form != nil {
			block.Child(e.Form)
		}
		card.Child(block)
	}

	root := Div().Set(clsLogin.AsAttr())
	root.Child(card)
	return root
}

func messageNode(sig *SignalString) *Element {
	// role="alert" hace que el lector de pantalla lo anuncie cuando
	// aparece, que es justo el momento en que importa.
	return Div().Set(clsMessage.AsAttr()).
		Attr("role", "alert").
		BindAttrBoolFunc("hidden", func() bool {
			return sig.Get() == ""
		}).
		BindText(sig)
}
