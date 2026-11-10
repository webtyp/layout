//go:build wasm

package app

import (
	"strings"

	"webtyp.com/components/usermenu"
	. "webtyp.com/dom"
	"webtyp.com/fetch"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/layout/login"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

var textLogOut = lang.Text("Log out")

const (
	pathSetup  = "/setup"
	pathLogout = "/logout"
)

// opMe is the operation that answers who the user is: the one place that names it.
var opMe = router.Operation{Module: "authority", Entity: "me", Action: model.Read}

type profileDTO struct {
	Id          string
	Name        string
	Email       string
	Avatar      string
	Roles       []string
	RoleNames   []string
	Permissions []string
	Locale      string
}

func (p profileDTO) IsNil() bool { return false }

func (p profileDTO) EncodeFields(w model.FieldWriter) {}

func (p *profileDTO) DecodeFields(r model.FieldReader) {
	p.Id, _ = r.String("id")
	p.Name, _ = r.String("name")
	p.Email, _ = r.String("email")
	p.Avatar, _ = r.String("avatar")
	p.Locale, _ = r.String("locale")
	if ar, ok := r.Array("roles"); ok {
		p.Roles = make([]string, ar.Len())
		for i := 0; i < ar.Len(); i++ {
			p.Roles[i] = ar.String(i)
		}
	}
	if arn, ok := r.Array("role_names"); ok {
		p.RoleNames = make([]string, arn.Len())
		for i := 0; i < arn.Len(); i++ {
			p.RoleNames[i] = arn.String(i)
		}
	}
	if ap, ok := r.Array("permissions"); ok {
		p.Permissions = make([]string, ap.Len())
		for i := 0; i < ap.Len(); i++ {
			p.Permissions[i] = ap.String(i)
		}
	}
}

func (p profileDTO) Can(resource model.Resource, action model.Action) bool {
	for _, perm := range p.Permissions {
		i := strings.LastIndex(perm, ":")
		if i < 0 {
			continue
		}
		actions, err := model.ParseAction(perm[i+1:])
		if err != nil {
			continue
		}
		if (model.Grant{Resource: model.Resource(perm[:i]), Actions: actions}).Matches(resource, action) {
			return true
		}
	}
	return false
}

// New is the layout an app hands to appkit.Browser: it draws the Session that appkit built.
//
//	func main() { appkit.Browser(config.App, layoutapp.New(devtools.Menu())) }
func New(tools Tools) view.Layout { return appLayout{tools: tools} }

type appLayout struct{ tools Tools }

// Show asks the server who the user is and shows the first-admin setup, the access screen or the chassis with the screens the user may open.
func (l appLayout) Show(s view.Session) {
	st := buildSession(s, l.tools)
	st.platform.Init(nil)
	Render("app", st)
}

func entryScreen(b view.Brand, title lang.Text, e view.Entry) *Element {
	if e.Form == nil {
		return Div()
	}
	l := &login.Login{Title: title, Subtitle: e.Subtitle, Form: e.Form, Message: e.Message, LogoMark: b.LogoMark}
	return l.Render()
}

// session switches between the three exclusive screens.
type session struct {
	Element
	brand           view.Brand
	needsSetup      *SignalBool // the server has no administrator yet (auth.PathSetup answers 200)
	showLogin       *SignalBool
	isAuthenticated *SignalBool
	platform        *platformd.Platform
	entry           view.EntryScreens
}

func (s *session) Render() *Element {
	setupView := Show(s.needsSetup, func() *Element {
		return entryScreen(s.brand, lang.Text(s.brand.Name), s.entry.Setup)
	})
	loginView := Show(s.showLogin, func() *Element {
		return entryScreen(s.brand, lang.Text(s.brand.Name), s.entry.Login)
	})
	// The chassis is built with the profile already known: platformd filters
	// the screens during its render, and Show renders its content only once.
	platformView := Show(s.isAuthenticated, func() *Element { return s.platform.Render() })
	return Div().Class("spa-root").Child(setupView, loginView, platformView)
}

// profileIdentity reads the profile live: "me" fills it after Platform.User is set.
type profileIdentity struct{ profile *profileDTO }

func (i profileIdentity) UserName() string    { return i.profile.Name }
func (i profileIdentity) UserAvatar() string  { return i.profile.Avatar }
func (i profileIdentity) UserRoles() []string { return i.profile.Roles }

func reload() { Navigate("") }

func logout() { fetch.Post(pathLogout).Send(func(*fetch.Response, error) { reload() }) }

func buildSession(s view.Session, tools Tools) *session {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err) // NewUnixID never fails in the browser
	}
	caller := s.Caller
	entry := s.Entry

	// Who the user is comes first: blocking here is valid because Browser is
	// called from main, not from a JS callback.
	profile := &profileDTO{}
	meDone := make(chan error, 1)
	caller.Call(opMe, &emptyArgs{}, profile, func(err error) { meDone <- err })
	authenticated := <-meDone == nil

	// Without a session only the server knows whether the installation is new;
	// on any doubt the access screen is shown.
	setupRequired := false
	if !authenticated {
		setupDone := make(chan int, 1)
		fetch.Get(pathSetup).Send(func(resp *fetch.Response, err error) {
			if err != nil || resp == nil {
				setupDone <- 0
				return
			}
			setupDone <- resp.Status
		})
		setupRequired = <-setupDone == 200
	}

	isAuthenticated := NewBool(authenticated)
	needsSetup := NewBool(setupRequired)
	showLogin := DeriveBool(func() bool { return !isAuthenticated.Get() && !needsSetup.Get() })

	menu, err := BuildMenu(
		s.Screens,
		ids,
		func(r model.Resource) bool { return profile.Can(r, model.Read) },
		tools,
	)
	if err != nil {
		panic(err)
	}

	idle := 0
	if authenticated {
		idle = entry.IdleSeconds
		if idle <= 0 {
			idle = 30 * 60
		}
	}
	p := &platformd.Platform{
		AppName: s.Brand.Name,
		Modules: menu.Modules,
		User:    profileIdentity{profile: profile},
		UserActions: func() Component {
			return Button().Text(lang.Translate(textLogOut).String()).
				Class(string(usermenu.NameUserMenu.Class(usermenu.PartAction))).
				OnClick(func(Event) { logout() })
		},
		CanView:     menu.CanView,
		IdleTimeout: idle,
		OnIdle:      logout,
	}

	return &session{
		brand: s.Brand, needsSetup: needsSetup, showLogin: showLogin, isAuthenticated: isAuthenticated,
		platform: p, entry: entry,
	}
}

type emptyArgs struct{}

func (emptyArgs) IsNil() bool                      { return false }
func (emptyArgs) EncodeFields(w model.FieldWriter) {}
func (emptyArgs) DecodeFields(r model.FieldReader) {}
func (emptyArgs) Validate(action byte) error       { return nil }
