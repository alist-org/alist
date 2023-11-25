package http

import (
	"fmt"
	"mime"
)

func parseFilenameFromContentDisposition(contentDisposition string) (string, error) {
	if contentDisposition == "" {
		return "", fmt.Errorf("Content-Disposition is empty")
	}
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return "", err
	}
	filename := params["filename"]
	if filename == "" {
		return "", fmt.Errorf("filename not found in Content-Disposition: [%s]", contentDisposition)
	}
	return filename, nil
}
