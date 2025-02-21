package employee

import "github.com/hadroncorp/geck/errors/syserr"

var (
	ErrNotFound = syserr.NewResourceNotFound[Employee]()
)
