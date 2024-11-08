package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

type userAPI struct {
	app *core.App
}

type updateUserRequestPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func NewUserAPI(app *core.App) *userAPI {
	return &userAPI{
		app: app,
	}
}

func (api *userAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/me", api.MeHandler())
		r.Put("/me", api.MeUpdateHandler())
		r.Get("/{id}", api.GetUserHandler())
	}
}

func (api *userAPI) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := core.IdentityFromContext(r.Context())
		user, err := api.app.Repository.User.FindById(r.Context(), identity.UserId())
		if err != nil {
			renderError(w, r, &ApiError{
				StatusCode: http.StatusInternalServerError,
				Error:      err,
			})
		}
		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"user":    user,
		})
	}
}

func (api *userAPI) MeUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := core.IdentityFromContext(r.Context())
		user, err := func() (*model.User, *ApiError) {
			var payload updateUserRequestPayload
			err := json.NewDecoder(r.Body).Decode(&payload)
			if err != nil {
				return nil, &ApiError{
					StatusCode: http.StatusBadRequest,
					Error:      err,
				}
			}
			user, err := api.app.Repository.User.FindById(r.Context(), identity.UserId())
			if err != nil {
				return nil, &ApiError{
					StatusCode: http.StatusInternalServerError,
					Error:      err,
				}
			}
			if user == nil {
				return nil, &ApiError{
					StatusCode: http.StatusNotFound,
					Error:      err,
				}
			}
			user.FirstName = payload.FirstName
			user.LastName = payload.LastName
			err = api.app.Repository.User.Update(r.Context(), user)
			if err != nil {
				return nil, &ApiError{
					StatusCode: http.StatusInternalServerError,
					Error:      err,
				}
			}

			return user, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"user":    user,
		})
	}
}

func (api *userAPI) GetUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := func() (*model.User, *ApiError) {
			userId, err := strconv.Atoi(chi.URLParam(r, "id"))
			if err != nil {
				return nil, &ApiError{
					StatusCode: http.StatusBadRequest,
					Error:      err,
				}
			}
			uid := uid.NewUID(int64(userId))
			user, err := api.app.Repository.User.FindById(r.Context(), *uid)
			if err != nil {
				return nil, &ApiError{
					StatusCode: http.StatusInternalServerError,
					Error:      err,
				}
			}
			return user, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"user":    user,
		})
	}
}
