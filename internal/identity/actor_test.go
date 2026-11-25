package identity

import (
	"testing"
	"time"
)

func TestLocalActor(t *testing.T) {
	got := LocalActor("geoff")
	if got.String() != "local:geoff" {
		t.Errorf("LocalActor(%q).String() = %q, want %q", "geoff", got.String(), "local:geoff")
	}
	if got.Channel() != ChannelCommandLine {
		t.Errorf("LocalActor(%q).Channel() = %v, want %v", "geoff", got.Channel(), ChannelCommandLine)
	}
	if got.ClientIP() != "local" {
		t.Errorf("LocalActor(%q).ClientIP() = %q, want %q", "geoff", got.ClientIP(), "local")
	}
}

func TestLocalActor_EmptyUsernamePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("LocalActor(\"\") did not panic")
		}
	}()
	LocalActor("")
}

// TestActor_ZeroNamesNobody proves the zero Actor renders as no actor
// at all, never as a member with the zero id.
func TestActor_ZeroNamesNobody(t *testing.T) {
	var zero Actor
	if got := zero.String(); got != "" {
		t.Errorf("Actor{}.String() = %q, want the empty string", got)
	}
	if zero.Channel().Valid() {
		t.Errorf("Actor{}.Channel() = %v, want an invalid channel", zero.Channel())
	}
}

// validIdentity is a cookie-verified login, valid by construction.
func validIdentity() Identity {
	return Identity{
		Address:  "geoff@example.com",
		UserID:   UserID{0x01},
		Username: "geoff",
		Channel:  ChannelCookie,
	}
}

// testAnchor is the instant every network actor here was anchored.
var testAnchor = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

// TestNetworkActor pins the wire form, the channel, and the three
// read-only accessors for a valid login. The wire form is the
// lowercase hex of the UserID, written out here rather than computed.
func TestNetworkActor(t *testing.T) {
	got := NetworkActor(validIdentity(), testAnchor, "192.0.2.1")
	if want := "member:01000000000000000000000000000000"; got.String() != want {
		t.Errorf("NetworkActor(...).String() = %q, want %q", got.String(), want)
	}
	if got.ClientIP() != "192.0.2.1" {
		t.Errorf("NetworkActor(...).ClientIP() = %q, want %q", got.ClientIP(), "192.0.2.1")
	}
	if got.Channel() != ChannelCookie {
		t.Errorf("NetworkActor(...).Channel() = %v, want %v", got.Channel(), ChannelCookie)
	}
	if want := (UserID{0x01}); got.UserID() != want {
		t.Errorf("NetworkActor(...).UserID() = %x, want %x", got.UserID(), want)
	}
	if !got.Anchor().Equal(testAnchor) {
		t.Errorf("NetworkActor(...).Anchor() = %v, want %v", got.Anchor(), testAnchor)
	}
}

// TestNetworkActor_WireFormFollowsUserID proves the wire form is the
// durable key alone: two logins with one UserID and different
// usernames render one form, in lowercase, and a login with another
// UserID renders another.
func TestNetworkActor_WireFormFollowsUserID(t *testing.T) {
	login := func(user UserID, username string) Identity {
		i := validIdentity()
		i.UserID, i.Username = user, username
		return i
	}
	shared := UserID{0xab, 0xcd, 0xef}
	const wantShared = "member:abcdef00000000000000000000000000"
	tests := []struct {
		name string
		id   Identity
		want string
	}{
		{"username geoff", login(shared, "geoff"), wantShared},
		{"username renamed", login(shared, "geoffrey"), wantShared},
		{"another member", login(UserID{0x02}, "geoff"), "member:02000000000000000000000000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NetworkActor(tt.id, testAnchor, "192.0.2.1").String(); got != tt.want {
				t.Errorf("NetworkActor(%+v).String() = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

// TestLocalActor_CarriesNoSession proves the command line has no
// durable id and no anchor to read.
func TestLocalActor_CarriesNoSession(t *testing.T) {
	got := LocalActor("geoff")
	if got.UserID() != (UserID{}) {
		t.Errorf("LocalActor(...).UserID() = %x, want the zero UserID", got.UserID())
	}
	if !got.Anchor().IsZero() {
		t.Errorf("LocalActor(...).Anchor() = %v, want the zero time", got.Anchor())
	}
}

// TestNetworkActor_RefusesInvalidIdentity proves construction refuses
// both the zero identity and a command-line identity: neither is a
// verified login a network caller can assert.
func TestNetworkActor_RefusesInvalidIdentity(t *testing.T) {
	tests := []struct {
		name string
		id   Identity
	}{
		{"zero identity", Identity{}},
		{"command-line identity", func() Identity {
			id := validIdentity()
			id.Channel = ChannelCommandLine
			return id
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("NetworkActor(%+v) did not panic", tt.id)
				}
			}()
			NetworkActor(tt.id, testAnchor, "192.0.2.1")
		})
	}
}
