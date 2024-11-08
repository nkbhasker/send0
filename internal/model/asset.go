package model

import "github.com/usesend0/send0/internal/uid"

type Asset struct {
	Base
	Name        string     `json:"name" db:"name"`
	MimeType    string     `json:"mimeType" db:"mime_type"`
	Slug        string     `json:"slug" db:"slug" gorm:"unique_index"`
	Tags        JSONBArray `json:"tags" db:"tags" gorm:"type:jsonb;not null;default '[]'"`
	WorkspaceId uid.UID    `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
