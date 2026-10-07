package stack

import (
	"bufio"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	pinNavidrome = "docker.io/deluan/navidrome@sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52"
	pinGonic     = "docker.io/sentriz/gonic@sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49"
)

// TestMain re-executes the test binary as a stand-in for podman when
// STACK_FAKE_PODMAN is set, so Start's subprocess calls are observable.
func TestMain(m *testing.M) {
	if os.Getenv("STACK_FAKE_PODMAN") == "1" {
		os.Exit(fakePodman(os.Args[1:]))
	}
	os.Exit(m.Run())
}

// fakeCall is one recorded podman invocation.
type fakeCall struct {
	Args   []string `json:"args"`
	Secret string   `json:"secret"` // ND_DEVAUTOCREATEADMINPASSWORD as the process saw it
}

func fakePodman(args []string) int {
	secret := os.Getenv("ND_DEVAUTOCREATEADMINPASSWORD")
	line, _ := json.Marshal(fakeCall{Args: args, Secret: secret})
	f, err := os.OpenFile(os.Getenv("FAKE_LOG"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	_, _ = fmt.Fprintf(f, "%s\n", line)
	_ = f.Close()
	switch args[0] {
	case "run":
		if secret != "" {
			_ = os.WriteFile(os.Getenv("FAKE_SECRET_FILE"), []byte(secret), 0o644)
		}
		fmt.Println("0123456789ab")
	case "port":
		name := args[1]
		key := "FAKE_ADDR_NAVIDROME"
		if strings.HasSuffix(name, "-gonic") {
			key = "FAKE_ADDR_GONIC"
		}
		fmt.Println(os.Getenv(key))
	case "container":
		// container inspect --format {{.Image}} NAME
		name := args[len(args)-1]
		if strings.HasSuffix(name, "-gonic") {
			fmt.Println("image-gonic")
		} else {
			fmt.Println("image-navidrome")
		}
	case "image":
		// image inspect ID
		key := "FAKE_INSPECT_NAVIDROME"
		if args[len(args)-1] == "image-gonic" {
			key = "FAKE_INSPECT_GONIC"
		}
		fmt.Println(os.Getenv(key))
	case "logs":
		b, _ := os.ReadFile(os.Getenv("FAKE_SECRET_FILE"))
		fmt.Printf("starting\nadmin password is %s\n", b)
	}
	return 0
}

// oracleServer answers ping and getScanStatus like an OpenSubsonic server.
func oracleServer(t *testing.T, scanning bool, count int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/ping"):
			_, _ = fmt.Fprint(w, `{"subsonic-response":{"status":"ok","version":"1.16.1"}}`)
		case strings.HasSuffix(r.URL.Path, "/getScanStatus"):
			_, _ = fmt.Fprintf(w, `{"subsonic-response":{"status":"ok","scanStatus":{"scanning":%t,"count":%d}}}`, scanning, count)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.Listener.Addr().String()
}

// scan is the scan state an oracleServer reports.
type scan struct {
	scanning bool
	count    int
}

type env struct {
	log, music, digests string
	cfg                 Config
}

func newEnv(t *testing.T, navidromeAddr, gonicAddr string) env {
	t.Helper()
	dir := t.TempDir()
	digests := filepath.Join(dir, "oracles.json")
	b, err := json.Marshal(images{Navidrome: pinNavidrome, Gonic: pinGonic})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(digests, b, 0o644); err != nil {
		t.Fatal(err)
	}
	e := env{log: filepath.Join(dir, "calls.jsonl"), music: filepath.Join(dir, "music"), digests: digests}
	t.Setenv("STACK_FAKE_PODMAN", "1")
	t.Setenv("FAKE_LOG", e.log)
	t.Setenv("FAKE_SECRET_FILE", filepath.Join(dir, "secret"))
	t.Setenv("FAKE_ADDR_NAVIDROME", navidromeAddr)
	t.Setenv("FAKE_ADDR_GONIC", gonicAddr)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	e.cfg = Config{
		MusicDir: e.music, Tracks: 3, DigestFile: digests, Podman: exe,
		ReadyTimeout: 400 * time.Millisecond, Poll: 10 * time.Millisecond,
	}
	return e
}

func (e env) calls(t *testing.T) []fakeCall {
	t.Helper()
	f, err := os.Open(e.log)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var calls []fakeCall
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var c fakeCall
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			t.Fatalf("decode fake log: %v", err)
		}
		calls = append(calls, c)
	}
	return calls
}

func (e env) verb(t *testing.T, verb string) []fakeCall {
	t.Helper()
	var out []fakeCall
	for _, c := range e.calls(t) {
		if c.Args[0] == verb {
			out = append(out, c)
		}
	}
	return out
}

// flagValues returns every value that follows flag in args.
func flagValues(args []string, flag string) []string {
	var vs []string
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			vs = append(vs, args[i+1])
		}
	}
	return vs
}

func TestStartRunArguments(t *testing.T) {
	e := newEnv(t, oracleServer(t, false, 3), oracleServer(t, false, 3))
	s, err := Start(t.Context(), e.cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	runs := e.verb(t, "run")
	if len(runs) != 2 {
		t.Fatalf("run calls = %d, want 2", len(runs))
	}
	nd, gonic := runs[0], runs[1]

	hostPublish := regexp.MustCompile(`^127\.0\.0\.1::\d+$`)
	label := regexp.MustCompile(`^dubplate-harness-run=[0-9a-f]{12}$`)
	cases := []struct {
		name    string
		call    fakeCall
		ref     string
		port    string
		envWant []string
		tmpfs   []string
		bareEnv []string
	}{
		{
			name: "navidrome", call: nd, ref: pinNavidrome, port: "4533",
			envWant: []string{
				"ND_MUSICFOLDER=/music", "ND_ENABLEINSIGHTSCOLLECTOR=false",
				"ND_DEEZER_ENABLED=false", "ND_LASTFM_ENABLED=false",
				"ND_LISTENBRAINZ_ENABLED=false", "ND_SCANNER_SCANONSTARTUP=true",
			},
			bareEnv: []string{"ND_DEVAUTOCREATEADMINPASSWORD"},
		},
		{
			name: "gonic", call: gonic, ref: pinGonic, port: "4747",
			envWant: []string{
				"GONIC_MUSIC_PATH=/music", "GONIC_LISTEN_ADDR=0.0.0.0:4747",
				"GONIC_SCAN_AT_START_ENABLED=true",
				"GONIC_MULTI_VALUE_ARTIST=multi", "GONIC_MULTI_VALUE_ALBUM_ARTIST=multi",
			},
			tmpfs: []string{"/data", "/cache", "/podcasts", "/playlists"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.call.Args
			if got := flagValues(args, "-v"); !slices.Equal(got, []string{e.music + ":/music:ro,z"}) {
				t.Errorf("-v = %q, want the read-only shared-label mount", got)
			}
			pubs := flagValues(args, "-p")
			if len(pubs) != 1 || !hostPublish.MatchString(pubs[0]) || !strings.HasSuffix(pubs[0], "::"+tc.port) {
				t.Errorf("-p = %q, want one loopback publish of %s with no fixed host port", pubs, tc.port)
			}
			if ls := flagValues(args, "--label"); len(ls) != 1 || !label.MatchString(ls[0]) {
				t.Errorf("--label = %q, want one run label", ls)
			}
			if got := args[len(args)-1]; got != tc.ref {
				t.Errorf("image = %q, want %q", got, tc.ref)
			}
			envs := flagValues(args, "-e")
			wantEnv := append(slices.Clone(tc.envWant), tc.bareEnv...)
			slices.Sort(envs)
			slices.Sort(wantEnv)
			if !slices.Equal(envs, wantEnv) {
				t.Errorf("-e = %q, want %q", envs, wantEnv)
			}
			gotTmpfs := flagValues(args, "--tmpfs")
			if !slices.Equal(gotTmpfs, tc.tmpfs) {
				t.Errorf("--tmpfs = %q, want %q", gotTmpfs, tc.tmpfs)
			}
		})
	}
	if nd.Secret == "" || nd.Secret != s.Navidrome.Password {
		t.Errorf("navidrome process env password = %q, want the run password %q", nd.Secret, s.Navidrome.Password)
	}
	for _, a := range nd.Args {
		if strings.Contains(a, s.Navidrome.Password) {
			t.Errorf("password appears in argv: %q", a)
		}
	}
	if s.Gonic.User != "admin" || s.Gonic.Password != "admin" || s.Navidrome.User != "admin" {
		t.Errorf("credentials = %+v / %+v", s.Navidrome, s.Gonic)
	}
	if !strings.HasPrefix(s.Navidrome.BaseURL, "http://127.0.0.1:") {
		t.Errorf("base url = %q", s.Navidrome.BaseURL)
	}
}

func TestCloseRemovesContainersWithVolumes(t *testing.T) {
	e := newEnv(t, oracleServer(t, false, 3), oracleServer(t, false, 3))
	s, err := Start(t.Context(), e.cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := len(e.verb(t, "rm")); got != 0 {
		t.Fatalf("rm calls before Close = %d, want 0", got)
	}
	s.Close(t.Context())
	assertRemovedWithVolumes(t, e)
}

// assertRemovedWithVolumes fails unless one rm call carries -v and names
// every container a run call started.
func assertRemovedWithVolumes(t *testing.T, e env) {
	t.Helper()
	var started []string
	for _, c := range e.verb(t, "run") {
		started = append(started, flagValues(c.Args, "--name")...)
	}
	if len(started) == 0 {
		t.Fatal("fake saw no run call")
	}
	rms := e.verb(t, "rm")
	if len(rms) == 0 {
		t.Fatal("fake saw no rm call")
	}
	for _, name := range started {
		found := false
		for _, rm := range rms {
			if slices.Contains(rm.Args, "-v") && slices.Contains(rm.Args, "-f") && slices.Contains(rm.Args, name) {
				found = true
			}
		}
		if !found {
			t.Errorf("no rm -f -v call names %s: %v", name, rms)
		}
	}
}

func TestStartTearsDownOnCancellation(t *testing.T) {
	// Never finishes scanning, so Start waits until the context is cancelled.
	e := newEnv(t, oracleServer(t, true, 0), oracleServer(t, true, 0))
	e.cfg.ReadyTimeout = time.Minute
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		for len(e.verb(t, "port")) < 2 {
			time.Sleep(5 * time.Millisecond)
		}
		cancel()
	}()
	s, err := Start(ctx, e.cfg)
	if err == nil {
		s.Close(ctx)
		t.Fatal("Start succeeded under a cancelled context")
	}
	assertRemovedWithVolumes(t, e)
}

func TestStartRefusals(t *testing.T) {
	badDigest := func(t *testing.T, e env, content string) {
		t.Helper()
		if err := os.WriteFile(e.digests, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tests := []struct {
		name    string
		prep    func(*testing.T, *env)
		want    error
		mention []string
	}{
		{
			name: "tag-only reference",
			prep: func(t *testing.T, e *env) {
				badDigest(t, *e, `{"navidrome":"docker.io/deluan/navidrome:0.64.2","gonic":"`+pinGonic+`"}`)
			},
			want:    errNotPinned,
			mention: []string{"docker.io/deluan/navidrome:0.64.2"},
		},
		{
			name: "both references tag-only",
			prep: func(t *testing.T, e *env) {
				badDigest(t, *e, `{"navidrome":"a:1","gonic":"b:2"}`)
			},
			want:    errNotPinned,
			mention: []string{`"a:1"`, `"b:2"`},
		},
		{
			name: "short digest",
			prep: func(t *testing.T, e *env) {
				badDigest(t, *e, `{"navidrome":"a@sha256:abc","gonic":"`+pinGonic+`"}`)
			},
			want: errNotPinned,
		},
		{
			name: "digest file absent",
			prep: func(t *testing.T, e *env) {
				e.cfg.DigestFile = filepath.Join(filepath.Dir(e.digests), "none.json")
			},
			want:    errImagesFile,
			mention: []string{"none.json"},
		},
		{
			name:    "digest file empty",
			prep:    func(t *testing.T, e *env) { badDigest(t, *e, "") },
			want:    errImagesFile,
			mention: []string{"oracles.json"},
		},
		{
			name:    "digest file malformed",
			prep:    func(t *testing.T, e *env) { badDigest(t, *e, `{"navidrome":`) },
			want:    errImagesFile,
			mention: []string{"oracles.json"},
		},
		{
			name: "digest file with unknown member",
			prep: func(t *testing.T, e *env) {
				badDigest(t, *e, `{"navidrome":"`+pinNavidrome+`","gonic":"`+pinGonic+`","extra":1}`)
			},
			want: errImagesFile,
		},
		{
			name: "podman absent",
			prep: func(t *testing.T, e *env) { e.cfg.Podman = "podman-absent-from-path" },
			want: ErrNoPodman, mention: []string{"podman-absent-from-path"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t, "127.0.0.1:1", "127.0.0.1:1")
			tt.prep(t, &e)
			_, err := Start(t.Context(), e.cfg)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Start error = %v, want %v", err, tt.want)
			}
			for _, m := range tt.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q does not mention %q", err, m)
				}
			}
			if got := e.calls(t); len(got) != 0 {
				t.Errorf("podman ran %d times before the refusal: %v", len(got), got)
			}
		})
	}
}

func TestStartReadiness(t *testing.T) {
	tests := []struct {
		name       string
		nd, gonic  scan
		want       []error
		mention    []string
		notMention []string
	}{
		{
			name: "scan never finishes",
			nd:   scan{true, 0}, gonic: scan{false, 3},
			want:    []error{errNotReady},
			mention: []string{"navidrome"}, notMention: []string{"gonic"},
		},
		{
			name: "finished scan counts zero",
			nd:   scan{false, 3}, gonic: scan{false, 0},
			want:    []error{errTrackCount},
			mention: []string{"gonic counted 0, manifest has 3"}, notMention: []string{"navidrome"},
		},
		{
			name: "finished scan counts too many",
			nd:   scan{false, 4}, gonic: scan{false, 3},
			want:    []error{errTrackCount},
			mention: []string{"navidrome counted 4, manifest has 3"},
		},
		{
			name: "both fail",
			nd:   scan{false, 0}, gonic: scan{true, 0},
			want:    []error{errTrackCount, errNotReady},
			mention: []string{"navidrome counted 0", "gonic scan did not finish"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t,
				oracleServer(t, tt.nd.scanning, tt.nd.count),
				oracleServer(t, tt.gonic.scanning, tt.gonic.count))
			_, err := Start(t.Context(), e.cfg)
			if err == nil {
				t.Fatal("Start succeeded")
			}
			for _, w := range tt.want {
				if !errors.Is(err, w) {
					t.Errorf("error is not %v: %v", w, err)
				}
			}
			for _, m := range tt.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q does not mention %q", err, m)
				}
			}
			for _, m := range tt.notMention {
				if strings.Contains(err.Error(), m) {
					t.Errorf("error %q mentions %q", err, m)
				}
			}
			assertRemovedWithVolumes(t, e)
		})
	}
}

func TestStartRedactsLogTail(t *testing.T) {
	e := newEnv(t, oracleServer(t, true, 0), oracleServer(t, false, 3))
	_, err := Start(t.Context(), e.cfg)
	if err == nil {
		t.Fatal("Start succeeded")
	}
	secret, rerr := os.ReadFile(filepath.Join(filepath.Dir(e.log), "secret"))
	if rerr != nil || len(secret) < 8 {
		t.Fatalf("fake recorded no run password: %q %v", secret, rerr)
	}
	msg := err.Error()
	if !strings.Contains(msg, "admin password is [redacted]") {
		t.Errorf("error lacks the redacted log line:\n%s", msg)
	}
	if strings.Contains(msg, string(secret)) {
		t.Errorf("error carries the run password:\n%s", msg)
	}
}

func TestImageDigest(t *testing.T) {
	const other = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	pinDigest := pinNavidrome[strings.Index(pinNavidrome, "@")+1:]
	tests := []struct {
		name    string
		inspect string
		want    string
		wantErr bool
	}{
		{"digest field", `[{"Digest":"` + pinDigest + `","RepoDigests":[]}]`, pinDigest, false},
		{"repo digest beside a different digest",
			`[{"Digest":"` + other + `","RepoDigests":["registry.example/nd@` + pinDigest + `"]}]`, pinDigest, false},
		{"no pin anywhere",
			`[{"Digest":"` + other + `","RepoDigests":["registry.example/nd@` + other + `"]}]`, "", true},
		{"tag-only repo digest does not match by prefix",
			`[{"Digest":"","RepoDigests":["docker.io/deluan/navidrome:latest"]}]`, "", true},
		{"empty inspect list", `[]`, "", true},
		{"malformed inspect", `not json`, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t, oracleServer(t, false, 3), oracleServer(t, false, 3))
			s, err := Start(t.Context(), e.cfg)
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			t.Setenv("FAKE_INSPECT_NAVIDROME", tc.inspect)
			got, err := s.ImageDigest(t.Context(), s.Navidrome)
			if tc.wantErr {
				if !errors.Is(err, ErrImageNotPinned) {
					t.Fatalf("error = %v, want ErrImageNotPinned", err)
				}
				if !strings.Contains(err.Error(), pinNavidrome) {
					t.Errorf("error %q does not name the image %s", err, pinNavidrome)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("digest = %q, want %q", got, tc.want)
			}
		})
	}
}
