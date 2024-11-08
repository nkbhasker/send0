package model

import (
	"fmt"

	"github.com/usesend0/send0/internal/uid"
)

type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "DRAFT"
	BroadcastStatusScheduled BroadcastStatus = "SCHEDULED"
	BroadcastStatusRunning   BroadcastStatus = "RUNNING"
	BroadcastStatusCompleted BroadcastStatus = "COMPLETED"
	BroadcastStatusCanceled  BroadcastStatus = "CANCELED"
)

var BroadcastStatusTypeCreateQuery = fmt.Sprintf(
	`CREATE TYPE %s AS ENUM ('%s','%s','%s','%s','%s');`,
	DBTypeCampaignStatus,
	BroadcastStatusDraft,
	BroadcastStatusScheduled,
	BroadcastStatusRunning,
	BroadcastStatusCompleted,
	BroadcastStatusCanceled,
)

type Broadcast struct {
	Base
	Name           string          `json:"name"`
	From           string          `json:"from" db:"from"`
	ReplyTo        *string         `json:"replyTo" db:"reply_to"`
	TemplateId     int             `json:"templateId" db:"template_id" gorm:"not null"`
	Delay          int             `json:"delay"`
	DelayTimeZone  string          `json:"delayTimeZone" db:"delay_time_zone"`
	Events         JSONBArray      `json:"events" db:"events" gorm:"type:jsonb;not null;default '[]'"` // Events to track
	Status         BroadcastStatus `json:"status" db:"status" gorm:"type:campaign_status;not null;default:'DRAFT'"`
	Segments       JSONBArray      `json:"segments" db:"segments" gorm:"type:jsonb;not null;default '[]'"`
	OpenTracking   bool            `json:"openTracking" db:"open_tracking"`
	ClickTracking  bool            `json:"clickTracking" db:"click_tracking"`
	CCAddresses    JSONBArray      `json:"ccAddresses" db:"cc_addresses" gorm:"type:jsonb;not null;default:'[]'"`
	BCCAddresses   JSONBArray      `json:"bccAddresses" db:"bcc_addresses" gorm:"type:jsonb;not null;default:'[]'"`
	OrganizationId uid.UID         `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID         `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type BroadcastStat struct {
	Base
	BroadcastId  uid.UID `json:"broadcastId" db:"broadcast_id" gorm:"not null"`
	Total        int     `json:"total"`
	Sent         int     `json:"sent"`
	Opened       int     `json:"opened"`
	Clicked      int     `json:"clicked"`
	Unsubscribed int     `json:"unsubscribed"`
	Complained   int     `json:"complained"`
	Bounced      int     `json:"bounced"`
	Delivered    int     `json:"delivered"`
	WorkspaceId  uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
