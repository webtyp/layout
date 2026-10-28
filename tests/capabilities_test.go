package tests

import (
	"testing"

	"webtyp.com/components/countbadge"
	"webtyp.com/dom"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/layout/platformd"
)

type describedModule struct {
	platformd.UIModule
}

func (d *describedModule) Description() lang.Text {
	return "test description"
}

func (d *describedModule) Steps() []platformd.Step {
	return []platformd.Step{
		{Label: "step 1", Done: true},
	}
}

type alertedModule struct {
	platformd.UIModule
}

func (a *alertedModule) UnseenAlerts() []platformd.ModuleAlert {
	return []platformd.ModuleAlert{
		{
			ID:      "alert-1",
			Subject: "test alert",
			OnView:  func() {},
			OnSeen:  func() {},
		},
	}
}

func TestCapabilities(t *testing.T) {
	var _ platformd.Described = (*describedModule)(nil)
	var _ platformd.AlertedModule = (*alertedModule)(nil)

	desc := &describedModule{}
	if desc.Description() != "test description" {
		t.Errorf("expected 'test description', got '%s'", desc.Description())
	}
	steps := desc.Steps()
	if len(steps) != 1 || steps[0].Label != "step 1" || !steps[0].Done {
		t.Errorf("unexpected steps: %+v", steps)
	}

	alert := &alertedModule{}
	alerts := alert.UnseenAlerts()
	if len(alerts) != 1 || alerts[0].ID != "alert-1" || alerts[0].Subject != "test alert" {
		t.Errorf("unexpected alerts: %+v", alerts)
	}
}

type badgedComponent struct {
	dom.Element
	badge *countbadge.CountBadge
}

func (b *badgedComponent) Badge() *countbadge.CountBadge { return b.badge }

func TestUIModule_BadgedForwarding(t *testing.T) {
	cb := &countbadge.CountBadge{}
	c := &badgedComponent{badge: cb}
	m := layout.NewUIModule("chat", "Chat", "", c)
	b, ok := m.(layout.Badged)
	if !ok {
		t.Fatalf("expected NewUIModule to implement Badged")
	}
	if b.Badge() != cb {
		t.Fatalf("expected Badge() to return component's badge")
	}
}
