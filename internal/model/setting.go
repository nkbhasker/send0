package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

type SettingRepository interface {
	Create(ctx context.Context, setting *Setting) error
	FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) (*Setting, error)
}

type Setting struct {
	Base
	IndividualTracking bool    `json:"individualTracking" db:"individual_tracking" gorm:"not null;default:false"`
	OpenTracking       bool    `json:"openTracking" db:"open_tracking" gorm:"not null;default:false"`
	ClickTracking      bool    `json:"clickTracking" db:"click_tracking" gorm:"not null;default:false"`
	WorkspaceId        uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type settingRepository struct {
	*baseRepository
}

func NewSettingRepository(baseRepository *baseRepository) SettingRepository {
	return &settingRepository{baseRepository}
}

func (r *settingRepository) Create(ctx context.Context, setting *Setting) error {
	stmt := `INSERT INTO settings (
		id,
		individual_tracking,
		open_tracking,
		click_tracking,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		r.UID(setting.Id),
		setting.IndividualTracking,
		setting.OpenTracking,
		setting.ClickTracking,
		setting.WorkspaceId,
	)

	return err
}

func (r *settingRepository) FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) (*Setting, error) {
	var setting Setting
	err := r.DB.Connection().QueryRow(ctx, `SELECT * FROM settings WHERE workspace_id = $1`, workspaceId).Scan(&setting)
	if err != nil {
		return nil, err
	}

	return &setting, nil
}
