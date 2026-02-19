package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolAdaptor struct {
	Pool *pgxpool.Pool
}

func (p PoolAdaptor) Exec(ctx context.Context, sql string, arguments ...any) (any, error) {
	tag, err := p.Pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (p PoolAdaptor) Begin(ctx context.Context) (Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return txAdaptor{tx: tx}, nil
}

func (p PoolAdaptor) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}

type txAdaptor struct {
	tx pgx.Tx
}

func (t txAdaptor) Exec(ctx context.Context, sql string, arguments ...any) (any, error) {
	tag, err := t.tx.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (t txAdaptor) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t txAdaptor) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }
