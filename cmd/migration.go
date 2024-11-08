package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/urfave/cli/v2"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
)

var migrationCmd = &cli.Command{
	Name:  "migration",
	Usage: "Run migrations",
	Subcommands: []*cli.Command{
		migrationGenerateCmd,
		migrationUpCmd,
	},
}

var migrationGenerateCmd = &cli.Command{
	Name:  "generate",
	Usage: "Generate migrations",
	Action: func(ctx *cli.Context) error {
		return generate(ctx.Context)
	},
}

var migrationUpCmd = &cli.Command{
	Name:  "up",
	Usage: "Run up migrations",
	Action: func(ctx *cli.Context) error {
		return up(ctx.Context)
	},
}

func generate(_ context.Context) error {
	sb := &strings.Builder{}
	sb = loadEnums(sb)
	sb = loadModels(sb)
	_, err := io.WriteString(os.Stdout, sb.String())
	if err != nil {
		return err
	}

	return nil
}

func up(_ context.Context) error {
	return fmt.Errorf("Not implemented")
}

func loadEnums(sb *strings.Builder) *strings.Builder {
	enums := []string{
		model.BroadcastStatusTypeCreateQuery,
		model.DomainStatusTypeCreateQuery,
		model.EventTypeCreateQuery,
		model.IdentityProviderTypeCreateQuery,
		model.TeamUserStatusTypeCreateQuery,
	}
	for _, enum := range enums {
		sb.WriteString(enum)
	}

	return sb
}

func loadModels(sb *strings.Builder) *strings.Builder {
	models := []interface{}{
		&model.Client{},
		&model.Domain{},
		&model.Email{},
		&model.Event{},
		&model.Organization{},
		&model.SNSTopic{},
		&model.Team{},
		&model.TeamUser{},
		&model.Template{},
		&model.User{},
		&model.Workspace{},
	}
	stmts, err := gormschema.New(constant.DialectPostgres).Load(models...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	sb.WriteString(stmts)

	return sb
}
