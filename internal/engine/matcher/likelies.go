package matcher

// Likelies is beets' get_most_common_tags's first return value
// (util/__init__.py:810-846), narrowed from Python's untyped dict to
// a typed struct: every one of the thirteen fields the album-level
// distance path reads already has a known Go type (info.go's
// ItemFields), so a map[string]any here would throw that typing away
// for no benefit the album components could use. Each field holds
// the plurality value across items (contract 11: ties break by
// first-seen insertion order, i.e. by item order).
type Likelies struct {
	Artist        string
	Album         string
	AlbumArtist   string
	Year          int
	DiscTotal     int
	MBAlbumID     string
	Label         string
	Barcode       string
	CatalogNum    string
	Country       string
	Media         string
	AlbumDisambig string
	DataSource    string
}

// MostCommonTags is beets' get_most_common_tags (util/__init__.py:
// 810-846): the plurality value of each of the thirteen album-level
// fields across items, in beets' own field order, plus a consensus
// map reporting whether every item that contributed a value to a
// field agreed on it. An albumartist consensus, non-empty, overrides
// the returned Likelies.Artist (util/__init__.py:843-844).
//
// distance() itself discards its own copy of consensus
// (`likelies, _ = get_most_common_tags(items)`, distance.py:442), so
// no album component reads the map this function returns; it exists
// for parity with beets' two-value return and because the override
// above is provable only by inspecting it, which this file's own test
// does.
//
// Python's get_most_common_tags filters falsy items before counting
// (util/__init__.py:838, `values = [item.get(field) for item in items
// if item]`) because a beets Item object can itself evaluate false.
// ItemFields is a Go struct value, never nil and never falsy, so
// every item in items is counted here; the filter has no Go analogue
// to reproduce, and consensus's denominator is simply len(items).
func MostCommonTags(items []ItemFields) (Likelies, map[string]bool) {
	consensus := make(map[string]bool, 13)
	total := len(items)

	str := func(key string, get func(ItemFields) string) string {
		value, freq := plurality(projectField(items, get))
		consensus[key] = freq == total
		return value
	}
	num := func(key string, get func(ItemFields) int) int {
		value, freq := plurality(projectField(items, get))
		consensus[key] = freq == total
		return value
	}

	likelies := Likelies{
		Artist:        str(fieldArtist, func(i ItemFields) string { return i.Artist }),
		Album:         str(fieldAlbum, func(i ItemFields) string { return i.Album }),
		AlbumArtist:   str(fieldAlbumArtist, func(i ItemFields) string { return i.AlbumArtist }),
		Year:          num(fieldYear, func(i ItemFields) int { return i.Year }),
		DiscTotal:     num(fieldDiscTotal, func(i ItemFields) int { return i.DiscTotal }),
		MBAlbumID:     str(fieldMBAlbumID, func(i ItemFields) string { return i.MBAlbumID }),
		Label:         str(fieldLabel, func(i ItemFields) string { return i.Label }),
		Barcode:       str(fieldBarcode, func(i ItemFields) string { return i.Barcode }),
		CatalogNum:    str(fieldCatalogNum, func(i ItemFields) string { return i.CatalogNum }),
		Country:       str(fieldCountry, func(i ItemFields) string { return i.Country }),
		Media:         str(fieldMedia, func(i ItemFields) string { return i.Media }),
		AlbumDisambig: str(fieldAlbumDisambig, func(i ItemFields) string { return i.AlbumDisambig }),
		DataSource:    str(fieldDataSource, func(i ItemFields) string { return i.DataSource }),
	}

	if consensus[fieldAlbumArtist] && likelies.AlbumArtist != "" {
		likelies.Artist = likelies.AlbumArtist
	}

	return likelies, consensus
}

// projectField runs get over items, in order, the Go form of
// Python's list comprehension over item.get(field).
func projectField[T any](items []ItemFields, get func(ItemFields) T) []T {
	values := make([]T, len(items))
	for i, item := range items {
		values[i] = get(item)
	}
	return values
}

// plurality is Counter.most_common(1)[0] (util/__init__.py:799-807):
// the most common value in values and its count. Ties break by the
// value's first-seen index (contract 11) rather than by Go's
// randomized map order: values is walked once to record each
// distinct value's first-seen position, and the highest count wins
// with ties resolved in favor of the earliest position. An empty
// values returns T's zero value and a count of zero, reachable only
// when items itself is empty.
func plurality[T comparable](values []T) (T, int) {
	type tally struct {
		count     int
		firstSeen int
	}
	tallies := make(map[T]*tally, len(values))
	var order []T
	for i, v := range values {
		t, seen := tallies[v]
		if !seen {
			t = &tally{firstSeen: i}
			tallies[v] = t
			order = append(order, v)
		}
		t.count++
	}
	if len(order) == 0 {
		var zero T
		return zero, 0
	}
	best := order[0]
	for _, v := range order[1:] {
		if tallies[v].count > tallies[best].count {
			best = v
		}
	}
	return best, tallies[best].count
}
