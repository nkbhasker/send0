package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/usesend0/send0/internal/core"
)

type webhookAPI struct {
	app *core.App
}

func NewWebhookAPI(app *core.App) *webhookAPI {
	return &webhookAPI{app: app}
}

func (w *webhookAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/webhooks", w.CreateWebhook())
		r.Get("/webhooks", w.GetWebhooks())
		r.Get("/webhooks/{webhookId}", w.GetWebhook())
		r.Patch("/webhooks/{webhookId}", w.UpdateWebhook())
		r.Delete("/webhooks/{webhookId}", w.DeleteWebhook())
	}
}

func (w *webhookAPI) CreateWebhook() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create webhook
	}
}

func (w *webhookAPI) GetWebhooks() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get webhooks
	}
}

func (w *webhookAPI) GetWebhook() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get webhook
	}
}

func (w *webhookAPI) UpdateWebhook() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Update webhook
	}
}

func (w *webhookAPI) DeleteWebhook() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Delete webhook
	}
}
