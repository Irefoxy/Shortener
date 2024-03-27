package postgres

import (
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RepoSuite struct {
	suite.Suite
	pool    *Postgres
	storage DbIFace
}

func (s *RepoSuite) SetupTest() {
	pool, err := pgxmock.NewPool()
	assert.NoError(s.T(), err)
	s.storage = pool
	s.pool = &Postgres{pool: s.storage}
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoSuite))
}
