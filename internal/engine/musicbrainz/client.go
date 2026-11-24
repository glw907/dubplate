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

// defaultBaseURL is the live MusicBrainz host beets hard-codes for the
// musicbrainz.org branch (beetsplug/_utils/musicbrainz.py:598-599).
const defaultBaseURL = "https://musicbrainz.org"

// wsPathPrefix is the WS/2 root beets always appends to api_host
// (beetsplug/_utils/musicbrainz.py:609-610).
const wsPathPrefix = "/ws/2"

// releaseIncludes is beets' RELEASE_INCLUDES
// (beetsplug/_utils/musicbrainz.py:50-67), in its own order, joined with
// a literal "+" at request time (:634-635).
//
// This is deliberately a SECOND definition of a string
// internal/harness/cassette also carries (maximalReleaseInclude), and it
// must stay one. cassette.ReleaseLookup is what the capture tool and the
// synth tool use to PRODUCE a committed request; the committed request
// log is the expectation and this list is the system under test. One
// shared definition would make the conformance run's request check a
// self-derived expectation, which is the anti-pattern that run exists to
// catch. The two copies are independent on purpose and the conformance
// run is what proves they agree. The same holds for wsPathPrefix.
var releaseIncludes = []string{
	"artists", "media", "recordings", "release-groups", "labels",
	"artist-credits", "aliases", "recording-level-rels", "work-rels",
	"work-level-rels", "artist-rels", "isrcs", "url-rels",
	"release-rels", "genres", "tags",
}

// defaultTimeout is 30s rather than beets' own 10s, for the reason the
// capture tool's liveFetchTimeout gives: maximal-inc release lookups are
// large.
const defaultTimeout = 30 * time.Second

// maxBodyBytes caps a response body well above the largest maximal-inc
// lookup, still bounding what a misbehaving server can make the client
// hold.
const maxBodyBytes = 32 << 20

// errBodyTooLarge reports a response body over [maxBodyBytes]. It is
// never retried.
var errBodyTooLarge = fmt.Errorf("response body exceeds the %d MiB cap", maxBodyBytes>>20)

// maxAttempts is one plus six retries, beets' own Retry(total=6)
// (beetsplug/_utils/requests.py:78-88).
const maxAttempts = 7

// retryableStatuses is urllib3's status_forcelist: 429, 500,
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
// Its \w is Unicode-aware in Python and Go's RE2 \w is ASCII-only: no
// committed mb_albumid is non-ASCII, so nothing moves today.
var releaseIDPattern = regexp.MustCompile(`\w{8}(?:-\w{4}){3}-\w{12}`)

// ErrNotFound reports that a lookup's release does not exist. beets'
// own flow branches on it: album_for_id catches HTTPNotFoundError and
// drops the candidate silently rather than failing the import
// (beetsplug/musicbrainz.py:777-781), so callers must be able to tell a
// deleted release from a real failure.
var ErrNotFound = errors.New("release not found")

// ErrUnavailable reports that every retry was spent on a retryable
// status or a transport error, 429 and 503 alike. A caller branches on
// it with errors.Is to tell an unreachable or overloaded MusicBrainz
// from a release that does not exist (ErrNotFound) or one that cannot
// be read (ErrMalformedRelease). It is the shape of beets' RetryError,
// which is not an HTTPError and so is never mapped to a beets exception
// (beetsplug/_utils/requests.py:78-89).
var ErrUnavailable = errors.New("musicbrainz unavailable")

// sleep is the backoff wait, a variable so a test observes the schedule
// without spending wall time. It is a timer rather than a bare
// time.Sleep so a caller's context can cancel the wait.
var sleep = sleepReal

// ClientConfig configures a [Client]. UserAgent is required and the
// engine never builds one: MusicBrainz requires an identified client
// and the identity is site-specific, so it belongs to internal/config
// and reaches here through cmd/dubplate-server's composition
// (dubplate-server/<version> ( <contact> ), on cmd/harness/capture.go's
// own precedent, with the contact from Config.AdminAddress).
type ClientConfig struct {
	BaseURL   string // empty means defaultBaseURL
	UserAgent string // required
}

// Client is a WS/2 client: fmt=json forced on every request, beets'
// own retry policy with three named divergences (the 30s timeout, the
// 413 non-port, and the HTTP-date Retry-After), an injected base URL, a
// required User-Agent, and the pseudo-release double fetch.
type Client struct {
	baseURL   string
	userAgent string
	http      *http.Client
}

// NewClient validates cfg and returns a client. It fails rather than
// defaulting on a missing UserAgent, because MusicBrainz requires an
// identified client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.UserAgent == "" {
		return nil, errors.New("ClientConfig.UserAgent is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		baseURL:   baseURL,
		userAgent: cfg.UserAgent,
		http:      &http.Client{Timeout: defaultTimeout},
	}, nil
}

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
// sixteen-element include set joined with a literal "+", encoded
// through url.Values so the wire form matches the committed corpus
// byte for byte with no adapter.
func (c *Client) releaseRequestURL(id string) string {
	v := url.Values{
		"fmt": {"json"},
		"inc": {strings.Join(releaseIncludes, "+")},
	}
	return c.baseURL + wsPathPrefix + "/release/" + id + "?" + v.Encode()
}

// searchRequestURL builds a search's URL: fmt=json forced and no inc
// at all.
func (c *Client) searchRequestURL(s AlbumSearch) string {
	v := url.Values{
		"fmt":   {"json"},
		"query": {s.Query},
		"limit": {strconv.Itoa(s.Limit)},
	}
	return c.baseURL + wsPathPrefix + "/release?" + v.Encode()
}

// release fetches one release at beets' own maximal inc= set, mirroring
// get_release (beetsplug/_utils/musicbrainz.py:687-690). It returns
// ErrNotFound on a 404, an error wrapping ErrUnavailable when the
// retries are spent, and an error wrapping ErrMalformedRelease when the
// body does not decode. mbid is run through the same id extraction beets
// applies first (beets/util/id_extractors.py:20), so a malformed id
// costs no request at all, exactly as album_for_id costs none
// (beetsplug/musicbrainz.py:771-773). It decodes only: TranslateAlbum
// validates.
func (c *Client) release(ctx context.Context, mbid string) (*Release, error) {
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
// releases: the per-id fan-out and the candidate ordering belong to the
// caller. A search body is an object whose hits live under "releases":
// search ends by indexing the decoded body at f"{entity}s" (:684-685).
//
// Its errors are its own, and they do not mirror beets. beets swallows
// a search failure into an empty result inside _search_api's bare
// except (beets/metadata_plugins.py:419-427, which re-raises first when
// raise_on_error is set at :422-423); this client returns the failure
// instead, because swallowing it here would hide a search that never
// happened from the request grading. A 404 returns ErrNotFound, and
// every other failure returns a non-nil error. A caller that wants
// beets' shape maps them to an empty candidate list.
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
// caller maps it to "no candidate", reproducing album_for_id's
// return None. A 404 on the SECOND, pseudo-release lookup returns a
// wrapped error rather than a dropped candidate, because beets' second
// get_release sits outside the try and its HTTPNotFoundError escapes to
// maybe_handle_plugin_error, which logs and drops the plugin's whole
// candidate list. The shapes differ; the outcome for the caller is the
// same candidate loss.
//
// Both AlbumInfo values are built inside this call and only the merged
// one escapes, so MergePseudoRelease's aliasing of the pseudo value's
// slices has no second retaining reader.
func (c *Client) AlbumForID(ctx context.Context, mbid string, tun matcher.Tunables) (*matcher.AlbumInfo, error) {
	rel, err := c.release(ctx, mbid)
	if err != nil {
		return nil, err
	}
	info, err := TranslateAlbum(rel, tun)
	if err != nil {
		return nil, err
	}

	// An absent status decodes to "" and reads as a non-pseudo release,
	// where beets' direct index (beetsplug/musicbrainz.py:785) raises
	// KeyError and drops the plugin's whole candidate list.
	targetID, ok := PseudoReleaseTarget(rel)
	if !ok {
		return info, nil
	}

	actualRel, err := c.release(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: fetching pseudo-release %s's actual release %s: %w", rel.ID, targetID, err)
	}
	actualInfo, err := TranslateAlbum(actualRel, tun)
	if err != nil {
		return nil, err
	}
	return MergePseudoRelease(info, actualInfo), nil
}

// doWithRetry issues one GET at requestURL, retrying as beets' own
// policy does. It returns the response body on a 200, ErrNotFound on a
// 404, an error wrapping ErrUnavailable when the retries are spent, and
// a generic error for any other status. A context that ends before or
// between attempts returns the context's own error, never
// ErrUnavailable.
func (c *Client) doWithRetry(ctx context.Context, requestURL string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("musicbrainz: building request: %w", err)
		}
		req.Header.Set("User-Agent", c.userAgent)

		var wait time.Duration
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			wait = backoff(attempt)
		} else {
			body, err := readBody(resp)
			if err != nil {
				return nil, err
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
			wait = retryWait(resp.Header.Get("Retry-After"), attempt)
		}

		if attempt == maxAttempts {
			break
		}
		if err := sleep(ctx, wait); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("%w: %w", ErrUnavailable, lastErr)
}

// readBody reads and closes resp's body, failing on one over
// maxBodyBytes rather than truncating it.
func readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: reading response body: %w", err)
	}
	if len(body) > maxBodyBytes {
		return nil, errBodyTooLarge
	}
	return body, nil
}

// retryWait picks the wait before the next attempt: a delta-seconds
// Retry-After header wins over the computed backoff, and an HTTP-date
// Retry-After falls through to the computed backoff untouched:
// reducing an HTTP date to a duration needs the current time, which the
// engine never reads, and the computed backoff is already a
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

// backoff is urllib3's own schedule (urllib3/util/retry.py:309-326):
// zero when the consecutive-error run is one or less, then
// 500ms * 2^(n-1), capped at 120s with no jitter. attempt is the
// 1-indexed attempt that just failed, which is that run count.
func backoff(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}
	return min(500*time.Millisecond*time.Duration(1<<(attempt-1)), 120*time.Second)
}
