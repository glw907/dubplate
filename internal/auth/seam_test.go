package auth

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// seamBirth, seamLifetime, and seamCap pin every timestamp below. The
// cap is far longer than the lifetime, the same shape session_test.go
// and epoch_test.go already use, so a refresh has room to move the
// expiry without ever touching the cap.
var (
	seamBirth    = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	seamLifetime = 30 * 24 * time.Hour
	seamCap      = 180 * 24 * time.Hour
)

// seamFixture bundles one test's roster, epoch store, and seam, built
// fresh so a chmod or a bump in one test never leaks into another.
type seamFixture struct {
	roster *Roster
	epoch  *EpochStore
	seam   *MagicLinkSeam
	dir    string // stateDir, for reaching the epoch subdirectory directly
}

// newSeamFixture opens a roster carrying rosterJSON's members array and
// a fresh epoch store under t.TempDir(), and wires both into a
// MagicLinkSeam signing under currentFixture and previousFixture
// (token_test.go).
func newSeamFixture(t *testing.T, rosterJSON string) seamFixture {
	t.Helper()
	stateDir := t.TempDir()

	rosterPath := writeRosterFile(t, `{"schema_version":1,"members":`+rosterJSON+`}`)
	roster, err := OpenRoster(rosterPath)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	epoch := newEpochStore(t, stateDir)

	seam := NewMagicLinkSeam(roster, epoch, []byte(currentFixture), []byte(previousFixture), seamLifetime, seamCap)
	return seamFixture{roster: roster, epoch: epoch, seam: seam, dir: stateDir}
}

// requestWithSession builds a request carrying value as the session
// cookie, the shape VerifyIdentity reads.
func requestWithSession(t *testing.T, value string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.test/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: value})
	return req
}

// TestMagicLinkSeam_VerifyIdentity covers the identity a valid session
// yields: the channel, the admin flag sourced from the roster, and the
// refusal for an address the roster no longer carries. Named mutation 1
// (read the admin flag from the session value) fails the "address named
// admin is not itself an admin claim" row, because that row's address
// carries the string "admin" while its roster record does not carry the
// flag.
func TestMagicLinkSeam_VerifyIdentity(t *testing.T) {
	tests := []struct {
		name       string
		rosterJSON string
		address    string
		wantAdmin  bool
		wantErr    error
	}{
		{
			name:       "member on roster",
			rosterJSON: `[{"address":"member@example.com","display_name":"Member","admin":false}]`,
			address:    "member@example.com",
			wantAdmin:  false,
		},
		{
			name:       "admin member on roster",
			rosterJSON: `[{"address":"lead@example.com","display_name":"Lead","admin":true}]`,
			address:    "lead@example.com",
			wantAdmin:  true,
		},
		{
			name:       "address named admin is not itself an admin claim",
			rosterJSON: `[{"address":"admin@example.com","display_name":"Not Actually Admin","admin":false}]`,
			address:    "admin@example.com",
			wantAdmin:  false,
		},
		{
			name:       "address removed from roster",
			rosterJSON: `[]`,
			address:    "gone@example.com",
			wantErr:    ErrMemberNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newSeamFixture(t, tt.rosterJSON)
			value, _, err := MintSession([]byte(currentFixture), tt.address, seamBirth, seamLifetime)
			if err != nil {
				t.Fatalf("MintSession: %v", err)
			}
			req := requestWithSession(t, value)

			got, err := fx.seam.VerifyIdentity(req, seamBirth.Add(time.Hour), false)
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
			if got.Admin != tt.wantAdmin {
				t.Errorf("Admin = %v, want %v", got.Admin, tt.wantAdmin)
			}
		})
	}
}

// TestMagicLinkSeam_VerifyIdentity_EpochRevokedAfterRefresh proves that
// refreshing a session cannot outrun a revocation aimed at its anchor:
// the anchor a refresh never moves is what a later epoch bump compares
// against, the property 5b-i's CheckSession doc records. The test
// performs a refresh before the bump, so a check reading IssuedAt
// instead of AnchoredAt would miss it.
func TestMagicLinkSeam_VerifyIdentity_EpochRevokedAfterRefresh(t *testing.T) {
	fx := newSeamFixture(t, `[{"address":"member@example.com","display_name":"Member","admin":false}]`)
	value, _, err := MintSession([]byte(currentFixture), "member@example.com", seamBirth, seamLifetime)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	refreshTime := seamBirth.Add(time.Hour)
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
	bumpTime := seamBirth.Add(30 * time.Minute)
	if err := fx.epoch.Bump("member@example.com", bumpTime); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	req2 := requestWithSession(t, refreshed.RefreshedValue)
	_, err = fx.seam.VerifyIdentity(req2, refreshTime.Add(time.Minute), true)
	if !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("VerifyIdentity (post-bump) error = %v, want ErrSessionRevoked", err)
	}
}

// TestMagicLinkSeam_VerifyIdentity_CrossOriginDoesNotRefresh proves the
// refresh runs only for a same-origin request: the session VerifyIdentity
// returns for a cross-origin request carries the same expiry the minted
// cookie did, and RefreshedValue is empty. Named mutation 4 (refresh
// unconditionally) fails this row, because it would move ExpiresAt
// forward regardless of sameOrigin.
func TestMagicLinkSeam_VerifyIdentity_CrossOriginDoesNotRefresh(t *testing.T) {
	fx := newSeamFixture(t, `[{"address":"member@example.com","display_name":"Member","admin":false}]`)
	value, minted, err := MintSession([]byte(currentFixture), "member@example.com", seamBirth, seamLifetime)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	req := requestWithSession(t, value)
	got, err := fx.seam.VerifyIdentity(req, seamBirth.Add(time.Hour), false)
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

// TestMagicLinkSeam_VerifyIdentity_UnreadableEpochDenies proves that a
// non-sentinel error from EpochStore.Read (via CheckSession) denies
// exactly as a sentinel refusal would, the fail-closed discipline
// security constraint 2 states. Named mutation 3 (treat a non-sentinel
// error as "not refused") fails this row, because it would let a valid
// session and a valid roster member through despite the read failure.
func TestMagicLinkSeam_VerifyIdentity_UnreadableEpochDenies(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not block reads")
	}
	fx := newSeamFixture(t, `[{"address":"member@example.com","display_name":"Member","admin":false}]`)
	value, _, err := MintSession([]byte(currentFixture), "member@example.com", seamBirth, seamLifetime)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	epochDir := filepath.Join(fx.dir, leafDirName, epochSubdir)
	if err := os.Chmod(epochDir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(epochDir, 0o700) })

	req := requestWithSession(t, value)
	_, err = fx.seam.VerifyIdentity(req, seamBirth.Add(time.Hour), false)
	if err == nil {
		t.Fatalf("VerifyIdentity: got nil error, want a denial from the unreadable epoch directory")
	}
	if errors.Is(err, ErrMemberNotFound) || errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("VerifyIdentity error = %v, want a non-sentinel I/O error", err)
	}
}

// TestMagicLinkSeam_Read proves Read is the authority a caller can
// consult directly: a bump with no subscriber registered still shows up
// on Read, proving Read does not depend on the notification path.
func TestMagicLinkSeam_Read(t *testing.T) {
	fx := newSeamFixture(t, `[]`)
	bumpTime := seamBirth.Add(time.Minute)
	if err := fx.epoch.Bump("member@example.com", bumpTime); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	got, ok, err := fx.seam.Read("member@example.com")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !ok {
		t.Fatalf("Read: ok = false, want true")
	}
	if !got.Equal(bumpTime.Truncate(time.Second)) {
		t.Errorf("Read = %v, want %v", got, bumpTime.Truncate(time.Second))
	}
}
