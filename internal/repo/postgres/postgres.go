package postgres

import (
	"Yandex/internal/models"
	"Yandex/internal/service/gin_srv"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

var _ gin_srv.Repo = (*Postgres)(nil)

type Postgres struct {
	dsn  string
	conn *pgx.Conn
}

func New(dsn string) *Postgres {
	return &Postgres{dsn: dsn}
}

func (p *Postgres) GetAllUrls(ctx context.Context, unit models.ServiceUnit) (result []models.ServiceUnit, err error) {
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, _ := p.conn.Query(newCtx, "SELECT original, short FROM urls WHERE uuid=$1", unit.Id)
	var shortUrl, originalUrl string
	_, err = pgx.ForEachRow(rows, []any{&originalUrl, &shortUrl}, func() error {
		result = append(result, models.ServiceUnit{
			Id:          unit.Id,
			OriginalUrl: originalUrl,
			ShortUrl:    shortUrl,
		})
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return
}

func (p *Postgres) SetBatch(ctx context.Context, units []models.ServiceUnit) (err error) {
	newCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	copyCount, err := p.conn.CopyFrom(
		newCtx,
		pgx.Identifier{"Urls"},
		[]string{"uuid", "short", "original"},
		pgx.CopyFromSlice(len(units), func(i int) ([]any, error) {
			return []any{units[i].Id, units[i].ShortUrl, units[i].OriginalUrl}, nil
		}),
	)
	if err != nil {
		return err
	}
	if int(copyCount) != len(units) {
		return models.ErrorConflict
	}
	return
}

func (p *Postgres) Get(ctx context.Context, unit models.ServiceUnit) (*models.ServiceUnit, error) {
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	row := p.conn.QueryRow(newCtx, "SELECT original FROM urls WHERE short=$1 and uuid=$2", unit.ShortUrl, unit.Id)
	var shortUrl string
	switch err := row.Scan(&shortUrl); {
	case err == nil:
		unit.ShortUrl = shortUrl
		return &unit, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, err
	}
}

func (p *Postgres) Set(ctx context.Context, units models.ServiceUnit) error { // TODO update set for batches
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tag, err := p.conn.Exec(newCtx, "INSERT INTO Urls(uuid, short, original) VALUES ($1, $2, $3)"+
		"ON CONFLICT(uuid, original) DO NOTHING", units.Id, units.ShortUrl, units.OriginalUrl)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return models.ErrorConflict
	}
	return nil
}

func (p *Postgres) Init(ctx context.Context) error {
	var err error
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if p.conn, err = pgx.Connect(newCtx, p.dsn); err != nil {
		return err
	}
	if err = p.prepareDb(); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) Close(ctx context.Context) error {
	if p.conn != nil {
		newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := p.conn.Close(newCtx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.conn.Ping(newCtx); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) prepareDb(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	createScript := `
        CREATE TABLE IF NOT EXISTS Urls (
            uuid TEXT NOT NULL,
            short TEXT NOT NULL,
            original TEXT NOT NULL,
            UNIQUE (uuid, original)
        );
    `
	_, err := p.conn.Exec(newCtx, createScript)
	if err != nil {
		return err
	}
	return nil
}
