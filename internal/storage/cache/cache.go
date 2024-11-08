package cache

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/health"
)

type Cache interface {
	health.Checker
	Connection() *redis.Client
	Close() error
	WithTTL(ttl time.Duration) Cache
	WithKeepTTL() Cache
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) error
}

type cache struct {
	connection *redis.Client
	ttl        time.Duration
}

func NewCache(ctx context.Context, cnf *config.Config) (Cache, error) {
	logger := zerolog.Ctx(ctx)
	opts, err := redis.ParseURL(cnf.Redis.URL)
	if err != nil {
		return nil, err
	}
	connection := redis.NewClient(opts)

	// check if the connection is working
	err = connection.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	logger.Info().Msg("Redis cache connected")

	return &cache{connection: connection}, nil
}

func (s cache) Connection() *redis.Client {
	return s.connection
}

func (s *cache) Close() error {
	return s.connection.Close()
}

func (s *cache) Set(ctx context.Context, key string, value interface{}) error {
	v, err := toValue(value)
	if err != nil {
		return err
	}

	return s.connection.Set(ctx, key, v, s.ttl).Err()
}

func (s *cache) Get(ctx context.Context, key string) error {
	return s.connection.Get(ctx, key).Err()
}

func (s cache) WithTTL(ttl time.Duration) Cache {
	return &cache{
		connection: s.connection,
		ttl:        ttl,
	}
}

func (s cache) WithKeepTTL() Cache {
	return &cache{
		connection: s.connection,
		ttl:        redis.KeepTTL,
	}
}

func (s *cache) Health() *health.Health {
	h := health.NewHealth()
	res := s.connection.InfoMap(context.Background(), "server")
	if res.Err() != nil {
		h.SetStatus(health.HealthStatusDown)
		h.SetInfo("error", res.Err().Error())
	} else {
		h.SetStatus(health.HealthStatusUp)
		h.SetInfo("version", res.Val()["Server"]["redis_version"])
	}

	return h
}

func toValue(value interface{}) (interface{}, error) {
	k := reflect.Indirect(reflect.ValueOf(value)).Kind()
	if k == reflect.Struct {
		return json.Marshal(value)
	}

	return value, nil
}
