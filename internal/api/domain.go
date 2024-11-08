package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

type createDomainRequestPayload struct {
	Name   string             `json:"name" validate:"required"`
	Region constant.AwsRegion `json:"region" validate:"required"`
}

type domainApi struct {
	app *core.App
}

func NewDomainAPI(app *core.App) *domainApi {
	return &domainApi{
		app: app,
	}
}

func (api *domainApi) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", api.CreateDomainHandler())
		r.Delete("/{domainId}", api.DeleteDomainHandler())
	}
}

func (api *domainApi) CreateDomainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// identity := core.IdentityFromContext(r.Context())
		payload := new(createDomainRequestPayload)
		domain, err := func() (*model.Domain, *ApiError) {
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
			domain := &model.Domain{
				Name:   payload.Name,
				Region: payload.Region,
			}
			err = api.app.Service.Domain.Create(r.Context(), domain)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			return domain, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}
		render.JSON(w, r, domain)
	}
}

func (api *domainApi) DeleteDomainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// identity := core.IdentityFromContext(r.Context())
		err := func() *ApiError {
			domainId, err := uid.NewUIDFromString(chi.URLParam(r, "domainId"))
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			err = api.app.Service.Domain.Delete(r.Context(), *domainId)
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
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
		})
	}
}
