package validator

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
)

var SupportedTypes = []string{"csv"}

type Validator interface {
	Verify() error
}

func Validate(validators ...func() error) error {
	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}
	return nil
}

func VerifyFilePathIsEmpty(path string, message string) error {
	if path == "" {
		return errors.New(message)
	}

	return nil
}

func VerifyFilePathTypeIsValid(path string) error {
	ext := filepath.Ext(path)

	if !slices.Contains(SupportedTypes, ext) {
		return errors.New(fmt.Sprintf("format %s not supported. Supported formats: %s", ext,
			strings.Join(SupportedTypes, ",")))
	}

	return nil
}

func VerifyFilePathExists(path string) error {
	_, err := os.Stat(path)

	return err
}

func VerifyCSVHeader(expectedHeader []string, actualHeader []string) error {
	if !slices.Equal(expectedHeader, actualHeader) {
		return errors.New(fmt.Sprintf("Expected header not matches actual. Expected: %s, Actual: %s",
			strings.Join(expectedHeader, ","), strings.Join(actualHeader, ",")))
	}

	return nil
}
