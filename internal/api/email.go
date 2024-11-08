package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/service"
)

type sendEmailRequestPayload struct {
	From           string                  `json:"from" validate:"required"`
	Recipients     []string                `json:"recipients" validate:"required"`
	CC             []string                `json:"cc"`
	BCC            []string                `json:"bcc"`
	Delay          int                     `json:"delay"`
	Subject        *string                 `json:"subject" validate:"required"`
	Html           *string                 `json:"html" validate:"required"`
	Text           *string                 `json:"text"`
	Data           *map[string]interface{} `json:"data"`
	ReplyTo        *string                 `json:"replyTo"`
	Headers        []map[string]string     `json:"headers"`
	TemplateId     *string                 `json:"templateId"`
	Attachments    []model.Attachment      `json:"attachments"`
	OrganizationId *string                 `json:"organizationId"`
	MetaData       *map[string]interface{} `json:"metaData"`
}

type sendEmailRequestHeaders struct {
	IdempotencyKey *string `json:"Idempotency-Key"`
}

type EmailAPI struct {
	app *core.App
}

func NewEmailAPI(app *core.App) *EmailAPI {
	return &EmailAPI{
		app: app,
	}
}

func (api *EmailAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", api.SendEmailHandler())
	}
}

func (api *EmailAPI) SendEmailHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// identity := core.IdentityFromContext(r.Context())
		payload := new(sendEmailRequestPayload)
		headers := new(sendEmailRequestHeaders)
		email, err := func() (*model.Email, *ApiError) {
			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			err = json.NewDecoder(r.Body).Decode(headers)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			email := &model.Email{
				From: payload.From,
			}
			if headers.IdempotencyKey != nil {
				email.RequestId = *headers.IdempotencyKey
			} else {
				email.RequestId = api.app.UIDGenerator.Next().String()
			}
			email.Recipients, err = service.ParseRecipients(payload.Recipients)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			email.CCRecipients, err = service.ParseRecipients(payload.CC)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			email.BCCRecipients, err = service.ParseRecipients(payload.BCC)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			// TODO: Check default organization id
			// TODO: Check and validate template
			_, err = api.app.Service.Email.Send(r.Context(), email.RequestId, []*model.Email{
				email,
			})
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return email, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}
		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"email":   email,
		})
	}
}
