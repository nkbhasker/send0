package service

import (
	"context"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

type WorkspaceService interface {
	Create(ctx context.Context, workspace *model.Workspace) error
	CreateTeam(ctx context.Context, team *model.Team) error
	CreateTeamUser(ctx context.Context, teamUser *model.TeamUser) error
	Seed(ctx context.Context) error
}

type workspaceService struct {
	*baseService
	organizationService OrganizationService
}

func NewWorkspaceService(
	baseService *baseService,
	organizationService OrganizationService,
) WorkspaceService {
	return &workspaceService{
		baseService,
		organizationService,
	}
}

func (s *workspaceService) Create(ctx context.Context, workspace *model.Workspace) error {
	return nil
}

func (s *workspaceService) Seed(ctx context.Context) error {
	workspaceId := uid.NewUID(int64(s.config.WorkspaceId))
	organizationId := uid.NewUID(int64(s.config.OrganizationId))
	workspace, err := s.repository.Workspace.FindById(ctx, *workspaceId)
	if err != nil {
		return err
	}
	if workspace != nil {
		return nil
	}
	user := &model.User{
		Email:            s.config.AdminEmail,
		IdentityProvider: model.IdentityProviderLocal,
	}
	workspace = &model.Workspace{
		Base: model.Base{
			Id: *workspaceId,
		},
		Name:  "Main",
		Owner: user.Id,
	}
	organization := &model.Organization{
		Base: model.Base{
			Id: *organizationId,
		},
		Name:      "Main",
		IsDefault: true,
		Subdomain: constant.AppName,
	}
	team := &model.Team{
		Name:           "Admins",
		WorkspaceId:    *workspaceId,
		OrganizationId: *organizationId,
		IsActive:       true,
	}
	err = s.Transact(ctx, func(ctx context.Context, service *Service) error {
		err = s.repository.Workspace.Save(ctx, workspace)
		if err != nil {
			return err
		}
		err = s.repository.User.Save(ctx, user)
		if err != nil {
			return err
		}
		err = s.CreateTeam(ctx, team)
		if err != nil {
			return err
		}
		err = s.repository.Team.SaveTeamUser(ctx, &model.TeamUser{})
		if err != nil {
			return err
		}

		return s.organizationService.Create(
			ctx,
			organization,
		)
	})

	return err
}

func (s *workspaceService) CreateTeam(ctx context.Context, team *model.Team) error {
	return s.repository.Team.Save(ctx, team)
}

func (s *workspaceService) CreateTeamUser(ctx context.Context, teamUser *model.TeamUser) error {
	return s.repository.Team.SaveTeamUser(ctx, teamUser)
}
