package grade

import (
	"cmp"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

// orderSignificant lists the array paths whose element order is part of
// the response's meaning. Every other array compares as a set of paired
// elements, since a server may list them in any order.
var orderSignificant = []struct{ path, why string }{
	{
		"subsonic-response/album/song",
		"an album lists its songs in disc and track order; both pinned oracles do so in the seed session's getAlbum answers",
	},
}

// presenceOnly lists the scalar members that every server fills with its
// own value, so two sources agree on them when both carry one of the
// same type and a differing literal is no finding. Ledger row L-08 and
// the contract's note on the envelope's serverVersion and type give the
// reason for each.
var presenceOnly = []struct{ path, why string }{
	{"subsonic-response/serverVersion", "each server reports its own version"},
	{"subsonic-response/type", "each server names itself"},
	{"subsonic-response/error/message", "each server words its failures; the code is what clients branch on"},
}

// idMembers are the members whose value is a server's own opaque id.
// They compare through the id pairing, never by their literals.
var idMembers = []string{"id", "parent", "albumId", "artistId", "coverArt"}

// idSpace separates the ids of entities from the ids of cover art, since
// a server may use one string for an album and for its cover.
type idSpace uint8

const (
	spaceEntity idSpace = iota
	spaceCover
)

func spaceOf(member string) idSpace {
	if member == "coverArt" {
		return spaceCover
	}
	return spaceEntity
}

type tableKey struct {
	scope string // the subtree row id, or "" for the session-wide table
	space idSpace
}

// idTable is the pairing between the two sources' ids in one scope and
// space, in both directions.
type idTable struct {
	fwd, back map[string][]string
}

// pairing holds every id pair the grader has learned. A pair learned
// inside a subtree row's scope is filed under that row and is
// visible only to comparisons inside the same scope.
type pairing struct {
	tables   map[tableKey]*idTable
	reported map[string]bool
}

func newPairing() *pairing {
	return &pairing{tables: map[tableKey]*idTable{}, reported: map[string]bool{}}
}

func (p *pairing) add(scope string, space idSpace, left, right string) {
	if left == "" || right == "" {
		return
	}
	k := tableKey{scope, space}
	t := p.tables[k]
	if t == nil {
		t = &idTable{fwd: map[string][]string{}, back: map[string][]string{}}
		p.tables[k] = t
	}
	if !slices.Contains(t.fwd[left], right) {
		t.fwd[left] = append(t.fwd[left], right)
	}
	if !slices.Contains(t.back[right], left) {
		t.back[right] = append(t.back[right], left)
	}
}

// partners returns the ids id pairs with from scope: the session-wide
// pairs and the scope's own.
func (p *pairing) partners(scope string, space idSpace, id string, forward bool) []string {
	var out []string
	for _, s := range slices.Compact([]string{"", scope}) {
		t := p.tables[tableKey{s, space}]
		if t == nil {
			continue
		}
		m := t.back
		if forward {
			m = t.fwd
		}
		for _, v := range m[id] {
			if !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
	}
	slices.Sort(out)
	return out
}

// once reports whether key is new, recording it.
func (p *pairing) once(key string) bool {
	if p.reported[key] {
		return false
	}
	p.reported[key] = true
	return true
}

// differ compares the two sources' trees for one request. The same walk
// runs twice over a session: first with learn set, recording the id pairs
// every paired entity yields and reporting nothing, then to report, once
// every pair the session yields is known.
type differ struct {
	lib     *library
	ids     *pairing
	scopes  []Row // the subtree rows applied to the graded pair
	learn   bool
	request string
	endpt   string
	format  client.Format
	param   string
	sources [2]string
	found   []Finding
}

func (d *differ) scope(path string) string {
	for _, r := range d.scopes {
		if r.selects(d.endpt, d.format, path) {
			return r.ID
		}
	}
	return ""
}

func (d *differ) report(kind Kind, path string, left, right string) {
	if d.learn {
		return
	}
	d.found = append(d.found, Finding{
		Request: d.request, Endpoint: d.endpt, Format: d.format,
		Sources: d.sources[:], Kind: kind, Path: path, Values: []string{left, right},
	})
}

func join(path, seg string) string {
	if path == "" {
		return seg
	}
	return path + "/" + seg
}

// compare diffs the left and right values of member at path.
func (d *differ) compare(path, member string, a, b *node, ca, cb identCtx) {
	if a == nil || b == nil {
		if a != b {
			d.report(KindPresence, path, a.render(), b.render())
		}
		return
	}
	if d.format == client.XML {
		a, b = wrapSingle(a, b), wrapSingle(b, a)
	}
	if a.kind != b.kind || a.form != b.form {
		d.report(KindType, path, a.render(), b.render())
		return
	}
	switch a.kind {
	case kindObject:
		d.compareObject(path, member, a, b, ca, cb)
	case kindArray:
		d.compareArray(path, member, a, b, ca, cb)
	default:
		if slices.Contains(idMembers, member) {
			d.compareID(path, member, a, b)
		} else if a.lit != b.lit && !d.presenceOnly(path) {
			d.report(KindValue, path, a.render(), b.render())
		}
	}
}

// wrapSingle returns n as a one-element array when other is an array and
// n a lone element: XML cannot tell a single repeated element from a
// single one.
func wrapSingle(n, other *node) *node {
	if other.kind == kindArray && n.kind != kindArray && n.form == formElem {
		return &node{kind: kindArray, form: formElem, items: []*node{n}}
	}
	return n
}

func (d *differ) compareObject(path, member string, a, b *node, ca, cb identCtx) {
	ka, kb := kindOf(d.endpt, member, a), kindOf(d.endpt, member, b)
	if ka != entityNone || kb != entityNone {
		ia := d.lib.resolve(ka, a, ca, member)
		ib := d.lib.resolve(kb, b, cb, member)
		if ia.key == "" || ia.key != ib.key {
			d.report(KindIdentity, path, ia.describe(a), ib.describe(b))
			return
		}
		if d.learn {
			d.ids.add(d.scope(path), spaceEntity, a.field("id"), b.field("id"))
		}
		ca, cb = ca.enter(ia, a), cb.enter(ib, b)
	}
	if d.learn {
		d.ids.add(d.scope(path), spaceCover, a.field("coverArt"), b.field("coverArt"))
	}
	names := slices.Collect(maps.Keys(a.fields))
	names = slices.AppendSeq(names, maps.Keys(b.fields))
	slices.Sort(names)
	for _, name := range slices.Compact(names) {
		d.compare(join(path, name), name, a.fields[name], b.fields[name], ca, cb)
	}
}

// elementKey is what pairs an array element with its counterpart. A
// strict key names one entity or one keyed record, so two elements
// sharing it on one side cannot be told apart; any other key is the
// element's whole value, and equal values pair in order.
type elementKey struct {
	key    string
	strict bool
}

// naturalKeys are the members that key a record which is not an entity.
var naturalKeys = []string{"name", "value", "disc"}

func (d *differ) key(member string, n *node, c identCtx) elementKey {
	if n.kind != kindObject {
		return elementKey{key: n.canonical()}
	}
	if kind := kindOf(d.endpt, member, n); kind != entityNone {
		id := d.lib.resolve(kind, n, c, member)
		if id.key != "" {
			return elementKey{key: id.key, strict: true}
		}
		return elementKey{key: "unresolved " + string(kind) + " " + n.title(), strict: true}
	}
	for _, k := range naturalKeys {
		if v := n.field(k); v != "" {
			return elementKey{key: k + "=" + v, strict: true}
		}
	}
	return elementKey{key: n.canonical()}
}

// compareArray pairs the two arrays' elements by key, never by position.
// An element with no counterpart is a presence finding, a strict key held
// by more than one element on a side is one identity finding, and on an
// order-significant path the paired elements must keep one order.
func (d *differ) compareArray(path, member string, a, b *node, ca, cb identCtx) {
	var keys [2][]elementKey
	byKey := map[string][2][]int{}
	var order []string
	ctxs := [2]identCtx{ca, cb}
	for side, arr := range [2]*node{a, b} {
		keys[side] = make([]elementKey, len(arr.items))
		for i, n := range arr.items {
			k := d.key(member, n, ctxs[side])
			keys[side][i] = k
			g := byKey[k.key]
			if g[0] == nil && g[1] == nil {
				order = append(order, k.key)
			}
			g[side] = append(g[side], i)
			byKey[k.key] = g
		}
	}
	type match struct{ i, j int }
	var pairs []match
	for _, k := range order {
		g := byKey[k]
		strict := (len(g[0]) > 0 && keys[0][g[0][0]].strict) || (len(g[1]) > 0 && keys[1][g[1][0]].strict)
		if strict && (len(g[0]) > 1 || len(g[1]) > 1) {
			d.ambiguous(path, member, k, a, b, g)
			continue
		}
		n := min(len(g[0]), len(g[1]))
		for x := range n {
			pairs = append(pairs, match{g[0][x], g[1][x]})
		}
		for _, i := range g[0][n:] {
			d.report(KindPresence, join(path, strconv.Itoa(i)), a.items[i].render(), "absent")
		}
		for _, j := range g[1][n:] {
			d.report(KindPresence, join(path, strconv.Itoa(j)), "absent", b.items[j].render())
		}
	}
	slices.SortFunc(pairs, func(x, y match) int { return cmp.Compare(x.i, y.i) })
	if d.ordered(path) && !slices.IsSortedFunc(pairs, func(x, y match) int { return cmp.Compare(x.j, y.j) }) {
		left := make([]string, len(pairs))
		for n, p := range pairs {
			left[n] = keys[0][p.i].key
		}
		right := slices.Clone(left)
		slices.SortStableFunc(right, func(x, y string) int {
			return cmp.Compare(byKey[x][1][0], byKey[y][1][0])
		})
		d.report(KindOrder, path, strings.Join(left, ", "), strings.Join(right, ", "))
	}
	for _, p := range pairs {
		d.compare(join(path, strconv.Itoa(p.i)), member, a.items[p.i], b.items[p.j], ca, cb)
	}
}

func (d *differ) presenceOnly(path string) bool {
	for _, p := range presenceOnly {
		if p.path == path {
			return true
		}
	}
	return false
}

func (d *differ) ordered(path string) bool {
	for _, o := range orderSignificant {
		if pathCovers(o.path, path, false) {
			return true
		}
	}
	return false
}

// ambiguous reports a strict key held by more than one element on a
// side. When learning, every id the key's entities carry on one side
// pairs with every id on the other, so the ambiguity also surfaces
// wherever those ids are compared.
func (d *differ) ambiguous(path, member, key string, a, b *node, g [2][]int) {
	if !d.learn {
		d.report(KindIdentity, path, heldBy(key, len(g[0])), heldBy(key, len(g[1])))
		return
	}
	scope := d.scope(path)
	for _, i := range g[0] {
		for _, j := range g[1] {
			if kindOf(d.endpt, member, a.items[i]) != entityNone {
				d.ids.add(scope, spaceEntity, a.items[i].field("id"), b.items[j].field("id"))
			}
			d.ids.add(scope, spaceCover, a.items[i].field("coverArt"), b.items[j].field("coverArt"))
		}
	}
}

// compareID checks that two ids are each other's only partner. An id
// with two partners is reported once per scope, and so is a pair of ids
// neither of which pairs with anything; a reference to the wrong partner
// is reported wherever it appears.
func (d *differ) compareID(path, member string, a, b *node) {
	if d.learn {
		return
	}
	scope, space := d.scope(path), spaceOf(member)
	right := d.ids.partners(scope, space, a.lit, true)
	left := d.ids.partners(scope, space, b.lit, false)
	if slices.Equal(right, []string{b.lit}) && slices.Equal(left, []string{a.lit}) {
		return
	}
	at := scope + "\x00" + strconv.Itoa(int(space)) + "\x00"
	switch {
	case len(right) > 1 && !d.ids.once(at+"L\x00"+a.lit):
		return
	case len(right) <= 1 && len(left) > 1 && !d.ids.once(at+"R\x00"+b.lit):
		return
	case len(right) == 0 && len(left) == 0 && !d.ids.once(at+"U\x00"+a.lit+"\x00"+b.lit):
		return
	}
	d.report(KindIdentity, path, a.render()+partnersNote(right), b.render()+partnersNote(left))
}

func heldBy(key string, n int) string {
	if n == 1 {
		return key + " on 1 element"
	}
	return key + " on " + strconv.Itoa(n) + " elements"
}

func partnersNote(ids []string) string {
	if len(ids) == 0 {
		return " pairs with nothing"
	}
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = strconv.Quote(id)
	}
	return " pairs with " + strings.Join(quoted, " and ")
}
