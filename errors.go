package pgr

import "errors"

var (
	ErrNotConnection      = errors.New("pgr: connection is not set")
	ErrNotFound           = errors.New("pgr: not found")
	ErrNotSupported       = errors.New("pgr: not supported")
	ErrTableNotSpecified  = errors.New("pgr: table not specified")
	ErrColumnNotSpecified = errors.New("pgr: column not specified")
	ErrInvalidPointer     = errors.New("pgr: attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("pgr: wrong placeholder count")
	ErrInvalidSliceLength = errors.New("pgr: length of slice is 0. length must be >= 1")
)
