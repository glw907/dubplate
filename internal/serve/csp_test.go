package serve

import (
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
			got, err := DocumentCSP("login.html", []byte(tt.doc))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("DocumentCSP error = %v, want %v", err, tt.wantErr)
				}
				if !strings.Contains(err.Error(), `"login.html"`) {
					t.Errorf("DocumentCSP error = %q, want it to name the document", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("DocumentCSP error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("DocumentCSP =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

// TestDocumentCSP_UnhashedScriptNamesItsHash proves the refusal names
// the hash it looked for, so an operator reading the startup error can
// find the script without rerunning the build.
func TestDocumentCSP_UnhashedScriptNamesItsHash(t *testing.T) {
	doc := cspMeta("script-src 'self' '"+bootScriptHash+"'") + "<script>" + nextScript + "</script>"
	_, err := DocumentCSP("members.html", []byte(doc))
	if err == nil || !strings.Contains(err.Error(), nextScriptHash) {
		t.Errorf("DocumentCSP error = %v, want it to name %s", err, nextScriptHash)
	}
}
