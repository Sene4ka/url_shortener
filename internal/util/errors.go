package util

import "errors"

var (
	ErrLinkNotFound       = errors.New("link not found")
	ErrIDExists           = errors.New("generated id is already used, please retry")
	ErrFailedToGenerateID = errors.New("failed to generate id")
	ErrInvalidIDLength    = errors.New("generated id has invalid length")

	ErrEmptyURL          = errors.New("url field is empty")
	ErrInvalidURL        = errors.New("invalid url")
	ErrUnsupportedScheme = errors.New("unsupported url scheme")
	ErrMissingHost       = errors.New("url is missing host")
	ErrURLTooLong        = errors.New("url is too long")
)
