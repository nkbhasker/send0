package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

type ApiLogInterface interface {
	Create(ctx context.Context, apiLog *ApiLog) error
	FindById(ctx context.Context, id uid.UID) (*ApiLog, error)
	FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*ApiLog, error)
}

type ApiLog struct {
	Base
	Endpoint       string  `json:"endpoint"`
	Method         string  `json:"method"`
	Payload        string  `json:"payload"`
	Status         string  `json:"status"`
	Response       string  `json:"response"`
	Subject        uid.UID `json:"subjectId" db:"subject" gorm:"not null"`
	OrganizationId uid.UID `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type apiLogRepository struct {
	*baseRepository
}

func NewApiLogRepository(baseRepository *baseRepository) ApiLogInterface {
	return &apiLogRepository{baseRepository}
}

func (r *apiLogRepository) Create(ctx context.Context, apiLog *ApiLog) error {
	stmt := `INSERT INTO api_logs (
		id,
		endpoint,
		method,
		payload,
		status,
		response,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		apiLog.Id,
		apiLog.Endpoint,
		apiLog.Method,
		apiLog.Payload,
		apiLog.Status,
		apiLog.Response,
		apiLog.WorkspaceId,
	)

	return err
}

func (r *apiLogRepository) FindById(ctx context.Context, id uid.UID) (*ApiLog, error) {
	var apiLog ApiLog
	err := r.DB.Connection().QueryRow(ctx, `SELECT * FROM api_logs WHERE id = $1`, id).Scan(
		&apiLog.Id,
		&apiLog.Endpoint,
		&apiLog.Method,
		&apiLog.Payload,
		&apiLog.Status,
		&apiLog.Response,
		&apiLog.WorkspaceId,
	)
	if err != nil {
		return nil, err
	}

	return &apiLog, nil
}

func (r *apiLogRepository) FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*ApiLog, error) {
	apiLogs := make([]*ApiLog, 0)
	stmt := `SELECT * FROM api_logs WHERE workspace_id = $1`
	rows, err := r.DB.Connection().Query(ctx, stmt, workspaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var apiLog ApiLog
		err = rows.Scan(&apiLog.Id, &apiLog.Endpoint, &apiLog.Method, &apiLog.Payload, &apiLog.Status, &apiLog.Response, &apiLog.WorkspaceId)
		if err != nil {
			return nil, err
		}
		apiLogs = append(apiLogs, &apiLog)
	}

	return apiLogs, nil
}
