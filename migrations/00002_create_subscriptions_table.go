package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateSubscriptionsTable, downCreateSubscriptionsTable)
}

func upCreateSubscriptionsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE contacts (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		username TEXT UNIQUE NOT NULL,
		favorite_coins TEXT[] DEFAULT '{}',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX idx_users_username ON users(username);
	`)

	return err
}

func downCreateSubscriptionsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS contacts CASCADE;
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS users CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
