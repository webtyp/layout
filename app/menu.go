package app

import (
	"webtyp.com/fmt"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/layout/apiexplorer"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
	"webtyp.com/view"
)

const helpAPIExplorer lang.Text = "Every operation of this app: who may call it, and a form to try it."

// Menu is what layout/app hands to any layout chassis (such as platformd or a custom layout):
// the modules and the visibility rule.
type Menu struct {
	Modules []layout.UIModule
	CanView func(id string) bool
}

type screenPerm struct {
	id         string
	permission model.Resource
}

// BuildMenu validates every screen and builds its layout.UIModule tree. allows answers whether the
// signed-in user holds a permission (in an app: func(r model.Resource) bool { return profile.Can(r, model.Read) }).
// The API explorer is always added last, visible to whoever may read router.ResourceIntrospection; the app never adds it.
func BuildMenu(screens []view.Screen, ids model.IDGenerator, allows func(model.Resource) bool) (Menu, error) {
	for _, s := range screens {
		if err := s.Validate(); err != nil {
			return Menu{}, err
		}
	}

	var perms []screenPerm
	mods := make([]layout.UIModule, 0, len(screens))
	for _, s := range screens {
		m, err := buildScreen(s, ids, &perms)
		if err != nil {
			return Menu{}, err
		}
		mods = append(mods, m)
	}

	explorer, err := apiexplorer.New(apiexplorer.Config{})
	if err != nil {
		return Menu{}, err
	}
	mods = append(mods, layout.Describe(explorer, helpAPIExplorer))
	perms = append(perms, screenPerm{id: explorer.ModelName(), permission: router.ResourceIntrospection})

	canView := func(id string) bool {
		if allows == nil {
			return false
		}
		for _, p := range perms {
			if p.id == id {
				return allows(p.permission)
			}
		}
		return false
	}

	return Menu{
		Modules: mods,
		CanView: canView,
	}, nil
}

func buildScreen(s view.Screen, ids model.IDGenerator, perms *[]screenPerm) (layout.UIModule, error) {
	if len(s.Children) > 0 {
		children := make([]layout.UIModule, 0, len(s.Children))
		for _, c := range s.Children {
			ch, err := buildScreen(c, ids, perms)
			if err != nil {
				return nil, err
			}
			children = append(children, ch)
		}
		grp := layout.NewGroupUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), children)
		return layout.Describe(grp, s.Help), nil
	}

	if s.Component != nil {
		*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
		m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), s.Component)
		return layout.Describe(m, s.Help), nil
	}

	v, err := crudview.New(crudview.Config{
		ParentID:  s.ID,
		Presenter: s.Presenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}

	lookups := s.Lookups
	activate := onFirstActivation(v, func() {
		for _, l := range lookups {
			fillOptions(v, l.Field, l.Source)
		}
	})

	*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
	m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), v, activate)
	if desc := s.Presenter.Description(); desc != "" {
		m = layout.Describe(m, desc)
	}
	return m, nil
}

func onFirstActivation(v *crudview.CrudView, load func()) func() {
	loaded := false
	return func() {
		if !loaded {
			loaded = true
			load()
		}
		v.Activate()
	}
}

func fillOptions(v *crudview.CrudView, field string, source view.Presenter) {
	source.Reload(func(err error) {
		if err != nil {
			return
		}
		items := source.Items()
		opts := make([]fmt.KeyValue, 0, len(items))
		for _, it := range items {
			opts = append(opts, fmt.KeyValue{Key: it.ID, Value: it.Label})
		}
		v.SetOptions(field, opts...)
	})
}
