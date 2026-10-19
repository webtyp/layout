package rightpanel

import (
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/html"
)

const NameRightPanel widget.Name = "rp"

// Part names, declared once so rightpanel.go (markup) and css.go (rules)
// cannot drift apart by a typo in either string literal.
const (
	partMain         widget.Part = "main"
	partArticle      widget.Part = "article"
	partAside        widget.Part = "aside"
	partAsideHeader  widget.Part = "aside-header"
	partAsideContent widget.Part = "aside-content"
	partAsideFooter  widget.Part = "aside-footer"
)

var (
	clsWrapper      = NameRightPanel.Root()
	clsMain         = NameRightPanel.Class(partMain)
	clsArticle      = NameRightPanel.Class(partArticle)
	clsAside        = NameRightPanel.Class(partAside)
	clsAsideHeader  = NameRightPanel.Class(partAsideHeader)
	clsAsideContent = NameRightPanel.Class(partAsideContent)
	clsAsideFooter  = NameRightPanel.Class(partAsideFooter)
)

func (r *RightPanel) WidgetName() widget.Name { return NameRightPanel }
func (r *RightPanel) WidgetKind() widget.Kind { return widget.Region }

// RightPanel is a two-column layout skeleton:
//   - Left: main content area (the article). It has no title of its own: the
//     module's name is the platform header's <h1> (see platformd).
//   - Right: aside panel with its own header (controls) and content.
//
// All slots are optional. A nil slot is simply not rendered.
// The layout does not define what the slots contain — that is the consumer's job.
//
// IMPORTANT: All Component implementors passed as slots MUST embed Element as a value,
// not as a pointer. See webtyp/dom interface.dom.go for details.
//
// Usage:
//
//	panel := &rightpanel.RightPanel{
//	    Article:       myTable,
//	    AsideControls: myFilterBar,
//	    Aside:         myDetailPanel,
//	    AsideFooter:   myActionButton,
//	}
//	panel.Render()
type RightPanel struct {
	Element

	// Article is the main content area.
	Article Component

	// AsideControls is rendered at the top of the aside panel (e.g. search + filter).
	AsideControls Component

	// Aside is the content area of the aside panel (e.g. detail view, info card).
	Aside Component

	// AsideFooter is rendered at the bottom of the aside panel, below the
	// content (e.g. a primary action button). It keeps its size while the
	// content between it and AsideControls takes the slack.
	AsideFooter Component

	// element handles for scroll-snap targets
	wrapper *Element
	main    *Element
	aside   *Element
}

// Render builds the layout element tree.
// Implements ViewRenderer.
func (r *RightPanel) Render() *Element {
	// ── root wrapper ─────────────────────────────────────────────────────────
	wrapper := Div().Set(clsWrapper.AsAttr()).Key("strip")

	// ── main section ─────────────────────────────────────────────────────────
	main := Section().Set(clsMain.AsAttr()).Key("main")

	// article
	article := Article().Set(clsArticle.AsAttr())
	if r.Article != nil {
		article.Child(r.Article)
	}
	main.Child(article)

	wrapper.Child(main)

	// ── aside panel ──────────────────────────────────────────────────────────
	var aside *Element
	if r.AsideControls != nil || r.Aside != nil || r.AsideFooter != nil {
		aside = Aside().Set(clsAside.AsAttr()).Key("aside")

		if r.AsideControls != nil {
			aside.Child(Div().Set(clsAsideHeader.AsAttr()).Child(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Child(Div().Set(clsAsideContent.AsAttr()).Child(r.Aside))
		}
		if r.AsideFooter != nil {
			aside.Child(Div().Set(clsAsideFooter.AsAttr()).Child(r.AsideFooter))
		}

		wrapper.Child(aside)
	}

	r.wrapper = wrapper
	r.main = main
	r.aside = aside

	return wrapper
}

// ShowMain brings the main panel into view; ShowAside brings the aside back.
//
// Both are no-ops on a wide screen, and that guard is the whole point:
// ScrollIntoView walks EVERY scrollable ancestor, not just the nearest one the
// caller had in mind. Side by side there is nothing to scroll here, so an
// unguarded call reached the platform's module deck instead and slid the whole
// application to the next module.
func (r *RightPanel) ShowMain()  { r.showPanel(r.main) }
func (r *RightPanel) ShowAside() { r.showPanel(r.aside) }

func (r *RightPanel) showPanel(target *Element) {
	if r.wrapper == nil || target == nil {
		return
	}
	strip, ok := r.wrapper.Ref()
	if !ok || !strip.ScrollsX() {
		return
	}
	if el, ok := target.Ref(); ok {
		el.ScrollIntoView()
	}
}
