package core

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/health"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/service"
	"github.com/usesend0/send0/internal/storage/cache"
	"github.com/usesend0/send0/internal/storage/db"
	"github.com/usesend0/send0/internal/uid"
)

var _ health.Checker = (*App)(nil)

type App struct {
	startedAt    time.Time
	Version      string
	Config       *config.Config
	Validate     *validator.Validate
	Logger       *zerolog.Logger
	JWT          *crypto.JWT
	DB           db.DB
	Cache        cache.Cache
	Repository   *model.Repository
	Service      *service.Service
	UIDGenerator uid.UIDGenerator
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logger := zerolog.Ctx(ctx)
	uidGenerator := uid.NewUIDGenerator(cfg, logger)
	cache, err := cache.NewCache(ctx, cfg)
	if err != nil {
		return nil, err
	}
	db, err := db.NewDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	version, ok := VersionFromContext(ctx)
	if !ok {
		logger.Error().Msg("invalid version")
		return nil, errors.New("invalid version")
	}

	jwt, err := crypto.NewJWT(cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to setup JWT")
		return nil, err
	}
	baseRepository := model.NewBaseRepository(cache, db, uidGenerator, logger)
	repository := model.NewRepository(baseRepository)
	service, err := service.NewService(service.NewBaseService(cfg, uidGenerator, logger, repository))
	if err != nil {
		logger.Error().Err(err).Msg("failed to setup service")
		return nil, err
	}
	// seed workspace
	err = service.Workspace.Seed(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to seed workspace")
		return nil, err
	}

	return &App{
		startedAt:    time.Now(),
		Logger:       logger,
		Version:      version,
		Config:       cfg,
		DB:           db,
		Cache:        cache,
		JWT:          jwt,
		Repository:   repository,
		Service:      service,
		UIDGenerator: uidGenerator,
		Validate:     NewValidate(),
	}, nil
}

func (a *App) Health() *health.Health {
	h := health.NewHealth()
	h.SetStatus(health.HealthStatusUp)
	h.SetInfo("version", a.Version)
	h.SetInfo("uptime", time.Since(a.startedAt).String())
	h.SetInfo("cpus", runtime.NumCPU())
	h.SetInfo("os", runtime.GOOS)

	return h
}

func (a *App) Close() error {
	if a.Cache != nil {
		err := a.Cache.Close()
		if err != nil {
			return err
		}
	}

	return a.DB.Close()
}
