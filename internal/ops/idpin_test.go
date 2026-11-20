package ops

import (
	"errors"
	"testing"
)

// TestIDPinParseHandle pins the job handle's refusal sentence for each
// malformed shape, and that the all-zeros handle is accepted.
func TestIDPinParseHandle(t *testing.T) {
	for _, s := range []string{
		"",
		"0123456789abcde",
		"0123456789ABCDEF0123456789ABCDEF",
		"0123456789abcdef0123456789abcdeg",
	} {
		t.Run(s, func(t *testing.T) {
			_, err := ParseHandle(s)
			e, ok := errors.AsType[*Error](err)
			if !ok {
				t.Fatalf("ParseHandle(%q) = %v, want an *Error", s, err)
			}
			want := `job handle "` + s + `": want 32 lowercase hex characters`
			if e.Code != CodeInvalidRequest || e.Reason != want {
				t.Errorf("ParseHandle(%q) = code %q, reason %q; want code %q, reason %q", s, e.Code, e.Reason, CodeInvalidRequest, want)
			}
		})
	}

	const zeros = "00000000000000000000000000000000"
	got, err := ParseHandle(zeros)
	if err != nil || string(got) != zeros {
		t.Errorf("ParseHandle(zeros) = %q, %v; want the handle back and no error", got, err)
	}
}
