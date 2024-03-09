package repo

type Repo interface {
	Get(hash string) (string, bool)
	Set(hash, utl string) error
	Init() error
	Close()
}
