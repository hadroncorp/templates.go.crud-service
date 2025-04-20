package notification

import "errors"

// A Sender is a component used to send raw messages to a set of users.
type Sender interface {
	// Send sends a raw message to a set of users.
	Send(message []byte, userIDs ...string) error
}

// NoopSender is a no-operation implementation of the [Sender] interface.
type NoopSender struct{}

// compile-time assertion
var _ Sender = (*NoopSender)(nil)

// NewNoopSender creates a new instance of [NoopSender].
func NewNoopSender() NoopSender {
	return NoopSender{}
}

// Send does nothing and returns nil.
func (s NoopSender) Send(_ []byte, _ ...string) error {
	return errors.New("USE DLQ")
}
