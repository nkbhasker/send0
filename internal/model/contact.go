package model

import "github.com/usesend0/send0/internal/uid"

type Contact struct {
	Base
	FirstName     string                 `json:"firstName" db:"first_name"`
	LastName      string                 `json:"lastLame" db:"last_name"`
	Email         string                 `json:"email" db:"email" gorm:"not null;unique"`
	EmailVerified bool                   `json:"emailVerified" db:"email_verified" gorm:"not null;default false"`
	Attributes    map[string]interface{} `json:"attributes" db:"attributes" gorm:"type:jsonb;not null;default '{}'"`
	Tags          JSONBArray             `json:"tags" db:"tags" gorm:"type:jsonb;not null;default '[]'"`
	Unsubscribed  bool                   `json:"unsubscribed" db:"unsubscribed" gorm:"not null;default false"`
	WorkspaceId   uid.UID                `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
