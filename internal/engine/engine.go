package engine

type Engine interface {
	Get(url string) (string, error)
}
