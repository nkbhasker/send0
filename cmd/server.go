package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/usesend0/send0/internal/api"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/core"
	"github.com/usesend0/send0/internal/logger"
)

var serverCmd = &cli.Command{
	Name:  "server",
	Usage: "Manage the server",
	Subcommands: []*cli.Command{
		startServerCmd,
	},
}

var startServerCmd = &cli.Command{
	Name:  "start",
	Usage: "Start the server",
	Action: func(c *cli.Context) error {
		return startServer(c.Context)
	},
}

func startServer(ctx context.Context) error {
	cnf, err := config.LoadConfig()
	if err != nil {
		return err
	}
	logger := logger.NewLogger(cnf)
	ctx = logger.WithContext(ctx)
	app, err := core.NewApp(ctx, cnf)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to setup App")
		return err
	}
	defer func() {
		err := app.Close()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to close App")
		}
	}()
	api, err := api.NewAPI(app)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to setup APIs")
		return err
	}

	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cnf.Port),
		Handler: api.Handler(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	listner, err := net.Listen("tcp", server.Addr)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to listen")
		return err
	}
	defer listner.Close()
	logger.Info().Str("addr", server.Addr).Msg("Server started listening")
	// setup SNS topics after server starts listening
	err = app.Service.SNS.SetupTopics(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to setup SNS topics")
		return err
	}

	go func() {
		// start serving requests
		errCh <- server.Serve(listner)
	}()

	go func() {
		<-sigCh
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		go func() {
			<-ctx.Done()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				errCh <- errors.New("Graceful shutdown timed out.. forcing exit")
			}
			errCh <- ctx.Err()
		}()

		errCh <- server.Shutdown(ctx)
	}()

	return <-errCh
}
