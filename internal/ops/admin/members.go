package admin

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/ops"
)

// ListMembers answers the live roster in the order members were added.
func (s *Seam) ListMembers(ctx context.Context, _ ops.Actor, _ ops.ListMembersRequest) (ops.ListMembersResponse, error) {
	members, err := s.cfg.Roster.List(ctx)
	if err != nil {
		return ops.ListMembersResponse{}, err
	}
	// Non-nil, so an empty roster encodes as [] rather than null.
	out := make([]ops.Member, 0, len(members))
	for _, m := range members {
		out = append(out, wireMember(m))
	}
	return ops.ListMembersResponse{Members: out}, nil
}

// AddMember adds a member, or reactivates a removed one, without the
// admin flag, which the request cannot carry. It is gated, since the
// access it grants outlives the cookie that asked for it.
func (s *Seam) AddMember(ctx context.Context, actor ops.Actor, req ops.AddMemberRequest) (ops.AddMemberResponse, error) {
	if err := s.requireFresh(ctx, actor); err != nil {
		return ops.AddMemberResponse{}, err
	}
	candidate := auth.Member{Address: req.Address, DisplayName: req.DisplayName, Username: req.Username}
	if err := writeMember(ctx, func(ctx context.Context) error {
		return s.cfg.Roster.AddMember(ctx, candidate, rosterActor(actor), s.cfg.Logger)
	}); err != nil {
		return ops.AddMemberResponse{}, err
	}
	added, err := s.cfg.Roster.Lookup(ctx, req.Address)
	if err != nil {
		return ops.AddMemberResponse{}, fmt.Errorf("read the added member: %w", err)
	}
	return ops.AddMemberResponse{Member: wireMember(added)}, nil
}

// RemoveMember removes the member req names, which revokes their
// sessions. It is gated.
func (s *Seam) RemoveMember(ctx context.Context, actor ops.Actor, req ops.RemoveMemberRequest) (ops.RemoveMemberResponse, error) {
	id, err := parseUserID(req.UserID)
	if err != nil {
		return ops.RemoveMemberResponse{}, err
	}
	if err := s.requireFresh(ctx, actor); err != nil {
		return ops.RemoveMemberResponse{}, err
	}
	return ops.RemoveMemberResponse{}, writeMember(ctx, func(ctx context.Context) error {
		return s.cfg.Roster.RemoveMember(ctx, id, rosterActor(actor), s.cfg.Logger)
	})
}

// DemoteMember clears the admin flag of the member req names. It is
// gated, since the web cannot grant the flag back.
func (s *Seam) DemoteMember(ctx context.Context, actor ops.Actor, req ops.DemoteMemberRequest) (ops.DemoteMemberResponse, error) {
	id, err := parseUserID(req.UserID)
	if err != nil {
		return ops.DemoteMemberResponse{}, err
	}
	if err := s.requireFresh(ctx, actor); err != nil {
		return ops.DemoteMemberResponse{}, err
	}
	return ops.DemoteMemberResponse{}, writeMember(ctx, func(ctx context.Context) error {
		return s.cfg.Roster.DemoteMember(ctx, id, rosterActor(actor), s.cfg.Logger)
	})
}

// writeMember runs one roster write under writeWait and maps its
// refusal.
func writeMember(ctx context.Context, write func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	if err := write(ctx); err != nil {
		return refusal(err)
	}
	return nil
}

// wireMember is m as the admin operations carry it.
func wireMember(m auth.Member) ops.Member {
	return ops.Member{
		UserID:      hex.EncodeToString(m.ID[:]),
		Address:     m.Address,
		DisplayName: m.DisplayName,
		Username:    m.Username,
		Admin:       m.Admin,
	}
}
