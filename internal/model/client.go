package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

type Client struct {
	Base
	Description string     `json:"description" gorm:"type:varchar(255);"`
	Secret      string     `json:"secret" gorm:"type:varchar(255);not null"`
	LastUsedAt  *string    `json:"lastUsedAt" gorm:"type:timestamp with time zone" db:"last_used_at"`
	Permissions JSONBArray `json:"permissions" gorm:"type:jsonb;not null;default:'[]'" db:"permissions"`
	WorkspaceId uid.UID    `json:"workspaceId" gorm:"not null" db:"workspace_id"`
}

type ClientRepository interface {
	Create(ctx context.Context, client *Client) error
	FindByID(ctx context.Context, id uid.UID) (*Client, error)
	Delete(ctx context.Context, id uid.UID) error
}

type clientRepository struct {
	*baseRepository
}

func NewClientRepository(baseRepository *baseRepository) ClientRepository {
	return &clientRepository{
		baseRepository,
	}
}

func (r *clientRepository) Create(ctx context.Context, client *Client) error {
	stmt := `INSERT INTO clients (
		id,
		description,
		secret,
		last_used_at,
		permissions,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		client.Id,
		client.Description,
		client.Secret,
		client.LastUsedAt,
		client.Permissions,
		client.WorkspaceId,
	)

	return err
}

func (r *clientRepository) FindByID(ctx context.Context, id uid.UID) (*Client, error) {
	stmt := `SELECT
		id,
		description,
		secret,
		last_used_at,
		permissions,
		workspace_id
	FROM clients WHERE id = $1`

	client := new(Client)
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *clientRepository) FindByClientId(ctx context.Context, clientId string) (*Client, error) {
	return nil, nil
}

func (r *clientRepository) Delete(ctx context.Context, id uid.UID) error {
	return nil
}
