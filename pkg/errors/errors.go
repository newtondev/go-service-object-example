package errors

import "github.com/pkg/errors"

var (
	// ErrEmailExists returns when given email is present in storage.
	ErrEmailExists = errors.New("email already exists")
)