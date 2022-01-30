package client

import (
	"encoding/json"
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
	Client  *resty.Client
	Verbose bool
}

// NewFSClientConfig creates a new client configuration
func NewFSClientConfig(rawAddress string, verbose bool) (*FSClientConfig, error) {
	if !strings.Contains(rawAddress, "://") {
		rawAddress = "http://" + rawAddress
	}

	url, err := url.Parse(rawAddress)
	if err != nil {
		return nil, err
	}

	client := resty.New().OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		if verbose {
			fmt.Println(r.StatusCode(), string(r.Body()))
		}
		return nil
	})

	address := url.Scheme + "://" + url.Host

	return &FSClientConfig{
		Client:  client.SetBaseURL(address),
		Verbose: verbose,
	}, nil
}

// DeleteFile deletes a file
func (conf *FSClientConfig) DeleteFile(fileName string) error {
	genResponse := &GenericResponse{}
	resp, err := conf.Client.R().
		SetResult(&genResponse).
		SetQueryParam("filename", fileName).
		Delete("/files")

	if err != nil {
		return err
	} else if resp.IsError() {
		err := json.Unmarshal(resp.Body(), genResponse)
		if err != nil {
			return errors.New("unknown error")
		}
		return errors.New(genResponse.Message)
	}
	return nil
}

func (conf *FSClientConfig) UploadFile(fileName string, r io.Reader, overwrite bool) error {
	genResponse := &GenericResponse{}
	req := conf.Client.R()

	resp, err := req.
		SetQueryParam("overwrite", strconv.FormatBool(overwrite)).
		SetMultipartField("file", fileName, "application/octet-stream", r).
		SetResult(genResponse).
		Post("/files")

	if err != nil {
		return err
	} else if resp.IsError() {
		err := json.Unmarshal(resp.Body(), genResponse)
		if err != nil {
			return errors.New("unknown error")
		}
		return errors.New(genResponse.Message)
	}
	return nil
}

func (conf *FSClientConfig) ListFiles() ([]FileResponse, error) {
	var files []FileResponse
	resp, err := conf.Client.R().
		SetResult(&files).
		Get("/files")

	if err != nil {
		return nil, err
	} else if resp.IsError() {
		genResponse := &GenericResponse{}
		err := json.Unmarshal(resp.Body(), genResponse)
		if err != nil {
			return nil, errors.New("unknown error")
		}
		return nil, errors.New(genResponse.Message)
	}
	return files, nil

}
