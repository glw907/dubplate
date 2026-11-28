package webapp

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

// The fixture script bodies and their sha256 sources, each computed
// once with `printf '%s' <body> | openssl dgst -sha256 -binary |
// base64` rather than by any function in this package.
const (
	bootScript     = "boot();"
	bootScriptHash = "sha256-BoYdILVZAmAQfhxvCgsBo5rtKe9f5WOaIkiGVKr7Y94="
	nextScript     = "next();"
	nextScriptHash = "sha256-KCLlghij5ML5UWrSlZtsRXyzhp7zax5Du0ypFHCFDyw="
)

// cspMeta returns a CSP meta tag carrying content verbatim, in the
// attribute order the build writes.
func cspMeta(content string) string {
	return `<meta http-equiv="content-security-policy" content="` + content + `">`
}

func TestDocumentCSP(t *testing.T) {
	const hashedMeta = "default-src 'self'; script-src 'self' '" + bootScriptHash + "' '" + nextScriptHash + "'"
	tests := []struct {
		name    string
		doc     string
		want    string
		wantErr error
	}{
		{
			name: "every inline script hashed",
			doc:  cspMeta(hashedMeta) + "<script>" + bootScript + "</script><p>x</p><script>" + nextScript + "</script>",
			want: "default-src 'self'; script-src 'self' '" + bootScriptHash + "' '" + nextScriptHash + "'; " +
				"style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'",
		},
		{
			name: "an entity-escaped meta is unescaped before it is read",
			doc:  cspMeta("script-src &#39;self&#39; &#39;"+bootScriptHash+"&#39;") + "<script>" + bootScript + "</script>",
			want: "default-src 'self'; script-src 'self' '" + bootScriptHash + "'; " +
				"style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'",
		},
		{
			name: "a document with no script and a hashless script-src",
			doc:  cspMeta("script-src 'self'") + "<p>x</p>",
			want: "default-src 'self'; script-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'",
		},
		{
			name:    "no CSP meta",
			doc:     `<meta name="viewport" content="width=device-width"><script>` + bootScript + "</script>",
			wantErr: errNoCSPMeta,
		},
		{
			name:    "a meta with no script-src",
			doc:     cspMeta("default-src 'self'") + "<script>" + bootScript + "</script>",
			wantErr: errNoScriptSrc,
		},
		{
			name:    "script-src allowing unsafe-inline",
			doc:     cspMeta("script-src 'self' 'unsafe-inline' '"+bootScriptHash+"'") + "<script>" + bootScript + "</script>",
			wantErr: errUnsafeInlineScript,
		},
		{
			name:    "an inline script the meta does not hash",
			doc:     cspMeta("script-src 'self' '"+bootScriptHash+"'") + "<script>" + bootScript + "</script><script>" + nextScript + "</script>",
			wantErr: errUnhashedScript,
		},
		{
			name:    "scripts and no hash at all",
			doc:     cspMeta("script-src 'self'") + "<script>" + bootScript + "</script>",
			wantErr: errUnhashedScript,
		},
		{
			name:    "a module script",
			doc:     cspMeta(hashedMeta) + `<script type="module">` + bootScript + "</script>",
			wantErr: errUnrecognizedScript,
		},
		{
			name:    "an uppercase attributed script",
			doc:     cspMeta(hashedMeta) + `<SCRIPT src="x.js"></SCRIPT>`,
			wantErr: errUnrecognizedScript,
		},
		{
			name:    "a script never closed",
			doc:     cspMeta(hashedMeta) + "<script>" + bootScript,
			wantErr: errUnrecognizedScript,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := documentCSP("login.html", []byte(tt.doc))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("documentCSPerror = %v, want %v", err, tt.wantErr)
				}
				if !strings.Contains(err.Error(), `"login.html"`) {
					t.Errorf("documentCSPerror = %q, want it to name the document", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("documentCSPerror = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("documentCSP=\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

// TestDocumentCSP_UnhashedScriptNamesItsHash proves the refusal names
// the hash it looked for, so an operator reading the startup error can
// find the script without rerunning the build.
func TestDocumentCSP_UnhashedScriptNamesItsHash(t *testing.T) {
	doc := cspMeta("script-src 'self' '"+bootScriptHash+"'") + "<script>" + nextScript + "</script>"
	_, err := documentCSP("members.html", []byte(doc))
	if err == nil || !strings.Contains(err.Error(), nextScriptHash) {
		t.Errorf("documentCSPerror = %v, want it to name %s", err, nextScriptHash)
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
