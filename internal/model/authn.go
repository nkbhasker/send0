package model

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	AuthKeyOTP         AuthKey = "OTP"
	AuthKeyAccessToken AuthKey = "ACCESS_TOKEN"
)

type AuthKey string

type AuthnRepository interface {
	SaveOTP(ctx context.Context, key, otp string) error
	GetOTP(ctx context.Context, key string) (string, error)
	DeleteOTP(ctx context.Context, key string) error
	SaveAccessToken(ctx context.Context, jti, sub string) error
}

type authnRepository struct {
	*baseRepository
}

func NewAuthnRepository(baseRepository *baseRepository) AuthnRepository {
	return &authnRepository{
		baseRepository,
	}
}

func (a *authnRepository) SaveOTP(ctx context.Context, key, otp string) error {
	return a.Cache.Set(ctx, key, otp)
}

func (a *authnRepository) GetOTP(ctx context.Context, key string) (string, error) {
	result := a.Cache.Connection().Get(ctx, key)
	if result.Err() == redis.Nil {
		return "", fmt.Errorf("otp expired")
	}
	if result.Err() != nil {
		return "", result.Err()
	}

	return result.Val(), nil
}

func (a *authnRepository) DeleteOTP(ctx context.Context, key string) error {
	return a.Cache.Connection().Del(ctx, key).Err()
}

func (a *authnRepository) SaveAccessToken(ctx context.Context, jti, sub string) error {
	key := fmt.Sprintf("%s:%s:%s", AuthKeyAccessToken, sub, jti)

	return a.Cache.Set(ctx, key, "1")
}
