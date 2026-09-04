package conform

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/analysis"
)

// componentMethods is conform's own decoder-method set: the mirror of
// analysis.CheckComponentTable's coverage check, maintained separately
// because analysis.componentMethods is unexported and behind the
// territory wall (fold, M-g). Two independently maintained lists
// stating the same fact is the point: a disagreement between them is a
// gate failure, not a silent skip. Every method here must have at
// least one replayed row (TestComponentTableReplaysAllRows), and every
// row's method must be one of these (ReplayRow).
var componentMethods = []string{
	"add",
	"add_data_source",
	"add_equality",
	"add_expr",
	"add_number",
	"add_priority",
	"add_ratio",
	"add_string",
	"generic_penalty_keys",
	"string_dist",
	"track_distance",
}

// ReplayOutcome is the result of replaying one components.json row
// against the matcher primitive its method names.
type ReplayOutcome struct {
	// RowID is the row's own id, named in every error this outcome
	// carries.
	RowID string
	// Method is the row's method, once it resolved to one conform
	// knows; empty when the method itself was unrecognized.
	Method string
	// Err is non-nil when the row failed outright: an unknown method,
	// an unresolvable argument shape, or an output that disagrees with
	// the committed row beyond the 1e-12 fallback (contract 12).
	Err error
	// NearMiss is true when every field matched only within the
	// 1e-12 relative fallback and none was bit-identical: a finding to
	// report, not a failure (contract 12).
	NearMiss bool
	// Detail names the field or fields that were near misses, for the
	// pass report's finding list.
	Detail string
}

// ReplaySummary is the outcome of replaying an entire component table.
type ReplaySummary struct {
	// CountByMethod is the number of rows successfully replayed under
	// each method conform dispatched to. A row that failed outright is
	// not counted here, so a decoder that silently swallows a bad row
	// cannot inflate this total.
	CountByMethod map[string]int
	// NearMisses lists every row that matched only within the 1e-12
	// fallback (contract 12's own reporting duty).
	NearMisses []ReplayOutcome
	// TotalRows is len(table.Rows), for the "nothing was skipped"
	// check: total replayed must equal this.
	TotalRows int
}

// ReplayTable replays every row in t and returns the per-method tally,
// the near-miss findings, and every row failure joined into one error.
// A row that fails is excluded from CountByMethod, which is what makes
// "replayed count == table row count" catch a silently skipped row.
func ReplayTable(t *analysis.ComponentTable) (ReplaySummary, error) {
	summary := ReplaySummary{CountByMethod: make(map[string]int), TotalRows: len(t.Rows)}
	var errs []error
	for _, row := range t.Rows {
		outcome := ReplayRow(row)
		if outcome.Err != nil {
			errs = append(errs, outcome.Err)
			continue
		}
		summary.CountByMethod[outcome.Method]++
		if outcome.NearMiss {
			summary.NearMisses = append(summary.NearMisses, outcome)
		}
	}
	return summary, errors.Join(errs...)
}

// ReplayRow decodes row per its method (dispatching on method, and
// inside method "add" on the argument shape: eleven methods, twelve
// shapes), replays it against the matcher, and compares the result
// against row.Output. An unknown method or an unresolvable shape is a
// failure naming row.ID, never a skip.
func ReplayRow(row analysis.ComponentRow) (outcome ReplayOutcome) {
	defer func() {
		if r := recover(); r != nil {
			outcome = ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: panic: %v", row.ID, r)}
		}
	}()

	if !slices.Contains(componentMethods, row.Method) {
		return ReplayOutcome{RowID: row.ID, Err: fmt.Errorf("component %s: unknown method %q", row.ID, row.Method)}
	}

	switch row.Method {
	case "add":
		return replayAddMethod(row)
	case "add_equality", "add_expr", "add_number", "add_priority", "add_ratio", "add_string":
		var args adderArgs
		if err := json.Unmarshal(row.Args, &args); err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
		}
		return replayAdderShape(row, args)
	case "generic_penalty_keys":
		var args aggregateArgs
		if err := json.Unmarshal(row.Args, &args); err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
		}
		return replayGenericPenaltyKeys(row, args)
	case "string_dist":
		return replayStringDist(row)
	case "track_distance":
		return replayTrackDistance(row)
	case "add_data_source":
		return replayAddDataSource(row)
	default:
		// Unreachable: componentMethods was checked above, and every
		// entry in it is handled by one of the cases.
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: no dispatcher registered for method %q", row.ID, row.Method)}
	}
}

// replayAddMethod splits method "add"'s two incompatible argument
// shapes: args.key names the adder shape (a replay of Distance.Add
// itself), args.weights names the aggregate shape (the raw, max, and
// normalized arithmetic over a set of added penalties). A row carrying
// neither, or both, is a failure naming the row id (fold, B1).
func replayAddMethod(row analysis.ComponentRow) ReplayOutcome {
	var probe struct {
		Key     *string             `json:"key"`
		Weights *map[string]float64 `json:"weights"`
	}
	if err := json.Unmarshal(row.Args, &probe); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
	}
	switch {
	case probe.Key != nil && probe.Weights == nil:
		var args adderArgs
		if err := json.Unmarshal(row.Args, &args); err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
		}
		return replayAdderShape(row, args)
	case probe.Weights != nil && probe.Key == nil:
		var args aggregateArgs
		if err := json.Unmarshal(row.Args, &args); err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
		}
		return replayAggregateShape(row, args)
	default:
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf(`component %s: method "add" args carry neither key nor weights, or both`, row.ID)}
	}
}

// The adder shape: method "add" with args.key, plus add_equality,
// add_expr, add_number, add_priority, add_ratio, and add_string. Each
// row is a replay of the full call prefix against a fresh accumulator,
// and the output is that key's raw penalty slice in append order
// (derive_component.py:241-254). args.key is the literal penalty key
// the adder was called with, not a beets weight key, so replay scores
// against plain default tunables: no view in this shape ever consults
// a weight.

type adderArgs struct {
	Key   string            `json:"key"`
	Calls []json.RawMessage `json:"calls"`
}

type adderOutput struct {
	Penalties []float64 `json:"penalties"`
}

// adderCall applies one call, already split into its raw JSON
// elements, to dist under key.
type adderCall func(dist *matcher.Distance, key string, elems []json.RawMessage) error

var adderCalls = map[string]adderCall{
	"add":          callAdd,
	"add_equality": callAddEquality,
	"add_expr":     callAddExpr,
	"add_number":   callAddNumber,
	"add_priority": callAddPriority,
	"add_ratio":    callAddRatio,
	"add_string":   callAddString,
}

func replayAdderShape(row analysis.ComponentRow, args adderArgs) ReplayOutcome {
	call := adderCalls[row.Method]
	dist := matcher.NewDistance(matcher.DefaultTunables())
	for i, raw := range args.Calls {
		elems, err := decodeCallElems(raw)
		if err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: calls[%d]: %w", row.ID, i, err)}
		}
		if err := call(dist, args.Key, elems); err != nil {
			return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: calls[%d]: %w", row.ID, i, err)}
		}
	}

	var want adderOutput
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}
	checks, err := checkFloatSlice(row.ID, "penalties", want.Penalties, dist.Penalties(args.Key))
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: err}
	}
	return checkFields(row.ID, row.Method, checks)
}

func decodeCallElems(raw json.RawMessage) ([]json.RawMessage, error) {
	var elems []json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, err
	}
	return elems, nil
}

func wantArity(method string, elems []json.RawMessage, n int) error {
	if len(elems) != n {
		return fmt.Errorf("%s: want %d arguments, got %d", method, n, len(elems))
	}
	return nil
}

func callAdd(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add", elems, 1); err != nil {
		return err
	}
	var v float64
	if err := json.Unmarshal(elems[0], &v); err != nil {
		return fmt.Errorf("add: value: %w", err)
	}
	dist.Add(key, v)
	return nil
}

func callAddExpr(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_expr", elems, 1); err != nil {
		return err
	}
	var v bool
	if err := json.Unmarshal(elems[0], &v); err != nil {
		return fmt.Errorf("add_expr: value: %w", err)
	}
	dist.AddExpr(key, v)
	return nil
}

func callAddNumber(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_number", elems, 2); err != nil {
		return err
	}
	var n1, n2 int
	if err := json.Unmarshal(elems[0], &n1); err != nil {
		return fmt.Errorf("add_number: n1: %w", err)
	}
	if err := json.Unmarshal(elems[1], &n2); err != nil {
		return fmt.Errorf("add_number: n2: %w", err)
	}
	dist.AddNumber(key, n1, n2)
	return nil
}

func callAddRatio(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_ratio", elems, 2); err != nil {
		return err
	}
	var n1, n2 float64
	if err := json.Unmarshal(elems[0], &n1); err != nil {
		return fmt.Errorf("add_ratio: n1: %w", err)
	}
	if err := json.Unmarshal(elems[1], &n2); err != nil {
		return fmt.Errorf("add_ratio: n2: %w", err)
	}
	dist.AddRatio(key, n1, n2)
	return nil
}

func callAddString(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_string", elems, 2); err != nil {
		return err
	}
	var s1, s2 *string
	if err := json.Unmarshal(elems[0], &s1); err != nil {
		return fmt.Errorf("add_string: str1: %w", err)
	}
	if err := json.Unmarshal(elems[1], &s2); err != nil {
		return fmt.Errorf("add_string: str2: %w", err)
	}
	dist.AddString(key, s1, s2)
	return nil
}

func callAddEquality(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_equality", elems, 2); err != nil {
		return err
	}
	var value string
	if err := json.Unmarshal(elems[0], &value); err != nil {
		return fmt.Errorf("add_equality: value: %w", err)
	}
	options, err := decodeOptions(elems[1])
	if err != nil {
		return fmt.Errorf("add_equality: options: %w", err)
	}
	dist.AddEquality(key, value, options...)
	return nil
}

func callAddPriority(dist *matcher.Distance, key string, elems []json.RawMessage) error {
	if err := wantArity("add_priority", elems, 2); err != nil {
		return err
	}
	var value string
	if err := json.Unmarshal(elems[0], &value); err != nil {
		return fmt.Errorf("add_priority: value: %w", err)
	}
	options, err := decodeOptions(elems[1])
	if err != nil {
		return fmt.Errorf("add_priority: options: %w", err)
	}
	dist.AddPriority(key, value, options...)
	return nil
}

// decodeOptions decodes an add_equality or add_priority option
// argument: either a bare option (a string, or a compiled-regex object)
// or a JSON array of options. A bare option is wrapped into a
// one-element slice, mirroring beets' own scalar wrap (contract 6a);
// AddEquality and AddPriority's variadic parameter is that wrap, so
// passing the decoded slice through options... reproduces it exactly.
func decodeOptions(raw json.RawMessage) ([]any, error) {
	var probe any
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}
	if list, ok := probe.([]any); ok {
		options := make([]any, len(list))
		for i, item := range list {
			opt, err := decodeOption(item)
			if err != nil {
				return nil, fmt.Errorf("options[%d]: %w", i, err)
			}
			options[i] = opt
		}
		return options, nil
	}
	opt, err := decodeOption(probe)
	if err != nil {
		return nil, err
	}
	return []any{opt}, nil
}

// decodeOption decodes one option value: a compiled-regex object
// ({"regex", "ignorecase"}, fold M-e) becomes a matcher.PatternOption,
// anchored at position 0 per contract 7; anything else is a literal
// value compared with Go equality.
func decodeOption(raw any) (any, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return raw, nil
	}
	pattern, _ := m["regex"].(string)
	ignorecase, _ := m["ignorecase"].(bool)
	opt, err := matcher.CompilePatternOption(pattern, ignorecase)
	if err != nil {
		return nil, fmt.Errorf("compile pattern option %q: %w", pattern, err)
	}
	return opt, nil
}

// The aggregate shape: method "add" with args.weights, plus
// generic_penalty_keys, which shares the same args. args.weights is an
// OVERRIDE layered key by key onto matcher.DefaultTunables's
// DistanceWeights, never a complete table (fold, B2):
// derive_component.py's weights() context manager patches
// Distance._weights and restores it afterward, so every key the row
// does not name keeps beets' shipped default. A key beets does not
// ship (zero_weighted) is tolerated: it simply becomes a new map
// entry.

type aggregateArgs struct {
	Weights map[string]float64 `json:"weights"`
	Calls   []json.RawMessage  `json:"calls"` // each a [key, value] pair
}

type aggregateOutput struct {
	Distance    float64            `json:"distance"`
	RawDistance float64            `json:"raw_distance"`
	MaxDistance float64            `json:"max_distance"`
	PerKey      map[string]float64 `json:"per_key"`
	Items       json.RawMessage    `json:"items"`
}

// tunablesWithOverrides returns beets' shipped defaults with overrides
// layered on top key by key, never replacing the table wholesale.
func tunablesWithOverrides(overrides map[string]float64) matcher.Tunables {
	tun := matcher.DefaultTunables()
	maps.Copy(tun.DistanceWeights, overrides)
	return tun
}

// applyAggregateCalls replays calls, each a [key, value] pair, against
// dist in order.
func applyAggregateCalls(dist *matcher.Distance, calls []json.RawMessage) error {
	for i, raw := range calls {
		elems, err := decodeCallElems(raw)
		if err != nil {
			return fmt.Errorf("calls[%d]: %w", i, err)
		}
		if err := wantArity("add", elems, 2); err != nil {
			return fmt.Errorf("calls[%d]: %w", i, err)
		}
		var key string
		var value float64
		if err := json.Unmarshal(elems[0], &key); err != nil {
			return fmt.Errorf("calls[%d]: key: %w", i, err)
		}
		if err := json.Unmarshal(elems[1], &value); err != nil {
			return fmt.Errorf("calls[%d]: value: %w", i, err)
		}
		dist.Add(key, value)
	}
	return nil
}

func replayAggregateShape(row analysis.ComponentRow, args aggregateArgs) ReplayOutcome {
	tun := tunablesWithOverrides(args.Weights)
	dist := matcher.NewDistance(tun)
	if err := applyAggregateCalls(dist, args.Calls); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: %w", row.ID, err)}
	}

	var want aggregateOutput
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}
	wantItems, err := decodeItems(want.Items)
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output.items: %w", row.ID, err)}
	}

	checks := []fieldCheck{
		{Name: "distance", Want: want.Distance, Got: dist.Distance()},
		{Name: "raw_distance", Want: want.RawDistance, Got: dist.RawDistance()},
		{Name: "max_distance", Want: want.MaxDistance, Got: dist.MaxDistance()},
	}
	perKeyChecks, err := checkFloatMap(row.ID, "per_key", want.PerKey, dist.PerKey())
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: err}
	}
	checks = append(checks, perKeyChecks...)

	gotItems := make([]keyDist, 0, len(want.PerKey))
	for _, item := range dist.NonzeroItems() {
		gotItems = append(gotItems, keyDist{Key: item.Key, Dist: item.Distance})
	}
	itemChecks, err := checkItems(row.ID, "items", wantItems, gotItems)
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: err}
	}
	checks = append(checks, itemChecks...)

	return checkFields(row.ID, row.Method, checks)
}

func replayGenericPenaltyKeys(row analysis.ComponentRow, args aggregateArgs) ReplayOutcome {
	tun := tunablesWithOverrides(args.Weights)
	dist := matcher.NewDistance(tun)
	if err := applyAggregateCalls(dist, args.Calls); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: %w", row.ID, err)}
	}

	var want struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}
	got := dist.GenericPenaltyKeys()
	if !slices.Equal(want.Keys, got) {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: keys: got %v, want %v", row.ID, got, want.Keys)}
	}
	return ReplayOutcome{RowID: row.ID, Method: row.Method}
}

// replayStringDist replays a string_dist row: args is a bare
// two-element array of possibly-null strings, output is {"distance"}.
func replayStringDist(row analysis.ComponentRow) ReplayOutcome {
	elems, err := decodeCallElems(row.Args)
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
	}
	if err := wantArity("string_dist", elems, 2); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: %w", row.ID, err)}
	}
	var s1, s2 *string
	if err := json.Unmarshal(elems[0], &s1); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args[0]: %w", row.ID, err)}
	}
	if err := json.Unmarshal(elems[1], &s2); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args[1]: %w", row.ID, err)}
	}

	var want struct {
		Distance float64 `json:"distance"`
	}
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}
	got := matcher.StringDist(matcher.DefaultTunables(), s1, s2)
	return checkFields(row.ID, row.Method, []fieldCheck{{Name: "distance", Want: want.Distance, Got: got}})
}

// track_distance's args: incl_artist, item, and track_info always;
// weights, data_source_mismatch_penalty, and multiple_data_sources
// only on the data-source rows, which is why the last two are
// pointers: their absence, not their zero value, is what tells the
// decoder to leave Tunables' metadata-source set at its own default.

type trackDistanceArgs struct {
	InclArtist                bool               `json:"incl_artist"`
	Item                      map[string]any     `json:"item"`
	TrackInfo                 map[string]any     `json:"track_info"`
	Weights                   map[string]float64 `json:"weights"`
	DataSourceMismatchPenalty *float64           `json:"data_source_mismatch_penalty"`
	MultipleDataSources       *bool              `json:"multiple_data_sources"`
}

// metadataSourceSet builds the get_penalty two-plugin stub's loaded-
// source set (fold, B3): {"Original": 0.5, "Other": penalty}, or just
// {"Other": penalty} when multiple is false. The fallback get_penalty
// resolves to when after names no loaded source is always the class
// default 0.5, never the row's own penalty field.
func metadataSourceSet(penalty float64, multiple bool) map[string]float64 {
	sources := map[string]float64{"Other": penalty}
	if multiple {
		sources["Original"] = 0.5
	}
	return sources
}

func decodeTrackInfo(raw map[string]any) *matcher.TrackInfo {
	return &matcher.TrackInfo{
		Title:       stringAny(raw, "title"),
		Artist:      stringAny(raw, "artist"),
		Length:      floatAny(raw, "length"),
		Index:       intAny(raw, "index"),
		MediumIndex: intAny(raw, "medium_index"),
		Medium:      intAny(raw, "medium"),
		TrackID:     stringAny(raw, "track_id"),
		DataSource:  stringAny(raw, "data_source"),
	}
}

func stringAny(raw map[string]any, key string) string {
	s, _ := raw[key].(string)
	return s
}

func intAny(raw map[string]any, key string) int {
	switch v := raw[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func floatAny(raw map[string]any, key string) float64 {
	switch v := raw[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}

func replayTrackDistance(row analysis.ComponentRow) ReplayOutcome {
	var args trackDistanceArgs
	if err := json.Unmarshal(row.Args, &args); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
	}

	tun := tunablesWithOverrides(args.Weights)
	if args.DataSourceMismatchPenalty != nil {
		multiple := args.MultipleDataSources == nil || *args.MultipleDataSources
		tun.DataSourcePenalties = metadataSourceSet(*args.DataSourceMismatchPenalty, multiple)
		tun.DataSourceMismatchPenalty = 0.5
	}

	item := matcher.ProjectItemFields(args.Item)
	info := decodeTrackInfo(args.TrackInfo)
	dist := matcher.DistanceForTrack(tun, item, info, args.InclArtist)

	var want struct {
		Distance  float64              `json:"distance"`
		Penalties map[string][]float64 `json:"penalties"`
	}
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}

	got := rawPenaltyMap(dist)
	checks := []fieldCheck{{Name: "distance", Want: want.Distance, Got: dist.Distance()}}
	penaltyChecks, err := checkPenaltyMap(row.ID, "penalties", want.Penalties, got)
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: err}
	}
	checks = append(checks, penaltyChecks...)
	return checkFields(row.ID, row.Method, checks)
}

// add_data_source's args: before and after may each be null (Python
// None), and the penalty and multiple-source shape is the same
// get_penalty stub track_distance's data-source rows use.

type addDataSourceArgs struct {
	Before                    *string            `json:"before"`
	After                     *string            `json:"after"`
	Weights                   map[string]float64 `json:"weights"`
	DataSourceMismatchPenalty float64            `json:"data_source_mismatch_penalty"`
	MultipleDataSources       bool               `json:"multiple_data_sources"`
}

func replayAddDataSource(row analysis.ComponentRow) ReplayOutcome {
	var args addDataSourceArgs
	if err := json.Unmarshal(row.Args, &args); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: args: %w", row.ID, err)}
	}

	tun := tunablesWithOverrides(args.Weights)
	tun.DataSourcePenalties = metadataSourceSet(args.DataSourceMismatchPenalty, args.MultipleDataSources)
	tun.DataSourceMismatchPenalty = 0.5

	dist := matcher.NewDistance(tun)
	dist.AddDataSource(derefOr(args.Before, ""), derefOr(args.After, ""))

	var want struct {
		Penalties map[string][]float64 `json:"penalties"`
	}
	if err := json.Unmarshal(row.Output, &want); err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: fmt.Errorf("component %s: output: %w", row.ID, err)}
	}

	got := rawPenaltyMap(dist)
	checks, err := checkPenaltyMap(row.ID, "penalties", want.Penalties, got)
	if err != nil {
		return ReplayOutcome{RowID: row.ID, Method: row.Method, Err: err}
	}
	return checkFields(row.ID, row.Method, checks)
}

func derefOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// rawPenaltyMap reads dist's raw, unweighted penalty slice for every
// key it holds, in the shape track_distance and add_data_source rows
// commit their "penalties" output as.
func rawPenaltyMap(dist *matcher.Distance) map[string][]float64 {
	got := make(map[string][]float64, len(dist.AddedKeys()))
	for _, key := range dist.AddedKeys() {
		got[key] = dist.Penalties(key)
	}
	return got
}

// keyDist is one (key, weighted distance) pair, the Go form of one
// entry in a components.json aggregate row's "items" array.
type keyDist struct {
	Key  string
	Dist float64
}

// decodeItems decodes an aggregate row's "items" field: a JSON array
// of [key, distance] pairs.
func decodeItems(raw json.RawMessage) ([]keyDist, error) {
	var pairs [][]any
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return nil, err
	}
	items := make([]keyDist, len(pairs))
	for i, pair := range pairs {
		if len(pair) != 2 {
			return nil, fmt.Errorf("items[%d]: want 2 elements, got %d", i, len(pair))
		}
		key, ok := pair[0].(string)
		if !ok {
			return nil, fmt.Errorf("items[%d]: key is not a string", i)
		}
		dist, ok := pair[1].(float64)
		if !ok {
			return nil, fmt.Errorf("items[%d]: distance is not a number", i)
		}
		items[i] = keyDist{Key: key, Dist: dist}
	}
	return items, nil
}

// fieldCheck is one committed value and the value the matcher
// produced for it, named for the error and the near-miss report a
// mismatch carries.
type fieldCheck struct {
	Name string
	Want float64
	Got  float64
}

// checkFields runs compareFloats over every check, failing outright on
// the first check beyond the 1e-12 fallback and otherwise reporting a
// near miss when any check matched only within it (contract 12).
func checkFields(rowID, method string, checks []fieldCheck) ReplayOutcome {
	var nearMissFields []string
	for _, c := range checks {
		near, err := compareFloats(rowID, c.Name, c.Want, c.Got)
		if err != nil {
			return ReplayOutcome{RowID: rowID, Method: method, Err: err}
		}
		if near {
			nearMissFields = append(nearMissFields, c.Name)
		}
	}
	outcome := ReplayOutcome{RowID: rowID, Method: method}
	if len(nearMissFields) > 0 {
		outcome.NearMiss = true
		outcome.Detail = fmt.Sprintf("%s: bit-identity failed but matched within 1e-12 relative tolerance; suspected floating-point operation-order divergence", strings.Join(nearMissFields, ", "))
	}
	return outcome
}

// compareFloats reports whether want and got are bit-identical, and if
// not, whether they still agree within 1e-12 relative tolerance
// (contract 12). A row that passes only via the tolerance is not
// silently treated as bit-identical: near is true and the caller
// records it as a finding.
func compareFloats(rowID, field string, want, got float64) (near bool, err error) {
	if want == got {
		return false, nil
	}
	scale := math.Max(math.Abs(want), 1)
	if math.IsNaN(got) || math.Abs(got-want)/scale > 1e-12 {
		return false, fmt.Errorf("component %s: %s: got %v, want %v", rowID, field, got, want)
	}
	return true, nil
}

// checkFloatSlice compares two float64 slices elementwise, failing
// outright on a length mismatch (a components.json row never omits or
// adds a penalty) before any float comparison runs.
func checkFloatSlice(rowID, field string, want, got []float64) ([]fieldCheck, error) {
	if len(want) != len(got) {
		return nil, fmt.Errorf("component %s: %s: got %d values, want %d", rowID, field, len(got), len(want))
	}
	checks := make([]fieldCheck, len(want))
	for i := range want {
		checks[i] = fieldCheck{Name: fmt.Sprintf("%s[%d]", field, i), Want: want[i], Got: got[i]}
	}
	return checks, nil
}

// checkFloatMap compares two key -> float64 maps, failing outright on
// a key-set mismatch (missing or extra key) before any float
// comparison runs.
func checkFloatMap(rowID, field string, want, got map[string]float64) ([]fieldCheck, error) {
	if len(want) != len(got) {
		return nil, fmt.Errorf("component %s: %s: got %d keys, want %d", rowID, field, len(got), len(want))
	}
	keys := slices.Sorted(maps.Keys(want))
	checks := make([]fieldCheck, 0, len(keys))
	for _, key := range keys {
		gotValue, ok := got[key]
		if !ok {
			return nil, fmt.Errorf("component %s: %s: missing key %q", rowID, field, key)
		}
		checks = append(checks, fieldCheck{Name: field + "." + key, Want: want[key], Got: gotValue})
	}
	return checks, nil
}

// checkPenaltyMap compares two key -> []float64 maps (a raw penalty
// map, track_distance's and add_data_source's own output shape),
// failing outright on a key-set or per-key length mismatch before any
// float comparison runs.
func checkPenaltyMap(rowID, field string, want, got map[string][]float64) ([]fieldCheck, error) {
	if len(want) != len(got) {
		return nil, fmt.Errorf("component %s: %s: got %d keys, want %d", rowID, field, len(got), len(want))
	}
	var checks []fieldCheck
	for _, key := range slices.Sorted(maps.Keys(want)) {
		gotSlice, ok := got[key]
		if !ok {
			return nil, fmt.Errorf("component %s: %s: missing key %q", rowID, field, key)
		}
		sub, err := checkFloatSlice(rowID, field+"."+key, want[key], gotSlice)
		if err != nil {
			return nil, err
		}
		checks = append(checks, sub...)
	}
	return checks, nil
}

// checkItems compares two ordered (key, distance) lists (an
// aggregate row's "items" view), failing outright on a length or key
// mismatch at any position before any float comparison runs: order is
// part of the contract (NonzeroItems' sort), not incidental.
func checkItems(rowID, field string, want, got []keyDist) ([]fieldCheck, error) {
	if len(want) != len(got) {
		return nil, fmt.Errorf("component %s: %s: got %d items, want %d", rowID, field, len(got), len(want))
	}
	checks := make([]fieldCheck, len(want))
	for i := range want {
		if want[i].Key != got[i].Key {
			return nil, fmt.Errorf("component %s: %s[%d]: got key %q, want %q", rowID, field, i, got[i].Key, want[i].Key)
		}
		checks[i] = fieldCheck{Name: fmt.Sprintf("%s[%d]", field, i), Want: want[i].Dist, Got: got[i].Dist}
	}
	return checks, nil
}

// itemsEqual reports whether two ordered (key, distance) lists are
// exactly equal: used by the arithmetic-mutation tests to show a wrong
// sort produces something other than the committed order.
func itemsEqual(a, b []keyDist) bool {
	return slices.EqualFunc(a, b, func(x, y keyDist) bool {
		return x.Key == y.Key && x.Dist == y.Dist
	})
}
