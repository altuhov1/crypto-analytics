package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateAppUser, downCreateAppUser)
}

func quoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

func upCreateAppUser(ctx context.Context, tx *sql.Tx) error {
	username := os.Getenv("APP_USER")
	password := os.Getenv("APP_PASSWORD")

	if username == "" || password == "" {
		return fmt.Errorf("APP_USER and APP_PASSWORD must be set")
	}

	var exists bool
	err := tx.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = $1)",
		username,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		safeUser := quoteIdentifier(username)
		safePass := quoteLiteral(password)
		_, err = tx.ExecContext(
			ctx,
			fmt.Sprintf("CREATE USER %s WITH PASSWORD %s", safeUser, safePass),
		)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}
	return nil
}

func downCreateAppUser(ctx context.Context, tx *sql.Tx) error {
	username := os.Getenv("APP_USER")
	if username == "" {
		username = "myapp_user"
	}

	safeUser := quoteIdentifier(username)
	_, err := tx.ExecContext(
		ctx,
		fmt.Sprintf("DROP USER IF EXISTS %s", safeUser),
	)
	return err
}
