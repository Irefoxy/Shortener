package postgres

import (
	"Yandex/internal/models"
	"Yandex/internal/repo/in_memory/mocks"
	"github.com/stretchr/testify/suite"
)

type RepoSuite struct {
	suite.Suite
	repo    *Postgres
	storage *mocks.MockFileStorage[models.Entry]
}

func (s *RepoSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.storage = mocks.NewMockFileStorage[models.Entry](ctrl)
	s.repo = New(s.storage)
}
