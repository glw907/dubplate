package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake"
)

func TestRegisterIntakeConfig_StagingWallRow(t *testing.T) {
	reg := newProductionRegistry()

	var row config.WallRow
	var found bool
	for _, r := range reg.OwnedWall() {
		if r.Name == "DUBPLATE_STAGING_DIR" {
			row, found = r, true
		}
	}
	if !found {
		t.Fatal("DUBPLATE_STAGING_DIR is not an owned wall row")
	}
	if row.Mark != config.Sensitive {
		t.Errorf("mark = %v, want Sensitive", row.Mark)
	}
	if want := "the directory on the music volume where uploads stage before admission"; row.Description != want {
		t.Errorf("description = %q, want %q", row.Description, want)
	}
}

// TestRegisterIntakeConfig_RetiresTheIntakeClass proves the class edits
// of the commit that declares the wall row: the intake reserved class is
// gone, its two names are superseded rows naming the staging row, and
// rung 9's pipeline class no longer claims staging. A name two families
// claim panics at registration, so building the registry at all is the
// first half of the proof.
func TestRegisterIntakeConfig_RetiresTheIntakeClass(t *testing.T) {
	reg := newProductionRegistry()

	for _, class := range reg.ReservedClasses() {
		if class.Name == "intake" {
			t.Errorf("ReservedClasses() still carries the intake class %v, want it retired", class.Keys)
		}
		for _, name := range []string{"DUBPLATE_STAGING_DIR", "DUBPLATE_UPLOADS_DIR", "DUBPLATE_TRIGGER_DIR"} {
			if slices.Contains(class.Keys, name) {
				t.Errorf("reserved class %q still claims %s", class.Name, name)
			}
		}
	}
	for _, name := range []string{"DUBPLATE_UPLOADS_DIR", "DUBPLATE_TRIGGER_DIR"} {
		if reason := reg.Superseded()[name]; !strings.Contains(reason, "DUBPLATE_STAGING_DIR") {
			t.Errorf("Superseded()[%s] = %q, want a reason naming DUBPLATE_STAGING_DIR", name, reason)
		}
	}
}

func TestLoadConfig_StagingFixtures(t *testing.T) {
	base := authEnviron(t)
	with := func(library, staging string) []string {
		return replaceKey(replaceKey(base, "DUBPLATE_LIBRARY_ROOT", library), "DUBPLATE_STAGING_DIR", staging)
	}
	tests := []struct {
		name    string
		environ []string
		want    []string // each must appear in the refusal; none for an accepted case
	}{
		{
			name:    "staging unset",
			environ: dropKey(base, "DUBPLATE_STAGING_DIR"),
			want:    []string{"DUBPLATE_STAGING_DIR", "required setting unset"},
		},
		{
			name:    "staging inside the library root",
			environ: with("/srv/music/library", "/srv/music/library/staging"),
			want:    []string{"DUBPLATE_STAGING_DIR", "/srv/music/library/staging", "DUBPLATE_LIBRARY_ROOT", "/srv/music/library"},
		},
		{
			name:    "the library root inside staging",
			environ: with("/srv/music/staging/library", "/srv/music/staging"),
			want:    []string{"DUBPLATE_STAGING_DIR", "/srv/music/staging", "DUBPLATE_LIBRARY_ROOT", "/srv/music/staging/library"},
		},
		{
			name:    "the two equal",
			environ: with("/srv/music/shared", "/srv/music/shared"),
			want:    []string{"DUBPLATE_STAGING_DIR", "DUBPLATE_LIBRARY_ROOT", "/srv/music/shared"},
		},
		{
			name:    "equal after Clean",
			environ: with("/srv/music/shared/", "/srv/music//shared/../shared"),
			want:    []string{"DUBPLATE_STAGING_DIR", "DUBPLATE_LIBRARY_ROOT", "/srv/music/shared"},
		},
		{
			name:    "inside after Clean resolves a dot-dot",
			environ: with("/srv/music/library", "/srv/music/other/../library/staging"),
			want:    []string{"DUBPLATE_STAGING_DIR", "/srv/music/library/staging"},
		},
		{
			name:    "siblings sharing a name prefix, library shorter",
			environ: with("/srv/music/lib", "/srv/music/library"),
		},
		{
			name:    "siblings sharing a name prefix, staging shorter",
			environ: with("/srv/music/library", "/srv/music/lib"),
		},
		{
			name:    "staging absent on disk loads",
			environ: with("/nonexistent-dubplate-test/library", "/nonexistent-dubplate-test/staging"),
		},
		{
			name:    "a relative staging directory",
			environ: with("/srv/music/library", "staging"),
			want:    []string{"DUBPLATE_STAGING_DIR", "absolute", "staging"},
		},
		{
			name:    "a relative library root",
			environ: with("library", "/srv/music/staging"),
			want:    []string{"DUBPLATE_STAGING_DIR", "absolute", "library"},
		},
		{
			name:    "the retired uploads directory",
			environ: append(slices.Clone(base), "DUBPLATE_UPLOADS_DIR=/srv/music/uploads"),
			want:    []string{"DUBPLATE_UPLOADS_DIR", "DUBPLATE_STAGING_DIR"},
		},
		{
			name:    "the retired trigger directory",
			environ: append(slices.Clone(base), "DUBPLATE_TRIGGER_DIR=/run/dubplate"),
			want:    []string{"DUBPLATE_TRIGGER_DIR", "DUBPLATE_STAGING_DIR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadConfig(tt.environ, productionRegistry())
			if len(tt.want) == 0 {
				if err != nil {
					t.Fatalf("loadConfig() error = %v, want nil", err)
				}
				if cfg == nil {
					t.Fatal("loadConfig() config = nil, want the loaded config")
				}
				return
			}
			if err == nil {
				t.Fatalf("loadConfig() error = nil, want a refusal naming %v", tt.want)
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("loadConfig() error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestResolveStaging_ReadsTheWallValue(t *testing.T) {
	base := authEnviron(t)
	environ := replaceKey(base, "DUBPLATE_STAGING_DIR", "/srv/music/library/staging")
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("fixture state: config.Load: %v", err)
	}
	err = resolveStaging(cfg)
	if err == nil {
		t.Fatal("resolveStaging() error = nil, want the nesting refused: Load alone accepts it")
	}
	if !strings.HasPrefix(err.Error(), "DUBPLATE_STAGING_DIR: ") {
		t.Errorf("resolveStaging() error = %q, want it keyed by DUBPLATE_STAGING_DIR for doctor's row", err)
	}
}

// TestNestedStagingRefusedByEveryVerb proves each verb that loads the
// wall loads it through loadConfig: the library tree is what backup
// syncs and restore refuses to fill, so the nesting rule must hold at
// every one of them, not only serve.
func TestNestedStagingRefusedByEveryVerb(t *testing.T) {
	environ := serveAuthEnviron(t)
	library := environValue(environ, "DUBPLATE_LIBRARY_ROOT")
	nested := replaceKey(environ, "DUBPLATE_STAGING_DIR", library+"/staging")

	reg := productionRegistry()
	verbs := map[string]func() error{
		"serve": func() error {
			cmd := newServeCmd(nested, reg, io.Discard, io.Discard)
			cmd.SetArgs([]string{})
			return cmd.Execute()
		},
		"check": func() error {
			cmd := newCheckCmd(nested, reg, io.Discard, io.Discard)
			cmd.SetArgs([]string{})
			return cmd.Execute()
		},
		"backup": func() error {
			return runBackup(context.Background(), nested, reg, io.Discard, io.Discard)
		},
		"restore": func() error {
			return runRestore(context.Background(), nested, reg, restoreFlags{fromFile: "x.db"}, io.Discard, io.Discard)
		},
		"member": func() error {
			cmd := newMemberCmd(nested, reg, io.Discard, io.Discard)
			cmd.SetArgs([]string{"list"})
			return cmd.Execute()
		},
	}
	for name, run := range verbs {
		t.Run(name, func(t *testing.T) {
			err := run()
			if err == nil {
				t.Fatal("error = nil, want the nesting refused")
			}
			for _, want := range []string{"DUBPLATE_STAGING_DIR", library + "/staging", "DUBPLATE_LIBRARY_ROOT"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

// TestDoctor_StagingNestingReadsInvalid proves doctor loads the wall and
// resolves staging apart, so the nesting lands on the staging row as
// "invalid" with its reason rather than abandoning the report.
func TestDoctor_StagingNestingReadsInvalid(t *testing.T) {
	environ := replaceKey(doctorEnviron(t), "DUBPLATE_STAGING_DIR", "/srv/music/library/staging")

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want a nonzero status: staging nests in the library root")
	}

	var line string
	for l := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.HasPrefix(l, "DUBPLATE_STAGING_DIR ") {
			line = l
		}
	}
	prefix := fmt.Sprintf("%-32s invalid: ", "DUBPLATE_STAGING_DIR")
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("staging row = %q, want it to start with %q\n%s", line, prefix, stdout.String())
	}
	if !strings.Contains(line, "/srv/music/library") {
		t.Errorf("staging row = %q, want the reason to name the library root", line)
	}
	if !strings.Contains(stdout.String(), fmt.Sprintf("%-32s ok", "DUBPLATE_LIBRARY_ROOT")) {
		t.Errorf("doctor output lost the library row, want the report to continue past the nesting\n%s", stdout.String())
	}
}

// TestIntakeSettings pins each setting's kind, default, bounds, apply
// mode, and mark as literals, so a changed default or bound fails here
// rather than passing because the test restates the code's own table.
func TestIntakeSettings(t *testing.T) {
	const (
		next    = config.ApplyNextOperation
		restart = config.ApplyRestart
		fixed   = config.NotSensitive
	)
	tests := []struct {
		key     string
		kind    config.ThresholdKind
		def     string
		floor   string
		ceiling string
		apply   config.ApplyMode
		mark    config.Sensitivity
	}{
		{"DUBPLATE_INTAKE_MAX_ALBUM_BYTES", config.KindInt, "4294967296", "104857600", "17179869184", next, fixed},
		{"DUBPLATE_INTAKE_MAX_FILES", config.KindInt, "500", "1", "2000", next, fixed},
		{"DUBPLATE_INTAKE_MAX_PATH_DEPTH", config.KindInt, "8", "1", "32", next, fixed},
		{"DUBPLATE_INTAKE_MAX_TAG_BYTES", config.KindInt, "1048576", "4096", "16777215", next, fixed},
		{"DUBPLATE_INTAKE_MAX_PICTURE_BYTES", config.KindInt, "8388608", "65536", "16777215", next, fixed},
		{"DUBPLATE_INTAKE_MIN_SAMPLE_RATE", config.KindInt, "44100", "44100", "192000", next, fixed},
		{"DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE", config.KindInt, "16", "16", "24", next, fixed},
		{"DUBPLATE_INTAKE_SESSIONS_PER_MEMBER", config.KindInt, "3", "1", "20", next, fixed},
		{"DUBPLATE_INTAKE_FREE_SPACE_FLOOR", config.KindInt, "5368709120", "1073741824", "1099511627776", next, config.Sensitive},
		{"DUBPLATE_INTAKE_GROUP_QUOTA_BYTES", config.KindInt, "0", "0", "1152921504606846976", next, fixed},
		{"DUBPLATE_INTAKE_RESUME_WINDOW", config.KindDuration, "168h", "1h", "720h", next, fixed},
		{"DUBPLATE_INTAKE_HOLD", config.KindDuration, "168h", "24h", "720h", next, fixed},
		{"DUBPLATE_INTAKE_STALL_AFTER", config.KindDuration, "30m", "5m", "24h", next, fixed},
		{"DUBPLATE_INTAKE_SWEEP_INTERVAL", config.KindDuration, "10m", "1m", "1h", restart, fixed},
		{"DUBPLATE_INTAKE_CHECK_WORKERS", config.KindInt, "1", "1", "8", restart, fixed},
	}

	registered := make(map[string]config.Setting)
	for _, s := range newProductionRegistry().Settings() {
		if strings.HasPrefix(s.Key, "DUBPLATE_INTAKE_") {
			registered[s.Key] = s
		}
	}
	if len(registered) != len(tests) {
		t.Errorf("production registry holds %d DUBPLATE_INTAKE_* settings, want %d", len(registered), len(tests))
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			s, ok := registered[tt.key]
			if !ok {
				t.Fatalf("%s is not a registered setting", tt.key)
			}
			if s.Kind != tt.kind || s.Default != tt.def || s.Floor != tt.floor || s.Ceiling != tt.ceiling || s.Apply != tt.apply || s.Mark != tt.mark {
				t.Errorf("setting = {kind %v, default %q, floor %q, ceiling %q, apply %v, mark %v}, want {%v, %q, %q, %q, %v, %v}",
					s.Kind, s.Default, s.Floor, s.Ceiling, s.Apply, s.Mark,
					tt.kind, tt.def, tt.floor, tt.ceiling, tt.apply, tt.mark)
			}
			for _, v := range []string{tt.floor, tt.ceiling, tt.def} {
				if err := s.Validate(v); err != nil {
					t.Errorf("Validate(%q) = %v, want nil", v, err)
				}
			}
		})
	}
}

// TestIntakeSettings_OutOfBounds proves the bounds refuse: one value
// below each floor and above each ceiling that has headroom.
func TestIntakeSettings_OutOfBounds(t *testing.T) {
	byKey := make(map[string]config.Setting)
	for _, s := range newProductionRegistry().Settings() {
		byKey[s.Key] = s
	}
	tests := []struct {
		key, value, want string
	}{
		{"DUBPLATE_INTAKE_MAX_ALBUM_BYTES", "104857599", "below the floor"},
		{"DUBPLATE_INTAKE_MAX_ALBUM_BYTES", "17179869185", "above the ceiling"},
		{"DUBPLATE_INTAKE_MAX_FILES", "0", "below the floor"},
		{"DUBPLATE_INTAKE_MAX_FILES", "2001", "above the ceiling"},
		{"DUBPLATE_INTAKE_MIN_SAMPLE_RATE", "44099", "below the floor"},
		{"DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE", "15", "below the floor"},
		{"DUBPLATE_INTAKE_FREE_SPACE_FLOOR", "1073741823", "below the floor"},
		{"DUBPLATE_INTAKE_GROUP_QUOTA_BYTES", "-1", "below the floor"},
		{"DUBPLATE_INTAKE_HOLD", "23h59m", "below the floor"},
		{"DUBPLATE_INTAKE_RESUME_WINDOW", "721h", "above the ceiling"},
		{"DUBPLATE_INTAKE_CHECK_WORKERS", "9", "above the ceiling"},
	}
	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			err := byKey[tt.key].Validate(tt.value)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate(%q) = %v, want an error containing %q", tt.value, err, tt.want)
			}
		})
	}
}

func TestIntakeGroupQuotaDocSaysZeroMeansNoCap(t *testing.T) {
	for _, s := range newProductionRegistry().Settings() {
		if s.Key == "DUBPLATE_INTAKE_GROUP_QUOTA_BYTES" {
			if !strings.Contains(s.Doc, "0") || !strings.Contains(s.Doc, "no group cap") {
				t.Errorf("Doc = %q, want it to say that 0 means no group cap", s.Doc)
			}
			return
		}
	}
	t.Fatal("DUBPLATE_INTAKE_GROUP_QUOTA_BYTES is not a registered setting")
}

func TestIntakeFormatAllowlistThreshold(t *testing.T) {
	for _, th := range newProductionRegistry().Thresholds() {
		if th.Key != "DUBPLATE_INTAKE_FORMAT_ALLOWLIST" {
			continue
		}
		if th.Kind != config.KindString || th.Default != "flac" || th.Mark != config.NotSensitive {
			t.Errorf("threshold = {kind %v, default %q, mark %v}, want {KindString, \"flac\", NotSensitive}", th.Kind, th.Default, th.Mark)
		}
		return
	}
	t.Fatal("DUBPLATE_INTAKE_FORMAT_ALLOWLIST is not a registered threshold")
}

// TestIntakeSettings_SetInTheEnvironmentLoads proves a set intake key is
// a declared name: Load neither flags it unknown nor parses it, since
// the seed evaluation owns a store-homed key's environment value.
func TestIntakeSettings_SetInTheEnvironmentLoads(t *testing.T) {
	environ := append(authEnviron(t),
		"DUBPLATE_INTAKE_MAX_FILES=750",
		"DUBPLATE_INTAKE_FORMAT_ALLOWLIST=flac")
	if _, err := loadConfig(environ, productionRegistry()); err != nil {
		t.Errorf("loadConfig() error = %v, want nil", err)
	}
	_, err := loadConfig(append(authEnviron(t), "DUBPLATE_INTAKE_NOT_A_KEY=1"), productionRegistry())
	if err == nil || !strings.Contains(err.Error(), "DUBPLATE_INTAKE_NOT_A_KEY: unknown configuration key") {
		t.Errorf("an undeclared DUBPLATE_INTAKE_* name: loadConfig() error = %v, want unknown configuration key", err)
	}
}

// TestIntakeAllowlist: the allowlist splits into format names, and a
// name the gate's format table does not hold, an empty entry among
// them, is refused naming it.
func TestIntakeAllowlist(t *testing.T) {
	const key = "DUBPLATE_INTAKE_FORMAT_ALLOWLIST"
	tests := []struct {
		value string
		want  []string
		// refusal names the entry a refusal must quote.
		refusal string
	}{
		{value: "flac", want: []string{"flac"}},
		{value: " flac ", want: []string{"flac"}},
		{value: "flac,mp3", refusal: `"mp3"`},
		{value: "flac,", refusal: `""`},
		{value: ",", refusal: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := intakeAllowlist([]string{key + "=" + tt.value})
			if tt.refusal == "" {
				if err != nil || !slices.Equal(got, tt.want) {
					t.Errorf("intakeAllowlist = %q, %v; want %q", got, err, tt.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), key) || !strings.Contains(err.Error(), tt.refusal) {
				t.Errorf("intakeAllowlist error = %v, want a refusal naming %s and %s", err, key, tt.refusal)
			}
		})
	}
}

// TestServeRefusesAnAllowlistItCannotCheck: serve refuses to start over
// an allowlist naming a format it cannot check, naming the entry.
func TestServeRefusesAnAllowlistItCannotCheck(t *testing.T) {
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_INTAKE_FORMAT_ALLOWLIST", "flac,mp3")
	requireServeRefuses(t, environ, "DUBPLATE_INTAKE_FORMAT_ALLOWLIST", "mp3")
}

// TestServeChecksStagingBeforeTheStore is AC 2's serve half: serve over
// staging and a library root that are sibling temporary directories
// starts, and with staging absent it refuses naming the path before the
// store is opened, so the state directory holds no store.
func TestServeChecksStagingBeforeTheStore(t *testing.T) {
	requireServeStarts(t, serveAuthEnviron(t))

	environ := serveAuthEnviron(t)
	staging := environValue(environ, "DUBPLATE_STAGING_DIR")
	if err := os.Remove(staging); err != nil {
		t.Fatal(err)
	}
	requireServeRefuses(t, environ, staging, "does not exist")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	r := store.Inspect(t.Context(), store.InspectOptions{StateDir: stateDir, BusyTimeout: time.Second})
	if r.State != store.StateAbsent {
		t.Errorf("store state = %v after a staging refusal, want absent: the check must run before the store opens", r.State)
	}
}

// TestServeRefusedAtTheIntakeLease is AC 27's serve half: a second serve
// on one state directory refuses at the intake lease, naming it, while
// the first holds it, and starts once the first has closed.
func TestServeRefusedAtTheIntakeLease(t *testing.T) {
	environ := serveAuthEnviron(t)
	first, _, _, err := startServeShell(t, t.Context(), environ)
	if err != nil {
		t.Fatalf("first serve: %v", err)
	}

	second, _, _, err := startServeShell(t, t.Context(), environ)
	if !errors.Is(err, store.ErrLeaseBusy) || second != nil {
		t.Fatalf("second serve = %v, %v; want a refusal at the intake lease", second, err)
	}
	if !strings.Contains(err.Error(), "intake") {
		t.Errorf("refusal %q does not name the intake lease", err)
	}

	if err := first.Close(); err != nil {
		t.Fatalf("close the first serve: %v", err)
	}
	third, _, _, err := startServeShell(t, t.Context(), environ)
	if err != nil {
		t.Fatalf("serve after the first closed: %v", err)
	}
	if err := third.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestTusMaxSizeIsTheAlbumCapCeiling pins tusd's MaxSize to the album
// cap's ceiling: a creation one byte over the ceiling answers tusd's
// 413, and one at the ceiling passes MaxSize to the create callback,
// which refuses it for naming no file.
func TestTusMaxSizeIsTheAlbumCapCeiling(t *testing.T) {
	ceiling, err := strconv.ParseInt(intakeMaxAlbumBytesSetting.Ceiling, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	h, err := store.OpenServe(t.Context(), t.TempDir(), store.ServeOptions{
		BusyTimeout: 5 * time.Second, MigrationWait: 30 * time.Second, Keep: 1, Logger: slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Close() })
	in, err := intake.New(intake.Config{Store: h, Log: slog.New(slog.DiscardHandler), Staging: t.TempDir(), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = in.Close() })
	mount, err := in.TusMount(intake.TusConfig{
		Prefix: "/dubplate", PublicURL: "https://music.example",
		Verify: func(http.ResponseWriter, *http.Request) (identity.UserID, error) { return identity.UserID{1}, nil },
		Floor:  func(context.Context) (int64, error) { return 0, nil },
		TusLog: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	for length, want := range map[int64]int{ceiling + 1: http.StatusRequestEntityTooLarge, ceiling: http.StatusBadRequest} {
		r := httptest.NewRequest(http.MethodPost, "/dubplate/tus/", http.NoBody)
		r.Header.Set("Tus-Resumable", "1.0.0")
		r.Header.Set("Upload-Length", strconv.FormatInt(length, 10))
		rec := httptest.NewRecorder()
		mount.ServeHTTP(rec, r)
		if rec.Code != want {
			t.Errorf("Upload-Length %d = %d %q, want %d", length, rec.Code, rec.Body, want)
		}
	}
}

// TestStallMailGoesToLiveAdminsOnly is AC 28's recipients: the
// production address closure over a real roster holding two live
// admins, a demoted admin, and a removed admin reads the two live
// admins alone.
func TestStallMailGoesToLiveAdminsOnly(t *testing.T) {
	stateDir := t.TempDir()
	seedBackupStore(t, stateDir)
	seedMembers(t, stateDir,
		seedMember{address: "ann@example.com", displayName: "Ann", username: "ann", admin: true},
		seedMember{address: "demoted@example.com", displayName: "Dee", username: "dee"},
		seedMember{address: "removed@example.com", displayName: "Rae", username: "rae", admin: true, tombstoned: true},
		seedMember{address: "bo@example.com", displayName: "Bo", username: "bo", admin: true},
	)
	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Close() })

	got, err := liveAdminAddresses(auth.NewRoster(h, nil, nil))(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ann@example.com", "bo@example.com"}; !slices.Equal(got, want) {
		t.Errorf("stall mail recipients = %q, want %q", got, want)
	}
}
