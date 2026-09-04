//go:build oracle

package oracle

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

// fileTagsDocument decodes derive_file_tags.py's document just enough
// for this package's own tests to inspect it. The gate-resident
// ledger shape (internal/harness/analysis) is a later task's own type;
// this is a test-local view, not a second, competing home for it.
type fileTagsDocument struct {
	ShimVersion    string         `json:"shim_version"`
	BeetsVersion   string         `json:"beets_version"`
	WritableFields []string       `json:"writable_fields"`
	ReadonlyFields []string       `json:"readonly_fields"`
	Template       map[string]any `json:"template"`
	Rows           []struct {
		CaseID    string         `json:"case_id"`
		ItemIndex int            `json:"item_index"`
		Values    map[string]any `json:"values"`
	} `json:"rows"`
}

// wantReadonlyFields is the nine-key read-only remainder the plan
// names by name: read-only audio properties a read-back reports from
// the template blob's own STREAMINFO, never from a subject.
var wantReadonlyFields = []string{
	"bitdepth", "bitrate", "bitrate_mode", "channels", "encoder_info",
	"encoder_settings", "format", "length", "samplerate",
}

func deriveFileTagsDoc(t *testing.T, subjects []FileTagSubject) fileTagsDocument {
	t.Helper()
	raw, err := DeriveFileTags(context.Background(), venvOptions(t), subjects)
	if err != nil {
		t.Fatalf("DeriveFileTags: %v", err)
	}
	var doc fileTagsDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode file-tags document: %v", err)
	}
	return doc
}

// TestDeriveFileTagsOpensTemplateAsFLAC is the escalation-trigger
// bar's own proof: a subject that writes and reads back cleanly is
// only possible if mediafile actually opened the vendored template as
// a FLAC container, per mediafile/__init__.py's own FLAC dispatch.
func TestDeriveFileTagsOpensTemplateAsFLAC(t *testing.T) {
	doc := deriveFileTagsDoc(t, []FileTagSubject{
		{
			CaseID:    "smoke-test",
			ItemIndex: 0,
			AppliedFields: map[string]any{
				"title":      "Track One",
				"artist":     "Example Artist",
				"album":      "Example Album",
				"year":       2020,
				"mb_trackid": "3b6b0a4e-5e1a-4a52-9c31-3c0c1c1f0021",
			},
		},
	})

	if len(doc.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(doc.Rows))
	}
	row := doc.Rows[0]
	if row.Values["title"] != "Track One" {
		t.Errorf("row title = %v, want %q", row.Values["title"], "Track One")
	}
	if row.Values["artist"] != "Example Artist" {
		t.Errorf("row artist = %v, want %q", row.Values["artist"], "Example Artist")
	}
	if row.Values["year"] != float64(2020) {
		t.Errorf("row year = %v, want 2020", row.Values["year"])
	}
}

// TestDeriveFileTagsHeaderComputesTheKeySplit proves the writable set
// and the read-only remainder are computed off the venv, not assumed:
// 69 writable keys, the nine named read-only keys, no overlap between
// them, and the read-only values living only in the shared template
// block.
func TestDeriveFileTagsHeaderComputesTheKeySplit(t *testing.T) {
	doc := deriveFileTagsDoc(t, []FileTagSubject{
		{CaseID: "smoke-test", ItemIndex: 0, AppliedFields: map[string]any{"title": "x"}},
	})

	if got, want := len(doc.WritableFields), 69; got != want {
		t.Errorf("writable_fields count = %d, want %d", got, want)
	}
	gotReadonly := slices.Clone(doc.ReadonlyFields)
	slices.Sort(gotReadonly)
	wantSorted := slices.Clone(wantReadonlyFields)
	slices.Sort(wantSorted)
	if !slices.Equal(gotReadonly, wantSorted) {
		t.Errorf("readonly_fields = %v, want %v", gotReadonly, wantSorted)
	}

	writable := make(map[string]bool, len(doc.WritableFields))
	for _, f := range doc.WritableFields {
		writable[f] = true
	}
	for _, f := range doc.ReadonlyFields {
		if writable[f] {
			t.Errorf("field %q appears in both writable_fields and readonly_fields", f)
		}
	}

	for _, f := range doc.ReadonlyFields {
		if _, ok := doc.Template[f]; !ok {
			t.Errorf("template block is missing read-only field %q", f)
		}
	}
}

// TestDeriveFileTagsRowExcludesBookkeepingColumns proves no row's
// value dict ever carries an Item column that sits outside
// _media_fields entirely (id, album_id, mtime, path): these never
// reach a file, and a row carrying one would hand a consumer a value
// no writer can reproduce.
func TestDeriveFileTagsRowExcludesBookkeepingColumns(t *testing.T) {
	doc := deriveFileTagsDoc(t, []FileTagSubject{
		{CaseID: "smoke-test", ItemIndex: 0, AppliedFields: map[string]any{"title": "x", "genres": []any{"Rock", "Pop"}}},
	})

	for _, bookkeeping := range []string{"id", "album_id", "mtime", "path"} {
		if _, ok := doc.Rows[0].Values[bookkeeping]; ok {
			t.Errorf("row values carry bookkeeping column %q, want it absent", bookkeeping)
		}
	}
	if got, ok := doc.Rows[0].Values["genres"].([]any); !ok || !slices.Equal(toStrings(t, got), []string{"Rock", "Pop"}) {
		t.Errorf("row genres = %v, want [Rock Pop]", doc.Rows[0].Values["genres"])
	}
}

func toStrings(t *testing.T, vs []any) []string {
	t.Helper()
	out := make([]string, len(vs))
	for i, v := range vs {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("element %v is not a string", v)
		}
		out[i] = s
	}
	return out
}

// TestDeriveFileTagsRowExcludesReadOnlyRemainder is the fixture with
// real teeth: it proves that none of the nine read-only audio
// properties (present inside Item._media_fields today) ever leaks into
// a row's own value dict, only into the shared template block. Without
// this check, a row could silently widen to include one, handing rung
// 8 a per-subject value no writer could ever reproduce.
func TestDeriveFileTagsRowExcludesReadOnlyRemainder(t *testing.T) {
	doc := deriveFileTagsDoc(t, []FileTagSubject{
		{CaseID: "smoke-test", ItemIndex: 0, AppliedFields: map[string]any{"title": "x"}},
	})

	for _, readonly := range wantReadonlyFields {
		if _, ok := doc.Rows[0].Values[readonly]; ok {
			t.Errorf("row values carry read-only field %q, want it confined to the template block", readonly)
		}
	}
}

// TestDeriveFileTagsIsDeterministic proves two independent shim
// processes deriving the same subjects agree on the exact bytes, the
// same discipline TestShimStdoutIsByteIdentical applies to the primary
// oracle shim.
func TestDeriveFileTagsIsDeterministic(t *testing.T) {
	subjects := []FileTagSubject{
		{CaseID: "smoke-test", ItemIndex: 0, AppliedFields: map[string]any{"title": "x", "year": 2020}},
	}
	opts := venvOptions(t)

	first, err := DeriveFileTags(context.Background(), opts, subjects)
	if err != nil {
		t.Fatalf("first derivation: %v", err)
	}
	second, err := DeriveFileTags(context.Background(), opts, subjects)
	if err != nil {
		t.Fatalf("second derivation: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("file-tags stdout differs between processes:\nfirst:  %s\nsecond: %s", first, second)
	}
}

// TestDeriveFileTagsRejectsNonPrimitiveAppliedField proves the shim's
// own faithfulness assertion fires: an AppliedFields value that is not
// already a JSON primitive (nor a list of them) must fail the
// derivation loudly rather than pass through as a stringified repr.
func TestDeriveFileTagsRejectsNonPrimitiveAppliedField(t *testing.T) {
	_, err := DeriveFileTags(context.Background(), venvOptions(t), []FileTagSubject{
		{
			CaseID:        "smoke-test",
			ItemIndex:     0,
			AppliedFields: map[string]any{"title": map[string]any{"nested": "not a primitive"}},
		},
	})
	if err == nil {
		t.Fatal("DeriveFileTags with a non-primitive applied field: got nil error, want one")
	}
}

// TestDeriveFileTagsMultipleSubjectsEachGetOwnRow proves every subject
// in one process gets its own row and its own write, rather than one
// subject's write leaking into the next's read-back.
func TestDeriveFileTagsMultipleSubjectsEachGetOwnRow(t *testing.T) {
	doc := deriveFileTagsDoc(t, []FileTagSubject{
		{CaseID: "case-a", ItemIndex: 0, AppliedFields: map[string]any{"title": "First"}},
		{CaseID: "case-a", ItemIndex: 1, AppliedFields: map[string]any{"title": "Second"}},
	})

	if len(doc.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(doc.Rows))
	}
	if doc.Rows[0].Values["title"] != "First" || doc.Rows[1].Values["title"] != "Second" {
		t.Errorf("rows = %v, want distinct titles First then Second", doc.Rows)
	}
}
