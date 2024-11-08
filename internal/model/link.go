package model

import "github.com/usesend0/send0/internal/uid"

type Link struct {
	Base
	URL         string  `json:"url" db:"url" gorm:"not null"`
	WorkspaceId uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
