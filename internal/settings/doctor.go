package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/store"
)

// withheld stands in a doctor row for the value of a
// [config.Sensitive] setting. doctor's output is pasted into reports,
// and a setting whose change needs care is one whose value an operator
// should not have to scrub from them.
const withheld = "(withheld)"

// DoctorRows renders doctor's one row per registered setting, in
// declaration order, and one row per orphan, from report and the
// settings facts f its inspection read. A state [store.State.Vouches]
// refuses reports every setting "not checked"; every other state
// evaluates the seed rule against the facts [SeedFacts] picks. A store
// with no settings table yet shows each setting at its default and
// what serve's next start seeds, unless the seed would refuse, which
// fails here as it would there. A [config.Sensitive] setting's value
// never prints.
func DoctorRows(environ []string, reg *config.Registry, report store.Report, f Facts) []doctor.Row {
	list := reg.Settings()
	var rows []doctor.Row
	facts, ok := SeedFacts(report, f)
	if !report.State.Vouches() || !ok {
		for _, s := range list {
			rows = append(rows, doctor.Row{Name: s.Key, Status: notCheckedStatus(environ, s, report.State)})
		}
		return rows
	}

	ev := EvaluateFacts(reg, environ, facts)
	registered := make(map[string]config.Setting, len(list))
	for _, s := range list {
		registered[s.Key] = s
	}
	scrub := sensitiveValues(ev, registered)
	for _, o := range ev.Outcomes {
		s := registered[o.Key]
		row := doctor.Row{Name: o.Key}
		switch {
		case o.Err != nil:
			row.Status, row.Failed = scrub.Replace(o.Report), true
		case facts.Absent:
			row.Status = noTableStatus(environ, s)
		default:
			row.Status = outcomeStatus(s, o)
		}
		rows = append(rows, row)
	}
	for _, orphan := range ev.Orphans {
		rows = append(rows, doctor.Row{Name: orphan.Key, Status: "no longer registered; ignored"})
	}
	return rows
}

// sensitiveValues returns a replacer that withholds from a refusal's
// text every value a [config.Sensitive] setting could print in it: its
// default, its environment value, and its stored and seed values. Every
// refusal quotes the values it names with %q, a correlated rule's
// included, so the replacer matches the quoted form and leaves a value
// inside a longer one alone ("1h" within "17520h"). A non-sensitive
// value equal to a sensitive one is withheld too.
func sensitiveValues(ev Evaluation, registered map[string]config.Setting) *strings.Replacer {
	var pairs []string
	add := func(v string) {
		if v != "" {
			pairs = append(pairs, strconv.Quote(v), withheld)
		}
	}
	for _, o := range ev.Outcomes {
		s := registered[o.Key]
		if s.Sensitivity != config.Sensitive {
			continue
		}
		add(s.Default)
		if o.EnvSet {
			add(o.Env)
		}
		if o.Stored != nil {
			add(o.Stored.Value)
			add(o.Stored.SeedValue)
		}
	}
	return strings.NewReplacer(pairs...)
}

// shown renders value as s's row prints it: quoted, or withheld for a
// [config.Sensitive] setting.
func shown(s config.Setting, value string) string {
	if s.Sensitivity == config.Sensitive {
		return withheld
	}
	return strconv.Quote(value)
}

// outcomeStatus renders one from-facts outcome as the status for the
// registered setting s: the value in force, its source, and the apply
// mode, then any note about what changes at serve's next start. A
// refused outcome never reaches it.
func outcomeStatus(s config.Setting, o Outcome) string {
	value, source := s.Default, "default"
	if o.Stored != nil {
		value, source = o.Stored.Value, o.Stored.Source
	}
	status := shown(s, value) + " " + traits(s, source)
	if note := changeNote(s, o); note != "" {
		status += "; " + note
	}
	return status
}

// changeNote renders what o's change does at serve's next start, or ""
// for an outcome the seed rule leaves unchanged.
func changeNote(s config.Setting, o Outcome) string {
	switch o.Change {
	case Seed:
		return fmt.Sprintf("env %s seeds at serve's next start", shown(s, o.Env))
	case Reseed:
		return fmt.Sprintf("env %s re-seeds at serve's next start", shown(s, o.Env))
	case Remove:
		return "returns to the default at serve's next start"
	default:
		return ""
	}
}

// noTableStatus renders s's row for a store that holds no settings
// table yet, or no store at all: the default is in force, and an env
// value the seed accepts seeds at serve's next start.
func noTableStatus(environ []string, s config.Setting) string {
	status := shown(s, s.Default) + " " + traits(s, "default") + "; store holds no settings yet"
	if v, ok := config.Lookup(environ, s.Key); ok {
		status += fmt.Sprintf("; env %s seeds at serve's next start", shown(s, v))
	}
	return status
}

// notCheckedStatus renders s's row for a store state this binary
// cannot vouch for: the default is shown for reference, since no read
// reached the store's row, and an env value is named as not checked
// rather than compared. The zero state, which [store.Doctor] reports
// when it never ran [store.Inspect], names no reason, and the row says
// only "not checked".
func notCheckedStatus(environ []string, s config.Setting, state store.State) string {
	label := "not checked"
	if state != 0 {
		label = fmt.Sprintf("not checked (%s)", state)
	}
	status := shown(s, s.Default) + " " + traits(s, "default") + "; store " + label
	if v, ok := config.Lookup(environ, s.Key); ok {
		status += fmt.Sprintf("; env %s not checked against the store", shown(s, v))
	}
	return status
}

// traits renders the parenthesized traits doctor prints after a
// setting's value: the value's source, s's apply mode, and the word
// sensitive when s is [config.Sensitive].
func traits(s config.Setting, source string) string {
	if s.Sensitivity == config.Sensitive {
		return fmt.Sprintf("(source %s, %s, sensitive)", source, s.Apply)
	}
	return fmt.Sprintf("(source %s, %s)", source, s.Apply)
}
