package postgres

import (
	"Yandex/internal/models"
	"Yandex/internal/services/shortener"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var _ shortener.Repo = (*Postgres)(nil)

type DbIFace interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Postgres struct {
	dsn  string
	pool DbIFace
}

func New(dsn string) *Postgres {
	return &Postgres{dsn: dsn,
		pool: (*pgxpool.Pool)(nil)}
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
	newCtx, cancel := prepareContext(ctx, 5)
	defer cancel()
	return p.sendGetAllQuery(newCtx, script, uuid)
}

func (p *Postgres) sendGetAllQuery(newCtx context.Context, script string, uuid string) (result []models.Entry, err error) {
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

func (p *Postgres) Set(ctx context.Context, entries []models.Entry) error {
	script := `INSERT INTO Urls(uuid, short, original) VALUES ($1, $2, $3)
		ON CONFLICT(uuid, original) DO NOTHING`
	count, err := p.sendBatch(ctx, func() (batch *pgx.Batch) {
		batch = new(pgx.Batch)
		for _, entry := range entries {
			batch.Queue(script, entry.Id, entry.ShortUrl, entry.OriginalUrl)
		}
		return
	})
	if err != nil {
		return err
	}
	if count != len(entries) {
		return models.ErrorConflict
	}
	return nil
}

func (p *Postgres) Delete(ctx context.Context, entries []models.Entry) error {
	script := `UPDATE urls SET deleted = TRUE WHERE id = $1 and short = $2`
	_, err := p.sendBatch(ctx, func() (batch *pgx.Batch) {
		batch = new(pgx.Batch)
		for _, entry := range entries {
			batch.Queue(script, entry.Id, entry.ShortUrl)
		}
		return
	})
	return err
}

func (p *Postgres) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

func (p *Postgres) Get(ctx context.Context, entry models.Entry) (*models.Entry, error) {
	script := `SELECT original, deleted FROM urls WHERE short=$1 and uuid=$2`
	newCtx, cancel := prepareContext(ctx, 5)
	defer cancel()
	if err := p.Ping(newCtx); err != nil {
		return nil, err
	}
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
	newCtx, cancel := prepareContext(ctx, 2)
	defer cancel()
	if err := p.pool.Ping(newCtx); err != nil {
		return err
	}
	return nil
}

func prepareContext(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, duration*time.Second)
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

func (p *Postgres) sendBatch(ctx context.Context, prepareBatch func() *pgx.Batch) (int, error) {
	newCtx, cancel := prepareContext(ctx, 5)
	defer cancel()
	if err := p.Ping(newCtx); err != nil {
		return 0, err
	}
	tx, err := p.pool.Begin(newCtx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(newCtx)
	batch := prepareBatch()
	br := tx.SendBatch(newCtx, batch)
	defer br.Close()
	count, err := execBatch(batch, br)
	if err != nil {
		return 0, err
	}
	err = tx.Commit(newCtx)
	return count, err
}

func execBatch(batch *pgx.Batch, br pgx.BatchResults) (int, error) {
	numberOfAffectedRows := 0
	for i := 0; i < batch.Len(); i++ {
		tag, err := br.Exec()
		if err != nil {
			return 0, err
		}
		numberOfAffectedRows += int(tag.RowsAffected())
	}
	return numberOfAffectedRows, nil
}
