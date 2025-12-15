package db

import (
	"context"
	"database/sql"
)

var (
	_ Querier = &QuerierProxier{}
)

type QuerierProxier struct {
	Querier
	OverwrittenSystemCheckGrant                   func(ctx context.Context, arg SystemCheckGrantParams) (int32, error)
	OverwrittenGetPermissionsByUserID             func(ctx context.Context, usersIdusers int32) ([]*GetPermissionsByUserIDRow, error)
	OverwrittenSystemCheckRoleGrant               func(ctx context.Context, arg SystemCheckRoleGrantParams) (int32, error)
	OverwrittenSystemGetUserByID                  func(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error)
	OverwrittenListContentPublicLabels            func(ctx context.Context, arg ListContentPublicLabelsParams) ([]*ListContentPublicLabelsRow, error)
	OverwrittenListContentPrivateLabels           func(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error)
	OverwrittenListPrivateTopicParticipantsByTopicIDForUser func(ctx context.Context, arg ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error)
	OverwrittenListPrivateTopicsByUserID          func(ctx context.Context, userID sql.NullInt32) ([]*ListPrivateTopicsByUserIDRow, error)
}

func (q *QuerierProxier) SystemCheckGrant(ctx context.Context, arg SystemCheckGrantParams) (int32, error) {
	if q.OverwrittenSystemCheckGrant == nil {
		panic("SystemCheckGrant not implemented")
	}
	return q.OverwrittenSystemCheckGrant(ctx, arg)
}

func (q *QuerierProxier) GetPermissionsByUserID(ctx context.Context, usersIdusers int32) ([]*GetPermissionsByUserIDRow, error) {
	if q.OverwrittenGetPermissionsByUserID == nil {
		panic("GetPermissionsByUserID not implemented")
	}
	return q.OverwrittenGetPermissionsByUserID(ctx, usersIdusers)
}

func (q *QuerierProxier) SystemCheckRoleGrant(ctx context.Context, arg SystemCheckRoleGrantParams) (int32, error) {
	if q.OverwrittenSystemCheckRoleGrant == nil {
		panic("SystemCheckRoleGrant not implemented")
	}
	return q.OverwrittenSystemCheckRoleGrant(ctx, arg)
}

func (q *QuerierProxier) SystemGetUserByID(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error) {
	if q.OverwrittenSystemGetUserByID == nil {
		panic("SystemGetUserByID not implemented")
	}
	return q.OverwrittenSystemGetUserByID(ctx, idusers)
}

func (q *QuerierProxier) ListContentPublicLabels(ctx context.Context, arg ListContentPublicLabelsParams) ([]*ListContentPublicLabelsRow, error) {
	if q.OverwrittenListContentPublicLabels == nil {
		panic("ListContentPublicLabels not implemented")
	}
	return q.OverwrittenListContentPublicLabels(ctx, arg)
}

func (q *QuerierProxier) ListContentPrivateLabels(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error) {
	if q.OverwrittenListContentPrivateLabels == nil {
		panic("ListContentPrivateLabels not implemented")
	}
	return q.OverwrittenListContentPrivateLabels(ctx, arg)
}

func (q *QuerierProxier) ListPrivateTopicParticipantsByTopicIDForUser(ctx context.Context, arg ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
	if q.OverwrittenListPrivateTopicParticipantsByTopicIDForUser == nil {
		panic("ListPrivateTopicParticipantsByTopicIDForUser not implemented")
	}
	return q.OverwrittenListPrivateTopicParticipantsByTopicIDForUser(ctx, arg)
}

func (q *QuerierProxier) ListPrivateTopicsByUserID(ctx context.Context, userID sql.NullInt32) ([]*ListPrivateTopicsByUserIDRow, error) {
	if q.OverwrittenListPrivateTopicsByUserID == nil {
		panic("ListPrivateTopicsByUserID not implemented")
	}
	return q.OverwrittenListPrivateTopicsByUserID(ctx, userID)
}
