package organization

import (
	"reflect"
	"time"

	"github.com/hadroncorp/geck/event"
	"github.com/hadroncorp/geck/transport"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"event-schema-registry/iampb"
)

const (
	_eventSource = "/organizations"
)

var (
	// TopicCreated is the event topic for organization creation.
	TopicCreated = event.NewTopic("hadron", "organization", "created",
		event.WithPlatform("iam"))
	// TopicUpdated is the event topic for organization update.
	TopicUpdated = event.NewTopic("hadron", "organization", "updated",
		event.WithPlatform("iam"))
	// TopicDeleted is the event topic for organization deletion.
	TopicDeleted = event.NewTopic("hadron", "organization", "deleted",
		event.WithPlatform("iam"))
)

// CreatedEvent is an event that is emitted when an organization is created.
type CreatedEvent struct {
	src   Organization
	topic event.Topic
}

// compile-time assertion
var _ event.Event = (*CreatedEvent)(nil)

func newCreatedEvent(src Organization) CreatedEvent {
	return CreatedEvent{
		src:   src,
		topic: TopicCreated,
	}
}

func (e CreatedEvent) Topic() event.Topic {
	return e.topic
}

func (e CreatedEvent) Key() string {
	return e.src.id
}

func (e CreatedEvent) Bytes() ([]byte, error) {
	return proto.Marshal(&iampb.OrganizationCreatedEvent{
		OrganizationId: e.src.id,
		Name:           e.src.name,
		CreateTime:     timestamppb.New(e.src.CreateTime()),
		CreateBy:       e.src.CreateBy(),
	})
}

func (e CreatedEvent) BytesContentType() transport.MimeType {
	return transport.MimeTypeProtobuf
}

func (e CreatedEvent) Source() string {
	return _eventSource
}

func (e CreatedEvent) Subject() string {
	return e.src.id
}

func (e CreatedEvent) OccurrenceTime() time.Time {
	return e.src.CreateTime()
}

func (e CreatedEvent) SchemaSource() string {
	return reflect.TypeFor[iampb.OrganizationCreatedEvent]().PkgPath()
}

// UpdatedEvent is an event that is emitted when an organization is updated.
type UpdatedEvent struct {
	src   *Organization
	topic event.Topic
}

// compile-time assertion
var _ event.Event = (*UpdatedEvent)(nil)

func newUpdatedEvent(src *Organization) UpdatedEvent {
	return UpdatedEvent{
		src:   src,
		topic: TopicUpdated,
	}
}

func (e UpdatedEvent) Topic() event.Topic {
	return e.topic
}

func (e UpdatedEvent) Key() string {
	return e.src.id
}

func (e UpdatedEvent) Bytes() ([]byte, error) {
	return proto.Marshal(&iampb.OrganizationUpdatedEvent{
		OrganizationId: e.src.id,
		Name:           e.src.name,
		UpdateTime:     timestamppb.New(e.src.LastUpdateTime()),
		UpdateBy:       e.src.LastUpdateBy(),
	})
}

func (e UpdatedEvent) BytesContentType() transport.MimeType {
	return transport.MimeTypeProtobuf
}

func (e UpdatedEvent) Source() string {
	return "/organizations"
}

func (e UpdatedEvent) Subject() string {
	return e.src.id
}

func (e UpdatedEvent) OccurrenceTime() time.Time {
	return e.src.LastUpdateTime()
}

func (e UpdatedEvent) SchemaSource() string {
	return reflect.TypeFor[iampb.OrganizationUpdatedEvent]().PkgPath()
}

// DeletedEvent is an event that is emitted when an organization is deleted.
type DeletedEvent struct {
	src   *Organization
	topic event.Topic
}

// compile-time assertion
var _ event.Event = (*DeletedEvent)(nil)

func newDeletedEvent(src *Organization) DeletedEvent {
	return DeletedEvent{
		src:   src,
		topic: TopicDeleted,
	}
}

func (e DeletedEvent) Topic() event.Topic {
	return e.topic
}

func (e DeletedEvent) Key() string {
	return e.src.id
}

func (e DeletedEvent) Bytes() ([]byte, error) {
	return proto.Marshal(&iampb.OrganizationDeletedEvent{
		OrganizationId: e.src.id,
		DeleteTime:     timestamppb.New(e.src.LastUpdateTime()),
		DeleteBy:       e.src.LastUpdateBy(),
	})
}

func (e DeletedEvent) BytesContentType() transport.MimeType {
	return transport.MimeTypeProtobuf
}

func (e DeletedEvent) Source() string {
	return _eventSource
}

func (e DeletedEvent) Subject() string {
	return e.src.id
}

func (e DeletedEvent) OccurrenceTime() time.Time {
	return e.src.LastUpdateTime()
}

func (e DeletedEvent) SchemaSource() string {
	return reflect.TypeFor[iampb.OrganizationDeletedEvent]().PkgPath()
}
