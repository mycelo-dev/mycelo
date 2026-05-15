package account

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
)

const sessionTTL = 30 * 24 * time.Hour

// SignUpRepository creates a tenant and user atomically.
func SignUpRepository(ctx context.Context, tenantName string, userName string, email string, passwordHash string) (SignUpResponse, error) {
	tx, err := core.Get().Begin(ctx)
	if err != nil {
		fmt.Println("error starting signup transaction: ", err)
		return SignUpResponse{}, err
	}
	defer tx.Rollback(ctx)

	createdAt := time.Now().UnixMilli()
	updatedAt := createdAt

	var tenantId int64
	var tenantPublicId string
	if err := tx.QueryRow(
		ctx,
		insert_queries.GetInsertTenantQuery(),
		tenantName,
		createdAt,
		updatedAt,
	).Scan(&tenantId, &tenantPublicId); err != nil {
		fmt.Println("error creating tenant: ", err)
		return SignUpResponse{}, err
	}

	var userPublicId string
	if err := tx.QueryRow(
		ctx,
		insert_queries.GetInsertUserQuery(),
		tenantId,
		userName,
		email,
		passwordHash,
		createdAt,
		updatedAt,
	).Scan(&userPublicId); err != nil {
		fmt.Println("error creating user: ", err)
		return SignUpResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Println("error committing signup transaction: ", err)
		return SignUpResponse{}, err
	}

	return SignUpResponse{
		TenantPublicId: tenantPublicId,
		UserPublicId:   userPublicId,
		TenantName:     tenantName,
		UserName:       userName,
		Email:          email,
	}, nil
}

// LoginRepository loads account context and password hash from an email address.
func LoginRepository(ctx context.Context, email string) (loginRecord, error) {
	var record loginRecord
	err := core.Get().QueryRow(ctx, select_queries.GetAccountByEmailQuery(), email).Scan(
		&record.Account.TenantPublicId,
		&record.Account.UserPublicId,
		&record.Account.TenantName,
		&record.Account.UserName,
		&record.Account.Email,
		&record.PasswordHash,
	)

	if err != nil {
		fmt.Println("error logging in account: ", err)
		return loginRecord{}, err
	}

	return record, nil
}

// StoreSessionRepository stores a hashed operator-console session token.
func StoreSessionRepository(ctx context.Context, tenantPublicId string, userPublicId string, sessionHash string) error {
	createdAt := time.Now().UnixMilli()
	expiresAt := time.Now().Add(sessionTTL).UnixMilli()

	_, err := core.Get().Exec(
		ctx,
		insert_queries.GetInsertSessionQuery(),
		tenantPublicId,
		userPublicId,
		sessionHash,
		createdAt,
		expiresAt,
	)
	if err != nil {
		fmt.Println("error storing session: ", err)
		return err
	}

	return nil
}

// ReadSessionRepository returns account scope for an unexpired session hash.
func ReadSessionRepository(ctx context.Context, sessionHash string) (SessionContext, error) {
	var session SessionContext
	err := core.Get().QueryRow(
		ctx,
		select_queries.GetSessionContextQuery(),
		sessionHash,
		time.Now().UnixMilli(),
	).Scan(&session.TenantPublicId, &session.UserPublicId)
	if err != nil {
		fmt.Println("error reading session: ", err)
		return SessionContext{}, err
	}

	return session, nil
}

// CreateTeamRepository creates a team for a tenant user.
func CreateTeamRepository(ctx context.Context, tenantPublicId string, userPublicId string, teamName string) (TeamRecord, error) {
	createdAt := time.Now().UnixMilli()
	updatedAt := createdAt

	var team TeamRecord
	err := core.Get().QueryRow(
		ctx,
		insert_queries.GetInsertTeamForTenantUserQuery(),
		tenantPublicId,
		userPublicId,
		teamName,
		createdAt,
		updatedAt,
	).Scan(&team.TeamPublicId, &team.TeamName)

	if err != nil {
		fmt.Println("error creating team: ", err)
		return TeamRecord{}, err
	}

	return team, nil
}

// ListTeamsRepository lists teams for a tenant user.
func ListTeamsRepository(ctx context.Context, tenantPublicId string, userPublicId string) ([]TeamRecord, error) {
	rows, err := core.Get().Query(ctx, select_queries.GetTeamsForTenantUserQuery(), tenantPublicId, userPublicId)
	if err != nil {
		fmt.Println("error reading teams: ", err)
		return nil, err
	}
	defer rows.Close()

	teams := make([]TeamRecord, 0)
	for rows.Next() {
		var team TeamRecord
		if err := rows.Scan(&team.TeamPublicId, &team.TeamName); err != nil {
			fmt.Println("error scanning team: ", err)
			return nil, err
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("error reading team rows: ", err)
		return nil, err
	}

	return teams, nil
}
