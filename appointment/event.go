package appointment

import (
	"time"

	"github.com/hadroncorp/geck/transport/event"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"event-schema-registry/placespb"
)

var (
	_statusToProtobufMap = map[Status]placespb.AppointmentStatus{
		StatusUnknown:   placespb.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED,
		StatusScheduled: placespb.AppointmentStatus_APPOINTMENT_STATUS_SCHEDULED,
		StatusCancelled: placespb.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED,
		StatusCompleted: placespb.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED,
	}
)

// -- Scheduled --

type ScheduledEvent struct {
	ID           string
	Title        string
	PlaceID      string
	TargetedTo   string
	ScheduledBy  string
	ScheduleTime time.Time
	Notes        string
	Status       Status
	CreateTime   time.Time
	CreateBy     string
}

// compile-time assertion(s)
var _ event.Event = (*ScheduledEvent)(nil)

func newScheduledEvent(src Appointment) ScheduledEvent {
	return ScheduledEvent{
		ID:           src.id,
		Title:        src.title.String(),
		PlaceID:      src.placeID,
		TargetedTo:   src.targetedTo,
		ScheduledBy:  src.scheduledBy,
		ScheduleTime: src.scheduleTime,
		Notes:        src.notes,
		Status:       src.status,
		CreateTime:   src.CreateTime(),
		CreateBy:     src.CreateBy(),
	}
}

func (s ScheduledEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "scheduled",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s ScheduledEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s ScheduledEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentScheduledEvent{
		AppointmentId: s.ID,
		Title:         s.Title,
		PlaceId:       s.PlaceID,
		TargetedTo:    lo.EmptyableToPtr(s.TargetedTo),
		ScheduledBy:   s.ScheduledBy,
		ScheduledTime: timestamppb.New(s.ScheduleTime),
		Notes:         lo.EmptyableToPtr(s.Notes),
		Status:        _statusToProtobufMap[s.Status],
		CreateTime:    timestamppb.New(s.CreateTime),
		CreateBy:      s.CreateBy,
	})
}

// -- Updated --

type UpdatedEvent struct {
	ID           string
	Title        string
	PlaceID      string
	TargetedTo   string
	ScheduledBy  string
	ScheduleTime time.Time
	Notes        string
	Status       Status
	CreateTime   time.Time
	CreateBy     string
	UpdateTime   time.Time
	UpdateBy     string
	Version      uint64
}

// compile-time assertion(s)
var _ event.Event = (*UpdatedEvent)(nil)

func newUpdatedEvent(src *Appointment) UpdatedEvent {
	return UpdatedEvent{
		ID:           src.id,
		Title:        src.title.String(),
		PlaceID:      src.placeID,
		TargetedTo:   src.targetedTo,
		ScheduledBy:  src.scheduledBy,
		ScheduleTime: src.scheduleTime,
		Notes:        src.notes,
		Status:       src.status,
		CreateTime:   src.CreateTime(),
		CreateBy:     src.CreateBy(),
		UpdateTime:   src.LastUpdateTime(),
		UpdateBy:     src.LastUpdateBy(),
		Version:      src.Version(),
	}
}

func (s UpdatedEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "updated",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s UpdatedEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s UpdatedEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentUpdatedEvent{
		AppointmentId: s.ID,
		PlaceId:       s.PlaceID,
		Title:         s.Title,
		TargetedTo:    lo.EmptyableToPtr(s.TargetedTo),
		ScheduledBy:   s.ScheduledBy,
		ScheduledTime: timestamppb.New(s.ScheduleTime),
		Notes:         lo.EmptyableToPtr(s.Notes),
		Status:        _statusToProtobufMap[s.Status],
		CreateTime:    timestamppb.New(s.CreateTime),
		CreateBy:      s.CreateBy,
		UpdateTime:    timestamppb.New(s.UpdateTime),
		UpdateBy:      s.UpdateBy,
	})
}

// -- Canceled --

type CanceledEvent struct {
	ID         string
	PlaceID    string
	Notes      string
	Status     Status
	CancelTime time.Time
	CanceledBy string
}

// compile-time assertion(s)
var _ event.Event = (*CanceledEvent)(nil)

func newCanceledEvent(src *Appointment) CanceledEvent {
	return CanceledEvent{
		ID:         src.id,
		PlaceID:    src.placeID,
		Notes:      src.notes,
		Status:     src.status,
		CancelTime: src.LastUpdateTime(),
		CanceledBy: src.LastUpdateBy(),
	}
}

func (s CanceledEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "canceled",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s CanceledEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s CanceledEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentCanceledEvent{
		AppointmentId: s.ID,
		PlaceId:       s.PlaceID,
		Notes:         lo.EmptyableToPtr(s.Notes),
		Status:        _statusToProtobufMap[s.Status],
		CancelTime:    timestamppb.New(s.CancelTime),
		CancelBy:      s.CanceledBy,
	})
}

// -- Rescheduled --

type RescheduledEvent struct {
	ID             string
	PlaceID        string
	ScheduleTime   time.Time
	Notes          string
	Status         Status
	RescheduleTime time.Time
	RescheduleBy   string
}

// compile-time assertion(s)
var _ event.Event = (*RescheduledEvent)(nil)

func newRescheduledEvent(src *Appointment) RescheduledEvent {
	return RescheduledEvent{
		ID:             src.id,
		PlaceID:        src.placeID,
		ScheduleTime:   src.scheduleTime,
		Notes:          src.notes,
		Status:         src.status,
		RescheduleTime: src.LastUpdateTime(),
		RescheduleBy:   src.LastUpdateBy(),
	}
}

func (s RescheduledEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "rescheduled",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s RescheduledEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s RescheduledEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentRescheduledEvent{
		AppointmentId:  s.ID,
		PlaceId:        s.PlaceID,
		ScheduledTime:  timestamppb.New(s.ScheduleTime),
		Notes:          lo.EmptyableToPtr(s.Notes),
		Status:         _statusToProtobufMap[s.Status],
		RescheduleTime: timestamppb.New(s.RescheduleTime),
		RescheduleBy:   s.RescheduleBy,
	})
}

// -- Deleted --

type DeletedEvent struct {
	ID         string
	PlaceID    string
	DeleteTime time.Time
	DeleteBy   string
}

// compile-time assertion(s)
var _ event.Event = (*DeletedEvent)(nil)

func newDeletedEvent(src *Appointment) DeletedEvent {
	return DeletedEvent{
		ID:         src.id,
		PlaceID:    src.placeID,
		DeleteTime: src.LastUpdateTime(),
		DeleteBy:   src.LastUpdateBy(),
	}
}

func (s DeletedEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "deleted",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s DeletedEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s DeletedEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentDeletedEvent{
		AppointmentId: s.ID,
		PlaceId:       s.PlaceID,
		DeleteTime:    timestamppb.New(s.DeleteTime),
		DeleteBy:      s.DeleteBy,
	})
}

// -- Completed --

type CompletedEvent struct {
	ID                 string
	PlaceID            string
	Status             Status
	CompleteTime       time.Time
	MarkedAsCompleteBy string
}

// compile-time assertion(s)
var _ event.Event = (*CompletedEvent)(nil)

func newCompletedEvent(src *Appointment) CompletedEvent {
	return CompletedEvent{
		ID:                 src.id,
		PlaceID:            src.placeID,
		Status:             src.status,
		CompleteTime:       src.LastUpdateTime(),
		MarkedAsCompleteBy: src.LastUpdateBy(),
	}
}

func (s CompletedEvent) Topic() event.Topic {
	return event.NewTopic("appointment", "completed",
		event.WithOrganization("hadron"),
		event.WithPlatform("places"),
	)
}

func (s CompletedEvent) Key() string {
	// NOTE: Using place_id as event key. This can leverage in
	// having all scheduled events in an ordered manner (depends on actual event infrastructure).
	return s.PlaceID
}

func (s CompletedEvent) Bytes() ([]byte, error) {
	// NOTE: You can change this for a more simple codec like JSON.
	// Just add the field tags to ScheduledEvent.
	// e.g. json.Marshal(s).
	return proto.Marshal(&placespb.AppointmentCompletedEvent{
		AppointmentId:      s.ID,
		PlaceId:            s.PlaceID,
		Status:             _statusToProtobufMap[s.Status],
		CompleteTime:       timestamppb.New(s.CompleteTime),
		MarkedAsCompleteBy: s.MarkedAsCompleteBy,
	})
}
