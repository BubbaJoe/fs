// *build integration
package client_test

import (
	"fs-store/client"
	. "fs-store/types"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

// TestIntegrationPositive_DeleteFile tests the DeleteFile functionality
func TestIntegrationPositive_DeleteFile(t *testing.T) {
	domain := "http://domain"
	fileName := "text.txt"

	conf, err := client.NewFSClientConfig(domain, false)
	assert.NoError(t, err, "No error expected")
	// httpmock.Activate()
	httpmock.ActivateNonDefault(conf.Client.GetClient())

	httpmock.RegisterResponder("DELETE", conf.Client.BaseURL+"/files?filename="+fileName,
		func(req *http.Request) (*http.Response, error) {
			fn := req.URL.Query().Get("filename")
			assert.Equal(t, fileName, fn, "Expected file name to be %s, got %s", fileName, fn)
			return httpmock.NewJsonResponse(http.StatusAccepted, GenericResponse{
				Success: true, Message: "File deleted: " + fileName,
			})
		},
	)

	err = conf.DeleteFile(fileName)
	assert.NoError(t, err, "No error expected")

	assert.Equal(t, 1, httpmock.GetTotalCallCount(),
		"expected %d calls", 1)

	httpmock.DeactivateAndReset()

}

// TestIntegration_UploadFile tests the UploadFile functionality
func TestIntegration_UploadFile(t *testing.T) {
	domain := "http://domain"
	fileName := "text.txt"

	conf, err := client.NewFSClientConfig(domain, false)
	assert.NoError(t, err, "No error expected")
	// httpmock.Activate()
	httpmock.ActivateNonDefault(conf.Client.GetClient())

	httpmock.RegisterResponder("DELETE", conf.Client.BaseURL+"/files?filename="+fileName,
		func(req *http.Request) (*http.Response, error) {
			fn := req.URL.Query().Get("filename")
			assert.Equal(t, fileName, fn, "Expected file name to be %s, got %s", fileName, fn)
			return httpmock.NewJsonResponse(http.StatusAccepted, GenericResponse{
				Success: true, Message: "File deleted: " + fileName,
			})
		},
	)

	err = conf.DeleteFile(fileName)
	assert.NoError(t, err, "No error expected")

	assert.Equal(t, 1, httpmock.GetTotalCallCount(),
		"expected %d calls", 1)

	httpmock.DeactivateAndReset()
}

// TestIntegration_ListFiles tests the ListFiles functionality
func TestIntegration_ListFiles(t *testing.T) {
	domain := "http://domain"

	conf, err := client.NewFSClientConfig(domain, false)
	assert.NoError(t, err, "No error expected")
	// httpmock.Activate()
	httpmock.ActivateNonDefault(conf.Client.GetClient())

	curTime := time.Now()

	httpmock.RegisterResponder("GET", conf.Client.BaseURL+"/files",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, []FileResponse{
				{FileName: "file1.txt", FileSize: 123, CreatedAt: curTime},
				{FileName: "file2.txt", FileSize: 1234, CreatedAt: curTime.Add(5 * time.Second)},
				{FileName: "file3.txt", FileSize: 12345, CreatedAt: curTime.Add(5 * time.Minute)},
			})
		},
	)

	files, err := conf.ListFiles()
	assert.NoError(t, err, "No error expected")

	assert.True(t, len(files) == 3, "Expected number of files to be 2, got %d", len(files))

	// Assert that the file names are correct
	assert.Equal(t, files[0].FileName, "file1.txt", "Expected file name to be file1.txt, got %s", files[0].FileName)
	assert.Equal(t, files[1].FileName, "file2.txt", "Expected file name to be file2.txt, got %s", files[0].FileName)
	assert.Equal(t, files[2].FileName, "file3.txt", "Expected file name to be file3.txt, got %s", files[0].FileName)

	// Assert that the filesize is correct
	assert.Equal(t, int64(123), files[0].FileSize, "Expected file size to be 123, got %d", files[0].FileSize)
	assert.Equal(t, int64(1234), files[1].FileSize, "Expected file size to be 1234, got %d", files[0].FileSize)
	assert.Equal(t, int64(12345), files[2].FileSize, "Expected file size to be 12345, got %d", files[0].FileSize)

	// Assert that the created at time is correct (up to the millisecond)
	assert.Equal(t, curTime.UnixMilli(), files[0].CreatedAt.UnixMilli(), "Expected createdAt ts time to be %s, got %d", curTime, files[0].CreatedAt.UnixMilli())
	assert.Equal(t, curTime.Add(5*time.Second).UnixMilli(), files[1].CreatedAt.UnixMilli(), "Expected createdAt ts time to be %s, got %d", curTime.Add(5*time.Second), files[0].CreatedAt.UnixMilli())
	assert.Equal(t, curTime.Add(5*time.Minute).UnixMilli(), files[2].CreatedAt.UnixMilli(), "Expected createdAt ts time to be %s, got %d", curTime.Add(5*time.Minute), files[0].CreatedAt.UnixMilli())

	assert.Equal(t, 1, httpmock.GetTotalCallCount(),
		"expected %d calls", 1)

	httpmock.DeactivateAndReset()
}
