package in_memory

import (
	"errors"
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
		{"Empty", testSrcEmptyFile, nil, nil},
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
		{"Wrong", testSrcWrongFile, errors.New("invalid character '{' after top-level value"), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewJSONFileStorage[Sample](test.fileName)
			data, err := storage.LoadAll()
			if test.err != nil {
				assert.Error(t, err)
				assert.ErrorAs(t, err, &test.err)
			} else {
				assert.NoError(t, err)
			}
			assert.ElementsMatch(t, data, test.data)
		})
	}
}

func TestDump(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
		content  []Sample
		err      error
	}{
		{"File name empty", "", "", nil, nil},
		{"Create", testFileName, "[{\"id\":1,\"string\":\"123\"},{\"id\":2,\"string\":\"234\"}]\012", []Sample{
			{
				Id:     1,
				String: "123",
			},
			{
				Id:     2,
				String: "234",
			},
		}, nil},
		{"Overwrite", testFileName, "[{\"id\":3,\"string\":\"555\"}]\012", []Sample{
			{
				Id:     3,
				String: "555",
			},
		}, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewJSONFileStorage[Sample](test.fileName)
			err := storage.Dump(test.content)
			if test.err != nil {
				assert.Error(t, err)
				assert.ErrorAs(t, err, &test.err)
			} else {
				assert.NoError(t, err)
			}
			if test.expected != "" {
				data, err := os.ReadFile(testFileName)
				assert.NoError(t, err)
				assert.Equal(t, string(data), test.expected)
			}
		})
	}
	assert.NoError(t, os.Remove(testFileName))
}
