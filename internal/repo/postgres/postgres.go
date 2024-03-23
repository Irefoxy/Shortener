package postgres

import (
	"Yandex/internal/models"
	"Yandex/internal/services/shortener"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var _ shortener.Repo = (*Postgres)(nil)

type Postgres struct {
	dsn  string
	pool *pgxpool.Pool
}

func (p *Postgres) ConnectStorage() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if p.pool, err = pgxpool.New(ctx, p.dsn); err != nil {
		return err
	}
	if err = p.prepareDb(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) GetAllByUUID(ctx context.Context, uuid string) (result []models.Entry, err error) {
	script := `SELECT original, short, deleted FROM urls WHERE uuid=$1`
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := p.pool.Query(newCtx, script, uuid)
	if err != nil {
		return nil, err
	}
	var original, short string
	var deleted bool
	_, err = pgx.ForEachRow(rows, []any{&original, &short, &deleted}, func() error {
		result = append(result, models.Entry{
			Id:          uuid,
			OriginalUrl: original,
			ShortUrl:    short,
			DeletedFlag: deleted,
		})
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return
}

func (p *Postgres) Set(ctx context.Context, entries []models.Entry) (err error) {
	script := `INSERT INTO Urls(uuid, short, original) VALUES ($1, $2, $3)
		ON CONFLICT(uuid, original) DO NOTHING`
	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := p.Ping(newCtx); err != nil {
		return err
	}
	tx, err := p.pool.Begin(newCtx)
	if err != nil {
		return
	}
	defer tx.Rollback(newCtx)
	batch := pgx.Batch{}
	for _, entry := range entries {
		batch.Queue(script, entry.Id, entry.ShortUrl, entry.OriginalUrl)
	}
	br := tx.SendBatch(newCtx, &batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		tag, err := br.Exec()
		if err != nil {
			return
		}
		if tag.RowsAffected() != 1 {
			err = models.ErrorConflict
		}
	}
	err = tx.Commit(newCtx)
	return
}

func (p *Postgres) Delete(ctx context.Context, entries []models.Entry) error {
	script := `UPDATE urls SET deleted = TRUE WHERE id = $1`
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.Ping(newCtx); err != nil {
		return err
	}
	tx, err := p.pool.Begin(newCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(newCtx)
	batch := pgx.Batch{}
	for _, entry := range entries {
		batch.Queue(script, entry.Id)
	}
	br := tx.SendBatch(newCtx, &batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}
	err = tx.Commit(newCtx)
	return nil
}

func (p *Postgres) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

func New(dsn string) *Postgres {
	return &Postgres{dsn: dsn}
}

func (p *Postgres) Get(ctx context.Context, entry models.Entry) (*models.Entry, error) {
	script := `SELECT original, deleted FROM urls WHERE short=$1 and uuid=$2`
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	row := p.pool.QueryRow(newCtx, script, entry.ShortUrl, entry.Id)
	var original string
	var deleted bool
	switch err := row.Scan(&original, &deleted); {
	case err == nil:
		entry.OriginalUrl = original
		entry.DeletedFlag = deleted
		return &entry, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, err
	}
}

func (p *Postgres) Ping(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.pool.Ping(newCtx); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) prepareDb(ctx context.Context) error {
	createScript := `
        CREATE TABLE IF NOT EXISTS Urls (
            uuid TEXT NOT NULL,
            short TEXT NOT NULL,
            original TEXT NOT NULL,
            deleted BOOL NOT NULL DEFAULT FALSE
            UNIQUE (uuid, original)
        );
    `
	_, err := p.pool.Exec(ctx, createScript)
	if err != nil {
		return err
	}
	return nil
}
