package storage

import "errors"

var (
	ErrAddressNotFound = errors.New("address not found")
	NotEnoughBalance   = errors.New("not enough balance to send")
)
