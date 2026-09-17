package hash

import (
	"errors"
	"fmt"
	"strings"
)

func Compare(first, second string) (bool, error) {
	first = strings.TrimSpace(first)
	second = strings.TrimSpace(second)

	if first == "" {
		return false, errors.New("first hash cannot be empty")
	}

	if second == "" {
		return false, errors.New("second hash cannot be empty")
	}

	return strings.EqualFold(first, second), nil
}

func CompareFiles(firstPath, secondPath string) (bool, error) {
	if strings.TrimSpace(firstPath) == "" {
		return false, errors.New("first file path cannot be empty")
	}

	if strings.TrimSpace(secondPath) == "" {
		return false, errors.New("second file path cannot be empty")
	}

	firstHash, err := Calculate(firstPath)
	if err != nil {
		return false, fmt.Errorf("calculate hash for first file: %w", err)
	}

	secondHash, err := Calculate(secondPath)
	if err != nil {
		return false, fmt.Errorf("calculate hash for second file: %w", err)
	}

	return Compare(firstHash, secondHash)
}