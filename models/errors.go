package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
)

var (
	ErrEmailAlreadyExists = errors.New("models: email already exists")
	ErrNotFound           = errors.New("models: not found")
)

type FileError struct {
	Issue string
}

func (fe FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", fe.Issue)
}

const bufferSize = 512

func checkContentType(r io.ReadSeeker, allowedTypes []string) error {
	testBytes := make([]byte, bufferSize)

	_, err := r.Read(testBytes)
	if err != nil {
		return fmt.Errorf("checking content type: %w", err)
	}

	_, err = r.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("checking content type: %w", err)
	}

	contentType := http.DetectContentType(testBytes)

	if slices.Contains(allowedTypes, contentType) {
		return nil
	}

	return FileError{
		Issue: fmt.Sprintf("invalid content type: %v", contentType),
	}
}

func checkExtension(filename string, allowedExtensions []string) error {
	if !hasExtension(filename, allowedExtensions) {
		return FileError{
			Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
		}
	}
	return nil
}
