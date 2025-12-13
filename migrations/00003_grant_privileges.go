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
	goose.AddMigrationContext(upGrantPrivileges, downGrantPrivileges)
}

func quotePostgresIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func upGrantPrivileges(ctx context.Context, tx *sql.Tx) error {
	username := os.Getenv("APP_USER")
	if username == "" {
		return fmt.Errorf("APP_USER is not set")
	}
	quotedUser := quotePostgresIdentifier(username)

	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE contacts TO %s;
		GRANT USAGE, SELECT ON SEQUENCE contacts_id_seq TO %s;
	`, quotedUser, quotedUser))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE users TO %s;
		GRANT USAGE, SELECT ON SEQUENCE users_id_seq TO %s;
	`, quotedUser, quotedUser))
	if err != nil {
		return err
	}

	fmt.Printf("Granted privileges to user: %s\n", username)
	return nil
}

func downGrantPrivileges(ctx context.Context, tx *sql.Tx) error {
	username := os.Getenv("APP_USER")
	if username == "" {
		username = "myapp_user"
	}
	quotedUser := quotePostgresIdentifier(username)

	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		REVOKE ALL PRIVILEGES ON TABLE contacts FROM %s;
		REVOKE ALL PRIVILEGES ON SEQUENCE contacts_id_seq FROM %s;
	`, quotedUser, quotedUser))
	if err != nil {
		if strings.Contains(err.Error(), "undefined_object") {
			fmt.Printf("Privileges already revoked for user: %s\n", username)
			return nil
		}
		return err
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		REVOKE ALL PRIVILEGES ON TABLE users FROM %s;
		REVOKE ALL PRIVILEGES ON SEQUENCE users_id_seq FROM %s;
	`, quotedUser, quotedUser))
	if err != nil {
		if strings.Contains(err.Error(), "undefined_object") {
			fmt.Printf("Privileges already revoked for user: %s\n", username)
			return nil
		}
		return err
	}

	fmt.Printf("Revoked privileges from user: %s\n", username)
	return nil
}
