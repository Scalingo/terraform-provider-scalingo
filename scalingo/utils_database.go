package scalingo

import (
	"context"

	"github.com/Scalingo/go-scalingo/v11"
	"github.com/Scalingo/go-utils/errors/v3"
)

func toDatabaseTypeName(ctx context.Context, database scalingo.DatabaseNG) (string, error) {
	switch database.Technology {
	case "postgresql-ng":
		return "POSTGRESQL", nil
	case "mysql-ng":
		return "MYSQL", nil
	default:
		return "", errors.Newf(ctx, "no matching type for %q", database.Technology)
	}
}
