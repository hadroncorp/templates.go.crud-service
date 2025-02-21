package appointment

import (
	"context"
	"time"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/audit"

	"github.com/hadroncorp/service-template/domain/valueobject"
)

// NOTE: Entity (Appointment) is not exposing its fields to guarantee only
// exposed routines (e.g. New, Appointment.Update) are used, ensuring a valid domain state.

type Appointment struct {
	audit.Auditable
	id      string
	title   valueobject.Title
	placeID string
	// aka. employee id
	targetedTo   string
	scheduledBy  string
	scheduleTime time.Time
	notes        string
	status       Status
}

// compile-time assertion(s)
var _ persistence.Storable = (*Appointment)(nil)

func (a Appointment) ID() string {
	return a.id
}

func (a Appointment) Title() valueobject.Title {
	return a.title
}

func (a Appointment) PlaceID() string {
	return a.placeID
}

func (a Appointment) TargetedTo() string {
	return a.targetedTo
}

func (a Appointment) ScheduledBy() string {
	return a.scheduledBy
}

func (a Appointment) ScheduleTime() time.Time {
	return a.scheduleTime
}

func (a Appointment) Notes() string {
	return a.notes
}

func (a Appointment) Status() Status {
	return a.status
}

func (a Appointment) validate() error {
	now := time.Now().UTC()
	if a.scheduleTime.Before(now) {
		return ErrScheduledBeforeNow
	}
	if a.status == StatusUnknown {
		return ErrInvalidStatus
	}
	return nil
}

func (a *Appointment) Cancel(ctx context.Context, reason string) error {
	if a.status == StatusCompleted {
		return ErrIsAlreadyCompleted
	}

	a.notes += "CANCEL: " + reason + "\n"
	a.status = StatusCancelled
	if err := a.validate(); err != nil {
		return err
	}
	audit.UpdateAuditable(ctx, &a.Auditable)
	// TODO: Add events
	return nil
}

func (a *Appointment) Reschedule(ctx context.Context, reason string, newTime time.Time) error {
	if a.status == StatusCompleted {
		return ErrIsAlreadyCompleted
	}

	if newTime.Equal(a.scheduleTime) {
		return nil // no-op
	}
	a.notes += "RESCHEDULE: " + reason + "\n"
	a.status = StatusScheduled
	a.scheduleTime = newTime.UTC()
	if err := a.validate(); err != nil {
		return err
	}
	audit.UpdateAuditable(ctx, &a.Auditable)
	// TODO: Add events
	return nil
}

func (a *Appointment) Update(ctx context.Context, opts ...UpdateOption) error {
	for _, opt := range opts {
		opt(a)
	}
	if err := a.validate(); err != nil {
		return err
	}
	audit.UpdateAuditable(ctx, &a.Auditable)
	// TODO: Add events
	return nil
}

func (a *Appointment) Delete(ctx context.Context) {
	audit.DeleteAuditable(ctx, &a.Auditable)
	// TODO: Add events
}

func (a *Appointment) MarkAsCompleted(ctx context.Context) {
	a.status = StatusCompleted
	audit.UpdateAuditable(ctx, &a.Auditable)
	// TODO: Add events
}

// -- Options --

// NOTE: Only add options for entity fields that CAN be updated.

type UpdateOption func(*Appointment)

func WithTitleUpdate(t valueobject.Title) UpdateOption {
	return func(options *Appointment) {
		options.title = t
	}
}

func WithTargetUpdate(employeeID string) UpdateOption {
	return func(options *Appointment) {
		options.targetedTo = employeeID
	}
}

func WithScheduleTimeUpdate(t time.Time) UpdateOption {
	return func(options *Appointment) {
		options.scheduleTime = t
	}
}

func WithNotesUpdate(notes string) UpdateOption {
	return func(options *Appointment) {
		options.notes = notes
	}
}

func WithStatusUpdate(status Status) UpdateOption {
	return func(options *Appointment) {
		options.status = status
	}
}

// --> Factory Routines <--

type NewArgs struct {
	ID           string
	Title        valueobject.Title
	PlaceID      string
	ScheduledBy  string
	ScheduleTime time.Time
}

func New(args NewArgs, opts ...NewOption) (Appointment, error) {
	options := newOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	appointment := Appointment{
		id:           args.ID,
		title:        args.Title,
		placeID:      args.PlaceID,
		targetedTo:   options.targetedTo,
		scheduledBy:  args.ScheduledBy,
		scheduleTime: args.ScheduleTime.UTC(),
		status:       StatusScheduled,
	}
	if err := appointment.validate(); err != nil {
		return Appointment{}, err
	}

	// TODO: Add events
	return appointment, nil
}

// -- Options --

type newOptions struct {
	targetedTo string
}

type NewOption func(*newOptions)

func WithTargetNew(employeeID string) NewOption {
	return func(options *newOptions) {
		options.targetedTo = employeeID
	}
}
