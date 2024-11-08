package crypto

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/usesend0/send0/internal/config"
)

type JWT struct {
	issuer              string
	key                 *rsa.PrivateKey
	kid                 string
	accessTokenDuration time.Duration
}

type claims struct {
	jwt.RegisteredClaims
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
}

type JWTClaimOptions = func(*claims)

func NewJWT(cfg *config.Config) (*JWT, error) {
	key, err := EncodedToPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		return nil, err
	}
	bytes, err := PublicKeyToBytes(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	h := md5.New()
	_, err = h.Write(bytes)
	if err != nil {
		return nil, err
	}
	kid := hex.EncodeToString(h.Sum(nil))

	return &JWT{
		issuer:              cfg.Host,
		key:                 key,
		kid:                 kid,
		accessTokenDuration: time.Duration(cfg.JWT.AccessTokenExpiry) * time.Minute,
	}, nil
}

func (j *JWT) NewAccessToken(
	sub string,
	options ...JWTClaimOptions,
) (string, string, error) {
	claims := &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenDuration)),
		},
	}
	for _, opt := range options {
		opt(claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.kid

	// Create the JWT string.
	tokenString, err := token.SignedString(j.key)
	if err != nil {
		return "", "", err
	}

	return claims.ID, tokenString, nil
}

func (j *JWT) VerifyAccessToken(accessToken string) (*claims, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Name {
			return nil, fmt.Errorf(
				"unexpected access token signing method=%v, expect %v",
				t.Header["alg"],
				jwt.SigningMethodRS256,
			)
		}
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("unexpected kid")
		}
		if kid != j.kid {
			return nil, errors.New("invalid kid")
		}

		// Return public key pointer expected by rsa verify
		return &j.key.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

func WithEmail(email string) JWTClaimOptions {
	return func(c *claims) {
		c.Email = email
	}
}

func WithFirstName(firstName string) JWTClaimOptions {
	return func(c *claims) {
		c.FirstName = firstName
	}
}

func WithLastName(lastName string) JWTClaimOptions {
	return func(c *claims) {
		c.LastName = lastName
	}
}

func WithExpiresIn(expiresIn time.Duration) JWTClaimOptions {
	return func(c *claims) {
		c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresIn))
	}
}
