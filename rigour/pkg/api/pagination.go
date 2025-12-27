package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const (
	// DefaultPageSize is the default number of results per page
	DefaultPageSize = 50
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 500
)

// EncodePaginationToken encodes a pagination token to a base64 string.
func EncodePaginationToken(lastID string, sortField string) (string, error) {
	token := PaginationToken{
		LastID:    lastID,
		SortField: sortField,
	}

	data, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pagination token: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// DecodePaginationToken decodes a base64 pagination token.
func DecodePaginationToken(token string) (*PaginationToken, error) {
	if token == "" {
		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pagination token: %w", err)
	}

	var pt PaginationToken
	if err := json.Unmarshal(data, &pt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pagination token: %w", err)
	}

	return &pt, nil
}

// ValidatePageSize ensures the page size is within acceptable bounds.
func ValidatePageSize(size int) int {
	if size <= 0 {
		return DefaultPageSize
	}
	if size > MaxPageSize {
		return MaxPageSize
	}
	return size
}
