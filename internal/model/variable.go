package model

import "github.com/usesend0/send0/internal/uid"

const (
	VariableTypeString   VariableType = "STRING"
	VariableTypeLink     VariableType = "LINK"
	VariableTypeInteger  VariableType = "INTEGER"
	VariableTypeFloat    VariableType = "FLOAT"
	VariableTypeBoolean  VariableType = "BOOLEAN"
	VariableTypeDateTime VariableType = "DATETIME"
	VariableTypeArray    VariableType = "ARRAY"
	VariableTypeObject   VariableType = "OBJECT"
)

type VariableType string

type Variable struct {
	Base
	Name           string       `json:"name"`
	Type           VariableType `json:"type" db:"type" gorm:"type:variable_type"`
	DefaultValue   string       `json:"defaultValue" db:"default_value" gorm:"not null"`
	IsSocial       bool         `json:"isSocial" db:"is_social"`
	OrganizationId uid.UID      `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID      `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}
