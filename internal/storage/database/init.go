package database

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/url"
	"time"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Database struct {
	db *pgxpool.Pool
}

func (d *Database) GetDb() DBTX {
	return d.db
}

func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := d.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	txObj := tx
	return txObj, nil
}

func NewDatabase(ctx context.Context, cfg config.Database) *Database {

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	q := dsn.Query()

	q.Add("sslmode", "disable")

	dsn.RawQuery = q.Encode()
	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = 15
	poolConfig.MaxConnIdleTime = time.Minute * 10

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	return &Database{
		db: pgxPool,
	}
}
