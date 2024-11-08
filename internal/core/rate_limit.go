package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/usesend0/send0/internal/storage/cache"
)

type RateLimitKind string

const (
	RateLimitKindOtpVerify   RateLimitKind = "OTP_VERIFY"
	RateLimitKindOtpGenerate RateLimitKind = "OTP_GENERATE"
)

type RateLimiter interface {
	Evaluate(identifier string) (bool, error)
	Reset(identifier string) error
}

type rateLimiter struct {
	cache  cache.Cache
	kind   RateLimitKind
	limit  int
	window time.Duration
}

func NewRateLimiter(
	cache cache.Cache,
	kind RateLimitKind,
	limit int,
	timeWindowInSeconds int,
) RateLimiter {
	return &rateLimiter{
		cache:  cache,
		kind:   kind,
		limit:  limit,
		window: time.Duration(timeWindowInSeconds * int(time.Second)),
	}
}

func (r *rateLimiter) Evaluate(identifier string) (bool, error) {
	ctx := context.Background()
	timestamp := time.Now()
	key := strings.ToLower(fmt.Sprintf(`%s_%s`, r.kind, identifier))
	pipe := r.cache.Connection().TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0.0", strconv.FormatFloat(float64(timestamp.Add(-r.window).UnixMilli()), 'f', -1, 64))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(timestamp.UnixMilli()), Member: float64(timestamp.UnixMilli())})
	results := pipe.ZRange(ctx, key, 0, -1)
	pipe.Expire(ctx, key, r.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	if len(results.Val()) > r.limit {
		return false, nil
	}

	return true, nil
}

func (r *rateLimiter) Reset(identifier string) error {
	ctx := context.Background()
	key := strings.ToLower(fmt.Sprintf(`%s_%s`, r.kind, identifier))
	err := r.cache.Connection().Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}
