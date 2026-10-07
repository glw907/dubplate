package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
)

// The rendered fixture's member addresses and the admin's display name.
// The admin's display name carries `<` and `>` (universal at
// `example.test` scope), so a page that rendered it with `{@html}`
// would leave a stray element behind rather than the literal text a
// rendered row asserts. The three addresses are the reset baseline.
const (
	renderedAdminAddress     = "admin@example.test"
	renderedAdminDisplayName = "Fixture <Admin>"
	renderedMemberAddress    = "member@example.test"

	// renderedNewMemberAddress is the address the add-member row writes:
	// on no roster at the run's start, so `duplicate_address` on the
	// second submit proves the first write landed. It belongs to the
	// reset baseline, since that row runs before the capture matrix.
	renderedNewMemberAddress = "new-member@example.test"
)

// seedRenderedMembers writes the two fixture members, the admin first,
// into the store under stateDir. Their usernames match none of the
// usernames the browser script types, compared case-insensitively, or a
// rendered row's expected outcome would change.
func seedRenderedMembers(t *testing.T, stateDir string) {
	t.Helper()
	seedMembers(t, stateDir,
		seedMember{address: renderedAdminAddress, displayName: renderedAdminDisplayName, username: "fixtureadmin", admin: true},
		seedMember{address: renderedMemberAddress, displayName: "Fixture Member", username: "fixturemember"},
	)
}

// requirePOST answers 405 and reports false unless r is a POST, the
// shared guard the rendered leg's helper servers open their ServeHTTP
// with.
func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// rosterResetServer resets the fixture's users table to a named
// baseline on every POST, so each capture cell starts from the same
// member list and every cell of one members state renders the same
// count. The baseline is the set of addresses the harness names: the
// two seeded members and the add-member row's address. Each POST
// deletes, in one write, every row outside that set and answers
// {"members": n}, the number of rows left, so the answer does not
// depend on when the first POST arrives. It writes through its own
// write-mode handle over the fixture state directory, which the shell
// under test holds open beside it. Like [mintServer], it is its own
// loopback listener.
type rosterResetServer struct {
	stateDir string
}

func (s *rosterResetServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	members, err := s.reset(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Members int `json:"members"`
	}{members})
}

// reset deletes every users row whose address is outside the baseline
// and returns how many rows remain.
func (s *rosterResetServer) reset(ctx context.Context) (members int, err error) {
	h, err := store.OpenWrite(ctx, s.stateDir, testBackupBusy)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := h.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	err = h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM users WHERE address NOT IN (?, ?, ?)`,
			renderedAdminAddress, renderedMemberAddress, renderedNewMemberAddress); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&members)
	})
	return members, err
}

// TestRosterResetServer proves the reset server alone, with no browser.
// Fixture state: a store holding the two seeded rows. The first POST
// answers 2. Then the add-member row's address and one capture row are
// added, as the rendered run does before and during its capture matrix;
// the second POST must answer 3 and leave the capture row gone from the
// table. A baseline taken from the table at the first POST answers 2
// here, and a reset that deletes nothing leaves the capture row behind.
func TestRosterResetServer(t *testing.T) {
	stateDir := t.TempDir()
	seedBackupStore(t, stateDir)
	seedRenderedMembers(t, stateDir)
	srv := httptest.NewServer(&rosterResetServer{stateDir: stateDir})
	t.Cleanup(srv.Close)

	post := func() int {
		t.Helper()
		resp, err := http.Post(srv.URL, "text/plain", nil)
		if err != nil {
			t.Fatalf("POST the reset server: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var got struct {
			Members int `json:"members"`
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("reset answered %d, want 200", resp.StatusCode)
		}
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode the reset answer: %v", err)
		}
		return got.Members
	}

	if got := post(); got != 2 {
		t.Fatalf("first reset answered %d members, want the 2 seeded", got)
	}

	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	roster := auth.NewRoster(h)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	for _, m := range []auth.Member{
		{Address: renderedNewMemberAddress, DisplayName: "New", Username: "newmember"},
		{Address: "capture-added@example.test", DisplayName: "Capture", Username: "capture-added"},
	} {
		if err := roster.AddMember(ctx, m, "operator", identity.ChannelCommandLine, "", discardTestLogger()); err != nil {
			t.Fatalf("add %s: %v", m.Address, err)
		}
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close the fixture handle: %v", err)
	}

	if got := post(); got != 3 {
		t.Errorf("second reset answered %d members, want 3: the two seeded rows and the add-member row", got)
	}
	want := []string{renderedAdminAddress, renderedMemberAddress, renderedNewMemberAddress}
	if got := userAddresses(t, stateDir); !slices.Equal(got, want) {
		t.Errorf("users after the second reset = %v, want %v: the capture row must be gone", got, want)
	}
}
