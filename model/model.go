package model

import "errors"

type ID uint

func NewID(v uint) (ID, error) {
	if v == 0 {
		return 0, errors.New("invalid node id")
	}
	return ID(v), nil
}