package appointment

import "github.com/hadroncorp/geck/errors/syserr"

var (
	ErrScheduledBeforeNow = syserr.New(syserr.InvalidArgument, "appointment scheduled before current time",
		syserr.WithInternalCode("SCHEDULED_BEFORE_CURRENT_TIME"),
	)
	ErrNotFound           = syserr.NewResourceNotFound[Appointment]()
	ErrInvalidStatus      = syserr.NewNotOneOf("status", StatusScheduled.String(), StatusCancelled.String())
	ErrIsAlreadyCompleted = syserr.New(syserr.FailedPrecondition, "appointment is already completed",
		syserr.WithInternalCode("APPOINTMENT_ALREADY_COMPLETED"),
	)
)
