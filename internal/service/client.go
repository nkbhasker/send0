package service

import (
	"context"

	"github.com/usesend0/send0/internal/model"
)

type ClientService interface {
	Create(ctx context.Context, client *model.Client) error
}

type clientService struct {
	*baseService
}

func NewClientService(baseService *baseService) ClientService {
	return &clientService{
		baseService,
	}
}

func (s *clientService) Create(ctx context.Context, client *model.Client) error {
	return s.repository.Client.Create(ctx, client)
}
