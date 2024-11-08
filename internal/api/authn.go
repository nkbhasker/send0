package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/muhlemmer/httpforwarded"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

const (
	OTPScopeSignIn      OTPScope = "SIGN_IN"
	OTPScopeEmailUpdate OTPScope = "EMAIL_UPDATE"
)

type OTPScope string

type AuthnAPI struct {
	app                    *core.App
	otpGenerateRateLimiter core.RateLimiter
	otpVerifyRateLimiter   core.RateLimiter
}

type otpRequestPayload struct {
	Email string  `json:"email" validate:"required,email"`
	Scope *string `json:"scope" validate:"oneof=SIGN_IN EMAIL_UPDATE"`
}

type signInRequestPayload struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}

type clientTokenRequestPayload struct {
	ClientId     uid.UID `json:"clientId" validate:"required"`
	ClientSecret string  `json:"clientSecret" validate:"required"`
}

func NewAuthnAPI(app *core.App) *AuthnAPI {
	otpGenerateRateLimiter := core.NewRateLimiter(
		app.Cache,
		core.RateLimitKindOtpGenerate,
		app.Config.Authn.OtpGenerateRateLimit,
		app.Config.Authn.OtpGenerateRateLimitWindow,
	)
	otpVerifyRateLimiter := core.NewRateLimiter(
		app.Cache,
		core.RateLimitKindOtpVerify,
		app.Config.Authn.OtpVerifyRateLimit,
		app.Config.Authn.OtpVerifyRateLimitWindow,
	)

	return &AuthnAPI{
		app:                    app,
		otpGenerateRateLimiter: otpGenerateRateLimiter,
		otpVerifyRateLimiter:   otpVerifyRateLimiter,
	}
}

func (a *AuthnAPI) Route() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/otp", a.OtpHandler())
		r.Post("/signin", a.SignInHandler())
		r.Post("/token", a.ClientTokenHandler())
	}
}

func (a *AuthnAPI) OtpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		otp, err := func() (*string, *ApiError) {
			ok, err := a.otpGenerateRateLimiter.Evaluate(GetIP(r))
			if !ok || err != nil {
				return nil, &ApiError{
					Error:      errors.New("too many otp requests"),
					StatusCode: http.StatusTooManyRequests,
				}
			}
			payload := new(otpRequestPayload)
			err = json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			if payload.Scope == nil {
				*payload.Scope = string(OTPScopeSignIn)
			}
			err = a.app.Validate.Struct(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			otp, err := crypto.GenerateOtp()
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			email := strings.ToLower(payload.Email)

			key := fmt.Sprintf("%s:%s:%s", *payload.Scope, model.AuthKeyOTP, email)
			err = a.app.Repository.Authn.SaveOTP(r.Context(), key, otp)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			// reset otp verify rate limit
			err = a.otpVerifyRateLimiter.Reset(email)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return &otp, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}
		resp := map[string]interface{}{
			"success": true,
		}
		if a.app.Config.Env == constant.EnvDevelopment {
			resp["otp"] = *otp
		}

		render.JSON(w, r, resp)
	}
}

func (a *AuthnAPI) SignInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := func() (*string, *ApiError) {
			payload := &signInRequestPayload{}
			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			err = a.app.Validate.Struct(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			email := strings.ToLower(payload.Email)
			ok, err := a.otpVerifyRateLimiter.Evaluate(email)
			if !ok || err != nil {
				return nil, &ApiError{
					Error:      errors.New("too many otp verification requests"),
					StatusCode: http.StatusTooManyRequests,
				}
			}
			key := fmt.Sprintf("%s:%s:%s", OTPScopeSignIn, model.AuthKeyOTP, email)
			otp, err := a.app.Repository.Authn.GetOTP(r.Context(), key)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			if ok := crypto.ValidateOtp(otp, payload.OTP); !ok {
				return nil, &ApiError{
					Error:      errors.New("invalid otp"),
					StatusCode: http.StatusBadRequest,
				}
			}
			user, err := a.app.Repository.User.FindByEmail(r.Context(), email)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			// create new user
			if user == nil {
				user = &model.User{
					Base: model.Base{
						Id: *a.app.UIDGenerator.Next(),
					},
					Email:         email,
					EmailVerified: true,
				}
				err = a.app.Repository.User.Save(r.Context(), user)
				if err != nil {
					return nil, &ApiError{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}
				}
			}
			options := []crypto.JWTClaimOptions{
				crypto.WithEmail(email),
			}
			id, accessToken, err := a.app.JWT.NewAccessToken(user.Id.String(), options...)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			err = a.app.Repository.Authn.SaveAccessToken(r.Context(), id, email)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			// delete otp
			err = a.app.Repository.Authn.DeleteOTP(r.Context(), key)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return &accessToken, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success":     true,
			"accessToken": *accessToken,
		})
	}
}

func (a *AuthnAPI) ClientTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := func() (*string, *ApiError) {
			payload := &clientTokenRequestPayload{}
			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			err = a.app.Validate.Struct(payload)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusBadRequest,
				}
			}
			client, err := a.app.Repository.Client.FindByID(r.Context(), payload.ClientId)
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			if client == nil || client.Secret != payload.ClientSecret {
				return nil, &ApiError{
					Error:      errors.New("invalid client"),
					StatusCode: http.StatusBadRequest,
				}
			}
			id, token, err := a.app.JWT.NewAccessToken(client.Id.String())
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}
			err = a.app.Repository.Authn.SaveAccessToken(r.Context(), id, client.Id.String())
			if err != nil {
				return nil, &ApiError{
					Error:      err,
					StatusCode: http.StatusInternalServerError,
				}
			}

			return &token, nil
		}()
		if err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"token":   *token,
		})
	}
}

func GetIP(r *http.Request) string {
	fwd, err := httpforwarded.ParseFromRequest(r)
	if err == nil && len(fwd[http.CanonicalHeaderKey(constant.HeaderXFowardedFor)]) != 0 {
		return fwd[http.CanonicalHeaderKey(constant.HeaderXFowardedFor)][0]
	}

	return r.RemoteAddr
}
