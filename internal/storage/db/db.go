package db

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/health"
)

type Connection interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, src pgx.CopyFromSource) (int64, error)
}

type Transactioner interface {
	BeginTx(ctx context.Context) (pgx.Tx, DB, error)
}

type DB interface {
	Transactioner
	health.Checker
	Close() error
	Connection() Connection
	Builder() squirrel.StatementBuilderType
}

type db struct {
	logger  *zerolog.Logger
	pool    *pgxpool.Pool
	conn    Connection
	builder squirrel.StatementBuilderType
}

func NewDB(ctx context.Context, cnf *config.Config) (DB, error) {
	logger := zerolog.Ctx(ctx)
	if cnf.Postgres.URL == "" {
		logger.Error().Msg("Postgres url is empty")
		return nil, errors.New("postgres url is empty")
	}
	connConfig, err := pgxpool.ParseConfig(cnf.Postgres.URL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse postgres url")
		return nil, err
	}
	connConfig.MinConns = int32(cnf.Postgres.IdlePoolSize)
	connConfig.MaxConns = int32(cnf.Postgres.PoolSize)
	conn, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	logger.Info().Msg("Postgres database connected")
	return &db{
		logger:  logger,
		pool:    conn,
		conn:    conn,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (s db) Connection() Connection {
	return s.conn
}

func (s *db) Close() error {
	s.logger.Info().Msg("Closing database connection")
	s.pool.Close()
	return nil
}

func (s *db) BeginTx(ctx context.Context) (pgx.Tx, DB, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	return tx, &db{
		pool:    s.pool,
		builder: s.builder,
		conn:    tx.Conn(),
	}, nil
}

func (s *db) Builder() squirrel.StatementBuilderType {
	return s.builder
}

func (s *db) Health() *health.Health {
	var version string
	h := health.NewHealth()
	err := s.Connection().QueryRow(context.Background(), "SELECT VERSION();").Scan(&version)
	if err != nil {
		h.SetStatus(health.HealthStatusDown)
		h.SetInfo("error", err.Error())
	} else {
		h.SetStatus(health.HealthStatusUp)
		h.SetInfo("version", version)
	}

	return h
}
