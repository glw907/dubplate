package tagio

import (
	"fmt"
	"math"
	"strconv"
)

// KeyWrite is one field's contribution to a write, named by the field's
// ledger name in Field. Deletes lists a read-only style only when the
// field is absent, because read_only blocks a set but not a delete;
// Writes covers writable styles only.
type KeyWrite struct {
	Field   string
	Deletes []string
	Writes  []Comment
}

// Serialize turns ts into one KeyWrite per field, in the field table's
// write order. An absent member deletes every key its field owns; a
// present one, empty included, deletes the writable keys and writes
// its comments in their place. The year, month and day members compose
// into one KeyWrite named for the year, as do the three original_
// members, because one DATE or ORIGINALDATE comment carries all three
// components. Two TagSets that Serialize identically write identical
// comments, so a caller can compare TagSets through its result.
func Serialize(ts TagSet) []KeyWrite {
	out := make([]KeyWrite, 0, len(fields))
	for _, f := range fields {
		if f.codec.write == nil {
			continue
		}
		writes, present := f.codec.write(&ts, f.styles)
		if !present {
			out = append(out, KeyWrite{Field: f.name, Deletes: allKeys(f.styles)})
			continue
		}
		out = append(out, KeyWrite{Field: f.name, Deletes: writableKeys(f.styles), Writes: writes})
	}
	return out
}

// allKeys lists every key styles records, writable and read-only
// alike. A delete of an absent field removes all of them: deleting
// barcode removes UPC, EAN/UPN, EAN and UPN along with BARCODE.
func allKeys(styles []style) []string {
	keys := make([]string, len(styles))
	for i, s := range styles {
		keys[i] = s.key
	}
	return keys
}

// writableKeys lists the writable keys in styles, the set a present
// value replaces.
func writableKeys(styles []style) []string {
	var keys []string
	for _, s := range styles {
		if s.writable {
			keys = append(keys, s.key)
		}
	}
	return keys
}

// listComments writes every element of values under each writable
// style in turn, in list order.
func listComments(styles []style, values []string) []Comment {
	var out []Comment
	for _, s := range styles {
		if !s.writable {
			continue
		}
		for _, v := range values {
			out = append(out, Comment{Name: s.key, Value: v})
		}
	}
	return out
}

// formatBool writes comp as "1" or "0", never "true" or "false"
// (mediafile:535, StorageStyle.serialize).
func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// formatGain writes a ReplayGain gain to two decimal places, the
// default float_places, with the " dB" suffix that is the only suffix
// in the FLAC table (mediafile:836, :846).
func formatGain(v float64) string {
	return fmt.Sprintf("%.2f dB", v)
}

// formatPeak writes a ReplayGain peak to six decimal places, with no
// suffix (mediafile:856, :864).
func formatPeak(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

// formatQ8 writes an R128 gain as an integer string, the value shifted
// by fraction_bits=8: round(value * 2**8) (mediafile:874, :881;
// __init__.py:870-884). r128_track_gain = -1.5 writes
// R128_TRACK_GAIN=-384.
func formatQ8(v float64) string {
	return strconv.FormatInt(int64(math.Round(v*256)), 10)
}

// dateComments composes the DATE (or ORIGINALDATE) comment and, when
// styles carries a YEAR style, the bare-year fallback, from a year,
// month and day (mediafile/fields.py:182-201). The year is zero-padded
// to four digits, the month is appended zero-padded to two digits only
// when it is nonzero, and the day only when both the month and the day
// are. The caller passes an absent month or day as zero.
func dateComments(styles []style, y, m, d int) []Comment {
	date := fmt.Sprintf("%04d", y)
	if m != 0 {
		date += fmt.Sprintf("-%02d", m)
		if d != 0 {
			date += fmt.Sprintf("-%02d", d)
		}
	}

	var out []Comment
	for _, s := range styles {
		if !s.writable {
			continue
		}
		switch s.key {
		case "YEAR":
			out = append(out, Comment{Name: s.key, Value: strconv.Itoa(y)})
		default: // DATE, ORIGINALDATE
			out = append(out, Comment{Name: s.key, Value: date})
		}
	}
	return out
}

func valueOr(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
