package in_memory

import (
	"Yandex/internal/models"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"testing"
)

type Sample struct {
	Id     int    `json:"id"`
	String string `json:"string"`
}

const (
	testDir          = "./test_files/"
	testFileName     = "./test"
	testSrcFile      = testDir + "two_units.db"
	testSrcEmptyFile = testDir + "empty.db"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		err      error
	}{
		{"File name empty", "", models.ErrorFileNameNotGiven},
		{"Creating", testFileName, nil},
		{"Opening twice", testFileName, models.ErrorFileAlreadyOpened},
	}
	// File name is empty
	number := 0
	t.Run(tests[number].name, func(t *testing.T) {
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		assert.ErrorIs(t, storage.Open(), tests[number].err)
	})

	// Creating file
	number = 1
	t.Run(tests[number].name, func(t *testing.T) {
		asrt := assert.New(t)
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		_, err := os.Stat(tests[number].fileName)
		asrt.ErrorIs(err, fs.ErrNotExist)
		asrt.NoError(storage.Open())
		asrt.NoError(storage.Close())
		_, err = os.Stat(tests[number].fileName)
		asrt.NoError(err)
		asrt.NoError(os.Remove(tests[number].fileName))
	})

	// Trying to open file twice
	number = 2
	t.Run(tests[number].name, func(t *testing.T) {
		asrt := assert.New(t)
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		asrt.NoError(storage.Open())
		asrt.ErrorIs(storage.Open(), tests[number].err)
		asrt.NoError(storage.Close())
		asrt.NoError(os.Remove(tests[number].fileName))
	})
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
		content  Sample
		err      error
	}{
		{"File name empty", "", "", Sample{}, models.ErrorFileNotOpened},
		{"OK", testFileName, `{"id":1,"string":"123"}` + "\n", Sample{
			Id:     1,
			String: "123",
		}, nil},
	}
	// Trying to write without Open()
	number := 0
	t.Run(tests[number].name, func(t *testing.T) {
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		assert.ErrorIs(t, storage.Write(&tests[number].content), tests[number].err)
	})

	number = 1
	t.Run(tests[number].name, func(t *testing.T) {
		asrt := assert.New(t)
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		asrt.NoError(storage.Open())
		asrt.NoError(storage.Write(&tests[number].content))
		data, err := os.ReadFile(tests[number].fileName)
		asrt.NoError(err)
		asrt.Equal(string(data), tests[number].expected)
		asrt.NoError(os.Remove(tests[number].fileName))
		asrt.NoError(storage.Close())
		asrt.NoError(storage.Close())
	})
}

func TestLoadAll(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected []Sample
		err      error
	}{
		{"File name empty", "", nil, models.ErrorFileNotOpened},
		{"OK", testSrcFile, []Sample{
			{
				Id:     1,
				String: "123",
			},
			{
				Id:     2,
				String: "234",
			},
		}, nil},
		{"Empty src file", testSrcEmptyFile, nil, nil},
	}
	number := 0
	t.Run(tests[number].name, func(t *testing.T) {
		storage := NewJSONFileStorage[Sample](tests[number].fileName)
		_, err := storage.LoadAll()
		assert.ErrorIs(t, err, tests[number].err)
	})

	for number := 1; number < 3; number++ {
		t.Run(tests[number].name, func(t *testing.T) {
			asrt := assert.New(t)
			storage := NewJSONFileStorage[Sample](tests[number].fileName)
			asrt.NoError(storage.Open())
			data, err := storage.LoadAll()
			asrt.NoError(err)
			asrt.Equal(data, tests[number].expected)
			asrt.NoError(storage.Close())
		})
	}
}
