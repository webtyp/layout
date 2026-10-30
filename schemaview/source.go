package schemaview

import (
	"webtyp.com/erd"
	"webtyp.com/fetch"
	"webtyp.com/json"
	"webtyp.com/lang"

	. "webtyp.com/fmt"
)

const errNoSource = "schemaview: no source: use FromServer or FromReport"

// Source says where the diagram comes from. Build it with FromServer or FromReport.
type Source struct {
	load func(done func(erd.Report, error))
}

// FromServer reads the diagram the app's server serves at erd.SchemaPath.
// Example: app.Tools{Database: schemaview.FromServer()}
func FromServer() *Source { return &Source{load: fetchReport(erd.SchemaPath)} }

// FromReport shows a diagram built in process (a module's demo).
// Example: schemaview.FromReport(diagram.Report())
func FromReport(r erd.Report) *Source {
	return &Source{load: func(done func(erd.Report, error)) { done(r, nil) }}
}

func fetchReport(url string) func(done func(erd.Report, error)) {
	return func(done func(erd.Report, error)) {
		fetch.Get(url).Send(func(resp *fetch.Response, err error) {
			if err != nil {
				done(erd.Report{}, err)
				return
			}
			if resp.Status == 403 {
				done(erd.Report{}, Err(lang.Translate("no access to", "database diagram").String()))
				return
			}
			if resp.Status != 200 {
				done(erd.Report{}, Err(lang.Translate("Could not", "load", "database diagram").String()+" "+Sprintf("%d", resp.Status)))
				return
			}
			var rep erd.Report
			if err := json.Decode(resp.Body(), &rep); err != nil {
				done(erd.Report{}, Err(lang.Translate("Could not", "load", "database diagram").String()+": "+err.Error()))
				return
			}
			done(rep, nil)
		})
	}
}
