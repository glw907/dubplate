package notify

import (
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// TestConfigFrom_ResolvesTheSMTPKeys proves [ConfigFrom] reads each
// field from its own key of the selected identity source, with the port
// as an int. The registry declares one source whose Required keys are
// [Keys], so the test needs no other package's source.
func TestConfigFrom_ResolvesTheSMTPKeys(t *testing.T) {
	reg := config.NewRegistry()
	reg.AddWallRow(config.WallRow{Key: "DUBPLATE_IDENTITY_SOURCE", Description: "the source", Sensitivity: config.Sensitive})
	reg.AddIdentitySource(config.IdentitySource{Name: "mail", Selector: "DUBPLATE_IDENTITY_SOURCE", Required: Keys()})
	cfg, err := config.Load([]string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
		"DUBPLATE_IDENTITY_SOURCE=mail",
		"DUBPLATE_SMTP_HOST=smtp.example.org",
		"DUBPLATE_SMTP_PORT=587",
		"DUBPLATE_SMTP_USER=mailer",
		"DUBPLATE_SMTP_PASSWORD=fixture-password",
		"DUBPLATE_SMTP_FROM=Music <music@example.org>",
	}, reg)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	got, err := ConfigFrom(cfg)
	if err != nil {
		t.Fatalf("ConfigFrom: %v", err)
	}
	want := Config{Host: "smtp.example.org", Port: 587, User: "mailer", Password: "fixture-password", From: "Music <music@example.org>"}
	if got != want {
		t.Errorf("ConfigFrom() = %+v, want %+v", got, want)
	}
}

// TestValidateKeys proves the two value rules [ValidateKeys] applies,
// sender address first, each failure naming its key, and that a key
// absent from values is left alone.
func TestValidateKeys(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		want   string
	}{
		{"both valid", map[string]string{fromKey: "music@example.org", portKey: "587"}, ""},
		{"both absent", map[string]string{}, ""},
		{"port zero", map[string]string{portKey: "0"}, "DUBPLATE_SMTP_PORT: 0 is not in 1 through 65535"},
		{"port above range", map[string]string{portKey: "65536"}, "DUBPLATE_SMTP_PORT: 65536 is not in 1 through 65535"},
		{"from with a line feed", map[string]string{fromKey: "music@example.org\nBcc: x@example.org"}, "DUBPLATE_SMTP_FROM: invalid email address"},
		{"both bad", map[string]string{fromKey: "not an address", portKey: "abc"},
			"DUBPLATE_SMTP_FROM: invalid email address\nDUBPLATE_SMTP_PORT: not an integer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeys(tt.values)
			got := ""
			if err != nil {
				got = err.Error()
			}
			if got != tt.want {
				t.Errorf("ValidateKeys() = %q, want %q", got, tt.want)
			}
		})
	}
}
