package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/service"
	"github.com/usesend0/send0/internal/uid"
)

type createWorkspaceRequestPayload struct {
	Name string `json:"name" validate:"required"`
}

type createTeamRequestPayload struct {
	Name           string  `json:"name" validate:"required"`
	OrganizationId uid.UID `json:"organizationId" validate:"required"`
}

type inviteUserRequestPayload struct {
	Emails         []string `json:"email" validate:"required"`
	TeamId         uid.UID  `json:"teamId" validate:"required"`
	OrganizationId uid.UID  `json:"organizationId" validate:"required"`
}

type workspaceAPI struct {
	app *core.App
}

func NewWorkspaceAPI(app *core.App) *workspaceAPI {
	return &workspaceAPI{
		app: app,
	}
}

func (api *workspaceAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", api.CreateWorkspaceHandler())
		r.Get("/", api.GetWorkspacesHandler())
		r.Get("/{id}", api.GetWorkspacesHandler())
		r.Post("/{id}/invite", api.InviteUserHandler())
		r.Post("/{id}/team", api.CreateTeamHandler())
		r.Post("/{id}/team/{teamId}/users", api.InviteUserHandler())
	}
}

func (api *workspaceAPI) CreateWorkspaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := core.IdentityFromContext(r.Context())
		payload := new(createWorkspaceRequestPayload)
		workspace, err := func() (*model.Workspace, *ApiError) {
			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			workspace := &model.Workspace{
				Base: model.Base{
					Id: *api.app.UIDGenerator.Next(),
				},
				Name:  payload.Name,
				Owner: identity.UserId(),
			}
			err = api.createWorkspace(r.Context(), workspace)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return workspace, nil
		}()

		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success":   true,
			"workspace": workspace,
		})
	}
}

func (api *workspaceAPI) GetWorkspacesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := core.IdentityFromContext(r.Context())
		workspaces, err := api.app.Repository.Workspace.GetUserWorkspaces(r.Context(), identity.UserId())
		if err != nil {
			renderError(w, r, &ApiError{
				StatusCode: http.StatusInternalServerError,
				Error:      err,
			})
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success":    true,
			"workspaces": workspaces,
		})
	}
}

func (api *workspaceAPI) GetWorkspaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspace, err := func() (*model.Workspace, *ApiError) {
			workspaceId, err := uid.NewUIDFromString(chi.URLParam(r, "id"))
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			workspace, err := api.app.Repository.Workspace.FindById(r.Context(), *workspaceId)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			if workspace == nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusNotFound,
				}
			}

			return workspace, nil
		}()

		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success":   true,
			"workspace": workspace,
		})
	}
}

func (api *workspaceAPI) CreateTeamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team, err := func() (*model.Team, *ApiError) {
			workspaceId, err := uid.NewUIDFromString(chi.URLParam(r, "id"))
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			payload := new(createTeamRequestPayload)
			err = json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			team := &model.Team{
				Base: model.Base{
					Id: *api.app.UIDGenerator.Next(),
				},
				Name:        payload.Name,
				WorkspaceId: *workspaceId,
			}
			err = api.app.Repository.Team.Save(r.Context(), team)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return team, nil
		}()

		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"team":    team,
		})
	}
}

func (api *workspaceAPI) InviteUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() *ApiError {
			_, err := uid.NewUIDFromString(chi.URLParam(r, "id"))
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			payload := new(inviteUserRequestPayload)
			err = json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			team, err := api.app.Repository.Team.FindByID(r.Context(), payload.TeamId)
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			if team == nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusNotFound,
				}
			}
			for _, email := range payload.Emails {
				user, err := api.app.Repository.User.FindByEmail(r.Context(), email)
				if err != nil {
					return &ApiError{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}
				}
				if user == nil {
					user = &model.User{
						Base: model.Base{
							Id: *api.app.UIDGenerator.Next(),
						},
						Email:            email,
						IdentityProvider: model.IdentityProviderLocal,
					}
				}
				err = api.app.Repository.Team.SaveTeamUser(r.Context(), &model.TeamUser{
					TeamId:         team.Id,
					UserId:         user.Id,
					OrganizationId: team.OrganizationId,
					WorkspaceId:    team.WorkspaceId,
					Status:         model.TeamUserStatusActive,
				})
				if err != nil {
					return &ApiError{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}
				}
			}

			return nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"message": "user invited successfully.",
		})
	}
}

func (api *workspaceAPI) createWorkspace(ctx context.Context, workspace *model.Workspace) error {
	// tx, _, err := api.app.DB.BeginTx(ctx)
	// if err != nil {
	// 	return nil
	// }
	// err = func() error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// Create workspace

	// }()
	// if err != nil {
	// 	// rollback transaction if error occurs
	// 	_ = tx.Rollback(ctx)
	// 	return err
	// }

	// return tx.Commit(ctx)

	err := api.app.Service.Transact(ctx, func(ctx context.Context, service *service.Service) error {
		err := service.Workspace.Create(
			ctx,
			workspace,
		)
		if err != nil {
			return err
		}
		err = service.Organization.Create(
			ctx,
			&model.Organization{
				Base: model.Base{
					Id: *api.app.UIDGenerator.Next(),
				},
				Name:        "Main",
				IsDefault:   true,
				WorkspaceId: workspace.Id,
				CCAddresses: []string{},
			},
		)
		if err != nil {
			return err
		}
		team := &model.Team{
			Base: model.Base{
				Id: *api.app.UIDGenerator.Next(),
			},
			Name:        "Admins",
			WorkspaceId: workspace.Id,
			Permissions: []string{},
		}
		err = service.Workspace.CreateTeam(ctx, team)
		if err != nil {
			return err
		}
		err = service.Workspace.CreateTeamUser(ctx, &model.TeamUser{
			TeamId:         team.Id,
			UserId:         workspace.Owner,
			OrganizationId: team.OrganizationId,
			WorkspaceId:    team.WorkspaceId,
			Status:         model.TeamUserStatusActive,
		})
		if err != nil {
			return err
		}
		secret, err := crypto.GenerateSecret()
		if err != nil {
			return err
		}

		return service.Client.Create(
			ctx,
			&model.Client{
				Base: model.Base{
					Id: *api.app.UIDGenerator.Next(),
				},
				Description: "Default",
				Secret:      secret,
				Permissions: []string{"*"},
			},
		)
	})

	return err
}
