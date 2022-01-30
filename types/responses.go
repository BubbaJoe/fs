package types

import "time"

// FileResponse is the response for a file information
type FileResponse struct {
	FileName  string    `json:"fileName"`
	FileSize  int64     `json:"fileSize"`
	CreatedAt time.Time `json:"createdAt"`
}

// GeneralResponse is a general response for a request
type GenericResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
