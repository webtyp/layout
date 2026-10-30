// Root-level test (justified): the source's load function is unexported.
package schemaview

import (
	"testing"
	"webtyp.com/erd"
)

func TestNewWithoutSource(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for New(nil), got nil")
	}
	if err.Error() != errNoSource {
		t.Fatalf("expected error '%s', got '%s'", errNoSource, err.Error())
	}
}

func TestFromReportLoads(t *testing.T) {
	report := erd.Report{SVG: "<svg></svg>"}
	src := FromReport(report)

	var loaded erd.Report
	var loadedErr error
	called := false

	src.load(func(r erd.Report, err error) {
		loaded = r
		loadedErr = err
		called = true
	})

	if !called {
		t.Fatal("expected load to call done synchronously")
	}
	if loadedErr != nil {
		t.Fatalf("expected no error, got %v", loadedErr)
	}
	if loaded.SVG != report.SVG {
		t.Errorf("expected loaded SVG %q, got %q", report.SVG, loaded.SVG)
	}
}
