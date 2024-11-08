package model

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/uid"
)

const (
	WebhookStatusActive   WebhookStatus = "ACTIVE"
	WebhookStatusInactive WebhookStatus = "INACTIVE"
)

const (
	WebhookLogStatusSuccess WebhookLogStatus = "SUCCESS"
	WebhookLogStatusFailure WebhookLogStatus = "FAILURE"
	WebhookLogStatusPending WebhookLogStatus = "PENDING"
)

var _ sql.Scanner = (*WebhookPayload)(nil)
var _ driver.Valuer = (*WebhookPayload)(nil)

type WebhookRepository interface {
	Save(ctx context.Context, webhook *Webhook) error
	FindById(ctx context.Context, id uid.UID) (*Webhook, error)
	FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*Webhook, error)
	FindByEventType(ctx context.Context, eventType constant.EventType) (*Webhook, error)
	SaveLog(ctx context.Context, log *WebhookEvent) error
}

type WebhookStatus string
type WebhookEventType string
type WebhookLogStatus string

type WebhookPayload string

type Webhook struct {
	Base
	Status           WebhookStatus  `json:"status" db:"status" gorm:"not null"`
	URL              string         `json:"url" db:"url" gorm:"not null"`
	SigningKey       JSONPrivateKey `json:"-" db:"signing_key" gorm:"type:jsonb;not null;default '{}'"`
	SigningKeyPublic string         `json:"signingKeyPublic" db:"-" gorm:"-:all"`
	Events           JSONBArray     `json:"events" db:"events" gorm:"type:jsonb;not null;default '[]'"`
	WorkspaceId      uid.UID        `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type WehbookRequest struct {
	Base
	Signature      string `json:"signature"`
	SendAt         string `json:"sendAt" db:"send_at" gorm:"type:timestamp with time zone;not null;default:now()"`
	ResponseStatus int    `json:"responseStatus"`
	ResponseBody   string `json:"responseBody"`
}

type JSONBArrayWebhookRequests []WehbookRequest

type WebhookEvent struct {
	Base
	Status      WebhookLogStatus          `json:"status" db:"status" gorm:"not null"`
	EventType   WebhookEventType          `json:"eventType" db:"event_type" gorm:"not null"`
	Payload     string                    `json:"payload" db:"payload" gorm:"type:jsonb;not null;default '{}'"`
	Retries     int                       `json:"retries" db:"retries" gorm:"not null;default:0"`
	NextSendAt  string                    `json:"nextSendAt" db:"next_send_at" gorm:"type:timestamp with time zone;not null;default:now()"`
	Requests    JSONBArrayWebhookRequests `json:"requests" db:"requests" gorm:"type:jsonb;not null;default '[]'"`
	WebhookId   uid.UID                   `json:"webhookId" db:"webhook_id" gorm:"not null"`
	WorkspaceId uid.UID                   `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type webhookRepository struct {
	*baseRepository
}

func NewWebhookRepository(baseRepository *baseRepository) WebhookRepository {
	return &webhookRepository{baseRepository}
}

func (r *webhookRepository) Save(ctx context.Context, webhook *Webhook) error {
	stmt := `INSERT INTO webhooks (
		id,
		status,
		url,
		events,
		signing_key,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		webhook.Id,
		webhook.Status,
		webhook.URL,
		webhook.Events,
		webhook.SigningKey,
		webhook.WorkspaceId,
	)

	return err
}

func (r *webhookRepository) SaveLog(ctx context.Context, log *WebhookEvent) error {
	stmt := `INSERT INTO webhook_logs (
		id,
		status,
		event_type,
		payload,
		retries,
		next_send_at,
		requests,
		webhook_id,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		log.Id,
		log.Status,
		log.EventType,
		log.Payload,
		log.Retries,
		log.NextSendAt,
		log.Requests,
		log.WebhookId,
		log.WorkspaceId,
	)

	return err
}

func (r *webhookRepository) FindById(ctx context.Context, id uid.UID) (*Webhook, error) {
	var webhook Webhook
	stmt := `SELECT * FROM webhooks WHERE id = $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(
		&webhook.Id,
		&webhook.Status,
		&webhook.URL,
		&webhook.Events,
		&webhook.SigningKey,
		&webhook.WorkspaceId,
	)
	if err != nil {
		return nil, err
	}
	// Encode the public key for the response
	webhook.SigningKeyPublic, err = crypto.PublicKeyToEncoded(&webhook.SigningKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func (r *webhookRepository) FindByEventType(ctx context.Context, eventType constant.EventType) (*Webhook, error) {
	var webhook Webhook
	stmt := `SELECT * FROM webhooks WHERE events @> $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, eventType).Scan(
		&webhook.Id,
		&webhook.Status,
		&webhook.URL,
		&webhook.Events,
		&webhook.SigningKey,
		&webhook.WorkspaceId,
	)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func (r *webhookRepository) FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*Webhook, error) {
	var webhooks []*Webhook
	stmt := `SELECT * FROM webhooks WHERE workspace_id = $1`
	row, err := r.DB.Connection().Query(ctx, stmt, workspaceId)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		var webhook Webhook
		err := row.Scan(&webhook)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

func (w *WebhookPayload) Scan(value interface{}) error {
	return scanJSONB(w, value)
}

func (w WebhookPayload) Value() (driver.Value, error) {
	return valueJSONB(w)
}
