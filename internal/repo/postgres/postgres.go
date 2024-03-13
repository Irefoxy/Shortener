package postgres

import (
	"Yandex/internal/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type Postgres struct {
	dsn  string
	conn *pgx.Conn
}

func New(dsn string) *Postgres {
	return &Postgres{dsn: dsn}
}

func (p *Postgres) Get(unit models.ServiceUnit) ([]models.ServiceUnit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := p.conn.QueryRow(ctx, "SELECT original FROM urls WHERE short=$1 and user_id=$2", unit.ShortUrl, unit)
	var shortUrl string
	switch err := row.Scan(&shortUrl); {
	case err == nil:

		unit.ShortUrl = shortUrl
		return []models.ServiceUnit{unit}, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, err
	}
}

func (p *Postgres) Set(units ...models.ServiceUnit) error { // TODO update set for batches
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tag, err := p.conn.Exec(ctx, "INSERT INTO Urls(short, original) VALUES ($1, $2)"+
		"ON CONFLICT(original) DO NOTHING", hash, url)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("CONFLICT")
	}
	return nil
}

func (p *Postgres) Init() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if p.conn, err = pgx.Connect(ctx, p.dsn); err != nil {
		return err
	}
	if err = p.prepareDb(); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) Close() error {
	if p.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.conn.Close(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.conn.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) prepareDb() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	createScript := `
        CREATE TABLE IF NOT EXISTS Urls (
            id SERIAL PRIMARY KEY,
            short TEXT NOT NULL,
            original TEXT UNIQUE NOT NULL
        );
    `
	_, err := p.conn.Exec(ctx, createScript)
	if err != nil {
		return err
	}
	return nil
}
