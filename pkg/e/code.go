package e

import "errors"

var (
	ErrBlockKnown = errors.New("block already known")

	ErrBlockUnKnown = errors.New("block not found")
)
