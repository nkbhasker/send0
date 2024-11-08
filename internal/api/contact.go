package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/usesend0/send0/internal/core"
)

type contactAPI struct {
	app *core.App
}

func NewContactAPI(app *core.App) *contactAPI {
	return &contactAPI{app: app}
}

func (c *contactAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/contacts", c.CreateContact())
		r.Get("/contacts/{contactId}/segments", c.GetContactSegments())

		r.Post("/segments", c.CreateSegment())
		r.Post("/segments/{segmentId}/contacts", c.AddContactToSegment())
		r.Get("/segments/{segmentId}/contacts", c.GetContactsInSegment())
		r.Get("/segments/{segmentId}/contacts/{contactId}/subscribe", c.GetContactInSegment())
		r.Get("/segments/{segmentId}/contacts/{contactId}/unsubscribe", c.GetContactInSegment())
	}
}

func (c *contactAPI) CreateContact() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create contact
	}
}

func (c *contactAPI) ConfirmContact() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Confirm contact
	}
}

func (c *contactAPI) CreateSegment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create segment
	}
}

func (c *contactAPI) AddContactToSegment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add contact to segment
	}
}

func (c *contactAPI) GetContactsInSegment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get contacts in segment
	}
}

func (c *contactAPI) GetContactInSegment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get contact in segment
	}
}

func (c *contactAPI) GetContactSegments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get contact segments
	}
}
