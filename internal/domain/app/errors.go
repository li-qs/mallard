package app

import "fmt"

var (
	ErrInvalidSecret = fmt.Errorf("invalid secret")
	ErrAppExists     = fmt.Errorf("app already exists")
)
