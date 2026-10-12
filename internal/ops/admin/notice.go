package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
)

// noticeWait bounds the whole sensitive-change notice: the roster read
// and every send. The caller's answer waits on it, and the server's
// 30-second write timeout would otherwise cut the connection after a
// slow mail server and show the admin a failure for a write that
// committed.
const noticeWait = 10 * time.Second

// notice mails every live admin that actor changed key, a sensitive
// setting, as res reports. The caller runs it after the write commits
// and outside its transaction, so no writer waits on the mail server.
// Its context keeps none of ctx's cancellation or deadline and takes its
// own: a client that disconnects after the commit does not stop the
// send. A failure is logged and changes nothing. The write stays
// committed, and its settings audit row is the durable record.
func (s *Seam) notice(ctx context.Context, actor ops.Actor, key string, res settings.Result) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), noticeWait)
	defer cancel()
	log := s.cfg.Logger.With("key", key, "actor", actor.String())

	// The recipients are read after the commit, so an admin removed
	// while the write waited hears nothing.
	members, err := s.cfg.Roster.List(ctx)
	if err != nil {
		log.Error("sensitive-change notice not sent", "reason", err.Error())
		return
	}
	principal := actor.String()
	for _, m := range members {
		if m.ID == actor.UserID() {
			principal = m.Username
		}
	}
	subject, body := noticeText(principal, key, res)
	for _, m := range members {
		if !m.Admin {
			continue
		}
		if err := s.cfg.Send(ctx, m.Address, subject, body); err != nil {
			log.Error("sensitive-change notice failed", "reason", err.Error())
		}
	}
}

// noticeText is the notice's subject and body: who changed key, from
// what value to what.
func noticeText(principal, key string, res settings.Result) (subject, body string) {
	change := "changed"
	if res.Source == settings.SourceDefault {
		change = "reset"
	}
	subject = "Sensitive setting " + change + ": " + key
	body = fmt.Sprintf("%s %s %s, a sensitive setting, from the server's settings page.\n\n"+
		"Old value: %q\nNew value: %q (%s)\n\n"+
		"The settings audit records every change. If you did not expect this one, "+
		"check who in the group holds an admin sign-in.\n",
		principal, change, key, res.Previous, res.Value, res.Source)
	return subject, body
}
