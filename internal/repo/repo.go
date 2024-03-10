package repo

type Repo interface {
	Get(hash string) (string, error)
	Set(hash, utl string) error
	Init() error
	Close() error
}

type DbRepo interface {
	Repo
	Ping() error
}
