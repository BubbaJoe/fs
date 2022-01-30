package server

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// createTestDir creates a test directory.
func createTestDir(t *testing.T) string {
	// mkdir directory with permissions: rwx-rwx-rwx
	dir := "../.testdata/.tmp/"
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// cleanUpTestDir removes the test directory.
func cleanUpTestDir(t *testing.T, dir string) {
	// remove directory
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_FileStoreCreate(t *testing.T) {
	dir := createTestDir(t)
	defer cleanUpTestDir(t, dir)

	store := &FileStore{
		Version:   DefaultVersion,
		FileName:  "test.txt",
		DataSize:  4,
		CreatedAt: time.Now(),
		Reader:    strings.NewReader("data"),
	}

	err := store.createFileAt(dir, false)
	if err != nil {
		t.Error(err)
	}
}

func Test_FileStoreOverwrite(t *testing.T) {
	dir := createTestDir(t)
	defer cleanUpTestDir(t, dir)

	store := &FileStore{
		Version:   DefaultVersion,
		FileName:  "test.txt",
		DataSize:  4,
		CreatedAt: time.Now(),
		Reader:    strings.NewReader("data"),
	}

	err := store.createFileAt(dir, false)
	if assert.NoError(t, err, "Error creating file") {

		store2 := &FileStore{
			Version:   DefaultVersion,
			FileName:  "test.txt",
			DataSize:  8,
			CreatedAt: time.Now(),
			Reader:    strings.NewReader("new data"),
		}

		err := store2.createFileAt(dir, true)
		assert.NoError(t, err, "Error overwriting file")
	}
}

// Test_FileStoreParse tests the parsing of a file store.
func Test_FileStoreParse(t *testing.T) {
	data := "zxcasd asdaxz"
	store := &FileStore{
		Version:   DefaultVersion,
		FileName:  "test.txt",
		DataSize:  int64(len(data)),
		CreatedAt: time.Now(),
		Reader:    strings.NewReader(data),
	}

	writer := bytes.NewBuffer(make([]byte, 0))
	err := store.writeFileStore(writer)
	if !assert.NoError(t, err, "could not write to buffer") {
		return
	}

	reader := bytes.NewBuffer(writer.Bytes())

	// Parse the store
	testStore, err := parseFileStore(reader)
	if assert.NoError(t, err, "could not parse buffer") {
		assert.Equal(t, store.Version, testStore.Version)
		assert.Equal(t, store.FileName, testStore.FileName)
		assert.Equal(t, store.CreatedAt.UnixMilli(),
			testStore.CreatedAt.UnixMilli())
		assert.Equal(t, store.DataSize, testStore.DataSize)
		testStoreData, err := io.ReadAll(testStore)
		if assert.NoError(t, err, "could not read test store") {
			assert.Equal(t, data, string(testStoreData),
				"Data not equal to content")
		}
	}
}

// TODO: check on a lower level write and parse
