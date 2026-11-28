package identity

import "time"

// An Actor is who makes a change: an operator at the command line,
// named by their operating-system user, or a member over the network,
// named by their durable id. It is verified before a change reaches the
// code that makes it, and never a field a request carries, so no request
// body can widen its own privilege. Every audit record, the roster's
// log and the settings audit alike, names an actor by its
// [Actor.String] form. The zero value names nobody.
type Actor struct {
	channel  Channel
	userID   UserID
	name     string
	anchor   time.Time
	clientIP string
}

// LocalActor returns the actor for a change an operating-system user
// makes from the command line. It renders as "local:<username>", its
// channel is [ChannelCommandLine], and its client IP reads "local". It
// panics on an empty username, which names nobody.
func LocalActor(username string) Actor {
	if username == "" {
		panic("identity: LocalActor called with an empty username")
	}
	return Actor{channel: ChannelCommandLine, name: username, clientIP: "local"}
}

// NetworkActor returns the actor for a change a verified login makes
// over the network: anchor is when the session that carried login was
// first issued, and clientIP is the address the request came from. It
// renders as "member:<hex UserID>", the durable key and never the
// username, so a renamed member keeps one actor. It panics unless
// login.Valid() holds, so a command-line identity can never become a
// network actor.
func NetworkActor(login Identity, anchor time.Time, clientIP string) Actor {
	if !login.Valid() {
		panic("identity: NetworkActor called with an invalid identity")
	}
	return Actor{channel: login.Channel, userID: login.UserID, anchor: anchor, clientIP: clientIP}
}

// String returns how an audit record names a. The zero Actor returns
// the empty string.
func (a Actor) String() string {
	switch {
	case a.channel == ChannelCommandLine:
		return "local:" + a.name
	case a.channel.network():
		return "member:" + a.userID.String()
	default:
		return ""
	}
}

// Channel returns how a was asserted: [ChannelCommandLine] for a local
// actor, the verified identity's channel for a network actor, and the
// zero Channel for the zero Actor.
func (a Actor) Channel() Channel { return a.channel }

// UserID returns the durable id of the member a network actor speaks
// for. A local actor has none and returns the zero UserID.
func (a Actor) UserID() UserID { return a.userID }

// Anchor returns when the session behind a network actor was first
// issued, the instant a re-authentication window counts from. A local
// actor has no session and returns the zero time.
func (a Actor) Anchor() time.Time { return a.anchor }

// ClientIP returns the address the request behind a network actor came
// from, as the verifier that built it resolved it, or "local" for a
// local actor.
func (a Actor) ClientIP() string { return a.clientIP }
