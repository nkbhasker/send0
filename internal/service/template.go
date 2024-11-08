package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

type TemplateService interface {
	Create(ctx context.Context, template *model.Template) error
	CreateOptInTemplate(ctx context.Context, workspaceId, organizationId uid.UID) error
}

type templateService struct {
	*baseService
}

func NewTemplateService(baseService *baseService) TemplateService {
	return &templateService{baseService}
}

func (s *templateService) Create(ctx context.Context, template *model.Template) error {
	if template.IsOptIn {
		ok := strings.Contains(template.Content, string(constant.VariableOptInLink))
		if !ok {
			return errors.New("Opt-in template must contain the opt-in link variable")
		}
	}
	return s.repository.Template.Save(ctx, template)
}

func (s *templateService) CreateOptInTemplate(ctx context.Context, workspaceId, organizationId uid.UID) error {
	content := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Confirm your subscription</title>
			</head>
			<body>
				<h1>Confirm your subscription</h1>
				<p>Click the link below to confirm your subscription</p>
				<a href="{{.%s}}">Confirm Subscription</a>
			</body>
		</html>
	`, constant.VariableOptInLink)

	template := &model.Template{
		Name:            "Opt-In Confirmation",
		ContentEngine:   "HTML",
		IsTransactional: true,
		Subject:         `{{.Name}}, Confirm your subscription`,
		Content:         content,
		IsOptIn:         true,
		OrganizationId:  organizationId,
		WorkspaceId:     workspaceId,
	}

	return s.Create(ctx, template)
}
