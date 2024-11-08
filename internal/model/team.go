package model

import (
	"context"
	"fmt"

	"github.com/usesend0/send0/internal/uid"
)

const (
	TeamUserStatusActive   TeamUserStatus = "ACTIVE"
	TeamUserStatusInactive TeamUserStatus = "INACTIVE"
	TeamUserStatusPending  TeamUserStatus = "PENDING"
)

var TeamUserStatusTypeCreateQuery = fmt.Sprintf(
	`CREATE TYPE %s AS ENUM ('%s','%s','%s');`,
	DBTypeTeamUserStatus,
	TeamUserStatusActive,
	TeamUserStatusInactive,
	TeamUserStatusPending,
)

type TeamRepository interface {
	Save(ctx context.Context, team *Team) error
	SaveTeamUser(ctx context.Context, teamUser *TeamUser) error
	FindByID(ctx context.Context, id uid.UID) (*Team, error)
	FindByWorkspaceID(ctx context.Context, workspaceId uid.UID) ([]*Team, error)
}

type TeamUserStatus string

type Team struct {
	Base
	Name           string     `json:"name" db:"name"`
	IsActive       bool       `json:"isActive" db:"is_active"`
	Permissions    JSONBArray `json:"permissions" db:"permissions" gorm:"type:jsonb;not null;default:'[]'"`
	OrganizationId uid.UID    `json:"organization_id" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID    `json:"workspace_id" db:"workspace_id" gorm:"not null"`
}

type TeamUser struct {
	Base
	Status         TeamUserStatus `json:"status" db:"status" gorm:"not null;type:team_user_status;default:'PENDING'"`
	TeamId         uid.UID        `json:"teamId" db:"team_id" gorm:"not null"`
	UserId         uid.UID        `json:"userId" db:"user_id" gorm:"not null"`
	OrganizationId uid.UID        `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID        `json:"workspaceId" db:"workspace_id" gorm:"not null"`
	LastLoginAt    *string        `json:"lastLoginAt" db:"last_login_at" gorm:"type:timestamp with time zone"`
}

type teamRepository struct {
	*baseRepository
}

func NewTeamRepository(baseRepository *baseRepository) TeamRepository {
	return &teamRepository{
		baseRepository,
	}
}

func (r *teamRepository) Save(ctx context.Context, team *Team) error {
	stmt := `INSERT INTO teams (
		id, 
		name,
		is_active, 
		permissions, 
		organization_id,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Connection().Exec(ctx, stmt, r.UID(team.Id), team.Name, team.IsActive, team.Permissions, team.OrganizationId, team.WorkspaceId)

	return err
}

func (r *teamRepository) FindByID(ctx context.Context, id uid.UID) (*Team, error) {
	stmt := `SELECT * FROM teams WHERE id = $1`
	var team Team
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(&team)

	return &team, err
}

func (r *teamRepository) FindByWorkspaceID(ctx context.Context, workspaceId uid.UID) ([]*Team, error) {
	stmt := `SELECT * FROM teams WHERE workspace_id = $1`
	var teams []*Team
	rows, err := r.DB.Connection().Query(ctx, stmt, workspaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var team Team
		err := rows.Scan(&team)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	return teams, err
}

func (r *teamRepository) SaveTeamUser(ctx context.Context, teamUser *TeamUser) error {
	status := TeamUserStatusPending
	if teamUser.Status != "" {
		status = teamUser.Status
	}
	stmt := `INSERT INTO team_users (
		id,
		status,
		team_id, 
		user_id,
		organization_id,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		r.UID(teamUser.Id),
		status,
		teamUser.TeamId,
		teamUser.UserId,
		teamUser.OrganizationId,
		teamUser.WorkspaceId,
	)

	return err
}
