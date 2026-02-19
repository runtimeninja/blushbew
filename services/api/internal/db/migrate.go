package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, pool Execer) error {
	// Ensure schema_migrations exists
	if _, err := pool.Exec(ctx, `
		create table if not exists schema_migrations (
			filename text primary key,
			applied_at timestamptz not null default now()
		);
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	for _, f := range files {
		applied, err := isApplied(ctx, pool, f)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		b, err := migrationsFS.ReadFile("migrations/" + f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		// Apply in a transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(ctx, string(b)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", f, err)
		}

		if _, err := tx.Exec(ctx, `insert into schema_migrations(filename) values($1)`, f); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", f, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", f, err)
		}
	}

	return nil
}

// small interface so we can use pool or tx
type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (any, error)
	Begin(ctx context.Context) (Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

type Tx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (any, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Row interface {
	Scan(dest ...any) error
}

func isApplied(ctx context.Context, pool Execer, filename string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `select exists(select 1 from schema_migrations where filename=$1)`, filename).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check schema_migrations: %w", err)
	}
	return exists, nil
}
