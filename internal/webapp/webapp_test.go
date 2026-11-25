package webapp

import (
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/web"
)

func TestStampablePalettes(t *testing.T) {
	selectors := map[string]struct{}{
		"dubplate-label":       {},
		"dubplate-label-night": {},
		"dubplate-system":      {},
		"light":                {}, // no "dubplate-" prefix: DaisyUI's own bundled default, must be excluded
		"dark":                 {},
	}
	got := stampablePalettes(selectors, "dubplate")
	want := []string{"label", "label-night", "system"}
	if len(got) != len(want) {
		t.Fatalf("stampablePalettes = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("stampablePalettes[%d] = %q, want %q", i, got[i], w)
		}
	}
}

// TestStampablePalettes_RealBundle proves the stampable set the real
// built bundle's login documents yield is exactly {label, label-night,
// system}, the set cmd/dubplate-server's real-bundle stamp test reads
// its cookie rows against.
func TestStampablePalettes_RealBundle(t *testing.T) {
	const loginDocumentName, loginLinkDocumentName = "login.html", "login/link.html"
	buildFS := web.Build
	loginRaw, err := fs.ReadFile(buildFS, loginDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginDocumentName, err)
	}
	linkRaw, err := fs.ReadFile(buildFS, loginLinkDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginLinkDocumentName, err)
	}
	themes, err := extractThemes(buildFS,
		namedDocument{path: loginDocumentName, raw: loginRaw},
		namedDocument{path: loginLinkDocumentName, raw: linkRaw},
	)
	if err != nil {
		t.Fatalf("extractThemes: %v", err)
	}
	gotPalettes := stampablePalettes(themes.selectors, "dubplate")
	wantPalettes := []string{"label", "label-night", "system"}
	if !slices.Equal(gotPalettes, wantPalettes) {
		t.Fatalf("stampablePalettes(real bundle) = %v, want %v", gotPalettes, wantPalettes)
	}
}

// TestCheckVariantCSP_RefusesAPolicyUnlikeTheHeader proves a served
// variant must yield exactly the policy the header carries, computed
// from the build's own document: a variant whose meta lists one more
// hash still covers its every script, so only the equality check can
// refuse it.
func TestCheckVariantCSP_RefusesAPolicyUnlikeTheHeader(t *testing.T) {
	source := func(script string) string {
		sum := sha256.Sum256([]byte(script))
		return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	}
	doc := func(sources string) []byte {
		return []byte(`<meta http-equiv="content-security-policy" content="script-src 'self' ` + sources + `"><script>boot();</script>`)
	}
	boot := source("boot();")
	served, err := documentCSP("login.html", doc(boot))
	if err != nil {
		t.Fatalf("fixture: documentCSP:%v", err)
	}
	if err := checkVariantCSP("login.html", map[string][]byte{"label": doc(boot)}, served); err != nil {
		t.Fatalf("checkVariantCSP on a matching variant: %v", err)
	}
	variants := map[string][]byte{"label": doc(boot), "label-night": doc(boot + " " + source("next();"))}
	err = checkVariantCSP("login.html", variants, served)
	if err == nil {
		t.Fatal("checkVariantCSP: got nil error, want the policy mismatch refusal")
	}
	if want := `the "label-night" variant of document "login.html": its CSP differs from the served header`; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), want)
	}
}
