package service

import (
	"context"

	"github.com/usesend0/send0/internal/model"
)

type OrganizationService interface {
	Create(ctx context.Context, organization *model.Organization) error
	List(ctx context.Context) ([]*model.Organization, int, error)
}

type organizationService struct {
	*baseService
}

func NewOrganizationService(baseService *baseService) OrganizationService {
	return &organizationService{
		baseService,
	}
}

func (s *organizationService) Create(ctx context.Context, organization *model.Organization) error {
	return s.repository.Organization.Create(ctx, organization)
}

func (s *organizationService) List(ctx context.Context) ([]*model.Organization, int, error) {
	return s.repository.Organization.FindAll(ctx)
}
