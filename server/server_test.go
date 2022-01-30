package server

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getServerConfig returns a ServerConfig with a test data directory
func getServerConfig(t *testing.T) *ServerConfig {
	sc, err := NewServerConfig(":8080", "../.testdata/.tmp/", 10, 10, true)
	assert.NoError(t, err, "Error creating server config")
	return sc
}

func IsMtxLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&1 == 1
}

func IsMtxWriteLocked(rw *sync.RWMutex) bool {
	// RWMutex has a "w" sync.Mutex field for write lock
	state := reflect.ValueOf(rw).Elem().FieldByName("w").FieldByName("state")
	return state.Int()&1 == 1
}

func IsMtxReadLocked(rw *sync.RWMutex) bool {
	return reflect.ValueOf(rw).Elem().FieldByName("readerCount").Int() > 0
}

// Test_Mutex tests the mutex test helper functions
func Test_Mutex(t *testing.T) {
	// test helper functions
	for _, test := range []struct {
		description string
		testFunc    func(t *testing.T)
	}{
		{"MutexLocked", func(t *testing.T) {
			m := &sync.Mutex{}
			assert.False(t, IsMtxLocked(m), "Mutex is locked")
			m.Lock()
			assert.True(t, IsMtxLocked(m), "Mutex is not locked")
		}},
		{"RWMutexWriteLocked", func(t *testing.T) {
			rw := &sync.RWMutex{}
			assert.False(t, IsMtxWriteLocked(rw), "RWMutex is write locked")
			rw.Lock()
			assert.True(t, IsMtxWriteLocked(rw), "RWMutex is not write locked")
		}},
		{"RWMutexReadLocked", func(t *testing.T) {
			rw := &sync.RWMutex{}
			assert.False(t, IsMtxReadLocked(rw), "RWMutex is read locked")

			// Lock twice and make sure it's still read locked
			rw.RLock()
			assert.True(t, IsMtxReadLocked(rw), "RWMutex is not read locked")
			rw.RLock()
			assert.True(t, IsMtxReadLocked(rw), "RWMutex is not read locked")

			// read unlock and make sure it's still read locked
			rw.RUnlock()
			assert.True(t, IsMtxReadLocked(rw), "RWMutex is not read locked")

			// read unlock and make sure it's no locker read locked
			rw.RUnlock()
			assert.False(t, IsMtxReadLocked(rw), "RWMutex is read locked")
		}},
	} {
		t.Run(test.description, test.testFunc)
	}
}

// Test_AcquireLock test the locking and unlocking of mutexes
func Test_AcquireLockNonExistingKey(t *testing.T) {
	fileName := "test.txt"
	sc := getServerConfig(t)
	mtx := sc.acquireLock(fileName)
	if _, ok := sc.mtxMap[fileName]; !ok {
		t.Errorf("Mutex not found for file %s", fileName)
	} else {
		if assert.True(t, IsMtxLocked(mtx), "Mutex is not locked") {
			mtx.Unlock()
			assert.False(t, IsMtxLocked(mtx), "Mutex is locked")
		}
	}
}

// Test_AcquireLockExistingKey test the locking of an existing key
func Test_AcquireLockExistingKey(t *testing.T) {
	fileName := "test.txt"
	sc := getServerConfig(t)

	testMutex := &sync.Mutex{}
	sc.mtxMap["test.txt"] = testMutex

	mtx := sc.acquireLock(fileName)
	if _, ok := sc.mtxMap[fileName]; !ok {
		t.Errorf("Mutex not found for file %s", fileName)
	} else {
		if assert.True(t, IsMtxLocked(mtx), "Mutex is not locked") {
			mtx.Unlock()
			assert.False(t, IsMtxLocked(mtx), "Mutex is locked")
			assert.Equal(t, testMutex, mtx, "Mutex is not the same")
		}
	}
}

// Test_ServerConfig_createFile test the creation of a file
func Test_ServerConfig_createFile(t *testing.T) {
	sc := getServerConfig(t)

	fn := "create_file_test.txt"
	data := "test data"

	r := strings.NewReader(data)

	err := sc.createFile(fn, int64(len(data)), r, false)

	if assert.NoError(t, err, "Error creating file") {
		defer os.RemoveAll(sc.DataDir)
		filePath := filepath.Join(sc.DataDir, generateFileName(fn))
		file, err := os.Open(filePath)
		if assert.NoError(t, err, "Error opening file") {
			defer file.Close()

			fs, err := parseFileStore(file)
			assert.NoError(t, err, "Error creating file store")

			dataBytes, err := io.ReadAll(fs)

			assert.NoError(t, err, "Error reading content from file")
			if assert.NoError(t, err, "Error reading file store") {
				// make sure the file is the same
				assert.Equal(t, data, string(dataBytes), "File data is not the same")

			}
		}
	}
}

// Test_ServerConfig_createFile_Overwrite test the overwriting of a file
func Test_ServerConfig_createFile_Overwrite(t *testing.T) {
	sc := getServerConfig(t)

	fn := "create_file_test.txt"
	data := "test data"
	r := strings.NewReader(data)

	err := sc.createFile(fn, int64(len(data)), r, false)

	if !assert.NoError(t, err, "Error creating file") {
		return
	}

	data = "test data2"
	r = strings.NewReader(data)
	err = sc.createFile(fn, int64(len(data)), r, true)

	if assert.NoError(t, err, "Error creating file") {
		defer os.RemoveAll(sc.DataDir)
		filePath := filepath.Join(sc.DataDir, generateFileName(fn))
		file, err := os.Open(filePath)
		if assert.NoError(t, err, "Error opening file") {
			defer file.Close()

			fs, err := parseFileStore(file)
			assert.NoError(t, err, "Error creating file store")

			dataBytes, err := io.ReadAll(fs)

			assert.NoError(t, err, "Error reading content from file")
			if assert.NoError(t, err, "Error reading file store") {
				// make sure the file is the same
				assert.Equal(t, data, string(dataBytes), "File data is not the same")

			}
		}
	}
}

// Test_ServerConfig_createFile_Overwrite_Fails test if overwriting of a file fails
func Test_ServerConfig_createFile_Overwrite_Fails(t *testing.T) {
	sc := getServerConfig(t)

	fn := "create_file_test.txt"
	data := "test data"

	err := sc.createFile(fn, int64(len(data)), strings.NewReader(data), false)
	if !assert.NoError(t, err, "Error creating file") {
		return
	}
	defer os.RemoveAll(sc.DataDir)

	err = sc.createFile(fn, int64(len(data)), strings.NewReader(data), false)

	assert.ErrorIs(t, err, ErrFileAlreadyExists, "No error returned when overwriting a file that already exists")
}

// Test_ServerConfig_deleteFile test the deletion of a file
func Test_ServerConfig_deleteFile(t *testing.T) {
	sc := getServerConfig(t)

	fn := "delete_file_test.txt"
	data := "test data"

	err := sc.createFile(fn, int64(len(data)), strings.NewReader(data), false)
	if !assert.NoError(t, err, "Error creating file") {
		return
	}
	defer os.RemoveAll(sc.DataDir)

	err = sc.deleteFile(fn)
	assert.NoError(t, err, "Error when deleting file")

	exists, err := fileExists(sc.DataDir, fn)
	assert.NoError(t, err, "Error when checking if file exists")
	assert.False(t, exists, "File was not deleted")

	_, hasMtx := sc.mtxMap[fn]
	assert.False(t, hasMtx, "Mutex is still in map after file was deleted")
}

// Test_ServerConfig_getFileList test the listing of files
func Test_ServerConfig_getFileList(t *testing.T) {
	// test list files + delete concurreny
	sc := getServerConfig(t)
	defer os.RemoveAll(sc.DataDir)

	fileList := []string{
		"list_test1.txt",
		"list_test2.txt",
		"list_test3.txt",
	}

	for _, fn := range fileList {
		err := sc.createFile(fn, int64(len(fn)), strings.NewReader(fn), false)
		if !assert.NoError(t, err, "Error creating file") {
			return
		}
	}

	files, err := sc.getFileList(len(fileList))
	if assert.NoError(t, err, "Error listing files") {
		assert.Equal(t, len(fileList), len(files), "File list is not the same")
		for _, fileMeta := range files {
			assert.Contains(t, fileList, fileMeta.FileName, "File list does not contain file")
		}
	}
}
