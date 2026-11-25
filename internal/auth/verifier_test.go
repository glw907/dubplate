package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// verifierBirth, verifierLifetime, and verifierCap pin every timestamp below. The
// cap is far longer than the lifetime, the same shape session_test.go
// and epoch_test.go already use, so a refresh has room to move the
// expiry without ever touching the cap.
var (
	verifierBirth    = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	verifierLifetime = 30 * 24 * time.Hour
	verifierCap      = 180 * 24 * time.Hour
)

// verifierFixture bundles one test's roster, epoch store, and seam, built
// fresh so a chmod or a bump in one test never leaks into another.
type verifierFixture struct {
	roster *member.Roster
	users  *store.Handle
	epoch  *EpochStore
	seam   *MagicLinkVerifier
	dir    string // stateDir, for reaching the epoch subdirectory directly
}

// newVerifierFixture opens a roster over a fresh store holding rows and a
// fresh epoch store under t.TempDir(), and wires both into a
// MagicLinkVerifier signing under currentFixture and previousFixture
// (token_test.go).
func newVerifierFixture(t *testing.T, rows ...userRow) verifierFixture {
	t.Helper()
	return newVerifierFixtureReading(t, fixed(verifierLifetime), fixed(verifierCap), rows...)
}

// newVerifierFixtureReading is [newVerifierFixture] with the verifier reading its
// lifetime and absolute cap through lifetime and absoluteCap.
func newVerifierFixtureReading(t *testing.T, lifetime, absoluteCap settings.Reader[time.Duration], rows ...userRow) verifierFixture {
	t.Helper()
	stateDir := t.TempDir()
	_, users := newRosterOver(t, rows...)
	epoch := newEpochStore(t, stateDir)
	roster := member.NewRoster(users, epoch, fixedNow(rosterNow))

	seam := NewMagicLinkVerifier(roster, epoch, []byte(currentFixture), []byte(previousFixture), lifetime, absoluteCap)
	return verifierFixture{roster: roster, users: users, epoch: epoch, seam: seam, dir: stateDir}
}

// requestWithSession builds a request carrying value as the session
// cookie, the shape VerifyIdentity reads.
func requestWithSession(t *testing.T, value string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.test/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: value})
	return req
}

// TestMagicLinkVerifier_VerifyIdentity covers the identity a valid session
// yields: the channel, the admin flag sourced from the roster, and the
// refusal for an address the roster no longer carries. Named mutation 1
// (read the admin flag from the session value) fails the "address named
// admin is not itself an admin claim" row, because that row's address
// carries the string "admin" while its roster record does not carry the
// flag.
func TestMagicLinkVerifier_VerifyIdentity(t *testing.T) {
	tests := []struct {
		name      string
		rows      []userRow
		address   string
		wantAdmin bool
		wantPrinc string
		wantErr   error
	}{
		{
			name:      "member on roster",
			rows:      []userRow{{address: "member@example.com", displayName: "Member", username: "member"}},
			address:   "member@example.com",
			wantAdmin: false,
			wantPrinc: "member",
		},
		{
			name:      "admin member on roster",
			rows:      []userRow{{address: "lead@example.com", displayName: "Lead", username: "chief", admin: true}},
			address:   "lead@example.com",
			wantAdmin: true,
			wantPrinc: "chief",
		},
		{
			name:      "address named admin is not itself an admin claim",
			rows:      []userRow{{address: "admin@example.com", displayName: "Not Actually Admin", username: "plainuser"}},
			address:   "admin@example.com",
			wantAdmin: false,
			wantPrinc: "plainuser",
		},
		{
			name:    "address removed from roster",
			address: "gone@example.com",
			wantErr: member.ErrMemberNotFound,
		},
		{
			name:    "address tombstoned",
			rows:    []userRow{{address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true}},
			address: "gone@example.com",
			wantErr: member.ErrMemberNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newVerifierFixture(t, tt.rows...)
			value, _, err := MintSession([]byte(currentFixture), tt.address, verifierBirth, verifierLifetime, verifierCap)
			if err != nil {
				t.Fatalf("MintSession: %v", err)
			}
			req := requestWithSession(t, value)

			got, err := fx.seam.VerifyIdentity(req, verifierBirth.Add(time.Hour), false)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("VerifyIdentity error = %v, want %v", err, tt.wantErr)
				}
				if got != (VerifiedIdentity{}) {
					t.Errorf("VerifyIdentity result = %+v, want zero value on refusal", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("VerifyIdentity: %v", err)
			}
			if got.Identity.Channel != identity.ChannelCookie {
				t.Errorf("Channel = %v, want cookie", got.Identity.Channel)
			}
			if got.Identity.Address != tt.address {
				t.Errorf("Address = %q, want %q", got.Identity.Address, tt.address)
			}
			if got.Identity.Principal != tt.wantPrinc {
				t.Errorf("Principal = %q, want %q", got.Identity.Principal, tt.wantPrinc)
			}
			if got.Admin != tt.wantAdmin {
				t.Errorf("Admin = %v, want %v", got.Admin, tt.wantAdmin)
			}
		})
	}
}

// TestMagicLinkVerifier_VerifyIdentity_NeverCommandLine proves every
// identity the seam verifies carries the cookie channel and passes
// identity.Identity.Valid, for an admin and a non-admin, with and
// without a refresh. The command-line channel names an operating-system
// user and is the one channel that may grant the admin flag, so a seam
// emitting it would hand a browser request the grant.
func TestMagicLinkVerifier_VerifyIdentity_NeverCommandLine(t *testing.T) {
	rows := []userRow{
		{address: "lead@example.com", displayName: "Lead", username: "lead", admin: true},
		{address: "member@example.com", displayName: "Member", username: "member"},
	}
	for _, address := range []string{"lead@example.com", "member@example.com"} {
		for _, sameOrigin := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s same-origin %v", address, sameOrigin), func(t *testing.T) {
				fx := newVerifierFixture(t, rows...)
				value, _, err := MintSession([]byte(currentFixture), address, verifierBirth, verifierLifetime, verifierCap)
				if err != nil {
					t.Fatalf("MintSession: %v", err)
				}
				got, err := fx.seam.VerifyIdentity(requestWithSession(t, value), verifierBirth.Add(time.Hour), sameOrigin)
				if err != nil {
					t.Fatalf("VerifyIdentity: %v", err)
				}
				if got.Identity.Channel != identity.ChannelCookie {
					t.Errorf("Channel = %v, want cookie", got.Identity.Channel)
				}
				if !got.Identity.Valid() {
					t.Errorf("Identity.Valid() = false for %+v, want true", got.Identity)
				}
			})
		}
	}
}

// TestMagicLinkVerifier_VerifyIdentity_EpochRevokedAfterRefresh proves that
// refreshing a session cannot outrun a revocation aimed at its anchor:
// the anchor a refresh never moves is what a later epoch bump compares
// against, the property 5b-i's CheckSession doc records. The test
// performs a refresh before the bump, so a check reading IssuedAt
// instead of AnchoredAt would miss it.
func TestMagicLinkVerifier_VerifyIdentity_EpochRevokedAfterRefresh(t *testing.T) {
	fx := newVerifierFixture(t, userRow{address: "member@example.com", displayName: "Member", username: "member"})
	value, _, err := MintSession([]byte(currentFixture), "member@example.com", verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	refreshTime := verifierBirth.Add(time.Hour)
	req := requestWithSession(t, value)
	refreshed, err := fx.seam.VerifyIdentity(req, refreshTime, true)
	if err != nil {
		t.Fatalf("VerifyIdentity (refresh): %v", err)
	}
	if refreshed.RefreshedValue == "" {
		t.Fatalf("RefreshedValue empty, want a refreshed cookie value")
	}

	// The bump lands after the original anchor and before the refresh,
	// which a check keyed on IssuedAt would miss and a check keyed on
	// AnchoredAt (unmoved by the refresh) still catches.
	bumpTime := verifierBirth.Add(30 * time.Minute)
	if err := fx.epoch.Bump("member@example.com", bumpTime); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	req2 := requestWithSession(t, refreshed.RefreshedValue)
	_, err = fx.seam.VerifyIdentity(req2, refreshTime.Add(time.Minute), true)
	if !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("VerifyIdentity (post-bump) error = %v, want ErrSessionRevoked", err)
	}
}

// TestMagicLinkVerifier_VerifyIdentity_CrossOriginDoesNotRefresh proves the
// refresh runs only for a same-origin request: the session VerifyIdentity
// returns for a cross-origin request carries the same expiry the minted
// cookie did, and RefreshedValue is empty. Named mutation 4 (refresh
// unconditionally) fails this row, because it would move ExpiresAt
// forward regardless of sameOrigin.
func TestMagicLinkVerifier_VerifyIdentity_CrossOriginDoesNotRefresh(t *testing.T) {
	fx := newVerifierFixture(t, userRow{address: "member@example.com", displayName: "Member", username: "member"})
	value, minted, err := MintSession([]byte(currentFixture), "member@example.com", verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	req := requestWithSession(t, value)
	got, err := fx.seam.VerifyIdentity(req, verifierBirth.Add(time.Hour), false)
	if err != nil {
		t.Fatalf("VerifyIdentity: %v", err)
	}
	if got.RefreshedValue != "" {
		t.Errorf("RefreshedValue = %q, want empty on a cross-origin request", got.RefreshedValue)
	}
	if !got.Session.ExpiresAt.Equal(minted.ExpiresAt) {
		t.Errorf("Session.ExpiresAt = %v, want unchanged %v", got.Session.ExpiresAt, minted.ExpiresAt)
	}
}

// TestMagicLinkVerifier_VerifyIdentity_UnreadableEpochDenies proves that a
// non-sentinel error from EpochStore.Read (via CheckSession) denies
// exactly as a sentinel refusal would, the fail-closed discipline
// security constraint 2 states. Named mutation 3 (treat a non-sentinel
// error as "not refused") fails this row, because it would let a valid
// session and a valid roster member through despite the read failure.
func TestMagicLinkVerifier_VerifyIdentity_UnreadableEpochDenies(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not block reads")
	}
	fx := newVerifierFixture(t, userRow{address: "member@example.com", displayName: "Member", username: "member"})
	value, _, err := MintSession([]byte(currentFixture), "member@example.com", verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	epochDir := filepath.Join(fx.dir, leafDirName, epochSubdir)
	if err := os.Chmod(epochDir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(epochDir, 0o700) })

	req := requestWithSession(t, value)
	_, err = fx.seam.VerifyIdentity(req, verifierBirth.Add(time.Hour), false)
	if err == nil {
		t.Fatalf("VerifyIdentity: got nil error, want a denial from the unreadable epoch directory")
	}
	if errors.Is(err, member.ErrMemberNotFound) || errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("VerifyIdentity error = %v, want a non-sentinel I/O error", err)
	}
}

// TestMagicLinkVerifier_VerifyIdentity_UsersTable proves the seam resolves a
// session through the users table: the identity carries the row's id as
// UserID and the row's username as Principal. The member's address and
// username differ, so a Principal built from the address fails.
func TestMagicLinkVerifier_VerifyIdentity_UsersTable(t *testing.T) {
	h := newUsersStore(t)
	insertUsers(t, h, userRow{id: aliceID, address: "alice@example.com", displayName: "Alice", username: "ally", createdAt: 100})
	seam := NewMagicLinkVerifier(member.NewRoster(h, nil, nil), newEpochStore(t, t.TempDir()), []byte(currentFixture), []byte(previousFixture), fixed(verifierLifetime), fixed(verifierCap))
	value, _, err := MintSession([]byte(currentFixture), "alice@example.com", verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	got, err := seam.VerifyIdentity(requestWithSession(t, value), verifierBirth.Add(time.Hour), false)
	if err != nil {
		t.Fatalf("VerifyIdentity: %v", err)
	}
	want := identity.Identity{
		Address:     "alice@example.com",
		UserID:      identity.UserID{0xa1, 0x1c, 0xe0},
		Principal:   "ally",
		DisplayName: "Alice",
		Source:      magicLinkSource,
		Channel:     identity.ChannelCookie,
	}
	if got.Identity != want {
		t.Errorf("Identity = %+v, want %+v", got.Identity, want)
	}
	if got.Admin {
		t.Error("Admin = true, want the row's false")
	}
}

// TestMagicLinkVerifier_ReadsTheCapAtEveryVerification proves the absolute
// cap applies to the next verification in both directions: one session,
// minted under a long cap and verified at the same instant, is refused
// once the cap in force drops below its age and admitted again once the
// cap rises. A cap cached when the seam was built would answer every
// row alike.
func TestMagicLinkVerifier_ReadsTheCapAtEveryVerification(t *testing.T) {
	member := userRow{address: "member@example.com", displayName: "Member", username: "member"}
	absoluteCap := newSettingFake(verifierCap)
	fx := newVerifierFixtureReading(t, fixed(verifierLifetime), absoluteCap.setting(), member)
	value, _, err := MintSession([]byte(currentFixture), member.address, verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	at := verifierBirth.Add(10 * 24 * time.Hour)

	for _, tt := range []struct {
		name    string
		cap     time.Duration
		wantErr error
	}{
		{"the cap in force admits the session's age", verifierCap, nil},
		{"a cap lowered below the session's age refuses it", 5 * 24 * time.Hour, ErrSessionCapReached},
		{"the cap raised again admits it", verifierCap, nil},
	} {
		absoluteCap.set(tt.cap, nil)
		_, err := fx.seam.VerifyIdentity(requestWithSession(t, value), at, false)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("%s: VerifyIdentity error = %v, want %v", tt.name, err, tt.wantErr)
		}
	}
}

// TestMagicLinkVerifier_ReadsTheLifetimeAtEachRefresh proves a refresh takes
// the lifetime in force at that refresh, never the one in force when
// the seam was built.
func TestMagicLinkVerifier_ReadsTheLifetimeAtEachRefresh(t *testing.T) {
	member := userRow{address: "member@example.com", displayName: "Member", username: "member"}
	lifetime := newSettingFake(verifierLifetime)
	fx := newVerifierFixtureReading(t, lifetime.setting(), fixed(verifierCap), member)
	value, _, err := MintSession([]byte(currentFixture), member.address, verifierBirth, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	at := verifierBirth.Add(time.Hour)
	const changed = 2 * 24 * time.Hour
	lifetime.set(changed, nil)

	got, err := fx.seam.VerifyIdentity(requestWithSession(t, value), at, true)
	if err != nil {
		t.Fatalf("VerifyIdentity: %v", err)
	}
	if want := at.Add(changed); !got.Session.ExpiresAt.Equal(want) {
		t.Errorf("refreshed expiry = %v, want %v", got.Session.ExpiresAt, want)
	}
}

// TestMagicLinkVerifier_SettingReadFailure proves a failed read of the cap,
// on every verification, or of the lifetime, on a refresh, refuses with
// an error wrapping errSettingUnreadable and the reader's own text,
// never a default; a lifetime that fails on a request that does not
// refresh is never read.
func TestMagicLinkVerifier_SettingReadFailure(t *testing.T) {
	member := userRow{address: "member@example.com", displayName: "Member", username: "member"}
	tests := []struct {
		name       string
		failCap    bool
		sameOrigin bool
		wantErr    bool
	}{
		{"the cap reader fails without a refresh", true, false, true},
		{"the cap reader fails on a refresh", true, true, true},
		{"the lifetime reader fails on a refresh", false, true, true},
		{"the lifetime reader fails without a refresh", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifetime, absoluteCap := newSettingFake(verifierLifetime), newSettingFake(verifierCap)
			fx := newVerifierFixtureReading(t, lifetime.setting(), absoluteCap.setting(), member)
			value, _, err := MintSession([]byte(currentFixture), member.address, verifierBirth, verifierLifetime, verifierCap)
			if err != nil {
				t.Fatalf("MintSession: %v", err)
			}
			if tt.failCap {
				absoluteCap.set(0, errFixtureRead)
			} else {
				lifetime.set(0, errFixtureRead)
			}

			got, err := fx.seam.VerifyIdentity(requestWithSession(t, value), verifierBirth.Add(time.Hour), tt.sameOrigin)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("VerifyIdentity: %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, errSettingUnreadable) || !errors.Is(err, errFixtureRead) {
				t.Fatalf("VerifyIdentity error = %v, want errSettingUnreadable wrapping the reader's error", err)
			}
			if got != (VerifiedIdentity{}) {
				t.Errorf("VerifyIdentity result = %+v, want zero value on refusal", got)
			}
		})
	}
}

// TestMagicLinkVerifier_CapReadFollowsTheMAC proves the absolute cap is read
// only for a cookie whose MAC verifies: a cookie that does not parse,
// or that parses under a secret the seam does not hold, reaches no
// setting read and so costs no database work, while a valid cookie
// still reads the cap at its verification.
func TestMagicLinkVerifier_CapReadFollowsTheMAC(t *testing.T) {
	member := userRow{address: "member@example.com", displayName: "Member", username: "member"}
	mint := func(t *testing.T, secret string) string {
		t.Helper()
		value, _, err := MintSession([]byte(secret), member.address, verifierBirth, verifierLifetime, verifierCap)
		if err != nil {
			t.Fatalf("MintSession: %v", err)
		}
		return value
	}
	tests := []struct {
		name      string
		value     func(t *testing.T) string
		wantErr   error
		wantReads int
	}{
		{"a value that does not parse", func(*testing.T) string { return "not-a-session" }, ErrSessionMalformed, 0},
		{"a MAC under neither secret", func(t *testing.T) string { return mint(t, "an-unrelated-secret-of-enough-length") }, ErrSessionMAC, 0},
		{"a valid session", func(t *testing.T) string { return mint(t, currentFixture) }, nil, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reads int
			absoluteCap := settings.Reader[time.Duration](func(context.Context) (time.Duration, error) {
				reads++
				return verifierCap, nil
			})
			fx := newVerifierFixtureReading(t, fixed(verifierLifetime), absoluteCap, member)

			_, err := fx.seam.VerifyIdentity(requestWithSession(t, tt.value(t)), verifierBirth.Add(time.Hour), false)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("VerifyIdentity: %v, want nil", err)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("VerifyIdentity error = %v, want %v", err, tt.wantErr)
			}
			if reads != tt.wantReads {
				t.Errorf("cap reads = %d, want %d", reads, tt.wantReads)
			}
		})
	}
}
