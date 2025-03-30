package organization

import (
	"context"

	"github.com/hadroncorp/geck/transport/stream"
	"google.golang.org/protobuf/proto"

	"event-schema-registry/iampb"
)

// ControllerStream is a stream controller for the organization domain.
type ControllerStream struct {
	readerManager stream.ReaderManager
}

// compile-time assertion
var _ stream.Controller = (*ControllerStream)(nil)

// NewControllerStream creates a new instance of [ControllerStream].
func NewControllerStream(readerManager stream.ReaderManager) ControllerStream {
	return ControllerStream{readerManager: readerManager}
}

func (c ControllerStream) RegisterReaders() {
	c.readerManager.Register(TopicCreated.String(), c.handleCreate)
}

func (c ControllerStream) handleCreate(ctx context.Context, message stream.Message) error {
	ev := &iampb.OrganizationCreatedEvent{}
	if err := proto.Unmarshal(message.Data, ev); err != nil {
		return err
	}

	ev.GetName()
	return nil
}
