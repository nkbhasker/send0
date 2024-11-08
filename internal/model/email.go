package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/usesend0/send0/internal/uid"
)

const (
	EmailStatusDelivered EmailStatus = "DELIVERED"
	EmailStatusBounced   EmailStatus = "BOUNCED"
	EmailStatusPending   EmailStatus = "PENDING"
	EmailStatusFailed    EmailStatus = "FAILED"
	EmailStatusOpened    EmailStatus = "OPENED"
	EmailStatusClicked   EmailStatus = "CLICKED"
	EmailStatusRejected  EmailStatus = "REJECTED"
)

var _ sql.Scanner = (*Recipient)(nil)
var _ driver.Valuer = (*Recipient)(nil)

var _ sql.Scanner = (*Attachment)(nil)
var _ driver.Valuer = (*Attachment)(nil)

type EmailRepository interface {
	Save(ctx context.Context, email *Email) error
	FindById(ctx context.Context, id uid.UID) (*Email, error)
	FindByMessageId(ctx context.Context, messageId string) (*Email, error)
}

type EmailStatus string

type Recipient struct {
	Address     string      `json:"address"`
	Status      EmailStatus `json:"status"`
	DeliveredAt string      `json:"deliveredAt"`
	BouncedAt   string      `json:"bouncedAt"`
	OpenedAt    string      `json:"openedAt"`
	ClickedAt   string      `json:"clickedAt"`
	FailedAt    string      `json:"failedAt"`
}

type Recipients []Recipient

type Attachment struct {
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
	Content     string `json:"content"`
}

type Email struct {
	Base
	MessageId      string       `json:"messageId" db:"message_id" gorm:"type:text;not null"`
	From           string       `json:"from" db:"from"`
	ReplyTo        *string      `json:"replyTo" db:"reply_to"`
	Recipients     Recipients   `json:"recipients" db:"recipients" gorm:"type:jsonb;not null;default '[]'"`
	CCRecipients   Recipients   `json:"ccRecipients" db:"cc_recipients" gorm:"type:jsonb;not null;default '[]'"`
	BCCRecipients  Recipients   `json:"bccRecipients" db:"bcc_recipients" gorm:"type:jsonb;not null;default '[]'"`
	Status         string       `json:"status"`
	Delay          int          `json:"delay"`
	DelayTimeZone  string       `json:"delayTimeZone" db:"delay_time_zone"`
	SentAt         string       `json:"sentAt" db:"sent_at" gorm:"type:timestamp with time zone"`
	RequestId      string       `json:"requestId" db:"request_id" gorm:"not null"` // Broadcast Id in case of broadcast email
	OrganizationId uid.UID      `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID      `json:"workspaceId" db:"workspace_id" gorm:"not null"`
	EmailContent   EmailContent `json:"emailContent" db:"-" gorm:"-:all"`
	MetaData       JSONBMap     `json:"metaData" db:"meta_data" gorm:"type:jsonb;not null;default '{}'"`
}

type EmailContent struct {
	Base
	Subject        *string             `json:"subject" db:"subject" gorm:"type:text"`
	Html           *string             `json:"html" db:"html" gorm:"type:text"`
	Text           *string             `json:"text" db:"text" gorm:"type:text"`
	Headers        []map[string]string `json:"headers" db:"headers" gorm:"type:jsonb;default '[]'"`
	Attachments    []Attachment        `json:"attachments" db:"attachments" gorm:"type:jsonb;default '[]'"`
	EmailId        uid.UID             `json:"emailId" db:"email_id" gorm:"not null"`
	OrganizationId uid.UID             `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID             `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type emailRepository struct {
	*baseRepository
}

func NewEmailRepository(baseRepository *baseRepository) EmailRepository {
	return &emailRepository{
		baseRepository,
	}
}

func (r *emailRepository) Save(ctx context.Context, email *Email) error {
	stmt, args, err := r.DB.Builder().Insert(string(TableNameEmail)).Columns(
		"id",
		"message_id",
		"from_address",
		"recipients",
		"cc_recipients",
		"bcc_recipients",
		"status",
		"delay",
		"delay_time_zone",
		"sent_at",
		"organization_id",
		"workspace_id",
	).Values(
		email.Id, // Always expect the id to be set
		email.MessageId,
		email.From,
		email.Recipients,
		email.CCRecipients,
		email.BCCRecipients,
		email.Status,
		email.Delay,
		email.DelayTimeZone,
		email.SentAt,
		email.OrganizationId,
		email.WorkspaceId,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	if err != nil {
		r.Logger.Error().Err(err).Msg("failed to save email")
		return err
	}
	stmt, args, err = r.DB.Builder().Insert(string(TableNameEmailContent)).Columns(
		"id",
		"subject",
		"html",
		"text",
		"attachments",
		"email_id",
		"organization_id",
		"workspace_id",
	).Values(
		r.UID(email.EmailContent.Id),
		email.EmailContent.Subject,
		email.EmailContent.Html,
		email.EmailContent.Text,
		email.EmailContent.Attachments,
		email.EmailContent.EmailId,
		email.EmailContent.OrganizationId,
		email.EmailContent.WorkspaceId,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	if err != nil {
		r.baseRepository.Logger.Error().Err(err).Msg("failed to save email content")
		return err
	}

	return nil
}

func (r *emailRepository) FindById(ctx context.Context, id uid.UID) (*Email, error) {
	var email Email
	var emailContent EmailContent
	stmt, args, err := r.DB.Builder().Select().From(string(TableNameEmail)).Where("id = ?", id).ToSql()
	if err != nil {
		return nil, err
	}
	err = r.DB.Connection().QueryRow(ctx, stmt, args...).Scan(
		&email.Id,
		&email.MessageId,
		&email.From,
		&email.Recipients,
		&email.CCRecipients,
		&email.BCCRecipients,
		&email.Status,
		&email.Delay,
		&email.DelayTimeZone,
		&email.SentAt,
		&email.OrganizationId,
		&email.WorkspaceId,
	)
	if err != nil {
		return nil, err
	}
	// TODO: Use a join query to fetch email content
	stmt, args, err = r.DB.Builder().Select().From(string(TableNameEmailContent)).Where("email_id = ?", email.Id).ToSql()
	if err != nil {
		return nil, err
	}
	err = r.DB.Connection().QueryRow(ctx, stmt, args...).Scan(
		&emailContent.Id,
		&emailContent.Subject,
		&emailContent.Html,
		&emailContent.Text,
		&emailContent.Attachments,
		&emailContent.EmailId,
		&emailContent.OrganizationId,
		&emailContent.WorkspaceId,
	)
	if err != nil {
		return nil, err
	}
	email.EmailContent = emailContent

	return &email, err
}

func (r *emailRepository) FindByMessageId(ctx context.Context, messageId string) (*Email, error) {
	var email Email
	stmt, args, err := r.DB.Builder().Select().From(string(TableNameEmail)).Where("message_id = ?", messageId).ToSql()
	if err != nil {
		return nil, err
	}
	err = r.DB.Connection().QueryRow(ctx, stmt, args...).Scan(
		&email.Id,
		&email.MessageId,
		&email.From,
		&email.Recipients,
		&email.CCRecipients,
		&email.BCCRecipients,
		&email.Status,
		&email.Delay,
		&email.DelayTimeZone,
		&email.SentAt,
		&email.OrganizationId,
		&email.WorkspaceId,
	)

	return &email, err
}

func (a *Recipient) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	default:
		return errors.New("invalid type")
	}
}

func (a Recipient) Value() (driver.Value, error) {
	if a == (Recipient{}) {
		return json.Marshal(Recipient{})
	}

	return json.Marshal(a)
}

func (a *Attachment) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	default:
		return errors.New("invalid type")
	}
}

func (a Attachment) Value() (driver.Value, error) {
	if a == (Attachment{}) {
		return json.Marshal(Attachment{})
	}

	return json.Marshal(a)
}

func (a *Recipients) Addresses() []string {
	var addresses []string
	for _, recipient := range *a {
		addresses = append(addresses, recipient.Address)
	}

	return addresses
}
