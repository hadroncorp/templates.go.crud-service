package appointment

import "github.com/hadroncorp/geck/errors/syserr"

var (
	ErrScheduledBeforeNow = syserr.New(syserr.InvalidArgument, "appointment scheduled before current time",
		syserr.WithInternalCode("SCHEDULED_BEFORE_CURRENT_TIME"),
	)
	ErrNotFound = syserr.NewResourceNotFound[Appointment]()
)
