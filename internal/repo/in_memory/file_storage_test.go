package in_memory

import (
	"github.com/stretchr/testify/assert"
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
	testSrcWrongFile = testDir + "wrong.db"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		err      error
		data     []Sample
	}{
		{"File name empty", "", nil, nil},
		{"Not exist", testFileName, os.ErrNotExist, nil},
		{"Not exist", testSrcEmptyFile, nil, nil},
		{"OK", testSrcFile, nil, []Sample{
			{
				Id:     1,
				String: "123",
			},
			{
				Id:     2,
				String: "234",
			},
		}},
		{"OK", testSrcWrongFile, nil, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewJSONFileStorage[Sample](test.fileName)
			data, err := storage.LoadAll()
			assert.ErrorIs(t, err, test.err)
			assert.ElementsMatch(t, data, test.data)
		})
	}
}

/*func TestWrite(t *testing.T) {
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
}*/
