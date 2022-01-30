package server

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	. "fs-store/types"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ServerConfig is the configuration for server properties
type ServerConfig struct {
	Address  string
	DataDir  string
	LogLevel string

	MaxFileSize int64
	MaxListSize int

	mapLock *sync.RWMutex
	mtxMap  map[string]*sync.Mutex
}

// Define Errors
var (
	// ErrFileAlreadyExists is returned when a file already exists
	ErrFileAlreadyExists = errors.New("file already exists")
)

func NewServerConfig(address, dataDir string, maxFileSize int64, maxListSize int, makeDir bool) (*ServerConfig, error) {
	if makeDir {
		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &ServerConfig{
		DataDir:     dataDir,
		Address:     address,
		MaxFileSize: maxFileSize,
		MaxListSize: maxListSize,
		mapLock:     new(sync.RWMutex),
		mtxMap:      make(map[string]*sync.Mutex, maxListSize),
	}, nil
}

func (sc *ServerConfig) StartServer() error {
	e := echo.New()

	// Hide initial messages
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetHeader("${time_rfc3339} ${remote_ip} ${method} ${uri} ${status} ${latency_human}")

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, error={${error}}\n",
	}))
	// e.Use(middleware.Recover())

	// Get File List
	e.GET("/files", listFilesRoute(sc))

	// Update File
	e.POST("/files", uploadFileRoute(sc))

	// Delete File
	e.DELETE("/files", deleteFileRoute(sc))

	// Start server
	return e.Start(sc.Address)
}

// acquireLock acquires a lock for a file, and return that lock
func (sc *ServerConfig) acquireLock(keyName string) *sync.Mutex {
	// Acquire read lock to check whether mutext exists
	sc.mapLock.RLock()
	if keyNameMtx, ok := sc.mtxMap[keyName]; !ok {
		// release read lock so write lock can be acquired
		sc.mapLock.RUnlock()

		// acquire lock, then check if mutex exists again (double-checked locking)
		sc.mapLock.Lock()
		if _, ok := sc.mtxMap[keyName]; !ok {
			newMutex := new(sync.Mutex)
			newMutex.Lock()
			sc.mtxMap[keyName] = newMutex
		}
		sc.mapLock.Unlock()
	} else {
		keyNameMtx.Lock()
	}

	return sc.mtxMap[keyName]
}

// removeLock removes the lock from the map, and returns the locked map lock
func (sc *ServerConfig) removeLock(keyName string) *sync.RWMutex {
	// Acquire read lock to check whether mutext exists
	sc.mapLock.RLock()
	if _, ok := sc.mtxMap[keyName]; ok {
		// release read lock so write lock can be acquired
		sc.mapLock.RUnlock()

		// acquire lock, then check if mutex exists again (double-checked locking)
		sc.mapLock.Lock()
		delete(sc.mtxMap, keyName)

	} else {
		// release (map) read lock after lock is acquired
		sc.mapLock.RUnlock()
	}

	return sc.mapLock
}

func (sc *ServerConfig) createFile(fileName string, size int64, data io.Reader, overwrite bool) error {
	// create file store
	store := &FileStore{
		Version:  DefaultVersion,
		FileName: fileName,
		Reader:   data,
		DataSize: size,
	}

	mutex := sc.acquireLock(fileName)
	defer mutex.Unlock()
	// After acquiring lock, check if file exists (double-checked locking)
	if exists, err := fileExists(sc.DataDir, fileName); err != nil {
		return err
	} else if exists && !overwrite {
		return ErrFileAlreadyExists
	}
	return store.createFileAt(sc.DataDir, overwrite)
}

func (sc *ServerConfig) deleteFile(fileName string) error {

	mapMutex := sc.removeLock(fileName)
	defer mapMutex.Unlock()

	// Create file from file store
	err := deleteFileAt(sc.DataDir, fileName)
	return err
}

//
func (sc *ServerConfig) getFileList(limit int) ([]FileResponse, error) {
	entries, err := os.ReadDir(sc.DataDir)

	if err != nil {
		return nil, err
	}

	files := make([]FileResponse, 0)
	for i, entry := range entries {
		if i >= limit {
			break
		}

		entryName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(entryName, ".fs") {
			continue
		}

		m := sc.acquireLock(entryName)
		path := filepath.Join(sc.DataDir, entryName)
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)

		// We acquired lock after reading director,
		// 	the file might not exist anymore, in this case we skip it
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			m.Unlock()
			return nil, err
		}
		store, err := parseFileStore(file)
		if err != nil {
			m.Unlock()
			return nil, err
		}

		files = append(files, FileResponse{
			FileName:  store.FileName,
			FileSize:  store.DataSize,
			CreatedAt: store.CreatedAt,
		})

		m.Unlock()
	}

	// release lock

	return files, nil
}
