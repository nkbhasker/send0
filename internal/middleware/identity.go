package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/crypto"
)

type identityInterceptor struct {
	jwt *crypto.JWT
}

const (
	Bearer string = "bearer"
)

func NewIdentityInterceptor(jwt *crypto.JWT) *identityInterceptor {
	return &identityInterceptor{
		jwt: jwt,
	}
}

func (a *identityInterceptor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		identity, err := func() (core.Identity, error) {
			accessToken, err := extractToken(r.Header)
			if err != nil {
				return nil, err
			}
			claims, err := a.jwt.VerifyAccessToken(accessToken)
			if err != nil {
				return nil, err
			}
			workspaceId := extractWorkspaceId(r.Header)

			return core.NewIdentity(core.IdentityOptions{
				JTI:         claims.ID,
				Sub:         claims.Subject,
				WorkspaceId: workspaceId,
			})
		}()
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		ctx = core.IdentityToContext(ctx, identity)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(h http.Header) (string, error) {
	authHeader := h.Get(constant.HeaderAuthorization)
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != Bearer {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}

func extractWorkspaceId(h http.Header) *string {
	workspaceId := h.Get(constant.HeaderWorkspaceId)
	if workspaceId == "" {
		return nil
	}

	return &workspaceId
}
