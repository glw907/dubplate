package admin

import (
	"context"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/ops"
)

// ListMembers answers the live roster in the order members were added.
func (s *Service) ListMembers(ctx context.Context, _ identity.Actor, _ ops.ListMembersRequest) (ops.ListMembersResponse, error) {
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
func (s *Service) AddMember(ctx context.Context, actor identity.Actor, req ops.AddMemberRequest) (ops.AddMemberResponse, error) {
	candidate := member.Member{Address: req.Address, DisplayName: req.DisplayName, Username: req.Username}
	var added member.Member
	if err := s.writeMember(ctx, actor, func(ctx context.Context) (err error) {
		added, err = s.cfg.Roster.AddMember(ctx, candidate, actor, s.cfg.Logger)
		return err
	}); err != nil {
		return ops.AddMemberResponse{}, err
	}
	return ops.AddMemberResponse{Member: wireMember(added)}, nil
}

// RemoveMember removes the member req names, which revokes their
// sessions. It is gated.
func (s *Service) RemoveMember(ctx context.Context, actor identity.Actor, req ops.RemoveMemberRequest) (ops.RemoveMemberResponse, error) {
	target, err := memberID(req.UserID)
	if err != nil {
		return ops.RemoveMemberResponse{}, err
	}
	return ops.RemoveMemberResponse{}, s.writeMember(ctx, actor, func(ctx context.Context) error {
		return s.cfg.Roster.RemoveMember(ctx, target, actor, s.cfg.Logger)
	})
}

// DemoteMember clears the admin flag of the member req names. It is
// gated, since the web cannot grant the flag back.
func (s *Service) DemoteMember(ctx context.Context, actor identity.Actor, req ops.DemoteMemberRequest) (ops.DemoteMemberResponse, error) {
	target, err := memberID(req.UserID)
	if err != nil {
		return ops.DemoteMemberResponse{}, err
	}
	return ops.DemoteMemberResponse{}, s.writeMember(ctx, actor, func(ctx context.Context) error {
		return s.cfg.Roster.DemoteMember(ctx, target, actor, s.cfg.Logger)
	})
}

// writeMember runs one roster write for actor: the gate, then write,
// bounded.
func (s *Service) writeMember(ctx context.Context, actor identity.Actor, write func(context.Context) error) error {
	if err := s.requireFresh(ctx, actor); err != nil {
		return err
	}
	return bounded(ctx, write)
}

// wireMember is m as the admin operations carry it.
func wireMember(m member.Member) ops.Member {
	return ops.Member{
		UserID:      id.ID(m.ID),
		Address:     m.Address,
		DisplayName: m.DisplayName,
		Username:    m.Username,
		Admin:       m.Admin,
	}
}
