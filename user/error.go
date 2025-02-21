package user

import "github.com/hadroncorp/geck/errors/syserr"

var (
	ErrNotFound = syserr.NewResourceNotFound[User]()
)
