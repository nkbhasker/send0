package core

import (
	"context"
	"embed"
)

type versionCtxKey struct{}

type migrationCtxKey struct{}

func VersionToContext(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionCtxKey{}, version)
}

func VersionFromContext(ctx context.Context) (string, bool) {
	version, ok := ctx.Value(versionCtxKey{}).(string)
	return version, ok
}

func MigrationToContext(ctx context.Context, migrations embed.FS) context.Context {
	return context.WithValue(ctx, migrationCtxKey{}, migrations)
}

func MigrationFromContext(ctx context.Context) (embed.FS, bool) {
	migrations, ok := ctx.Value(migrationCtxKey{}).(embed.FS)
	return migrations, ok
}
