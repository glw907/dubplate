package main

import (
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// reauthWindowFloor is the re-authentication window's floor. A session's
// anchor is set when its magic link is consumed, so the round trip that
// matters runs from that click to the retried write; five minutes covers
// it, and a lower window would let the web lock itself out of resetting
// the window. [reauthWindowSetting] spells it as its Floor, in the
// form a refusal prints.
const reauthWindowFloor = 5 * time.Minute

// reauthWindowSetting is how recently a session must have been signed in
// for a sensitive settings write, or any member add, removal, or
// demotion, over the web. Its two-hour default is GitHub sudo mode's
// window; a day is the ceiling, past which the gate stops meaning "a
// fresh sign-in". It is itself sensitive, since widening it weakens the
// gate.
var reauthWindowSetting = config.Setting{
	Key:         "DUBPLATE_REAUTH_WINDOW",
	Kind:        config.KindDuration,
	Default:     "2h",
	Doc:         "a Go duration string, for example 2h",
	Apply:       config.ApplyNextOperation,
	Floor:       "5m",
	Ceiling:     "24h",
	Description: "how recently an admin must have signed in to change who is in the group or a sensitive setting",
	Mark:        config.Sensitive,
}

// reauthWindowBelowCap holds the window strictly below the session
// absolute cap. No session lives past the cap, so a window at or above
// it would admit every live session and make the gate a no-op. The
// refusal names both remedies, unless the cap sits at or below the
// window's floor, where raising the cap is the only one.
var reauthWindowBelowCap = config.CorrelatedRule{
	Name: "re-authentication window below the absolute cap",
	Keys: []string{reauthWindowSetting.Key, sessionAbsoluteCapSetting.Key},
	Check: func(values map[string]string) error {
		window, err := time.ParseDuration(values[reauthWindowSetting.Key])
		if err != nil {
			return fmt.Errorf("%s: %w", reauthWindowSetting.Key, err)
		}
		limit, err := time.ParseDuration(values[sessionAbsoluteCapSetting.Key])
		if err != nil {
			return fmt.Errorf("%s: %w", sessionAbsoluteCapSetting.Key, err)
		}
		if window < limit {
			return nil
		}
		refusal := fmt.Sprintf("%s %q is not below %s %q", reauthWindowSetting.Key, values[reauthWindowSetting.Key],
			sessionAbsoluteCapSetting.Key, values[sessionAbsoluteCapSetting.Key])
		if limit <= reauthWindowFloor {
			return fmt.Errorf("%s: raise %s above %s, the window's floor", refusal, sessionAbsoluteCapSetting.Key, reauthWindowFloor)
		}
		return fmt.Errorf("%s: lower %s below the cap, or raise %s", refusal, reauthWindowSetting.Key, sessionAbsoluteCapSetting.Key)
	},
}

// registerAPIConfig adds the admin API's configuration to reg: the
// re-authentication window and its rule against the absolute cap, which
// registerAuthConfig declares and so must run first.
func registerAPIConfig(reg *config.Registry) {
	reg.AddSetting(reauthWindowSetting)
	reg.AddCorrelatedRule(reauthWindowBelowCap)
}
