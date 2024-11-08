package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/uid"
)

var EventTypeCreateQuery = fmt.Sprintf(
	`CREATE TYPE %s AS ENUM ('%s','%s','%s','%s','%s','%s','%s','%s');`,
	DBTypeEventType,
	constant.EventTypeEmailSend,
	constant.EventTypeEmailDelivered,
	constant.EventTypeEmailOpened,
	constant.EventTypeEmailClicked,
	constant.EventTypeEmailBounced,
	constant.EventTypeEmailUnsubsribed,
	constant.EventTypeEmailReported,
	constant.EventTypeEmailRejected,
)

var _ sql.Scanner = (*EventMetaData)(nil)
var _ driver.Valuer = (*EventMetaData)(nil)

type EventRepository interface {
	Save(ctx context.Context, event *Event) error
}

type EventMetaData map[string]interface{}
type Event struct {
	Base
	EventType      constant.EventType `json:"eventType" db:"event_type" gorm:"type:event_type;not null"`
	Receipients    JSONBArray         `json:"receipients" gorm:"type:jsonb;not null;default '[]'"`
	CCRecipients   JSONBArray         `json:"ccRecipients" gorm:"type:jsonb;not null;default '[]'"`
	BCCRecipients  JSONBArray         `json:"bccRecipients" gorm:"type:jsonb;not null;default '[]'"`
	MetaData       EventMetaData      `json:"metaData" gorm:"type:jsonb;not null;default '{}'"`
	OrganizationId uid.UID            `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID            `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type eventRepository struct {
	*baseRepository
}

func NewEventRepository(baseRepository *baseRepository) EventRepository {
	return &eventRepository{
		baseRepository,
	}
}

func (r *eventRepository) Save(ctx context.Context, event *Event) error {
	stmt, args, err := r.DB.Builder().Insert(string(TableNameEvent)).Columns(
		"id",
		"event_type",
		"receipients",
		"meta_data",
		"organization_id",
		"workspace_id",
	).Values(
		r.UID(event.Id),
		event.EventType,
		event.Receipients,
		event.MetaData,
		event.OrganizationId,
		event.WorkspaceId,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)

	return err
}

func (a *EventMetaData) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	default:
		return errors.New("invalid type")
	}
}

func (a EventMetaData) Value() (driver.Value, error) {
	return json.Marshal(a)
}
