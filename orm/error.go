package orm

import "errors"

var (
	ErrInvalidTransaction = errors.New("orm: invalid transaction")
)
