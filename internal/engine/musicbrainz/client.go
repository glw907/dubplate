package musicbrainz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// DefaultBaseURL is the live MusicBrainz host beets hard-codes for the
// musicbrainz.org branch (beetsplug/_utils/musicbrainz.py:598-599).
const DefaultBaseURL = "https://musicbrainz.org"

// wsPathPrefix is the WS/2 root beets always appends to api_host
// (beetsplug/_utils/musicbrainz.py:609-610). This is deliberately a
// second, independent copy of the "/ws/2" literal
// internal/harness/mbrequest also carries: the architectural rule lets
// internal/harness import internal/engine but never the reverse, and
// sharing one definition between the system under test and the
// committed-request producer would make Task 7's grading a
// self-derived expectation (query.go's own releaseIncludes doc comment
// states the same reasoning for that constant).
const wsPathPrefix = "/ws/2"

// defaultTimeout is 30s rather than beets' own 10s (R13 divergence 1),
// matching the capture tool's own reason: maximal-inc release lookups
// are large, and the largest committed response body is 8,894 KB
// (cmd/harness/capture.go:38-41).
const defaultTimeout = 30 * time.Second

// maxAttempts is one plus six retries, beets' own Retry(total=6)
// (beetsplug/_utils/requests.py:78-88).
const maxAttempts = 7

// retryableStatuses is urllib3's status_forcelist (C12): 429, 500,
// 502, 503, 504. 413 plus a Retry-After header is a sixth retryable
// case in urllib3 (RETRY_AFTER_STATUS_CODES,
// urllib3/util/retry.py:202), and it is a deliberate non-port: WS/2
// never answers 413 to a GET, so the omission is recorded rather than
// reachable.
var retryableStatuses = map[int]bool{
	http.StatusTooManyRequests:     true,
	http.StatusInternalServerError: true,
	http.StatusBadGateway:          true,
	http.StatusServiceUnavailable:  true,
	http.StatusGatewayTimeout:      true,
}

// releaseIDPattern is _extract_id's own pattern
// (beets/util/id_extractors.py:20), applied the same way beets applies
// it first, before any request (beets/metadata_plugins.py:272-278).
// Its \w is Unicode-aware in Python and Go's RE2 \w is ASCII-only,
// which is the fourth member of the case-mapping-and-regex family 3b's
// guard 11 opened (R14): no committed mb_albumid is non-ASCII, so
// nothing moves today.
var releaseIDPattern = regexp.MustCompile(`\w{8}(?:-\w{4}){3}-\w{12}`)

// ErrNotFound reports that a lookup's release does not exist. beets'
// own flow branches on it: album_for_id catches HTTPNotFoundError and
// drops the candidate silently rather than failing the import
// (beetsplug/musicbrainz.py:777-781), so callers must be able to tell a
// deleted release from a real failure.
var ErrNotFound = errors.New("release not found")

// ErrUnavailable reports that every retry was spent on a retryable
// status or a transport error. It is the shape of beets' own
// RetryError, which is deliberately NOT an HTTPError and so is never
// mapped to a beets exception (beetsplug/_utils/requests.py:78-89).
// Rung 4's shadow branches on it to defer a lookup rather than record a
// divergence.
//
// No rate-limit-specific sentinel ships beside it: %w is an API
// commitment justified by a caller that branches, and no caller
// branches on 429 separately from 503 today.
var ErrUnavailable = errors.New("musicbrainz unavailable")

// ClientConfig configures a [Client]. UserAgent is required and the
// engine never builds one: MusicBrainz requires an identified client
// and the identity is site-specific, so it belongs to internal/config
// and reaches here through cmd/dubplate-server's composition
// (dubplate-server/<version> ( <contact> ), on cmd/harness/capture.go's
// own precedent, with the contact from Config.AdminEmail). Budget is
// required too; offline callers pass [Unmetered].
type ClientConfig struct {
	BaseURL   string        // empty means DefaultBaseURL
	UserAgent string        // required
	Budget    *Budget       // required
	Timeout   time.Duration // zero means 30s
	HTTP      *http.Client  // zero value means one built from Timeout
}

// Client is a WS/2 client: fmt=json forced on every request, beets'
// own retry policy with R13's three named divergences, an injected
// base URL, a required User-Agent, and the pseudo-release double
// fetch.
type Client struct {
	baseURL   string
	userAgent string
	budget    *Budget
	http      *http.Client

	// sleep is the backoff wait, injected so a test observes the
	// schedule through the values it records rather than by spending
	// real wall time. Production uses a real timer; it is not a bare
	// time.Sleep because a caller's context must be able to cancel the
	// wait, and it is not a clock read: client_test.go calls time.Now
	// nowhere, matching TestEngineNeverReadsTheClock's walk over this
	// file (R2, fold m10).
	sleep func(context.Context, time.Duration) error
}

// NewClient validates cfg and returns a client. It fails rather than
// defaulting on a missing UserAgent or Budget, because both are
// correctness requirements rather than conveniences.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.UserAgent == "" {
		return nil, errors.New("musicbrainz: ClientConfig.UserAgent is required")
	}
	if cfg.Budget == nil {
		return nil, errors.New("musicbrainz: ClientConfig.Budget is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := cfg.HTTP
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		baseURL:   baseURL,
		userAgent: cfg.UserAgent,
		budget:    cfg.Budget,
		http:      httpClient,
		sleep:     sleepReal,
	}, nil
}

// sleepReal is the production sleep: a real timer a caller's context
// can still cancel.
func sleepReal(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// extractReleaseID mirrors _extract_id (beets/util/id_extractors.py:20):
// it finds the UUID substring anywhere in s, the same re.search beets
// itself runs, rather than requiring the whole string to be one.
func extractReleaseID(s string) (string, bool) {
	id := releaseIDPattern.FindString(s)
	if id == "" {
		return "", false
	}
	return id, true
}

// releaseRequestURL builds a lookup's URL: fmt=json forced, the
// sixteen-element include set joined with a literal "+" (C3), encoded
// through url.Values so the wire form matches the committed corpus
// byte for byte with no adapter (C18).
func (c *Client) releaseRequestURL(id string) string {
	v := url.Values{
		"fmt": {"json"},
		"inc": {strings.Join(releaseIncludes, "+")},
	}
	return c.baseURL + wsPathPrefix + "/release/" + id + "?" + v.Encode()
}

// searchRequestURL builds a search's URL: fmt=json forced and no inc
// at all (C4).
func (c *Client) searchRequestURL(s AlbumSearch) string {
	v := url.Values{
		"fmt":   {"json"},
		"query": {s.Query},
		"limit": {strconv.Itoa(s.Limit)},
	}
	return c.baseURL + wsPathPrefix + "/release?" + v.Encode()
}

// Release fetches one release at beets' own maximal inc= set, mirroring
// get_release (beetsplug/_utils/musicbrainz.py:687-690). It returns
// ErrNotFound on a 404, ErrUnavailable when the retries are spent, and
// an error wrapping ErrMalformedRelease when the body does not decode.
// mbid is run through the same id extraction beets applies first
// (beets/util/id_extractors.py:20), so a malformed id costs no request
// at all, exactly as album_for_id costs none (beetsplug/musicbrainz.py:
// 771-773).
//
// It decodes and does not validate. TranslateAlbum owns validation and
// calls Release.Validate as its first act (translate.go:46-48), so a
// second call here would validate every body twice and split ownership
// of one rule across two files.
func (c *Client) Release(ctx context.Context, mbid string) (*Release, error) {
	id, ok := extractReleaseID(mbid)
	if !ok {
		return nil, fmt.Errorf("musicbrainz: %q does not contain a release id beets' own extraction would recognize", mbid)
	}

	body, err := c.doWithRetry(ctx, c.releaseRequestURL(id))
	if err != nil {
		return nil, err
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("musicbrainz: decoding release %s: %w: %w", id, err, ErrMalformedRelease)
	}
	return &rel, nil
}

// SearchReleases issues one release search and returns the release ids
// in response order, mirroring search (beetsplug/_utils/musicbrainz.py:
// 665-685) and candidates' own use of the result
// (beets/metadata_plugins.py:442-446). It returns ids rather than
// releases: the per-id fan-out and the candidate ordering are 3d's.
//
// A search body is an OBJECT, and the hits live under "releases":
// search ends by indexing the decoded body at f"{entity}s" (:684-685),
// and all 48 committed search responses carry exactly count, created,
// offset and releases (contract C19). Decoding into a bare array fails
// every committed search.
//
// Its errors are its own, and they do not mirror beets. beets swallows
// a search failure into an empty result inside _search_api's bare
// except (beets/metadata_plugins.py:419-427, which re-raises first when
// raise_on_error is set at :422-423); this client returns the failure
// instead, because swallowing it here would make Task 7's zero-miss
// rule unable to see a search that never happened. A 404 returns
// ErrNotFound, exhausted retries return ErrUnavailable, and a body that
// does not decode returns an error wrapping ErrMalformedRelease. The
// caller that wants beets' shape maps them to an empty candidate list,
// which is 3d's flow.
func (c *Client) SearchReleases(ctx context.Context, s AlbumSearch) ([]string, error) {
	body, err := c.doWithRetry(ctx, c.searchRequestURL(s))
	if err != nil {
		return nil, err
	}

	var decoded struct {
		Releases []struct {
			ID string `json:"id"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("musicbrainz: decoding search response: %w: %w", err, ErrMalformedRelease)
	}

	ids := make([]string, len(decoded.Releases))
	for i, r := range decoded.Releases {
		ids[i] = r.ID
	}
	return ids, nil
}

// AlbumForID is album_for_id's whole body (beetsplug/musicbrainz.py:
// 765-800): fetch, translate, and, when the release is a pseudo-release
// with a backward transl-tracklisting relation, fetch the linked release
// and merge. Validation is TranslateAlbum's, called once there, never
// again here.
//
// A 404 on the first lookup returns ErrNotFound unwrapped, and the
// caller (3d) maps it to "no candidate", reproducing album_for_id's
// return None. A 404 on the SECOND, pseudo-release lookup returns a
// wrapped error rather than a dropped candidate, because beets' second
// get_release sits outside the try and its HTTPNotFoundError escapes to
// maybe_handle_plugin_error, which logs and drops the plugin's whole
// candidate list (C13, C15). The shapes differ; the outcome for the
// caller is the same candidate loss.
//
// Both AlbumInfo values are built inside this call and only the merged
// one escapes, closing 3b's MergePseudoRelease aliasing carry by
// construction for this path: the pseudo value has no second retaining
// reader (docs/HISTORY.md:265-272).
func (c *Client) AlbumForID(ctx context.Context, mbid string, tun matcher.Tunables) (*matcher.AlbumInfo, error) {
	rel, err := c.Release(ctx, mbid)
	if err != nil {
		return nil, err
	}
	info, err := TranslateAlbum(rel, tun)
	if err != nil {
		return nil, err
	}

	targetID, ok := PseudoReleaseTarget(rel)
	if !ok {
		return info, nil
	}

	actualRel, err := c.Release(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: fetching pseudo-release %s's actual release %s: %w", rel.ID, targetID, err)
	}
	actualInfo, err := TranslateAlbum(actualRel, tun)
	if err != nil {
		return nil, err
	}
	return MergePseudoRelease(info, actualInfo), nil
}

// doWithRetry issues one GET at requestURL, acquiring a limiter token
// before every attempt including retries (R13 divergence 2: beets gets
// per-request-only metering for free because Retry sits inside the
// adapter beneath both its limiters, and reproducing that accident here
// would make rung 4's shadow budget a lie). It returns the response
// body on a 200, ErrNotFound on a 404, ErrUnavailable when the retries
// are spent, and a generic error for any other status.
func (c *Client) doWithRetry(ctx context.Context, requestURL string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.budget.Wait(ctx); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("musicbrainz: building request: %w", err)
		}
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if attempt == maxAttempts {
				break
			}
			if serr := c.sleep(ctx, backoff(attempt)); serr != nil {
				return nil, serr
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("musicbrainz: reading response body: %w", readErr)
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			return body, nil
		case resp.StatusCode == http.StatusNotFound:
			return nil, ErrNotFound
		case retryableStatuses[resp.StatusCode]:
			lastErr = fmt.Errorf("musicbrainz: status %d", resp.StatusCode)
		default:
			return nil, fmt.Errorf("musicbrainz: unexpected status %d", resp.StatusCode)
		}

		if attempt == maxAttempts {
			break
		}
		if serr := c.sleep(ctx, retryWait(resp.Header.Get("Retry-After"), attempt)); serr != nil {
			return nil, serr
		}
	}
	return nil, fmt.Errorf("%w: %w", ErrUnavailable, lastErr)
}

// retryWait picks the wait before the next attempt: a delta-seconds
// Retry-After header wins over the computed backoff (C12), and an
// HTTP-date Retry-After falls through to the computed backoff
// untouched (R13 divergence 3): reducing an HTTP date to a duration
// needs the current time, which TestEngineNeverReadsTheClock forbids
// inside internal/engine, and the computed backoff is already a
// bounded monotone schedule, so the date form degrades to a correct
// wait rather than to no wait at all.
func retryWait(retryAfter string, attempt int) time.Duration {
	if retryAfter == "" {
		return backoff(attempt)
	}
	seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter))
	if err != nil || seconds < 0 {
		return backoff(attempt)
	}
	return time.Duration(seconds) * time.Second
}

// backoff is urllib3's own schedule (C12, urllib3/util/retry.py:309-326):
// zero when the consecutive-error run is one or less, then
// 500ms * 2^(n-1), capped at 120s with no jitter. attempt is the
// 1-indexed attempt that just failed, which is that run count.
func backoff(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}
	d := 500 * time.Millisecond * time.Duration(1<<(attempt-1))
	if d > 120*time.Second {
		return 120 * time.Second
	}
	return d
}
