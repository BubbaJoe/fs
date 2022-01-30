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
	"github.com/sirupsen/logrus"
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

	// ErrFileDoesntExist is returned when a file doesn't exists
	ErrFileDoesntExist = errors.New("file doesn't exist")
)

func NewServerConfig(address, dataDir string, maxFileSize int64, logLevel string) (*ServerConfig, error) {
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	logrus.WithFields(logrus.Fields{
		"address":     address,
		"dataDir":     dataDir,
		"maxFileSize": maxFileSize,
		"logLevel":    logLevel,
	}).Info("Creating server config")

	return &ServerConfig{
		DataDir:     dataDir,
		Address:     address,
		MaxFileSize: maxFileSize,
		MaxListSize: 255,
		mapLock:     &sync.RWMutex{},
		mtxMap:      make(map[string]*sync.Mutex, 255),
	}, nil
}

func (sc *ServerConfig) StartServer() error {
	e := echo.New()

	logrus.Info("Starting server at ", sc.Address)

	// Hide initial messages
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogLatency:   true,
		LogStatus:    true,
		LogUserAgent: true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			fields := logrus.Fields{
				"url":       values.URI,
				"status":    values.Status,
				"method":    values.Method,
				"latency":   values.Latency,
				"ip":        values.RemoteIP,
				"userAgent": values.UserAgent,
			}
			logrus.WithFields(fields).
				Info("request")

			return nil
		},
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
			newMutex := &sync.Mutex{}
			newMutex.Lock()
			sc.mtxMap[keyName] = newMutex
		}
		sc.mapLock.Unlock()
	} else {
		sc.mapLock.RUnlock()
		keyNameMtx.Lock()
	}

	return sc.mtxMap[keyName]
}

// createFile creates a file at the given path
func (sc *ServerConfig) createFile(fileName string, size int64, data io.Reader, overwrite bool) error {
	// create file store
	store := &FileStore{
		Version:  DefaultVersion,
		FileName: fileName,
		Reader:   data,
		DataSize: size,
	}
	logrus.Info("acquire lock for ", fileName)
	mutex := sc.acquireLock(fileName)
	logrus.Info("release lock for ", fileName)
	defer mutex.Unlock()
	// After acquiring lock, check if file exists (double-checked locking)
	if exists, err := fileExists(sc.DataDir, fileName); err != nil {
		return err
	} else if exists && !overwrite {
		return ErrFileAlreadyExists
	}
	return store.createFileAt(sc.DataDir, overwrite)
}

// deleteFile deletes a file at the given path
func (sc *ServerConfig) deleteFile(fileName string) error {
	sc.mapLock.Lock()
	delete(sc.mtxMap, fileName)
	defer sc.mapLock.Unlock()

	exists, err := fileExists(sc.DataDir, fileName)
	if err != nil {
		return err
	} else if exists {
		return deleteFileAt(sc.DataDir, fileName)
	}
	return ErrFileDoesntExist
}

// getFileList returns a list of files in the given directory
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
