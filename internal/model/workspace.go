package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/usesend0/send0/internal/uid"
)

var _ sql.Scanner = (*Address)(nil)
var _ driver.Valuer = (*Address)(nil)

type WorkspaceRepository interface {
	Save(ctx context.Context, workspace *Workspace) error
	FindById(ctx context.Context, id uid.UID) (*Workspace, error)
	GetUserWorkspaces(ctx context.Context, userId uid.UID) ([]*Workspace, error)
}
type Address struct {
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
	Street  string `json:"street,omitempty"`
	Zip     string `json:"zip,omitempty"`
}

type Workspace struct {
	Base
	Name    string  `json:"name,omitempty" db:"name"`
	Owner   uid.UID `json:"owner,omitempty" db:"owner"`
	Address Address `json:"address,omitempty" db:"address" gorm:"type:jsonb;not null;default:'{}'"`
}

type Environment struct {
	Base
	Name        string  `json:"name,omitempty" db:"name"`
	WorkspaceId uid.UID `json:"workspace_id,omitempty" db:"workspace_id"`
}

type workspaceRepository struct {
	*baseRepository
}

func NewWorkspaceRepository(baseRepository *baseRepository) WorkspaceRepository {
	return &workspaceRepository{
		baseRepository,
	}
}

func (r *workspaceRepository) Save(ctx context.Context, workspace *Workspace) error {
	stmt := `INSERT INTO workspaces (
		id, 
		name, 
		owner, 
		address
	) VALUES ($1, $2, $3, $4)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		r.UID(workspace.Id),
		workspace.Name,
		workspace.Owner,
		workspace.Address,
	)

	return err
}

func (r *workspaceRepository) FindById(ctx context.Context, id uid.UID) (*Workspace, error) {
	var workspace Workspace
	stmt := `SELECT id FROM workspaces WHERE id = $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(&workspace.Id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (r *workspaceRepository) GetUserWorkspaces(ctx context.Context, userId uid.UID) ([]*Workspace, error) {
	workspaces := make([]*Workspace, 0)
	stmt := `SELECT 
		id,
		name,
		owner,
		address 
	FROM workspaces WHERE id IN (
		SELECT workspace_id FROM team_users WHERE user_id = $1
	)`
	rows, err := r.DB.Connection().Query(ctx, stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var workspace Workspace
		err = rows.Scan(&workspace)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &workspace)
	}

	return workspaces, nil
}

func (a *Address) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	default:
		return errors.New("invalid address type")
	}
}

func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}
