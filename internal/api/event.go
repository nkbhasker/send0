package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/usesend0/send0/internal/core"
)

type eventAPI struct {
	app *core.App
}

func NewEventAPI(app *core.App) *eventAPI {
	return &eventAPI{app: app}
}

func (e *eventAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/events", e.GetEvents())
		r.Get("/events/{eventId}", e.GetEvent())
	}
}

func (e *eventAPI) GetEvents() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get events
	}
}

func (e *eventAPI) GetEvent() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get event
	}
}
