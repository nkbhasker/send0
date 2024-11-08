package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/health"
)

type HealthAPI struct {
	app *core.App
}

func NewHealthAPI(app *core.App) *HealthAPI {
	return &HealthAPI{app: app}
}

func (a *HealthAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", a.HealthzHandler())
	}
}

func (hh *HealthAPI) HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := health.NewHealth()
		h.SetStatus(health.HealthStatusUp)
		services := map[string]health.Checker{
			"app":   hh.app,
			"db":    hh.app.DB,
			"cache": hh.app.Cache,
		}
		type result struct {
			service string
			health  *health.Health
		}
		ch := make(chan result)
		defer close(ch)
		for k, s := range services {
			go func(k string, s health.Checker) {
				ch <- result{service: k, health: s.Health()}
			}(k, s)
		}

		for range services {
			result := <-ch
			if result.health.Status() != health.HealthStatusUp {
				render.Status(r, http.StatusServiceUnavailable)
				h.SetStatus(result.health.Status())
			}
			h.SetInfo(result.service, result.health.Info())
		}

		render.JSON(w, r, h)
	}
}
