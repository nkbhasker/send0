package model

import "github.com/usesend0/send0/internal/uid"

type Sender struct {
	Base
	Address        string  `json:"address" db:"address" gorm:"not null"`
	IsDefault      bool    `json:"isDefault" db:"is_default" gorm:"not null"`
	OrganizationId uid.UID `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
