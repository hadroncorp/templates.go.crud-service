package organization

import (
	"context"

	"github.com/hadroncorp/geck/event"
	"github.com/hadroncorp/geck/persistence/audit"
	"github.com/samber/lo"
)

// Organization is a group of people who work together to achieve a common goal.
//
// More over, an Organization in the context of this platform is a logical grouping of resources like
// employees and places. It is a way to organize resources in a way that makes sense for the business.
type Organization struct {
	audit.Auditable
	event.AggregatorTemplate
	id   string
	name string
}

// New creates a new [Organization] with the given ID and name.
func New(ctx context.Context, id string, name string) Organization {
	org := Organization{
		Auditable: audit.NewWithDefaults(ctx),
		id:        id,
		name:      name,
	}
	org.RegisterEvents(newCreatedEvent(org))
	return org
}

// ID returns the unique identifier of the organization.
func (o Organization) ID() string {
	return o.id
}

// Name returns the name of the organization.
func (o Organization) Name() string {
	return o.name
}

// Update updates the [Organization] with the given options.
//
// It returns true if the organization was updated, false otherwise.
func (o *Organization) Update(ctx context.Context, opts ...UpdateOption) bool {
	if len(opts) == 0 {
		return false // no-op
	}

	for _, opt := range opts {
		opt(o)
	}
	audit.Update(ctx, &o.Auditable)
	o.RegisterEvents(newUpdatedEvent(o))
	return true
}

// Delete deletes the [Organization].
func (o *Organization) Delete(ctx context.Context) {
	audit.Delete(ctx, &o.Auditable)
	o.RegisterEvents(newDeletedEvent(o))
}

// -- Option(s) --

// UpdateOption is a function that updates an [Organization].
type UpdateOption func(o *Organization)

// WithUpdatedName sets the new name of the [Organization].
func WithUpdatedName(name *string) UpdateOption {
	if name == nil {
		// no-op
		return func(_ *Organization) {}
	}
	return func(o *Organization) {
		o.name = lo.FromPtr(name)
	}
}
