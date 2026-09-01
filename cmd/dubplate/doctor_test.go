package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/checks"
)

// TestPrintDoctorReportConfigErrNotProbed proves the not-Probed branch
// renders ConfigErr's own text, not just the fact that config failed. A
// parse error (an invalid port, say) is not an [config.ErrMissing] entry,
// so this is the operator's only view into why: without the Fprintf at
// doctor.go:73, this test fails even though every FieldStatus is present
// and unchanged.
func TestPrintDoctorReportConfigErrNotProbed(t *testing.T) {
	report := checks.DoctorReport{
		ConfigErr: fmt.Errorf("DUBPLATE_SMTP_PORT: invalid port %q", "abc"),
	}

	var buf bytes.Buffer
	if err := printDoctorReport(&buf, report); err != nil {
		t.Fatalf("printDoctorReport: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `invalid port "abc"`) {
		t.Errorf("output = %q, want it to contain the ConfigErr text", out)
	}
}

// TestPrintDoctorReportMalformedPingURL proves a malformed PingURLErrs
// entry renders both the slug it came from and the never-fetched
// rationale: a ping slug is skipped for syntax reasons only, and an
// operator reading the report needs to know why doctor did not just GET
// it to find out.
func TestPrintDoctorReportMalformedPingURL(t *testing.T) {
	report := checks.DoctorReport{
		Probed: true,
		PingURLErrs: map[string]error{
			"import": errors.New(`parse "http://[bad": missing ']' in host`),
		},
	}

	var buf bytes.Buffer
	if err := printDoctorReport(&buf, report); err != nil {
		t.Fatalf("printDoctorReport: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "import") {
		t.Errorf("output = %q, want the malformed slug %q", out, "import")
	}
	if !strings.Contains(out, "never fetched") || !strings.Contains(out, "false success ping") {
		t.Errorf("output = %q, want the never-fetched rationale", out)
	}
}

// TestPrintDoctorReportMissingField proves a [checks.FieldStatus] with
// Missing set renders as MISSING rather than the resolved status.
func TestPrintDoctorReportMissingField(t *testing.T) {
	report := checks.DoctorReport{
		Probed: true,
		Fields: []checks.FieldStatus{
			{Name: "DUBPLATE_SMTP_HOST", Missing: true},
		},
	}

	var buf bytes.Buffer
	if err := printDoctorReport(&buf, report); err != nil {
		t.Fatalf("printDoctorReport: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DUBPLATE_SMTP_HOST") || !strings.Contains(out, "MISSING") {
		t.Errorf("output = %q, want DUBPLATE_SMTP_HOST reported MISSING", out)
	}
}
