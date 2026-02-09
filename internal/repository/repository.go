package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/newmo-oss/ergo"
	_ "modernc.org/sqlite"
)

type Repository struct {
	conn *sql.DB
	q    *db.Queries
}

// Create creates a new review DB (mkdir + open + migrate).
func Create(dbPath string, schema string) (*Repository, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, ergo.Wrap(err, "failed to create review directory",
			slog.String("path", filepath.Dir(dbPath)))
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to open database",
			slog.String("path", dbPath))
	}

	if err := setPragmas(conn); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, ergo.Wrap(err, "failed to create schema")
	}

	return &Repository{conn: conn, q: db.New(conn)}, nil
}

// Open opens an existing review DB.
func Open(dbPath string) (*Repository, error) {
	if _, err := os.Stat(dbPath); err != nil {
		return nil, ergo.Wrap(err, "review database not found",
			slog.String("path", dbPath))
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to open database",
			slog.String("path", dbPath))
	}

	if err := setPragmas(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &Repository{conn: conn, q: db.New(conn)}, nil
}

func (r *Repository) Queries() *db.Queries {
	return r.q
}

func (r *Repository) WithTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return ergo.Wrap(err, "failed to begin transaction")
	}

	if err := fn(r.q.WithTx(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func setPragmas(conn *sql.DB) error {
	if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return ergo.Wrap(err, "failed to set WAL mode")
	}
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return ergo.Wrap(err, "failed to enable foreign keys")
	}
	return nil
}
