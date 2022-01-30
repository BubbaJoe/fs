package client

import (
	"errors"
	"fmt"
	. "fs-store/types"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

// FSClientConfig is the configuration for the client
type FSClientConfig struct {
	Client *resty.Client
}

// NewFSClientConfig creates a new client configuration
func NewFSClientConfig(rawAddress string, secure bool) (*FSClientConfig, error) {
	if !strings.Contains(rawAddress, "://") {
		if secure {
			rawAddress = "https://" + rawAddress
		} else {
			rawAddress = "http://" + rawAddress
		}
	}
	url, err := url.Parse(rawAddress)
	if err != nil {
		return nil, err
	}

	client := resty.New()

	address := url.Scheme + "://" + url.Host

	return &FSClientConfig{
		Client: client.SetBaseURL(address),
	}, nil
}

// DeleteFile deletes a file
func (conf *FSClientConfig) DeleteFile(fileName string) (GenericResponse, error) {
	trimmedFileName := strings.TrimSpace(fileName)
	var genResponse GenericResponse
	_, err := conf.Client.R().
		SetResult(&genResponse).
		SetQueryParam("filename", trimmedFileName).
		Delete("/files")

	if err != nil {
		return genResponse, err
	}

	return genResponse, nil
}

func (conf *FSClientConfig) UploadFile(fileName string, r io.Reader, overwrite bool) error {
	genResponse := &GenericResponse{}
	req := conf.Client.R()

	resp, err := req.
		SetQueryParam("overwrite", strconv.FormatBool(overwrite)).
		SetMultipartField("file", fileName, "", r).
		SetResult(genResponse).
		Post("/files")

	fmt.Println(string(resp.Body()), resp.StatusCode())

	return err
}

func (conf *FSClientConfig) ListFiles() ([]FileResponse, error) {
	var files []FileResponse
	resp, err := conf.Client.R().
		SetResult(&files).
		Get("/files")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New("status - " + resp.Status())
	}

	return files, nil
}
