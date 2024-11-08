package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

const (
	ContentEngineHtml       = "HTML"
	CotentEngineText        = "TEXT"
	ContentEngineHandlebars = "HANDLEBARS"
	ContentEngineMarkdown   = "MARKDOWN"
	ContentEngineLiquid     = "LIQUID"
	ContentEnginePug        = "PUG"
	ContentEngineMustache   = "MUSTACHE"
)

type TemplateRepository interface {
	Save(ctx context.Context, template *Template) error
}

type ContentEngine string

type Template struct {
	Base
	Name            string            `json:"name"`
	ContentEngine   ContentEngine     `json:"contentEngine" db:"content_engine" gorm:"not null"`
	Headers         map[string]string `json:"headers"`
	Subject         string            `json:"subject"`
	Content         string            `json:"content"`
	TextContent     *string           `json:"textContent" db:"text_content"`
	ParsedContent   string            `json:"parsedContent" db:"parsed_content"`
	IsOptIn         bool              `json:"isOptIn" db:"is_opt_in" gorm:"not null;default:false"`
	IsTransactional bool              `json:"isTransactional" db:"is_transactional" gorm:"not null;default:true"`
	OrganizationId  uid.UID           `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId     uid.UID           `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type templateRepository struct {
	*baseRepository
}

func NewTemplateRepository(baseRepository *baseRepository) TemplateRepository {
	return &templateRepository{
		baseRepository,
	}
}

func (r *templateRepository) Save(ctx context.Context, template *Template) error {
	stmt := `INSERT INTO templates (
		id,
		name,
		content_engine,
		is_transactional,
		alt_subject,
		content,
		text_content,
		is_opt_in,
		organization_id,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		r.UID(template.Id),
		template.Name,
		template.ContentEngine,
		template.IsTransactional,
		template.Subject,
		template.Content,
		template.TextContent,
		template.IsOptIn,
		template.OrganizationId,
		template.WorkspaceId,
	)

	return err
}
