package postgres

import (
	"context"
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

func (p *Postgres) Get(hash string) (string, bool) {
	//TODO implement me
	panic("implement me")
}

func (p *Postgres) Set(hash, utl string) error {
	//TODO implement me
	panic("implement me")
}

func (p *Postgres) Init() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if p.conn, err = pgx.Connect(ctx, p.dsn); err != nil {
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

	return nil
}
