package server

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileStore stores metadata and content of a file
type FileStore struct {
	Version   FSVersion
	Reader    io.Reader
	FileName  string
	DataSize  int64
	CreatedAt time.Time
}

type FSVersion uint8

const (

	// V1 Order: version, fileNameSize, filename, createdAt, file size, content
	FSStoreV1 FSVersion = 1

	DefaultVersion FSVersion = FSStoreV1
)

// createFileAt creates file using file store at directory
func (store *FileStore) createFileAt(dataDir string, overwrite bool) error {
	fsFileName := generateFileName(store.FileName)
	flag := os.O_CREATE | os.O_WRONLY
	if overwrite {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := os.OpenFile(filepath.Join(dataDir, fsFileName), flag, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return store.writeFileStore(file)
}

// deleteFileAt deletes a file using file store at directory
func deleteFileAt(dataDir, fileName string) error {
	return os.Remove(filepath.Join(dataDir, generateFileName(fileName)))
}

// Exists checks if a file exists using the file store
func fileExists(dataDir, fileName string) (bool, error) {
	_, err := os.Stat(filepath.Join(dataDir,
		generateFileName(fileName)))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// generateFileName generates a file name from a file name
func generateFileName(fileName string) string {
	byteArr := md5.Sum([]byte(fileName))
	return hex.EncodeToString(byteArr[:]) + ".fs"
}

// parseFileStore creates a file store from an io.Reader
func parseFileStore(r io.Reader) (*FileStore, error) {
	var store FileStore

	// Read the version (1 byte)
	err := binary.Read(r, binary.BigEndian, &store.Version)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}

	// Read the file name size (1 byte)
	var fileNameSize uint8
	err = binary.Read(r, binary.BigEndian, &fileNameSize)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}

	// Read the filename (max 255 bytes)
	fileName := make([]byte, fileNameSize)
	n, err := r.Read(fileName)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if n != int(fileNameSize) {
		return nil, io.ErrUnexpectedEOF
	}
	store.FileName = string(fileName)

	// Read the createdAt by reading (8 bytes)
	var ts int64 = 0
	err = binary.Read(r, binary.BigEndian, &ts)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	store.CreatedAt = time.UnixMilli(ts)

	// Read the size (8 bytes)
	err = binary.Read(r, binary.BigEndian, &store.DataSize)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if store.DataSize == 0 {
		return nil, errors.New("invalid file store size")
	}

	store.Reader = r

	return &store, nil
}

// WriteFileStore writes the file store to an io.Writer
// V1 Order: version, fileNameSize, filename, createdAt, file size, content
func (store *FileStore) writeFileStore(w io.Writer) error {
	// Set default version if not set
	if store.Version == 0 {
		store.Version = DefaultVersion
	}

	// Write the version (1 byte)
	err := binary.Write(w, binary.BigEndian, store.Version)
	if err != nil {
		return err
	}

	// Write the file name size (1 byte)
	err = binary.Write(w, binary.BigEndian, uint8(len(store.FileName)))
	if err != nil {
		return err
	}

	// Write the filename (max 255 bytes)
	_, err = w.Write([]byte(store.FileName))
	if err != nil {
		return err
	}

	// Write the created
	err = binary.Write(w, binary.BigEndian, store.CreatedAt.UnixMilli())
	if err != nil {
		return err
	}

	// Write the real size
	err = binary.Write(w, binary.BigEndian, store.DataSize)
	if err != nil {
		return err
	}

	// buffer for storing the data
	buffer := make([]byte, 1<<16-1)

	// Write the content
	_, err = io.CopyBuffer(w, store, buffer)
	if err != nil {
		return err
	}
	if err == io.EOF {
		return nil
	}

	return nil
}

// Read reads the file store
func (r *FileStore) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}
