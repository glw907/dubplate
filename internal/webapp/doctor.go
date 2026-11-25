package webapp

import (
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
)

// DoctorRows renders the web row for cfg, which is nil when
// [config.Load] refused the environment, from resolve, the
// construction serve startup runs over the same inputs.
func DoctorRows(cfg *config.Config, resolve func(*config.Config) error) []doctor.Row {
	row := doctor.Row{Name: "web"}
	row.Status, row.Failed = rowStatus(cfg, resolve)
	return []doctor.Row{row}
}

// rowStatus renders the web row from resolve, the construction
// serve startup runs over the same inputs. The command supplies it, so
// doctor and serve cannot disagree about the theme match, the palette,
// the branding value, or the operator stylesheet. A nil error reads
// "ok". A refusal reads "invalid: " plus the error, with the lines of
// an [errors.Join] folded to "; ", and fails the row. A nil cfg
// ([config.Load] refused the environment) reads "not checked", adds no
// failure, and never calls resolve: Load's report already fails
// doctor.
func rowStatus(cfg *config.Config, resolve func(*config.Config) error) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	if err := resolve(cfg); err != nil {
		return "invalid: " + strings.ReplaceAll(err.Error(), "\n", "; "), true
	}
	return "ok", false
}
