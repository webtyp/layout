//go:build wasm

package app

import (
	"webtyp.com/components/usermenu"
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/layout/login"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

var (
	textLogOut     = lang.Text("Log out")
	textServerDown = lang.Text("The server did not answer. Check the connection and reload the page.")
)

// defaultIdleSeconds locks an idle session when the Session does not say.
const defaultIdleSeconds = 30 * 60

// New is the layout an app hands to browser.Run: it draws the Session that appkit built.
//
//	func main() { browser.Run(config.App, screens.Browser, layoutapp.New(devtools.Menu())) }
func New(tools Tools) view.Layout { return appLayout{tools: tools} }

type appLayout struct{ tools Tools }

// Show draws the Session: the server-down notice, the entry screens of the sign-in modes, or the chassis with the
// screens the user may open. It never asks the server who the user is: appkit already did (D75).
func (l appLayout) Show(s view.Session) {
	st := buildSession(s, l.tools)
	st.platform.Init(nil)
	Render("app", st)
}

// logoPath is the app's logo as the server serves it: sitec derives it from docs/logo/logo.png.
const logoPath = "/icon-192.png"

func entryScreen(title lang.Text, entries []view.Entry) *Element {
	var valid []view.Entry
	for _, e := range entries {
		if e.Form != nil {
			valid = append(valid, e)
		}
	}
	if len(valid) == 0 {
		return Div()
	}
	l := &login.Login{Title: title, Entries: valid, LogoMark: logoPath}
	return l.Render()
}

// session is one of three exclusive screens, chosen once from the Session.
type session struct {
	Element
	name       string
	serverDown bool
	signedIn   bool
	entries    []view.Entry
	platform   *platformd.Platform
}

func (s *session) Render() *Element {
	switch {
	case s.serverDown:
		return Div().Class("spa-root").Child(P().Text(lang.Translate(textServerDown).String()))
	case !s.signedIn:
		return Div().Class("spa-root").Child(entryScreen(lang.Text(s.name), s.entries))
	}
	return Div().Class("spa-root").Child(s.platform.Render())
}

// userIdentity is the signed-in user as the chassis' user menu reads it.
type userIdentity struct{ user view.User }

func (i userIdentity) UserName() string    { return i.user.Name }
func (i userIdentity) UserAvatar() string  { return i.user.Avatar }
func (i userIdentity) UserRoles() []string { return nil }

func buildSession(s view.Session, tools Tools) *session {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err) // NewUnixID never fails in the browser
	}
	user := s.User
	menu, err := BuildMenu(
		s.Screens,
		ids,
		func(r model.Resource) bool { return user.Can(r, model.Read) },
		tools,
	)
	if err != nil {
		panic(err)
	}

	signOut := s.SignOut
	idle := 0
	if user.ID != "" && signOut != nil {
		idle = s.IdleSeconds
		if idle <= 0 {
			idle = defaultIdleSeconds
		}
	}
	p := &platformd.Platform{
		AppName: s.Name,
		Modules: menu.Modules,
		User:    userIdentity{user: user},
		UserActions: func() Component {
			return Button().Text(lang.Translate(textLogOut).String()).
				Class(string(usermenu.NameUserMenu.Class(usermenu.PartAction))).
				OnClick(func(Event) {
					if signOut != nil {
						signOut()
					}
				})
		},
		CanView:     menu.CanView,
		IdleTimeout: idle,
		OnIdle: func() {
			if signOut != nil {
				signOut()
			}
		},
	}

	return &session{
		name: s.Name, serverDown: s.ServerDown, signedIn: user.ID != "",
		entries: s.Entries, platform: p,
	}
}
