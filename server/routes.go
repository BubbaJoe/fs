package server

import (
	. "fs-store/types"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// ListFilesRoute is the route for listing files
func listFilesRoute(sc *ServerConfig) echo.HandlerFunc {
	return func(c echo.Context) error {

		files, err := sc.getFileList(sc.MaxListSize)

		if err != nil {
			logrus.Error("Error while trying to get file list", err)
			return c.JSON(500, GenericResponse{
				Success: false,
				Message: "Internal server error",
			})
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
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "File not provided",
			})
		}

		if len(files) > 1 {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "Too many files provided",
			})
		}

		if len(files[0].Filename) > 255 {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "File name too long",
			})
		}

		logrus.Info("Uploading file: ", files[0].Size, " ", sc.MaxFileSize)
		if files[0].Size > sc.MaxFileSize {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "File too large",
			})
		}

		fileHeader := files[0]
		fileReader, err := fileHeader.Open()
		if err != nil {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "error reading file",
			})
		}
		logrus.Info("Uploading file: ", fileHeader.Filename)
		err = sc.createFile(fileHeader.Filename, fileHeader.Size, fileReader, overwrite)

		if err == ErrFileAlreadyExists {
			return c.JSON(409, GenericResponse{
				Success: false,
				Message: "File already exists",
			})
		}

		if err != nil {
			logrus.Error("Error while trying to read body", err)
			return c.JSON(500, GenericResponse{
				Success: false,
				Message: "File too large",
			})
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
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "No file name provided",
			})
		}

		if len(fileName) > 255 {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "File name too long",
			})
		}

		err := sc.deleteFile(fileName)
		if err == ErrFileDoesntExist {
			return c.JSON(400, GenericResponse{
				Success: false,
				Message: "File doesn't exist",
			})
		}

		if err != nil {
			return c.JSON(500, GenericResponse{
				Success: false,
				Message: "internal server error",
			})
		}

		return c.JSON(200, GenericResponse{
			Success: true,
			Message: "File deleted",
		})
	}
}
