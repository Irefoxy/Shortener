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

const (
	getAllQuery = `SELECT original, short, deleted FROM urls WHERE uuid=$1`
	setQuery    = `INSERT INTO Urls(uuid, short, original) VALUES ($1, $2, $3)
				ON CONFLICT(uuid, original) DO NOTHING`
	deleteQuery = `UPDATE urls SET deleted = TRUE WHERE id = $1 and short = $2`
	getQuery    = `SELECT original, deleted FROM urls WHERE short=$1 and uuid=$2`
)

type DbIFace interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type Postgres struct {
	dsn  string
	pool DbIFace
}

func New(dsn string) *Postgres {
	return &Postgres{
		dsn:  dsn,
		pool: (*pgxpool.Pool)(nil),
	}
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
	newCtx, cancel := prepareContext(ctx, 5)
	defer cancel()
	return p.sendGetAllQuery(newCtx, getAllQuery, uuid)
}

func (p *Postgres) sendGetAllQuery(newCtx context.Context, script string, uuid string) (result []models.Entry, err error) {
	if err := p.Ping(newCtx); err != nil {
		return nil, err
	}
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

func (p *Postgres) Set(ctx context.Context, entries []models.Entry) (int, error) {
	createBatch := func() (batch *pgx.Batch) {
		batch = new(pgx.Batch)
		for _, entry := range entries {
			batch.Queue(setQuery, entry.Id, entry.ShortUrl, entry.OriginalUrl)
		}
		return
	}
	count, err := p.sendBatch(ctx, createBatch)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *Postgres) Delete(ctx context.Context, entries []models.Entry) error {
	createBatch := func() (batch *pgx.Batch) {
		batch = new(pgx.Batch)
		for _, entry := range entries {
			batch.Queue(deleteQuery, entry.Id, entry.ShortUrl)
		}
		return
	}
	_, err := p.sendBatch(ctx, createBatch)
	return err
}

func (p *Postgres) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

func (p *Postgres) Get(ctx context.Context, entry models.Entry) (*models.Entry, error) {
	newCtx, cancel := prepareContext(ctx, 5)
	defer cancel()
	if err := p.Ping(newCtx); err != nil {
		return nil, err
	}
	row := p.pool.QueryRow(newCtx, getQuery, entry.ShortUrl, entry.Id)
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
	batch := prepareBatch()
	count, err := p.handleBatch(newCtx, batch)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *Postgres) handleBatch(newCtx context.Context, batch *pgx.Batch) (int, error) {
	br := p.pool.SendBatch(newCtx, batch)
	defer br.Close()
	count, err := execBatch(br, batch.Len())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func execBatch(br pgx.BatchResults, batchLen int) (int, error) {
	numberOfAffectedRows := 0
	for i := 0; i < batchLen; i++ {
		tag, err := br.Exec()
		if err != nil {
			return 0, err
		}
		numberOfAffectedRows += int(tag.RowsAffected())
	}
	return numberOfAffectedRows, nil
}
