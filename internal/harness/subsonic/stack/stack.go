// Package stack runs the pinned Navidrome and gonic oracle images under
// podman against the fixture library, so a session records what each
// oracle answers for the same files.
//
// Configuration comes from Navidrome's documentation
// (https://www.navidrome.org/docs/usage/configuration/options/) and
// gonic's README. Telemetry and every metadata agent are off, so an
// oracle's answers never depend on the network. Each container publishes
// to a loopback port podman assigns, and every container carries a
// run-unique name and label, so an orphan left by a killed process can be
// found with podman ps --filter label=dubplate-harness-run.
package stack

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

// ErrNoPodman marks a workstation without a podman executable.
var ErrNoPodman = errors.New("stack: podman not found in PATH")

// errNotReady marks an oracle whose scan never finished by the deadline.
var errNotReady = errors.New("stack: oracle not ready")

// errTrackCount marks an oracle whose finished scan counted a different
// number of tracks than the manifest declares.
var errTrackCount = errors.New("stack: scan track count differs from manifest")

// errImageNotPinned marks a running image whose inspect output does not
// carry the pinned digest.
var errImageNotPinned = errors.New("stack: running image does not carry the pinned digest")

const (
	labelKey     = "dubplate-harness-run"
	teardownWait = 30 * time.Second
	logTail      = "20"
)

// The podman executable, the bound on each oracle's wait for a finished
// scan, and the interval between readiness probes. Tests swap them.
var (
	podmanName   = "podman"
	readyTimeout = time.Minute
	pollInterval = 500 * time.Millisecond
)

// Config describes one stack run.
type Config struct {
	// MusicDir is the built fixture library, mounted read-only.
	MusicDir string
	// Tracks is the manifest's track count, which a finished scan must report.
	Tracks int
	// DigestFile is the committed file naming the pinned images.
	DigestFile string
}

// Oracle is one running server.
type Oracle struct {
	Name     string
	BaseURL  string
	User     string
	Password string

	image     string // the pinned reference the container was started from
	container string
}

// Client returns a client for o whose requests are each bounded by timeout.
func (o Oracle) Client(timeout time.Duration) *client.Client {
	return &client.Client{
		BaseURL: o.BaseURL, User: o.User, Password: o.Password,
		Name: "dubplate-harness", HTTP: &http.Client{Timeout: timeout},
	}
}

// Stack is the running pair.
type Stack struct {
	// Oracles holds Navidrome, then gonic.
	Oracles []Oracle

	podman string
	secret string
}

// spec is how one oracle image is run.
type spec struct {
	name       string
	ref        string
	port       int
	user       string
	password   string
	env        []string
	secretName string // set from the process environment, never argv
	tmpfs      []string
}

// Start runs both oracles and returns once each has finished scanning the
// library with the manifest's track count. On any failure it removes
// whatever it started before returning.
func Start(ctx context.Context, cfg Config) (*Stack, error) {
	imgs, err := loadImages(cfg.DigestFile)
	if err != nil {
		return nil, err
	}
	bin, err := exec.LookPath(podmanName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoPodman, err)
	}
	runID, err := randomHex(6)
	if err != nil {
		return nil, err
	}
	secret, err := randomHex(12)
	if err != nil {
		return nil, err
	}
	s := &Stack{podman: bin, secret: secret}

	specs := []spec{
		{
			name: "navidrome", ref: imgs.Navidrome, port: 4533,
			user: "admin", password: secret,
			env: []string{
				"ND_MUSICFOLDER=/music",
				"ND_ENABLEINSIGHTSCOLLECTOR=false",
				"ND_DEEZER_ENABLED=false",
				"ND_LASTFM_ENABLED=false",
				"ND_LISTENBRAINZ_ENABLED=false",
				"ND_SCANNER_SCANONSTARTUP=true",
			},
			secretName: "ND_DEVAUTOCREATEADMINPASSWORD",
		},
		{
			name: "gonic", ref: imgs.Gonic, port: 4747,
			user: "admin", password: "admin",
			env: []string{
				"GONIC_MUSIC_PATH=/music",
				"GONIC_LISTEN_ADDR=0.0.0.0:4747",
				"GONIC_SCAN_AT_START_ENABLED=true",
				"GONIC_MULTI_VALUE_ARTIST=multi",
				"GONIC_MULTI_VALUE_ALBUM_ARTIST=multi",
			},
			tmpfs: []string{"/data", "/cache", "/podcasts", "/playlists"},
		},
	}
	for i, sp := range specs {
		// The oracle joins s before its container runs, so a partial
		// start still tears down.
		s.Oracles = append(s.Oracles, Oracle{
			Name: sp.name, User: sp.user, Password: sp.password,
			image: sp.ref, container: "dubplate-oracle-" + runID + "-" + sp.name,
		})
		base, err := s.run(ctx, cfg, sp, s.Oracles[i].container, runID)
		if err != nil {
			s.Close(ctx)
			return nil, fmt.Errorf("start %s: %w", sp.name, err)
		}
		s.Oracles[i].BaseURL = base
	}

	var errs []error
	for _, o := range s.Oracles {
		if err := waitReady(ctx, cfg, o); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w%s", o.Name, err, s.logTail(ctx, o.container)))
		}
	}
	if err := errors.Join(errs...); err != nil {
		s.Close(ctx)
		return nil, err
	}
	return s, nil
}

// Close removes both containers and their anonymous volumes. It runs even
// when ctx is already cancelled, under its own timeout, because a command
// built on a cancelled context never starts.
func (s *Stack) Close(ctx context.Context) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownWait)
	defer cancel()
	args := []string{"rm", "-f", "-v", "-i"}
	for _, o := range s.Oracles {
		args = append(args, o.container)
	}
	// Teardown is best effort: the containers are loopback-only and
	// labelled, so a failure leaves an orphan the label finds.
	_, _ = s.exec(ctx, nil, args...)
}

// ImageDigest inspects the image o's container is running and returns the
// pinned digest, taken from the inspect output: the image's Digest field
// or the digest half of one of its RepoDigests. It refuses when neither
// carries the digest o's image reference pins, so a container started from
// a moved or substituted image never records.
func (s *Stack) ImageDigest(ctx context.Context, o Oracle) (string, error) {
	_, pin, _ := strings.Cut(o.image, "@")
	if pin == "" {
		// An empty pin would match an image whose inspect reports no digest.
		return "", fmt.Errorf("%w: %s names no digest", errImageNotPinned, o.image)
	}
	id, err := s.exec(ctx, nil, "container", "inspect", "--format", "{{.Image}}", o.container)
	if err != nil {
		return "", err
	}
	out, err := s.exec(ctx, nil, "image", "inspect", strings.TrimSpace(id))
	if err != nil {
		return "", err
	}
	var infos []struct {
		Digest      string
		RepoDigests []string
	}
	if err := json.Unmarshal([]byte(out), &infos); err != nil {
		return "", fmt.Errorf("%w: %s: parse inspect output: %v", errImageNotPinned, o.image, err)
	}
	var seen []string
	for _, info := range infos {
		seen = append(seen, fmt.Sprintf("Digest %q RepoDigests %q", info.Digest, info.RepoDigests))
		if info.Digest == pin {
			return pin, nil
		}
		for _, rd := range info.RepoDigests {
			if _, d, ok := strings.Cut(rd, "@"); ok && d == pin {
				return pin, nil
			}
		}
	}
	return "", fmt.Errorf("%w: %s: inspect reports %s", errImageNotPinned, o.image, strings.Join(seen, "; "))
}

func (s *Stack) run(ctx context.Context, cfg Config, sp spec, name, runID string) (string, error) {
	args := []string{
		"run", "-d",
		"--name", name,
		"--label", labelKey + "=" + runID,
		"-p", "127.0.0.1::" + strconv.Itoa(sp.port),
		"-v", cfg.MusicDir + ":/music:ro,z",
	}
	for _, t := range sp.tmpfs {
		args = append(args, "--tmpfs", t)
	}
	for _, e := range sp.env {
		args = append(args, "-e", e)
	}
	var extra []string
	if sp.secretName != "" {
		args = append(args, "-e", sp.secretName)
		extra = []string{sp.secretName + "=" + sp.password}
	}
	args = append(args, sp.ref)
	if _, err := s.exec(ctx, extra, args...); err != nil {
		return "", err
	}
	out, err := s.exec(ctx, nil, "port", name, strconv.Itoa(sp.port))
	if err != nil {
		return "", err
	}
	line, _, _ := strings.Cut(strings.TrimSpace(out), "\n")
	port := line[strings.LastIndex(line, ":")+1:]
	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("unparsable port mapping %q", line)
	}
	return "http://127.0.0.1:" + port, nil
}

// exec runs podman and returns its standard output. A failure carries the
// standard error, redacted.
func (s *Stack) exec(ctx context.Context, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, s.podman, args...)
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("podman %s: %w: %s", args[0], err, s.redact(strings.TrimSpace(stderr.String())))
	}
	return stdout.String(), nil
}

func (s *Stack) redact(text string) string {
	return strings.ReplaceAll(text, s.secret, "[redacted]")
}

// logTail returns the container's last log lines, redacted, as a
// parenthesized suffix for an error message.
func (s *Stack) logTail(ctx context.Context, name string) string {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownWait)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.podman, "logs", "--tail", logTail, name)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	_ = cmd.Run()
	text := strings.TrimSpace(s.redact(buf.String()))
	if text == "" {
		return ""
	}
	return "\nlast log lines:\n" + text
}

// waitReady polls until the oracle answers ping and a finished scan
// reports the manifest's track count, or the deadline passes.
func waitReady(ctx context.Context, cfg Config, o Oracle) error {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	c := o.Client(5 * time.Second)
	finished := -1 // count of the last finished scan, -1 when none was seen
	for {
		if _, err := c.Do(ctx, "ping", nil, client.JSON); err == nil {
			scanning, count, err := scanStatus(ctx, c)
			if err == nil && !scanning {
				if count == cfg.Tracks {
					return nil
				}
				finished = count
			}
		}
		select {
		case <-ctx.Done():
			if finished >= 0 {
				return fmt.Errorf("%w: oracle %s counted %d, manifest has %d", errTrackCount, o.Name, finished, cfg.Tracks)
			}
			return fmt.Errorf("%w: %s scan did not finish: %w", errNotReady, o.Name, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

func scanStatus(ctx context.Context, c *client.Client) (scanning bool, count int, err error) {
	resp, err := c.Do(ctx, "getScanStatus", nil, client.JSON)
	if err != nil {
		return false, 0, err
	}
	var body struct {
		Response struct {
			ScanStatus struct {
				Scanning bool `json:"scanning"`
				Count    int  `json:"count"`
			} `json:"scanStatus"`
		} `json:"subsonic-response"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return false, 0, fmt.Errorf("parse scan status: %w", err)
	}
	st := body.Response.ScanStatus
	return st.Scanning, st.Count, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
