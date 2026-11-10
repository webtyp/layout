// Root-level test (justified): pure logic tests must test unexported functions, which the tests/ package cannot access.
package apiexplorer

import (
	"testing"
	"webtyp.com/dom"
	"webtyp.com/router"
)

func TestExtractPathParams(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{"/api/sites/{site}/assets/{key}", []string{"site", "key"}},
		{"/api/users", nil},
		{"/{id}", []string{"id"}},
	}

	for _, c := range cases {
		got := extractPathParams(c.path)
		if len(got) != len(c.want) {
			t.Errorf("extractPathParams(%q) len = %d, want %d", c.path, len(got), len(c.want))
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("extractPathParams(%q)[%d] = %q, want %q", c.path, i, got[i], c.want[i])
			}
		}
	}
}

func TestBuildJSON(t *testing.T) {
	args := []router.ArgRecord{
		{Name: "username", Kind: "string"},
		{Name: "age", Kind: "int"},
		{Name: "active", Kind: "bool"},
	}

	sigText := dom.NewString("jules")
	sigInt := dom.NewString("42")
	sigBool := dom.NewString("true")

	got := buildJSON(args, []*dom.SignalString{sigText, sigInt, sigBool})
	want := `{"username":"jules","age":42,"active":true}`

	if got != want {
		t.Errorf("buildJSON = %q, want %q", got, want)
	}

	sigIntEmpty := dom.NewString("")
	gotEmpty := buildJSON([]router.ArgRecord{{Name: "age", Kind: "int"}}, []*dom.SignalString{sigIntEmpty})
	wantEmpty := `{"age":0}`
	if gotEmpty != wantEmpty {
		t.Errorf("buildJSON empty number = %q, want %q", gotEmpty, wantEmpty)
	}
}

func TestBuildQuery_EscapesAndSkipsEmpty(t *testing.T) {
	args := []router.ArgRecord{
		{Name: "username", Kind: "string"},
		{Name: "age", Kind: "int"},
		{Name: "active", Kind: "bool"},
		{Name: "note", Kind: "string"},
	}
	signals := []*dom.SignalString{dom.NewString("jules ñ"), dom.NewString("42"), dom.NewString("true"), dom.NewString("")}

	got := buildQuery(args, signals)
	want := "username=jules%20%C3%B1&age=42&active=true"
	if got != want {
		t.Errorf("buildQuery = %q, want %q", got, want)
	}
}

func TestBuildQuery_NoArgsIsEmpty(t *testing.T) {
	if got := buildQuery(nil, nil); got != "" {
		t.Errorf("buildQuery without args = %q, want empty", got)
	}
}
