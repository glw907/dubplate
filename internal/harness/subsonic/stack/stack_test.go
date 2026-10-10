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
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
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

// oracleServer answers ping, getMusicFolders (one folder), and getScanStatus
// like an OpenSubsonic server.
func oracleServer(t *testing.T, scanning bool, count int) string {
	t.Helper()
	return newOracle(t, oracleOptions{scan: scan{scanning, count}, folders: 1}).addr
}

// oracleOptions shapes a fakeOracle.
type oracleOptions struct {
	scan scan
	// folders is the music folder count before any library is created.
	folders int
	// loginFailures is how many login attempts answer 401 first.
	loginFailures int
	// createStatus is the library creation's status; 0 answers 200.
	createStatus int
}

// fakeOracle serves the OpenSubsonic routes the stack polls and Navidrome's
// login and library creation, recording what it was asked.
type fakeOracle struct {
	addr string
	opts oracleOptions

	mu      sync.Mutex
	logins  int
	events  []string
	created []map[string]string
	auths   []string
}

func newOracle(t *testing.T, opts oracleOptions) *fakeOracle {
	t.Helper()
	f := &fakeOracle{opts: opts}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case strings.HasSuffix(r.URL.Path, "/ping"):
			f.events = append(f.events, "ping")
			_, _ = fmt.Fprint(w, `{"subsonic-response":{"status":"ok","version":"1.16.1"}}`)
		case strings.HasSuffix(r.URL.Path, "/getMusicFolders"):
			f.events = append(f.events, "getMusicFolders")
			var folders []string
			for i := range f.opts.folders + len(f.created) {
				folders = append(folders, fmt.Sprintf(`{"id":%d,"name":"f%d"}`, i+1, i+1))
			}
			_, _ = fmt.Fprintf(w, `{"subsonic-response":{"status":"ok","musicFolders":{"musicFolder":[%s]}}}`, strings.Join(folders, ","))
		case strings.HasSuffix(r.URL.Path, "/getScanStatus"):
			f.events = append(f.events, "getScanStatus")
			_, _ = fmt.Fprintf(w, `{"subsonic-response":{"status":"ok","scanStatus":{"scanning":%t,"count":%d}}}`, f.opts.scan.scanning, f.opts.scan.count)
		case r.Method == http.MethodPost && r.URL.Path == "/auth/login":
			f.events = append(f.events, "login")
			f.logins++
			if f.logins <= f.opts.loginFailures {
				http.Error(w, "not yet", http.StatusUnauthorized)
				return
			}
			_, _ = fmt.Fprint(w, `{"id":"u1","isAdmin":true,"token":"jwt-token","username":"admin"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/library":
			f.events = append(f.events, "create")
			f.auths = append(f.auths, r.Header.Get("X-ND-Authorization"))
			var body map[string]string
			if err := json.UnmarshalRead(r.Body, &body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if f.opts.createStatus != 0 {
				http.Error(w, "path not allowed", f.opts.createStatus)
				return
			}
			f.created = append(f.created, body)
			_, _ = fmt.Fprint(w, `{"id":"2"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	f.addr = srv.Listener.Addr().String()
	return f
}

// eventIndex returns the position of the first event named name, or -1.
func (f *fakeOracle) eventIndex(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Index(f.events, name)
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
	e.cfg = Config{Collections: []Collection{{Name: "seed", Dir: e.music, Tracks: 3}}, DigestFile: digests}
	swap(t, &podmanName, exe)
	swap(t, &readyTimeout, 400*time.Millisecond)
	swap(t, &pollInterval, 10*time.Millisecond)
	return e
}

// swap sets *p to v for the rest of the test.
func swap[T any](t *testing.T, p *T, v T) {
	t.Helper()
	old := *p
	*p = v
	t.Cleanup(func() { *p = old })
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
	if len(s.Oracles) != 2 || s.Oracles[0].Name != "navidrome" || s.Oracles[1].Name != "gonic" {
		t.Fatalf("oracles = %+v, want navidrome then gonic", s.Oracles)
	}
	navidrome, gonicOracle := s.Oracles[0], s.Oracles[1]
	if nd.Secret == "" || nd.Secret != navidrome.Password {
		t.Errorf("navidrome process env password = %q, want the run password %q", nd.Secret, navidrome.Password)
	}
	for _, a := range nd.Args {
		if strings.Contains(a, navidrome.Password) {
			t.Errorf("password appears in argv: %q", a)
		}
	}
	if gonicOracle.User != "admin" || gonicOracle.Password != "admin" || navidrome.User != "admin" {
		t.Errorf("credentials = %+v / %+v", navidrome, gonicOracle)
	}
	for _, o := range s.Oracles {
		if !strings.HasPrefix(o.BaseURL, "http://127.0.0.1:") {
			t.Errorf("%s base url = %q", o.Name, o.BaseURL)
		}
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
	readyTimeout = time.Minute
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
			name:    "digest file with no members",
			prep:    func(t *testing.T, e *env) { badDigest(t, *e, "{}") },
			want:    errImagesFile,
			mention: []string{"oracles.json", "no navidrome member", "no gonic member"},
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
			prep: func(t *testing.T, e *env) { podmanName = "podman-absent-from-path" },
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
		image   string // overrides the pinned reference when set
		inspect string
		want    string
		wantErr bool
		mention []string
	}{
		{name: "digest field", inspect: `[{"Digest":"` + pinDigest + `","RepoDigests":[]}]`, want: pinDigest},
		{name: "repo digest beside a different digest",
			inspect: `[{"Digest":"` + other + `","RepoDigests":["registry.example/nd@` + pinDigest + `"]}]`, want: pinDigest},
		{name: "no pin anywhere",
			inspect: `[{"Digest":"` + other + `","RepoDigests":["registry.example/nd@` + other + `"]}]`, wantErr: true,
			mention: []string{pinNavidrome, `Digest "` + other + `"`, "registry.example/nd@" + other}},
		{name: "tag-only repo digest does not match by prefix",
			inspect: `[{"Digest":"","RepoDigests":["docker.io/deluan/navidrome:latest"]}]`, wantErr: true,
			mention: []string{pinNavidrome, "docker.io/deluan/navidrome:latest"}},
		{name: "unpinned reference against an image with no digest", image: "docker.io/deluan/navidrome:latest",
			inspect: `[{"Digest":"","RepoDigests":[]}]`, wantErr: true,
			mention: []string{"docker.io/deluan/navidrome:latest"}},
		{name: "empty inspect list", inspect: `[]`, wantErr: true, mention: []string{pinNavidrome}},
		{name: "malformed inspect", inspect: `not json`, wantErr: true, mention: []string{pinNavidrome}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t, oracleServer(t, false, 3), oracleServer(t, false, 3))
			s, err := Start(t.Context(), e.cfg)
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			t.Setenv("FAKE_INSPECT_NAVIDROME", tc.inspect)
			o := s.Oracles[0]
			if tc.image != "" {
				o.image = tc.image
			}
			got, err := s.ImageDigest(t.Context(), o)
			if tc.wantErr {
				if !errors.Is(err, errImageNotPinned) {
					t.Fatalf("error = %v, want errImageNotPinned", err)
				}
				for _, m := range tc.mention {
					if !strings.Contains(err.Error(), m) {
						t.Errorf("error %q does not mention %q", err, m)
					}
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

// twoCollections gives e two disjoint sibling roots holding 2 and 1 tracks.
func twoCollections(e *env) {
	e.cfg.Collections = []Collection{
		{Name: "first", Dir: filepath.Join(e.music, "first"), Tracks: 2},
		{Name: "second", Dir: filepath.Join(e.music, "second"), Tracks: 1},
	}
}

func TestStartMountsEveryCollection(t *testing.T) {
	ndOracle := newOracle(t, oracleOptions{scan: scan{false, 3}, folders: 1, loginFailures: 2})
	gonicOracle := newOracle(t, oracleOptions{scan: scan{false, 3}, folders: 2})
	e := newEnv(t, ndOracle.addr, gonicOracle.addr)
	twoCollections(&e)
	if _, err := Start(t.Context(), e.cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	runs := e.verb(t, "run")
	if len(runs) != 2 {
		t.Fatalf("run calls = %d, want 2", len(runs))
	}
	wantMounts := []string{
		filepath.Join(e.music, "first") + ":/collections/first:ro,z",
		filepath.Join(e.music, "second") + ":/collections/second:ro,z",
	}
	for _, c := range runs {
		if got := flagValues(c.Args, "-v"); !slices.Equal(got, wantMounts) {
			t.Errorf("-v = %q, want %q", got, wantMounts)
		}
	}
	if envs := flagValues(runs[0].Args, "-e"); !slices.Contains(envs, "ND_MUSICFOLDER=/collections/first") {
		t.Errorf("navidrome -e = %q, want its MusicFolder at the first collection", envs)
	}
	wantAlias := "GONIC_MUSIC_PATH=first->/collections/first,second->/collections/second"
	if envs := flagValues(runs[1].Args, "-e"); !slices.Contains(envs, wantAlias) {
		t.Errorf("gonic -e = %q, want %q", envs, wantAlias)
	}

	// Only the second collection is created; the first is the MusicFolder.
	wantBody := []map[string]string{{"name": "second", "path": "/collections/second"}}
	if !reflect.DeepEqual(ndOracle.created, wantBody) {
		t.Errorf("created libraries = %v, want %v", ndOracle.created, wantBody)
	}
	if !slices.Equal(ndOracle.auths, []string{"Bearer jwt-token"}) {
		t.Errorf("creation auth headers = %q, want the login token as a bearer", ndOracle.auths)
	}
	if len(gonicOracle.created) != 0 {
		t.Errorf("gonic was asked to create %v", gonicOracle.created)
	}
	if ndOracle.logins != 3 {
		t.Errorf("login attempts = %d, want 3 (two refused, then one accepted)", ndOracle.logins)
	}
	ping, login, create, folders := ndOracle.eventIndex("ping"), ndOracle.eventIndex("login"),
		ndOracle.eventIndex("create"), ndOracle.eventIndex("getMusicFolders")
	if ping < 0 || ping > login || login > create || create > folders {
		t.Errorf("navidrome events = %q, want ping, login, create, then the folder check", ndOracle.events)
	}
}

func TestStartCollectionRefusals(t *testing.T) {
	tests := []struct {
		name    string
		cols    func(e env) []Collection
		mention []string
	}{
		{name: "empty list", cols: func(env) []Collection { return nil }, mention: []string{"no collections"}},
		{
			name: "one name twice",
			cols: func(e env) []Collection {
				return []Collection{{"x", filepath.Join(e.music, "a"), 1}, {"x", filepath.Join(e.music, "b"), 1}}
			},
			mention: []string{`"x" is used twice`},
		},
		{
			name: "nested paths",
			cols: func(e env) []Collection {
				return []Collection{{"outer", e.music, 1}, {"inner", filepath.Join(e.music, "inner"), 1}}
			},
			mention: []string{`"outer"`, `"inner"`, "overlap"},
		},
		{
			name: "same path",
			cols: func(e env) []Collection {
				return []Collection{{"one", e.music, 1}, {"two", e.music, 1}}
			},
			mention: []string{`"one"`, `"two"`, "overlap"},
		},
		{name: "empty name", cols: func(e env) []Collection { return []Collection{{"", e.music, 1}} }, mention: []string{`""`}},
		{name: "dot name", cols: func(e env) []Collection { return []Collection{{".hidden", e.music, 1}} }, mention: []string{`".hidden"`}},
		{name: "comma name", cols: func(e env) []Collection { return []Collection{{"a,b", e.music, 1}} }, mention: []string{`"a,b"`}},
		{name: "arrow name", cols: func(e env) []Collection { return []Collection{{"a->b", e.music, 1}} }, mention: []string{`"a->b"`}},
		{name: "equals name", cols: func(e env) []Collection { return []Collection{{"a=b", e.music, 1}} }, mention: []string{`"a=b"`}},
		{name: "slash name", cols: func(e env) []Collection { return []Collection{{"a/b", e.music, 1}} }, mention: []string{`"a/b"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t, "127.0.0.1:1", "127.0.0.1:1")
			e.cfg.Collections = tt.cols(e)
			_, err := Start(t.Context(), e.cfg)
			if !errors.Is(err, errCollections) {
				t.Fatalf("Start error = %v, want errCollections", err)
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

func TestStartLibraryCreationError(t *testing.T) {
	ndOracle := newOracle(t, oracleOptions{scan: scan{false, 3}, folders: 1, createStatus: http.StatusBadRequest})
	e := newEnv(t, ndOracle.addr, oracleServer(t, false, 3))
	twoCollections(&e)
	_, err := Start(t.Context(), e.cfg)
	if !errors.Is(err, errLibrary) {
		t.Fatalf("Start error = %v, want errLibrary", err)
	}
	for _, m := range []string{"navidrome", "second", "400", "path not allowed"} {
		if !strings.Contains(err.Error(), m) {
			t.Errorf("error %q does not mention %q", err, m)
		}
	}
	assertRemovedWithVolumes(t, e)
}

func TestStartLoginNeverSucceeds(t *testing.T) {
	ndOracle := newOracle(t, oracleOptions{scan: scan{false, 3}, folders: 1, loginFailures: 1 << 20})
	e := newEnv(t, ndOracle.addr, oracleServer(t, false, 3))
	twoCollections(&e)
	_, err := Start(t.Context(), e.cfg)
	if !errors.Is(err, errNotReady) || !strings.Contains(err.Error(), "navidrome login did not succeed") {
		t.Fatalf("Start error = %v, want errNotReady naming the login", err)
	}
	if len(ndOracle.created) != 0 {
		t.Errorf("libraries created without a login: %v", ndOracle.created)
	}
	assertRemovedWithVolumes(t, e)
}

func TestStartFolderCount(t *testing.T) {
	tests := []struct {
		name          string
		two           bool
		nd, gonic     oracleOptions
		wantErr       bool
		is            error
		mention       []string
		wantNoMention []string
	}{
		{
			name: "one folder per collection",
			two:  true,
			nd:   oracleOptions{scan: scan{false, 3}, folders: 1}, gonic: oracleOptions{scan: scan{false, 3}, folders: 2},
		},
		{
			name: "gonic lists one folder for two collections",
			two:  true,
			nd:   oracleOptions{scan: scan{false, 3}, folders: 1}, gonic: oracleOptions{scan: scan{false, 3}, folders: 1},
			wantErr: true, is: errFolderCount, mention: []string{"gonic lists 1 music folders, run has 2 collections"}, wantNoMention: []string{"navidrome"},
		},
		{
			name: "navidrome lists three folders for two collections",
			two:  true,
			nd:   oracleOptions{scan: scan{false, 3}, folders: 2}, gonic: oracleOptions{scan: scan{false, 3}, folders: 2},
			wantErr: true, is: errFolderCount, mention: []string{"navidrome lists 3 music folders, run has 2 collections"},
		},
		{
			name: "one collection, two folders",
			nd:   oracleOptions{scan: scan{false, 3}, folders: 2}, gonic: oracleOptions{scan: scan{false, 3}, folders: 1},
			wantErr: true, is: errFolderCount, mention: []string{"navidrome lists 2 music folders, run has 1 collections"},
		},
		{
			name: "one collection, no folders",
			nd:   oracleOptions{scan: scan{false, 3}, folders: 1}, gonic: oracleOptions{scan: scan{false, 3}},
			wantErr: true, is: errFolderCount, mention: []string{"gonic lists 0 music folders, run has 1 collections"},
		},
		{
			name: "scan count short of the total",
			two:  true,
			nd:   oracleOptions{scan: scan{false, 2}, folders: 1}, gonic: oracleOptions{scan: scan{false, 3}, folders: 2},
			wantErr: true, is: errTrackCount, mention: []string{"navidrome counted 2, manifest has 3"}, wantNoMention: []string{"gonic"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t, newOracle(t, tt.nd).addr, newOracle(t, tt.gonic).addr)
			if tt.two {
				twoCollections(&e)
			}
			_, err := Start(t.Context(), e.cfg)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Start: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.is) {
				t.Fatalf("Start error = %v, want %v", err, tt.is)
			}
			for _, m := range tt.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q does not mention %q", err, m)
				}
			}
			for _, m := range tt.wantNoMention {
				if strings.Contains(err.Error(), m) {
					t.Errorf("error %q mentions %q", err, m)
				}
			}
			assertRemovedWithVolumes(t, e)
		})
	}
}
