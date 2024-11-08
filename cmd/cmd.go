package cmd

import (
	"context"
	"embed"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/core"
)

// Execute will run the root command
func Execute(version string, migrations embed.FS) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = core.VersionToContext(ctx, version)
	ctx = core.MigrationToContext(ctx, migrations)

	app := &cli.App{
		Name:    constant.AppName,
		Usage:   "CLI for " + constant.AppName,
		Version: version,
		Commands: []*cli.Command{
			serverCmd,
			migrationCmd,
		},
	}

	return app.RunContext(ctx, os.Args)
}
