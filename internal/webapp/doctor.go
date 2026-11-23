package webapp

import (
	"strings"

	"github.com/glw907/dubplate/internal/config"
)

// RowStatus renders doctor's "web" row from resolve, the construction
// serve startup runs over the same inputs. The command supplies it, so
// doctor and serve cannot disagree about the theme match, the palette,
// the branding value, or the operator stylesheet. A nil error reads
// "ok". A refusal reads "invalid: " plus the error, with the lines of
// an [errors.Join] folded to "; ", and fails the row. A nil cfg
// ([config.Load] refused the environment) reads "not checked", adds no
// failure, and never calls resolve: Load's report already fails
// doctor.
func RowStatus(cfg *config.Config, resolve func(*config.Config) error) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	if err := resolve(cfg); err != nil {
		return "invalid: " + strings.ReplaceAll(err.Error(), "\n", "; "), true
	}
	return "ok", false
}
