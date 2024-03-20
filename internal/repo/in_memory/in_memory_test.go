package in_memory

import (
	"Yandex/internal/models"
	"Yandex/internal/repo/in_memory/mocks"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

type TestCase struct {
	Name          string
	SetupMock     func(*mocks.MockFileStorage[models.Entry])
	Expected      []models.Entry
	ExpectedError error
}

func TestInit(t *testing.T) {
	expected := []models.Entry{
		{
			Id:          "1",
			OriginalUrl: "1234",
			ShortUrl:    "4321",
		},
		{
			Id:          "1",
			OriginalUrl: "2345",
			ShortUrl:    "5432",
		},
	}
	testCases := []TestCase{
		{
			Name: "Open error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().Open().Return(models.ErrorFileNameNotGiven)
			},
			ExpectedError: models.ErrorFileNameNotGiven,
		},
		{
			Name: "Load error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().Open().Return(nil)
				s.EXPECT().LoadAll().Return(nil, models.ErrorFileNotOpened)
			},
			ExpectedError: models.ErrorFileNotOpened,
		},
		{
			Name: "OK",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().Open().Return(nil)
				s.EXPECT().LoadAll().Return(expected, nil)
			},
			Expected:      expected,
			ExpectedError: nil,
		},
	}

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockFileStorage[models.Entry](ctrl)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			asrt := assert.New(t)
			testCase.SetupMock(storage)
			repo := New(storage)
			err := repo.ConnectStorage(context.Background())
			asrt.ErrorIs(err, testCase.ExpectedError)
			if err == nil {
				data, err := repo.GetAllUrlsByUUID(context.Background(), models.Entry{Id: "1"})
				asrt.NoError(err)
				asrt.ElementsMatch(data, expected)
			}
		})
	}
}

func TestSet(t *testing.T) {
	expected := []models.Entry{
		{
			Id:          "1",
			OriginalUrl: "1234",
			ShortUrl:    "4321",
		},
		{
			Id:          "1",
			OriginalUrl: "2345",
			ShortUrl:    "5432",
		},
		{
			Id:          "1",
			OriginalUrl: "1345",
			ShortUrl:    "1432",
		},
	}
	testCases := []TestCase{
		{
			Name: "FileStorage not opened",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(false)
			},
			Expected:      expected[:1],
			ExpectedError: nil,
		},
		{
			Name: "Write error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(true)
				s.EXPECT().Write(&expected[1]).Return(models.ErrorFileNotOpened)
			},
			Expected:      expected[1:2],
			ExpectedError: models.ErrorFileNotOpened,
		},
		{
			Name: "Conflict error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
			},
			Expected:      expected[:1],
			ExpectedError: models.ErrorConflict,
		},
		{
			Name: "OK",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(true)
				s.EXPECT().Write(&expected[2]).Return(nil)
			},
			Expected:      expected[2:3],
			ExpectedError: nil,
		},
	}

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockFileStorage[models.Entry](ctrl)
	repo := New(storage)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			asrt := assert.New(t)
			testCase.SetupMock(storage)
			err := repo.Set(context.Background(), models.Entry{
				Id:          testCase.Expected[0].Id,
				OriginalUrl: testCase.Expected[0].OriginalUrl,
				ShortUrl:    testCase.Expected[0].ShortUrl,
			})
			asrt.ErrorIs(err, testCase.ExpectedError)
			data, err := repo.Get(context.Background(), models.Entry{
				Id:          testCase.Expected[0].Id,
				OriginalUrl: testCase.Expected[0].OriginalUrl,
			})
			asrt.NoError(err)
			asrt.Equal(data.ShortUrl, testCase.Expected[0].ShortUrl)
		})
	}
}

func TestBatch(t *testing.T) {
	expected := []models.Entry{
		{
			Id:          "1",
			OriginalUrl: "1234",
			ShortUrl:    "4321",
		},
		{
			Id:          "1",
			OriginalUrl: "2345",
			ShortUrl:    "5432",
		},
		{
			Id:          "1",
			OriginalUrl: "1345",
			ShortUrl:    "1432",
		},
	}
	testCases := []TestCase{
		{
			Name: "FileStorage not opened",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(false)
			},
			Expected:      expected[:1],
			ExpectedError: nil,
		},
		{
			Name: "Write error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(true)
				s.EXPECT().Write(&expected[1]).Return(models.ErrorFileNotOpened)
			},
			Expected:      expected[1:2],
			ExpectedError: models.ErrorFileNotOpened,
		},
		{
			Name: "Conflict error",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
			},
			Expected:      expected[:1],
			ExpectedError: models.ErrorConflict,
		},
		{
			Name: "OK",
			SetupMock: func(s *mocks.MockFileStorage[models.Entry]) {
				s.EXPECT().IsOpened().Return(true)
				s.EXPECT().Write(&expected[2]).Return(nil)
			},
			Expected:      expected[2:3],
			ExpectedError: nil,
		},
	}

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockFileStorage[models.Entry](ctrl)
	repo := New(storage)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			asrt := assert.New(t)
			testCase.SetupMock(storage)
			err := repo.Set(context.Background(), models.Entry{
				Id:          testCase.Expected[0].Id,
				OriginalUrl: testCase.Expected[0].OriginalUrl,
				ShortUrl:    testCase.Expected[0].ShortUrl,
			})
			asrt.ErrorIs(err, testCase.ExpectedError)
			data, err := repo.Get(context.Background(), models.Entry{
				Id:          testCase.Expected[0].Id,
				OriginalUrl: testCase.Expected[0].OriginalUrl,
			})
			asrt.NoError(err)
			asrt.Equal(data.ShortUrl, testCase.Expected[0].ShortUrl)
		})
	}
}
