package service

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

var _ model.Transactioner[*Service] = (*Service)(nil)

type Service struct {
	*baseService
	Client       ClientService
	Domain       DomainService
	Email        EmailService
	Organization OrganizationService
	Workspace    WorkspaceService
	SNS          SNSService
	SES          SESService
}

type baseService struct {
	config       *config.Config
	repository   *model.Repository
	logger       *zerolog.Logger
	uidGenerator uid.UIDGenerator
}

func NewBaseService(config *config.Config, uidGenerator uid.UIDGenerator, logger *zerolog.Logger, repository *model.Repository) *baseService {
	return &baseService{
		config:       config,
		repository:   repository,
		logger:       logger,
		uidGenerator: uidGenerator,
	}
}

func NewService(baseService *baseService) (*Service, error) {
	orgaznizationService := NewOrganizationService(baseService)
	workspcaeService := NewWorkspaceService(baseService, orgaznizationService)
	eventService := NewEventService(baseService)
	snsService, err := NewSNSService(baseService, eventService)
	if err != nil {
		return nil, err
	}
	sesService, err := NewSESService(baseService, snsService)
	if err != nil {
		return nil, err
	}
	domainService := NewDomainService(baseService, sesService)
	emailService := NewEmailService(baseService, sesService, eventService)

	return &Service{
		baseService:  baseService,
		Domain:       domainService,
		Email:        emailService,
		Organization: orgaznizationService,
		Workspace:    workspcaeService,
		SNS:          snsService,
		SES:          sesService,
	}, nil
}

func (s *baseService) Transact(ctx context.Context, fn func(ctx context.Context, service *Service) error) error {
	return s.repository.Transact(ctx, func(ctx context.Context, repo *model.Repository) error {
		service, err := NewService(NewBaseService(s.config, s.uidGenerator, s.logger, repo))
		if err != nil {
			return err
		}

		return fn(ctx, service)
	})
}

func formatSubdomain(parts []string) string {
	filteredParts := []string{}
	for _, part := range parts {
		if part != "" {
			filteredParts = append(filteredParts, part)
		}
	}

	return strings.Join(filteredParts, ".")
}

func execRetry(fn func() error, limit int) error {
	var err error
	for i := 0; i < limit; i++ {
		err = fn()
		if err == nil {
			break
		}
	}

	return err
}
