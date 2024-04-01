package postgres

import (
	"Yandex/internal/models"
	"context"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"regexp"
	"testing"
)

// Unit tests for Set and Delete operations are skipped
// due to the complexity involved in testing batch processes.
// pgxmock/v3 does not easily support batch operation verifications,
// requiring a more sophisticated approach.

type Err string

func (e Err) Error() string {
	return string(e)
}

type RepoSuite struct {
	suite.Suite
	pool    pgxmock.PgxPoolIface
	storage *Postgres
}

func (s *RepoSuite) SetupTest() {
	var err error
	s.pool, err = pgxmock.NewPool()
	assert.NoError(s.T(), err)
	s.storage = &Postgres{pool: s.pool}
}

// OK case of 3 elements
func (s *RepoSuite) TestGetAll00() {
	test := struct {
		uuid     string
		expected []models.Entry
	}{
		uuid: "1",
		expected: []models.Entry{
			{
				Id:          "1",
				OriginalUrl: "yandex.com",
				ShortUrl:    "asdfs",
				DeletedFlag: true,
			},
			{
				Id:          "1",
				OriginalUrl: "sber.com",
				ShortUrl:    "reqweq",
				DeletedFlag: false,
			},
			{
				Id:          "1",
				OriginalUrl: "avito.com",
				ShortUrl:    "ggfasa",
				DeletedFlag: false,
			},
		},
	}
	rowsToReturn := pgxmock.NewRows([]string{"original", "short", "deleted"})
	for _, entry := range test.expected {
		rowsToReturn.AddRow(entry.OriginalUrl, entry.ShortUrl, entry.DeletedFlag)
	}

	s.pool.ExpectPing()
	s.pool.ExpectQuery(regexp.QuoteMeta(getAllQuery)).WithArgs(test.uuid).WillReturnRows(rowsToReturn)

	result, err := s.storage.GetAllByUUID(context.Background(), test.uuid)
	s.NoError(err)
	s.ElementsMatch(test.expected, result)
	s.NoError(s.pool.ExpectationsWereMet())
}

// OK case of 0 elements
func (s *RepoSuite) TestGetAll01() {
	rowsToReturn := pgxmock.NewRows([]string{"original", "short", "deleted"})

	s.pool.ExpectPing()
	s.pool.ExpectQuery(regexp.QuoteMeta(getAllQuery)).WithArgs(pgxmock.AnyArg()).WillReturnRows(rowsToReturn)

	result, err := s.storage.GetAllByUUID(context.Background(), "any")
	s.NoError(err)
	s.Nil(result)
	s.NoError(s.pool.ExpectationsWereMet())
}

// Returns Err
func (s *RepoSuite) TestGetAll02() {
	testErr := Err("test")

	s.pool.ExpectPing()
	s.pool.ExpectQuery(regexp.QuoteMeta(getAllQuery)).WithArgs(pgxmock.AnyArg()).WillReturnError(testErr)

	result, err := s.storage.GetAllByUUID(context.Background(), "any")
	s.ErrorIs(err, testErr)
	s.Nil(result)
	s.NoError(s.pool.ExpectationsWereMet())
}

func (s *RepoSuite) TestGet00() {
	rowsToReturn := pgxmock.NewRows([]string{"original", "short", "deleted"})

	s.pool.ExpectPing()
	s.pool.ExpectQuery(regexp.QuoteMeta(getAllQuery)).WithArgs(pgxmock.AnyArg()).WillReturnRows(rowsToReturn)

	result, err := s.storage.GetAllByUUID(context.Background(), "any")
	s.NoError(err)
	s.Nil(result)
	s.NoError(s.pool.ExpectationsWereMet())
}

// OK close()
func (s *RepoSuite) TestClose() {
	s.pool.ExpectClose()
	err := s.storage.Close()
	s.NoError(err)
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoSuite))
}
