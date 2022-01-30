package server

import (
	. "fs-store/types"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ListFilesRoute is the route for listing files
func listFilesRoute(sc *ServerConfig) echo.HandlerFunc {
	return func(c echo.Context) error {

		files, err := sc.getFileList(sc.MaxListSize)

		if err != nil {
			c.JSON(500, err.Error())
			return err
		}

		return c.JSON(200, files)
	}
}

// UploadFileRoute is the route for uploading files
func uploadFileRoute(sc *ServerConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		form, err := c.MultipartForm()
		if err != nil {
			// log
			return err
		}

		overwrite := c.QueryParams().
			Get("overwrite") == "true"

		files, ok := form.File["file"]
		if !ok || len(files) == 0 {
			return c.JSON(400, "No file found")
		}

		if len(files) > 1 {
			return c.JSON(400, "Too many files")
		}

		if len(files[0].Filename) > 255 {
			return c.JSON(400, "File name too long")
		}

		if files[0].Size > sc.MaxFileSize {
			return c.JSON(400, "File too large")
		}

		fileHeader := files[0]
		fileReader, err := fileHeader.Open()
		if err != nil {
			return c.JSON(400, "Could not read file")
		}

		err = sc.createFile(fileHeader.Filename, fileHeader.Size, fileReader, overwrite)

		if err == ErrFileAlreadyExists {
			return c.JSON(409, &GenericResponse{
				Success: false,
				Error:   "File already exists",
			})
		}

		if err != nil {
			log.Info("Error while trying to read body", err)
			return c.JSON(400, "Internal server error")
		}

		return c.JSON(200, GenericResponse{
			Success: true,
			Message: "File uploaded",
		})
	}
}

// DeleteFileRoute is the route for deleting files
func deleteFileRoute(sc *ServerConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		fileName := c.QueryParam("filename")
		if fileName == "" {
			return c.JSON(400, "No file name provided")
		}

		if len(fileName) > 255 {
			return c.JSON(400, "File name too long")
		}

		err := sc.deleteFile(fileName)
		if err != nil {
			return c.JSON(400, "Internal server error")
		}

		return c.JSON(200, GenericResponse{
			Success: true,
		})
	}
}
