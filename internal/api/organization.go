package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/model"
)

type createOrganizationRequestPayload struct {
	Name string `json:"name" validate:"required"`
}

type organizationApi struct {
	app *core.App
}

func NewOrganizationAPI(app *core.App) *organizationApi {
	return &organizationApi{
		app: app,
	}
}

func (api *organizationApi) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", api.CreateOrganizationHandler())
		r.Get("/", api.ListOrganizationsHandler())
	}
}

func (api *organizationApi) CreateOrganizationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := new(createOrganizationRequestPayload)
		organization, err := func() (*model.Organization, *ApiError) {
			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			err = api.app.Validate.Struct(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			organization := &model.Organization{
				Name: payload.Name,
			}
			err = api.app.Service.Organization.Create(r.Context(), organization)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			return organization, nil
		}()
		if err != nil {
			render.Status(r, err.StatusCode)
			render.JSON(w, r, map[string]interface{}{
				"success": false,
				"error":   err.Error.Error(),
			})
			return
		}
		render.JSON(w, r, map[string]interface{}{
			"success":      true,
			"organization": organization,
		})
	}
}

func (api *organizationApi) ListOrganizationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageOptions := NewPageOptions(r)
		organizations, count, err := api.app.Service.Organization.List(r.Context())
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		render.JSON(w, r, ToPaginated(organizations, pageOptions, count))
	}
}
