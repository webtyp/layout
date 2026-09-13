package platformd

import (
	"sync"

	"webtyp.com/components/usermenu"
	"webtyp.com/layout"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/time"
)

const NamePlatform widget.Name = "pd"

var (
	clsRoot           = NamePlatform.Root()
	clsHeader         = NamePlatform.Class("header")
	clsDrawerPanel    = NamePlatform.Class("drawer-panel")
	clsAppName        = NamePlatform.Class("app-name")
	clsDrawerIdentity = NamePlatform.Class("drawer-identity")
	clsDrawerBrand    = NamePlatform.Class("drawer-brand")
	clsBrand          = NamePlatform.Class("brand")
	clsBrandMark      = NamePlatform.Class("brand-mark")
	clsBrandName      = NamePlatform.Class("brand-name")
	clsMsgSlot        = NamePlatform.Class("msg-slot")
	clsMsg            = NamePlatform.Class("msg")
	clsHeaderRight    = NamePlatform.Class("header-right")
	clsBody           = NamePlatform.Class("body")
	clsStage          = NamePlatform.Class("stage")
	clsPanel          = NamePlatform.Class("panel")
	clsMenu           = NamePlatform.Class("menu")
	clsNavbar         = NamePlatform.Class("navbar")
	clsNavItem        = NamePlatform.Class("nav-item")
	clsNavLink        = NamePlatform.Class("nav-link")
	clsLinkText       = NamePlatform.Class("link-text")
	ClsNavIcon        = NamePlatform.Class("nav-icon")
	clsHamburger      = NamePlatform.Class("hamburger")
	clsNavOverlay     = NamePlatform.Class("nav-overlay")
	clsMsgStack       = NamePlatform.Class("msg-stack")
	clsMsgSlotMobile  = NamePlatform.Class("msg-slot-mobile")
)

const (
	IconUser  = svg.Icon("pd-user")
	IconBrand = svg.Icon("pd-brand")
	iconMenu  = svg.Icon("pd-menu")
)

// UIModule is a module that provides its UI to the platform chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	layout.Module    // identity: ModelName() → used as ID
	Label() string   // text in the nav rail
	Icon() svg.Icon  // chassis renders via the sprite
	View() Component // module content (often a *rightpanel.RightPanel)
}

func (p *Platform) WidgetName() widget.Name { return NamePlatform }
func (p *Platform) WidgetKind() widget.Kind { return widget.Menu }

// Identity is what the platform needs to know about whoever is logged in.
//
// It is a READ contract, not a store: platformd renders it and never mutates
// it. Whatever owns authentication — webtyp.com/user in a real
// application — supplies an implementation; the platform neither knows nor
// cares where the values come from.
//
// It asks for facts, not for presentation. The glyph drawn when there is no
// avatar is IconUser, which this package owns — an authentication package has
// no business choosing a sprite, and asking it to would put a rendering
// decision behind a login.
//
// Roles, plural, because they are plural: a user holds N of them and no
// ordering exists to say which one matters. Asking for a single "area" forced
// every implementation to pick arbitrarily, and that decision was invisible.
type Identity interface {
	// UserName is who is logged in.
	UserName() string
	// UserAvatar is the URL of their picture. Empty is normal and expected.
	UserAvatar() string
	// UserRoles are display names, not authorization codes. May be empty.
	UserRoles() []string
}

// Brand is what the platform calls itself in its own chrome. The shell asks for
// a mark and a name; how they are drawn, sized and spaced is platformd's
// business.
//
// It is a READ contract, not a store, mirroring Identity: platformd renders it
// and never mutates it. The consumer supplies facts, not presentation — a
// Brand never hands the shell an svg.Icon, because picking the sprite is a
// rendering decision this package owns (the same reasoning that keeps IconUser
// out of Identity).
type Brand interface {
	// BrandName is shown beside the mark, and is the mark's alt text.
	BrandName() string
	// BrandMark is a URL or inline SVG data URI. Empty is normal and expected:
	// the shell falls back to its own glyph, exactly as UserAvatar does.
	BrandMark() string
}

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName titles the drawer on a phone when there is no Brand to lead it —
	// the panel opens wholesale there and there is room for a line of text. The
	// collapsed rail never shows it: text with no icon beside it would have to
	// appear from nowhere when the rail expands.
	//
	// It is a fallback for a missing Brand, not a second header: a Brand
	// supersedes it in the drawer, since showing both would say the app's own
	// name twice in the same panel. A platform with neither renders no head at
	// all above the drawer's nav.
	AppName string

	// Brand is what the platform calls itself: the mark and the name at the
	// header's leading slot, mirrored at the head of the phone drawer where
	// there is no header to hold it. Tapping it — either surface — is the "go
	// home" control, landing on DefaultID (or the first viewable module).
	// Optional — a platform without a logo renders no brand slot, AppName
	// takes the drawer's head instead, and the header's message block stays
	// centred between nothing and the menu.
	Brand Brand

	// User is the logged-in identity. Required: the header's outer thirds and
	// the drawer's first entry are built from it.
	User Identity

	// UserActions slot — shown at the header RIGHT, next to the work-area name
	// (e.g. the light/dark theme toggle). Optional.
	// A factory, not a Component: the shell renders one menu per surface and a
	// single element cannot have two parents. Passing one instance put the same
	// element in both menus, which rendered twice with one id and left the copy
	// in the drawer inert.
	UserActions func() Component

	// Modules registered in order — appearance order in the nav rail.
	Modules []UIModule

	// CanView filters which modules the shell presents. nil = show all.
	CanView func(resource string) bool

	// DefaultID is the ModelName() of the module to show initially.
	// If empty, the first module is used.
	DefaultID string

	// IdleTimeout is the number of seconds without user activity inside the
	// platform before OnIdle fires once. Activity is the pointer entering the
	// shell, a click or tap, a key press, focus moving inside, or a document
	// scroll — see armIdle. 0 — the default — disables the idle lock entirely:
	// nothing is armed, no listener changes behavior.
	IdleTimeout int

	// OnIdle is called when IdleTimeout seconds pass without activity. Required
	// when IdleTimeout > 0 — Init panics otherwise (a configured idle lock with
	// no consequence would be a silent failure). The platform does NOT re-arm
	// after firing: the next activity starts a fresh period. Typical use: post
	// the logout route and reload.
	OnIdle func()

	// internal state
	active              *SignalString
	menuOpen            *SignalBool
	notifications       *SignalNodes // desktop toasts (header msg-slot)
	notificationsMobile *SignalNodes // mobile toasts (msg-stack under the hamburger)
	navIcon             *SignalNodes
	navStowed           *SignalBool
	idleTimer           time.Timer
	idleArmed           bool

	rawNotifications []notification
	mu               sync.Mutex

	// lastScrollTop es el testigo de la última posición leída por onScroll. Los
	// eventos de scroll llegan por el hilo de JS, igual que los click, así que no
	// va bajo p.mu — ese mutex existe por time.AfterFunc, que descarta
	// notificaciones desde otra goroutine.
	lastScrollTop float64
}

// scrollStowThreshold son los píxeles de desplazamiento que hacen falta para que
// el cromo reaccione. Sin umbral, un píxel de ruido lo haría entrar y salir; y
// como en la página conviven varios scrollers y sus posiciones se intercalan en
// un mismo handler, un salto pequeño puede venir de otro elemento y no de un
// gesto.
const scrollStowThreshold = 8

type notification struct {
	Type MessageType
	Msg  string
	ID   string
	// expiryNs is the auto-dismiss deadline (UnixNano); 0 = persistent. The
	// timer is stored on the notification so pause/resume can cancel and
	// re-arm it — see pauseToast/resumeToast.
	expiryNs int64
	timer    time.Timer
}

// Duration is how long a notification stays before dismissing itself. A
// Duration is a DECISION, not a number: 0-as-persistent and -1-as-automatic
// were magic numbers nobody could read at the call site, and auto-sizing
// needs the message text that only Notify has.
type Duration struct {
	millis func(msg string) int // nil = persistent
}

// Auto sizes the duration to the message: ~350ms per word plus ~1.2s to
// notice the toast, floored at 2s (a one-word confirmation is fully read in
// that time) and capped at 8s so a long message cannot hold the screen
// hostage. One word → 2s; three words → 2.25s; a 15-word error → 6.45s.
func Auto() Duration {
	return Duration{millis: func(msg string) int {
		words := Count(msg, " ") + 1
		ms := 1200 + words*350
		if ms < 2000 {
			return 2000
		}
		if ms > 8000 {
			return 8000
		}
		return ms
	}}
}

// Persistent keeps the notification until the user dismisses it. The case
// for it is error reporting: a message that vanishes before it is read
// defeats the report (WCAG 2.2.1, Timing Adjustable — the user must be able
// to extend or disable a time limit).
func Persistent() Duration { return Duration{} }

// For pins an exact duration in milliseconds.
func For(ms int) Duration {
	return Duration{millis: func(string) int { return ms }}
}

// Init initializes the platform state and routing.
func (p *Platform) Init(ctx Ctx) {
	p.active = NewString("")
	p.menuOpen = NewBool(false)
	p.notifications = NewNodes()
	p.notificationsMobile = NewNodes()
	p.navIcon = NewNodes()
	p.navStowed = NewBool(false)

	if p.IdleTimeout > 0 && p.OnIdle == nil {
		panic("platformd: IdleTimeout requires OnIdle")
	}

	OnHashChange(func(hash string) {
		if len(hash) > 0 && hash[0] == '#' {
			p.Activate(hash[1:])
		}
	})

	// El scroll no burbujea: en captura sobre el documento es la única forma de que
	// el chasis vea el desplazamiento de un contenedor que pertenece a otro paquete.
	OnScrollCapture(func(top float64) {
		p.onScroll(top)
		p.activity()
	})

	hash := GetHash()
	if hash != "" && len(hash) > 0 && hash[0] == '#' {
		id := hash[1:]
		if p.isViewable(id) {
			p.Activate(id)
		} else {
			p.fallback()
		}
	} else {
		p.goHome()
	}
}

// goHome activa el módulo raíz: el mismo cálculo que Init usa cuando la app
// abre sin hash, reutilizado por el clic en la marca — pulsar el logo lleva
// al mismo sitio que abrir la app desde cero.
func (p *Platform) goHome() {
	if p.DefaultID != "" && p.isViewable(p.DefaultID) {
		p.Activate(p.DefaultID)
		return
	}
	p.fallback()
}

func (p *Platform) isViewable(id string) bool {
	if p.CanView == nil {
		return true
	}
	return p.CanView(id)
}

// activeIcon es el glifo del módulo en el que estamos. iconMenu es el respaldo:
// UIModule permite que Icon() devuelva la cadena vacía, y un botón sin glifo no
// se puede pulsar porque no se ve.
func (p *Platform) activeIcon() svg.Icon {
	id := p.active.Get()
	for _, m := range p.Modules {
		if m.ModelName() == id {
			if ic := m.Icon(); ic != "" {
				return ic
			}
			break
		}
	}
	return iconMenu
}

// activity rearms the idle timer; called from every presence signal.
//
// idleTimer is guarded by p.mu: the AfterFunc callback below runs on its own
// goroutine (the stdlib backend lane; WASM is single-threaded, but the type
// is shared code) and writes idleTimer to nil concurrently with whatever
// goroutine calls activity() next — exactly the race p.mu was already
// documented, on lastScrollTop, as existing for.
func (p *Platform) activity() {
	if p.IdleTimeout <= 0 {
		return
	}
	p.mu.Lock()
	if p.idleTimer != nil {
		p.idleTimer.Stop()
	}
	p.idleTimer = time.AfterFunc(p.IdleTimeout*1000, func() {
		p.mu.Lock()
		p.idleTimer = nil
		p.mu.Unlock()
		p.OnIdle()
	})
	p.mu.Unlock()
}

// armIdle attaches the presence listeners to the platform root. Called
// once from Render (idleArmed guards re-renders).
func (p *Platform) armIdle(root *Element) {
	if p.IdleTimeout <= 0 || p.idleArmed {
		return
	}
	p.idleArmed = true
	root.OnMouseEnter(func(Event) { p.activity() })
	root.OnKeyDown(func(KeyEvent) { p.activity() })
	root.OnFocusIn(func(Event) { p.activity() })
	root.OnClick(func(Event) { p.activity() })
	p.activity() // opening the app counts as presence
}

// onScroll guarda el botón de menú mientras el usuario baja y lo devuelve en
// cuanto sube. Arriba del todo siempre está a mano: una pantalla que no llega a
// desplazarse — el listado de este demo, sin ir más lejos — dejaría el menú
// inalcanzable si el botón naciera guardado.
//
// Limitación conocida: el handler recibe la posición de CUALQUIER scroller. Si
// dos están en pantalla y se desplazan alternándose, sus posiciones se
// intercalan en un mismo lastScrollTop y el botón puede parpadear. En móvil
// solo hay una columna visible a la vez (MasterDetail enseña una), así que en
// la práctica no ocurre, y el umbral de 8px absorbe el resto.
func (p *Platform) onScroll(top float64) {
	if top <= 0 {
		p.lastScrollTop = 0
		p.navStowed.Set(false)
		return
	}
	switch {
	case top > p.lastScrollTop+scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(true)
	case top < p.lastScrollTop-scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(false)
	}
}

func (p *Platform) fallback() {
	for _, m := range p.Modules {
		id := m.ModelName()
		if p.isViewable(id) {
			p.Activate(id)
			return
		}
	}
}

// Render builds the DOM tree (implements ViewRenderer).
// userMenu adapts the Identity contract into the component's plain props. The
// component must not know the contract: it lives below layout in the graph, and
// typing Identity there would invert the dependency.
//
// A fresh instance per call on purpose — the shell renders one for the header
// and one for the drawer, and a single *UserMenu cannot have two parents.
// brand builds the header's leading slot from the Brand contract. The mark is
// an <img> when the contract returns a URL and the default glyph otherwise —
// the avatar's exact treatment, mirrored.
// Clickable: the brand doubles as the "go home" control, the standard
// meaning of tapping a logo. It fires the same navigation Init runs when the
// app opens with no hash, on the header and the drawer alike — one behaviour
// for one glyph, wherever it is drawn.
func (p *Platform) brand() *Element {
	slot := Div().Set(clsBrand.AsAttr())
	if url := p.Brand.BrandMark(); url != "" {
		slot.Child(NewElement("img").
			Set(clsBrandMark.AsAttr()).
			Attr("src", url).
			Attr("alt", p.Brand.BrandName()).
			Attr("loading", "lazy"))
	} else {
		slot.Child(IconBrand.Render(string(clsBrandMark)))
	}
	slot.Child(Span().Set(clsBrandName.AsAttr()).Text(p.Brand.BrandName()))
	slot.OnClick(func(Event) { p.goHome() })
	return slot
}

func (p *Platform) userMenu() Component {
	var actions Component
	if p.UserActions != nil {
		actions = p.UserActions()
	}
	return &usermenu.UserMenu{
		Name:     p.User.UserName(),
		Avatar:   p.User.UserAvatar(),
		Roles:    p.User.UserRoles(),
		Fallback: IconUser,
		Actions:  actions,
	}
}

func (p *Platform) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())

	// ── header ───────────────────────────────────────────────────────────────
	// Three parts: who the platform is (brand), what it is telling them
	// (messages), and who is logged in (user menu). None of it echoes the route
	// — the module names itself, and repeating that here says nothing new.
	header := Header().Set(clsHeader.AsAttr())

	// The leading slot: the brand's mark and name. Mirror of the user menu at
	// the other end, so the two edges of the header read as the same kind of
	// object — facts in, rendering owned by the shell.
	if p.Brand != nil {
		header.Child(p.brand())
	}

	// Chassis singleton: one per page by construction, so a stable global id is
	// the intended use of ID() (see webtyp/dom, "Element ids are owned by dom").
	msgSlot := Div().Set(clsMsgSlot.AsAttr()).ID("pd-msg-slot").
		BindChildren(p.notifications)
	// Because elementToHTML/SSR doesn't process "children" bindings, initial nodes must be manually added
	for _, n := range p.notifications.Get() {
		msgSlot.Child(n)
	}
	header.Child(msgSlot)

	right := Div().Set(clsHeaderRight.AsAttr())
	if p.User != nil {
		right.Child(p.userMenu())
	}

	header.Child(right)

	root.Child(header)

	// ── mobile message stack (mobile only) ─────────────────────────────────
	// The toasts' phone home, a sibling of the header for the same reason the
	// hamburger is one: on a phone the header is display:none, and a fixed
	// descendant of a hidden ancestor is not rendered either (that is exactly
	// why the desktop msg-slot, inside the header, never painted on mobile).
	//
	// The hamburger rides INSIDE this wrapper so the stack can claim the same
	// corner the button alone used to occupy — Docked to the top-end with the
	// same Space4 gap — with the toasts hanging just below it through
	// Stack(Space2). Two floating pieces pinned to the same corner would need
	// an offset calculation to stay apart; one stack with both inside is
	// positioned once.
	//
	// The slot is a dedicated child, never a sibling of the hamburger inside a
	// shared BindChildren: the keyed reconcile treats every existing child of
	// a bound container as a toast row, so a static sibling would be
	// reordered by it and finally removed as excess. The slot declares its own
	// Stack for the inter-toast gap — the wrapper's gap belongs between the
	// button and the block, not inside it.
	msgStack := Div().Set(clsMsgStack.AsAttr())

	// Chassis singleton: one per page by construction, so a stable global id is
	// the intended use of ID() (see webtyp/dom, "Element ids are owned by dom").
	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menu").
		// Open aquí significa "el cromo está desplegado": el botón se pinta
		// mientras NO está guardado por scroll Y el cajón NO está abierto. Con
		// el cajón abierto el botón sería redundante (el cajón ya ocupa dos
		// tercios de la pantalla) y se cerraría tocando el velo, no el botón.
		BindStateFunc(widget.Open, func() bool { return !p.navStowed.Get() && !p.menuOpen.Get() }).
		BindChildren(p.navIcon).
		ID("pd-hamburger-btn")
	hamburger.OnClick(func(Event) {
		p.menuOpen.Toggle()
	})

	msgSlotMobile := Div().Set(clsMsgSlotMobile.AsAttr()).
		BindChildren(p.notificationsMobile)
	// Because elementToHTML/SSR doesn't process "children" bindings, initial
	// nodes must be manually added — same as the desktop slot.
	for _, n := range p.notificationsMobile.Get() {
		msgSlotMobile.Child(n)
	}

	msgStack.Child(hamburger, msgSlotMobile)
	// Added as the root's LAST child below — see that site for why DOM order
	// matters there (the stack ties with the drawer/overlay at --z-dropdown).

	// ── nav overlay backdrop (mobile) ────────────────────────────────────────
	overlay := Div().Set(clsNavOverlay.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	overlay.OnClick(func(Event) {
		p.menuOpen.Set(false)
	})
	root.Child(overlay)

	// ── body (Sidebar wrapping stage and nav) ────────────────────────────────
	body := Div().Set(clsBody.AsAttr())

	// ── main stage ───────────────────────────────────────────────────────────
	stage := Main().Set(clsStage.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		panel := Section().Set(clsPanel.AsAttr()).
			Key(id).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if v := m.View(); v != nil {
			panel.Child(v)
		}
		stage.Child(panel)
	}
	body.Child(stage)

	// ── navigation menu (rail) ───────────────────────────────────────────────
	nav := Nav().Set(clsMenu.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	navbar := Ul().Set(clsNavbar.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if icon := m.Icon(); icon != "" {
			link.Child(icon.Render(string(ClsNavIcon)))
		}
		link.Child(Span().Set(clsLinkText.AsAttr()).Text(m.Label()))

		navbar.Child(Li().Set(clsNavItem.AsAttr()).Child(link))
	}

	// The drawer's head. On a phone this is where the brand rides — there is
	// no header to hold it — and on a wide screen the header carries it while
	// this panel floats out above the rail on hover.
	// One panel holding the head and the navbar. On a wide screen the whole
	// panel is what floats out over the content on hover — if the head expanded
	// inside the rail's flow instead, the rail would widen and push the stage.
	drawerPanel := Div().Set(clsDrawerPanel.AsAttr())

	// The brand leads the drawer and doubles as its "go home" control — tap
	// the mark, land back on the root module, the same gesture the header
	// offers. Wrapped in its own OnlyOn(Mobile) part: on a wide screen the
	// header already carries the brand, top-left, at all times — showing it
	// again in the hover-revealed rail panel would say the app's own name
	// twice on screen at once. AppName is the fallback for a platform with no
	// Brand: plain text, since there is no mark to make tappable. Brand and
	// AppName never render together — a phone showing both would say the
	// app's own name twice in the same panel — and AppName deliberately has
	// no CueWithin to reveal it on a wide screen either: the drawer opens
	// wholesale on a phone so a line of text costs nothing, while in the
	// collapsed rail it would have to materialise when the rail expands and
	// push everything under it down.
	if p.Brand != nil {
		drawerPanel.Child(Div().Set(clsDrawerBrand.AsAttr()).Child(p.brand()))
	} else if p.AppName != "" {
		drawerPanel.Child(Div().Set(clsAppName.AsAttr()).Text(p.AppName))
	}

	drawerPanel.Child(navbar)

	// The identity block trails the nav, not leads it: on open, what to DO
	// matters more than who is signed in. It is a second menu instance, for
	// the phone — there is no header there to hold the first one. This costs
	// nothing now that the contract returns strings — two elements built from
	// the same facts, of which exactly one is ever visible. It was impossible
	// while identity was a Component slot: one element has one parent.
	// Wrapped in a part this package can switch off: the component's own class
	// belongs to the component, and the shell needs somewhere to say "not on a
	// wide screen, where the header already has one".
	if p.User != nil {
		drawerPanel.Child(Div().Set(clsDrawerIdentity.AsAttr()).Child(p.userMenu()))
	}

	nav.Child(drawerPanel)
	body.Child(nav)

	root.Child(body)

	// The mobile toast stack is the root's LAST child: its Docked(Viewport)
	// ties with the drawer and the nav-overlay at --z-dropdown (platformd is
	// a Menu-kind widget, so --z-toast is not reachable from here), and the
	// cascade breaks the tie by DOM order — a toast must paint above the
	// drawer's veil, not under it.
	root.Child(msgStack)

	p.armIdle(root)

	return root
}

// toastNodes builds one toast element per raw notification. Called twice per
// refresh with different suffixes: the same notification renders into the
// desktop header slot AND the mobile stack, and the two must be distinct
// *Elements — one element cannot have two parents — so their IDs/keys differ
// by suffix and each BindChildren reconciles its own set.
//
// Accessibility: role=status (aria-live polite) announces info/success
// without stealing focus; role=alert (assertive) is for warnings and errors,
// which are announcements, not focus grabs. This is the only thing severity
// still drives — every message shares one visual class (clsMsg, styled once
// in css.go with no per-type variant) after a reported bug where the mobile
// and desktop toasts, and the delete-confirmation dialog, each showed a
// different color because each was its own declaration. role isn't a CSS
// class, so it never risked the same drift.
//
// The toast itself is the manual dismiss affordance — tapping it closes it.
// Pause-on-hover lets a long message hold the screen; focusin/focusout mirror
// it for the keyboard even though toasts are deliberately not focusable (a
// focusable toast would add tab stops that appear and vanish with each
// notification).
func (p *Platform) toastNodes(suffix string) []*Element {
	p.mu.Lock()
	defer p.mu.Unlock()

	nodes := make([]*Element, 0, len(p.rawNotifications))
	for _, n := range p.rawNotifications {
		role := "status"
		switch n.Type {
		case Msg.Warning, Msg.Error:
			role = "alert"
		}

		id := n.ID
		nodes = append(nodes, Div().Set(clsMsg.AsAttr()).
			Key(id+suffix).
			Attr("role", role).
			Text(n.Msg).
			OnClick(func(Event) { p.dismiss(id) }).
			OnMouseEnter(func(Event) { p.pauseToast(id) }).
			OnMouseLeave(func(Event) { p.resumeToast(id) }).
			OnFocusIn(func(Event) { p.pauseToast(id) }).
			OnFocusOut(func(Event) { p.resumeToast(id) }))
	}
	return nodes
}

// Notify queues a typed notification in both viewport slots (header on
// desktop, msg-stack on mobile). The duration is a decision, not a number:
// Auto() sizes it to the message, Persistent() leaves it until dismissed,
// For(ms) pins it. Errors are the one case that must not vanish on their own
// — hand them Persistent(), or a generous For().
func (p *Platform) Notify(t MessageType, msg string, d Duration) {
	p.mu.Lock()
	ms := 0
	if d.millis != nil {
		ms = d.millis(msg)
	}
	n := notification{
		Type: t,
		Msg:  msg,
		ID:   "pd-notification-" + p.GetID() + "-" + Sprint(time.Now()),
	}
	if ms > 0 {
		n.expiryNs = time.Now() + int64(ms)*1e6
		n.timer = time.AfterFunc(ms, func() { p.dismiss(n.ID) })
	}
	p.rawNotifications = append(p.rawNotifications, n)
	p.mu.Unlock()

	p.notifications.Set(p.toastNodes(""))
	p.notificationsMobile.Set(p.toastNodes("-m"))
}

// dismiss removes the notification with the given id — from a tap on the
// toast itself, or from the auto-dismiss timer firing. The timer is stopped
// on the manual path so it cannot fire later against a removed notification
// (the auto path's timer has already fired).
func (p *Platform) dismiss(id string) {
	p.mu.Lock()
	for i, n := range p.rawNotifications {
		if n.ID == id {
			if n.timer != nil {
				n.timer.Stop()
				n.timer = nil
			}
			p.rawNotifications = append(p.rawNotifications[:i], p.rawNotifications[i+1:]...)
			p.mu.Unlock()
			p.notifications.Set(p.toastNodes(""))
			p.notificationsMobile.Set(p.toastNodes("-m"))
			return
		}
	}
	p.mu.Unlock()
}

// pauseToast stops a notification's auto-dismiss timer, remembered so the
// pause is transparent: the deadline stays fixed, and resume re-arms with
// whatever remains of it. Hovering or focusing a toast is the user saying
// "I'm still reading this" — the timer must not run through that.
func (p *Platform) pauseToast(id string) {
	p.mu.Lock()
	for i := range p.rawNotifications {
		n := &p.rawNotifications[i]
		if n.ID != id {
			continue
		}
		if n.timer != nil {
			n.timer.Stop()
			n.timer = nil
		}
		break
	}
	p.mu.Unlock()
}

// resumeToast re-arms the auto-dismiss timer for the time remaining on the
// original deadline. A persistent notification (no deadline) or an already
// dismissed one is a no-op.
func (p *Platform) resumeToast(id string) {
	p.mu.Lock()
	for i := range p.rawNotifications {
		n := &p.rawNotifications[i]
		if n.ID != id {
			continue
		}
		if n.expiryNs > 0 && n.timer == nil {
			if remaining := n.expiryNs - time.Now(); remaining > 0 {
				n.timer = time.AfterFunc(int(remaining/1e6), func() { p.dismiss(n.ID) })
			}
		}
		break
	}
	p.mu.Unlock()
}

func (p *Platform) notificationCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.rawNotifications)
}

// Activate programmatically switches to a module by ID
// (also updates window.location.hash on wasm builds).
func (p *Platform) Activate(moduleID string) {
	if !p.isViewable(moduleID) {
		return
	}

	if p.active.Get() == moduleID && !p.menuOpen.Get() {
		return
	}

	p.active.Set(moduleID)
	p.menuOpen.Set(false)

	// El botón de menú lleva el estado de la navegación: en móvil no hay cabecera
	// ni rail visible, así que su glifo es lo único que dice en qué sección estás.
	p.navIcon.Set([]*Element{p.activeIcon().Render(string(ClsNavIcon)).Key(moduleID)})

	// Cambiar de sección reinicia el cromo: el módulo nuevo empieza desde arriba y
	// con el botón a mano.
	p.lastScrollTop = 0
	p.navStowed.Set(false)

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}
}
