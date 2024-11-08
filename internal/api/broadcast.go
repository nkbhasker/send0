package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/usesend0/send0/internal/core"
)

type broadcastAPI struct {
	app *core.App
}

func NewCampaignAPI(app *core.App) *broadcastAPI {
	return &broadcastAPI{app: app}
}

func (c *broadcastAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/broadcasts", c.GetBroadcasts())
		r.Post("/broadcasts", c.CreateBroadcast())
		r.Get("/broadcasts/{broadcastId}", c.GetBroadcast())
		r.Patch("/broadcasts/{broadcastId}", c.UpdateBroadcast())
		r.Delete("/broadcasts/{broadcastId}", c.DeleteBroadcast())
		r.Post("/broadcasts/{broadcastId}/start", c.StartBroadcast())
	}
}

func (c *broadcastAPI) GetBroadcasts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get broadcasts
	}
}

func (c *broadcastAPI) CreateBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create campaign
	}
}

func (c *broadcastAPI) GetBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get campaign
	}
}

func (c *broadcastAPI) UpdateBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Update campaign
	}
}

func (c *broadcastAPI) DeleteBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Delete campaign
	}
}

func (c *broadcastAPI) StartBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Send campaign
	}
}
