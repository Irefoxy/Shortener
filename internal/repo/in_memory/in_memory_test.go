package in_memory

import (
	"Yandex/internal/models"
	"Yandex/internal/repo/in_memory/mocks"
	"context"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type testErr string

func (e testErr) Error() string {
	return string(e)
}

const testError testErr = "TEST ERROR"

var entries = []models.Entry{
	{
		Id:          "1",
		OriginalUrl: "yandex.com",
		ShortUrl:    "yan",
		DeletedFlag: false,
	},
	{
		Id:          "1",
		OriginalUrl: "sber.com",
		ShortUrl:    "sb",
		DeletedFlag: false,
	},
	{
		Id:          "2",
		OriginalUrl: "sber.com",
		ShortUrl:    "sb",
		DeletedFlag: false,
	},
}

type RepoSuite struct {
	suite.Suite
	repo    *InMemory
	storage *mocks.MockFileStorage[models.Entry]
}

func (s *RepoSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.storage = mocks.NewMockFileStorage[models.Entry](ctrl)
	s.repo = New(s.storage)
}

func (s *RepoSuite) TestExportImport() {
	s.repo.importData(entries)
	data := s.repo.exportData()
	s.ElementsMatch(data, entries)
}

func (s *RepoSuite) TestFailedConnectAndClose() {
	s.storage.EXPECT().LoadAll().Return(nil, testError)
	s.storage.EXPECT().Dump(gomock.Any()).Return(testError)
	s.Assert().ErrorIs(s.repo.ConnectStorage(), testError)
	s.Assert().ErrorIs(s.repo.Close(), testError)
}

func (s *RepoSuite) TestOKConnect() {
	s.storage.EXPECT().LoadAll().Return(nil, nil)
	s.Assert().NoError(s.repo.ConnectStorage())
}

func (s *RepoSuite) TestSetAndGet00() {
	num, err := s.repo.Set(context.Background(), entries)
	s.Equal(len(entries), num)
	s.NoError(err)
	got, err := s.repo.Get(context.Background(), models.Entry{
		Id:       entries[0].Id,
		ShortUrl: entries[0].ShortUrl,
	})
	s.NoError(err)
	s.Assert().Contains(entries, *got)
}

func (s *RepoSuite) TestSetAndGet01() {
	num, err := s.repo.Set(context.Background(), entries)
	s.Equal(len(entries), num)
	s.NoError(err)
	num, err = s.repo.Set(context.Background(), entries)
	s.Equal(0, num)
}

func (s *RepoSuite) TestSetAndGet02() {
	got, err := s.repo.Get(context.Background(), models.Entry{
		Id:       entries[0].Id,
		ShortUrl: entries[0].ShortUrl,
	})
	s.NoError(err)
	s.Nil(got)
}

func (s *RepoSuite) TestSetAndGet03() {
	num, err := s.repo.Set(context.Background(), entries)
	s.Equal(len(entries), num)
	s.NoError(err)
	err = s.repo.Delete(context.Background(), entries)
	s.NoError(err)
	num, err = s.repo.Set(context.Background(), entries)
	s.Equal(len(entries), num)
	s.NoError(err)
}

func (s *RepoSuite) TestDelete00() {
	expectedEntry := entries[0]
	expectedEntry.DeletedFlag = true
	num, err := s.repo.Set(context.Background(), entries)
	s.Equal(len(entries), num)
	s.NoError(err)
	err = s.repo.Delete(context.Background(), entries)
	s.NoError(err)
	got, err := s.repo.Get(context.Background(), models.Entry{
		Id:       entries[0].Id,
		ShortUrl: entries[0].ShortUrl,
	})
	s.NoError(err)

	s.Equal(expectedEntry, *got)
}

func (s *RepoSuite) TestGetAll00() {
	entriesForUUID, err := s.repo.GetAllByUUID(context.Background(), "1")
	s.NoError(err)
	s.Nil(entriesForUUID)
}

func (s *RepoSuite) TestGetAll01() {
	_, err := s.repo.Set(context.Background(), entries)
	s.NoError(err)
	entriesForUUID, err := s.repo.GetAllByUUID(context.Background(), "1")
	s.NoError(err)
	s.ElementsMatch(entries[:2], entriesForUUID)
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoSuite))
}
