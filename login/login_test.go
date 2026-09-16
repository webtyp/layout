package login

import (
	"strings"
	"testing"

	. "webtyp.com/dom"
	. "webtyp.com/html"
)

func TestLogin_RendersTitleAndForm(t *testing.T) {
	form := Div().Attr("id", "the-form")
	html := (&Login{Title: "Demo CMS", Form: form}).Render().String()

	for _, want := range []string{"login", "login__card", "login__title", "Demo CMS", "the-form"} {
		if !strings.Contains(html, want) {
			t.Errorf("markup missing %q\n%s", want, html)
		}
	}
}

func TestLogin_MessageNilSlotNotRendered(t *testing.T) {
	html := (&Login{Title: "App"}).Render().String()
	if strings.Contains(html, "login__message") {
		t.Errorf("expected no login__message slot when Message is nil\n%s", html)
	}
}

func TestLogin_MessageEmptySignalRenderedHidden(t *testing.T) {
	msg := NewString("")
	html := (&Login{Title: "App", Message: msg}).Render().String()

	if !strings.Contains(html, "login__message") {
		t.Fatalf("expected login__message slot to be in DOM tree when Message signal is provided\n%s", html)
	}
	if !strings.Contains(html, "display:none") && !strings.Contains(html, "hidden") {
		t.Errorf("expected login__message slot to be hidden when signal value is empty\n%s", html)
	}
}

func TestLogin_MessageNonEmptySignalRenderedVisible(t *testing.T) {
	msg := NewString("Acceso denegado")
	html := (&Login{Title: "App", Message: msg}).Render().String()

	if !strings.Contains(html, "login__message") {
		t.Fatalf("expected login__message slot to be rendered\n%s", html)
	}
	if !strings.Contains(html, "Acceso denegado") {
		t.Errorf("expected message text 'Acceso denegado' in markup\n%s", html)
	}
	if strings.Contains(html, "display:none") {
		t.Errorf("expected message slot to be visible (no display:none)\n%s", html)
	}
}

func TestLogin_MessageSlotPrecedesForm(t *testing.T) {
	msg := NewString("Error")
	form := Div().Attr("id", "the-form")
	html := (&Login{Title: "App", Message: msg, Form: form}).Render().String()

	msgIdx := strings.Index(html, "login__message")
	formIdx := strings.Index(html, "the-form")

	if msgIdx == -1 || formIdx == -1 {
		t.Fatalf("expected both login__message and form in markup\n%s", html)
	}
	if msgIdx >= formIdx {
		t.Errorf("expected login__message (%d) to appear before form (%d)\n%s", msgIdx, formIdx, html)
	}
}

func TestLogin_MessageSlotCarriesRoleAlert(t *testing.T) {
	msg := NewString("Error")
	html := (&Login{Title: "App", Message: msg}).Render().String()

	if !strings.Contains(html, "role='alert'") && !strings.Contains(html, `role="alert"`) {
		t.Errorf("expected login__message slot to carry role=\"alert\"\n%s", html)
	}
}

func TestLogin_ConsumerShapeAsyncErrorUpdate(t *testing.T) {
	msg := NewString("")
	form := Div().Attr("id", "the-form")
	l := &Login{
		Title:   "App",
		Message: msg,
		Form:    form,
	}

	elem := l.Render()
	htmlBefore := elem.String()

	if strings.Contains(htmlBefore, "Credenciales inválidas") {
		t.Errorf("expected no error text initially, got:\n%s", htmlBefore)
	}

	// Simulate async server response setting error message after render
	msg.Set("Credenciales inválidas")

	htmlAfter := elem.String()
	if !strings.Contains(htmlAfter, "Credenciales inválidas") {
		t.Errorf("expected updated message 'Credenciales inválidas' after msg.Set, got:\n%s", htmlAfter)
	}
}

func TestLogin_SubtitleIsOptional(t *testing.T) {
	html := (&Login{Title: "App"}).Render().String()
	if strings.Contains(html, "login__subtitle") {
		t.Errorf("expected no subtitle line when Subtitle is empty\n%s", html)
	}

	html = (&Login{Title: "App", Subtitle: "Ingrese sus credenciales"}).Render().String()
	if !strings.Contains(html, "login__subtitle") || !strings.Contains(html, "Ingrese sus credenciales") {
		t.Errorf("expected the subtitle to render when set\n%s", html)
	}
}

func TestLogin_MarkIsOptional(t *testing.T) {
	html := (&Login{Title: "App"}).Render().String()
	if strings.Contains(html, "login__mark") {
		t.Errorf("expected no mark when LogoMark is empty\n%s", html)
	}

	html = (&Login{Title: "App", LogoMark: "data:image/svg+xml,x"}).Render().String()
	if !strings.Contains(html, "login__mark") {
		t.Errorf("expected a mark when LogoMark is set\n%s", html)
	}
	if !strings.Contains(html, "src='data:image/svg+xml,x'") {
		t.Errorf("expected the mark's src to carry LogoMark, got\n%s", html)
	}
}

func TestLogin_RenderHTMLZeroValueReturnsEmptyString(t *testing.T) {
	inst := &Login{}
	got := inst.RenderHTML()
	if got != "" {
		t.Errorf("expected RenderHTML on zero value to return empty string, got: %q", got)
	}
}

func TestLogin_RenderHTMLMatchesRenderString(t *testing.T) {
	buildLogin := func() *Login {
		form := Div().Attr("id", "ssr-form").Text("SSR content")
		return &Login{
			Title:    "Test SSR Login",
			Subtitle: "Please login",
			Form:     form,
		}
	}

	htmlRender := buildLogin().Render().String()
	htmlRenderHTML := buildLogin().RenderHTML()

	if htmlRenderHTML != htmlRender {
		t.Errorf("RenderHTML() and Render().String() mismatch:\nRenderHTML: %s\nRender().String(): %s", htmlRenderHTML, htmlRender)
	}

	for _, want := range []string{"login__card", "ssr-form", "SSR content"} {
		if !strings.Contains(htmlRenderHTML, want) {
			t.Errorf("RenderHTML output missing expected element/content %q:\n%s", want, htmlRenderHTML)
		}
	}
}
