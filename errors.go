package queryx

import "errors"

var (
	ErrNotFound           = errors.New("queryx: not found")
	ErrNotSupported       = errors.New("queryx: not supported")
	ErrTableNotSpecified  = errors.New("queryx: table not specified")
	ErrColumnNotSpecified = errors.New("queryx: column not specified")
	ErrInvalidPointer     = errors.New("queryx: attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("queryx: wrong placeholder count")
	ErrInvalidSliceLength = errors.New("queryx: length of slice is 0. length must be >= 1")
	ErrCantConvertToTime  = errors.New("queryx: can't convert to time.Time")
	ErrInvalidTimestring  = errors.New("queryx: invalid time string")
)
