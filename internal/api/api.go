package api

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/middleware"
)

const (
	PageOrderAsc  PageOrder = "asc"
	PageOrderDesc PageOrder = "desc"
)
const MaxTake = 100

const (
	QueryParamPage  = "page"
	QueryParamTake  = "take"
	QueryParamOrder = "order"
	QueryParamQ     = "q"
)

const (
	HeaderAmazonSNSMessageType = "x-amz-sns-message-type"
)

type API interface {
	Handler() http.Handler
}

type api struct {
	router chi.Router
}

type ApiError struct {
	Error      error `json:"error"`
	StatusCode int   `json:"code"`
}

type BaseHeader struct {
	WorkspaceId string `json:"x-workspace-id"`
}

type IdempotentHeader struct {
	IdempotencyKey string `json:"Idempotency-Key"`
}

type PageOrder string

type PaginatedOptions struct {
	Order PageOrder `json:"order"`
	Page  int       `json:"page"`
	Take  int       `json:"take"`
	Q     *string   `json:"q"`
}

type PaginatedMeta struct {
	ItemCount       int  `json:"itemCount"`
	PageCount       int  `json:"pageCount"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	HasNextPage     bool `json:"hasNextPage"`
}

type Paginated[T interface{}] struct {
	Data []*T          `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

func NewAPI(app *core.App) (API, error) {
	router := chi.NewRouter()
	authInterceptor := middleware.NewIdentityInterceptor(app.JWT)

	router.Route("/healthz", NewHealthAPI(app).Route())
	router.Route("/auth", NewAuthnAPI(app).Route())
	router.Post(constant.SNSEventPath, snsTopicHandler(app))

	router.Group(func(r chi.Router) {
		r.Use(authInterceptor.Handler)
		r.Route("/domains", NewDomainAPI(app).Route())
		r.Route("/users", NewUserAPI(app).Route())
		r.Route("/workspaces", NewWorkspaceAPI(app).Route())
	})

	return &api{
		router: router,
	}, nil
}

func (a *api) Handler() http.Handler {
	return a.router
}

func ToPaginated[T interface{}](data []*T, options *PaginatedOptions, itemCount int) *Paginated[T] {
	pageCount := itemCount / options.Take
	if itemCount%options.Take > 0 {
		pageCount++
	}
	hasPreviousPage := options.Page > 1
	hasNextPage := options.Page < pageCount
	return &Paginated[T]{
		Data: data,
		Meta: PaginatedMeta{
			ItemCount:       itemCount,
			PageCount:       pageCount,
			HasPreviousPage: hasPreviousPage,
			HasNextPage:     hasNextPage,
		},
	}
}

func NewPageOptions(r *http.Request) *PaginatedOptions {
	options := &PaginatedOptions{
		Page:  1,
		Take:  10,
		Order: PageOrderDesc,
		Q:     nil,
	}
	q := r.URL.Query().Get(QueryParamQ)
	if q != "" {
		options.Q = &q
	}
	page := r.URL.Query().Get(QueryParamPage)
	if page != "" {
		page, err := strconv.Atoi(page)
		if err == nil {
			options.Page = page
		}
	}
	take := r.URL.Query().Get(QueryParamTake)
	if take != "" {
		take, err := strconv.Atoi(take)
		if err == nil {
			if take <= MaxTake {
				options.Take = take
			}
		}
	}
	order := r.URL.Query().Get(QueryParamOrder)
	if order != "" && order == string(PageOrderAsc) {
		options.Order = PageOrderAsc
	}

	return options
}

func snsTopicHandler(app *core.App) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := func() *ApiError {
			messageType := r.Header.Get(HeaderAmazonSNSMessageType)
			if messageType == "" {
				return &ApiError{
					Error:      errors.New("Invalid request"),
					StatusCode: http.StatusBadRequest,
				}
			}
			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				return &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			app.Logger.Info().Str("messageType", messageType).Msg("Received SNS message")
			switch messageType {
			case "SubscriptionConfirmation":
				err = app.Service.SNS.ConfirmSubscribe(bytes)
				if err != nil {
					return &ApiError{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}
				}
			case "UnsubscribeConfirmation":
				err = app.Service.SNS.ConfirmUnsubscribe(bytes)
				if err != nil {
					return &ApiError{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}
				}
			case "Notification":
				err = app.Service.SNS.ProcessNotification(bytes)
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
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func (p *PaginatedOptions) Skip() int {
	return (p.Page - 1) * p.Take
}

func renderError(w http.ResponseWriter, r *http.Request, err *ApiError) {
	w.WriteHeader(err.StatusCode)
	render.JSON(w, r, map[string]interface{}{
		"error":   err.Error.Error(),
		"success": false,
	})
}
