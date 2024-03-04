package repo_impl

type RepoImpl struct {
	data map[string]string
}

func New() *RepoImpl {
	return &RepoImpl{make(map[string]string)}
}

func (s *RepoImpl) Get(hash string) (string, bool) {
	v, ok := s.data[hash]
	return v, ok
}

func (s *RepoImpl) Set(hash, url string) error {
	s.data[hash] = url
	return nil
}
