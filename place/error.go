package place

import "github.com/hadroncorp/geck/errors/syserr"

var (
	ErrNotFound = syserr.NewResourceNotFound[Place]()
)
