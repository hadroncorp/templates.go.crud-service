package organization

import (
	"context"
	"log/slog"

	"github.com/hadroncorp/geck/event"
	"github.com/hadroncorp/geck/transport/stream/kafka"
	kinterceptor "github.com/hadroncorp/geck/transport/stream/kafka/interceptor"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"

	"event-schema-registry/iampb"
	"github.com/hadroncorp/service-template/notification"
)

// ControllerKafka is the Apache Kafka controller listening to events related to the notifications domain context.
type ControllerKafka struct {
	logger         *slog.Logger
	sender         notification.Sender
	producerClient *kgo.Client
}

// compile-time assertion
var _ kafka.Controller = (*ControllerKafka)(nil)

// NewControllerKafka creates a new instance of [ControllerKafka].
func NewControllerKafka(logger *slog.Logger, sender notification.Sender, produceClient *kgo.Client) ControllerKafka {
	return ControllerKafka{
		logger:         logger,
		sender:         sender,
		producerClient: produceClient,
	}
}

func (c ControllerKafka) RegisterReaders(rm kafka.ReaderManager) {
	rm.MustRegister(TopicCreated.String(), c.sendEmailToOrgAdmin,
		kafka.WithReaderGroup(
			kafka.MustConsumerGroup("iam", "organization", "send_email",
				kafka.WithConsumerGroupEvent("org_created")),
		),
		kafka.WithReaderInterceptors(
			kinterceptor.UseDeadLetter(c.producerClient, ""),
		),
	)
}

func (c ControllerKafka) sendEmailToOrgAdmin(_ context.Context, record *kgo.Record) error {
	ev := &iampb.OrganizationCreatedEvent{}
	if err := proto.Unmarshal(record.Value, ev); err != nil {
		return err
	}

	headers := kafka.ParseHeaders(record)
	c.logger.Info("sending email to organization admin",
		slog.Group("message",
			slog.String("key", string(record.Key)),
			slog.String("topic", record.Topic),
			slog.Int("partition", int(record.Partition)),
			slog.Int64("offset", record.Offset),
		),
		slog.Group("headers",
			slog.String(event.HeaderEventID, headers.Get(event.HeaderEventID)),
			slog.String(event.HeaderSource, headers.Get(event.HeaderSource)),
			slog.String(event.HeaderSpecVersion, headers.Get(event.HeaderSpecVersion)),
			slog.String(event.HeaderEventType, headers.Get(event.HeaderEventType)),
			slog.String(event.HeaderDataContentType, headers.Get(event.HeaderDataContentType)),
			slog.String(event.HeaderDataSchema, headers.Get(event.HeaderDataSchema)),
			slog.String(event.HeaderSubject, headers.Get(event.HeaderSubject)),
			slog.String(event.HeaderEventTime, headers.Get(event.HeaderEventTime)),
		),
		slog.Group("organization",
			slog.String("id", ev.GetOrganizationId()),
			slog.String("name", ev.GetName()),
			slog.String("create_by", ev.GetCreateBy()),
			slog.String("create_time", ev.GetCreateTime().String()),
		),
	)
	return c.sender.Send([]byte("Welcome to your new organization!"), ev.GetCreateBy())
}
