package organization

import (
	"context"
	"errors"

	"github.com/hadroncorp/geck/event"
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
	"github.com/hadroncorp/geck/syserr"
)

// - Error(s) -

var (
	// ErrNotFound is returned when the organization is not found.
	ErrNotFound = syserr.NewResourceNotFound[Organization]()
	// ErrAlreadyExists is returned when the organization already exists.
	ErrAlreadyExists = syserr.NewResourceAlreadyExists[Organization]()
)

// - Domain Service(s) -

// getByID retrieves an [Organization] by its unique identifier.
func getByID(ctx context.Context, r persistence.ReadRepository[string, Organization], id string) (Organization, error) {
	org, err := r.FindByKey(ctx, id)
	if err != nil {
		return Organization{}, err
	} else if org == nil {
		return Organization{}, ErrNotFound
	}
	return *org, nil
}

// existByName checks if an [Organization] exists by its name.
func existByName(ctx context.Context, r Repository, name string) error {
	ok, err := r.ExistsByName(ctx, name)
	if err != nil {
		return err
	} else if ok {
		return ErrAlreadyExists
	}
	return nil
}

// - Application Service(s) -

// -- Manager --

// DEV-NOTE: Managers (and most services writing entities) should use the transactional repository to ensure
// strong consistency. This is because the manager is responsible for the business rules and the transactional
// boundaries. If the manager uses a non-transactional repository, the manager must ensure that the repository
// is consistent with the business rules and the transactional boundaries.

// A Manager is the service that manages the [Organization] administrative operations such as
// create, update and delete.
type Manager interface {
	// Register creates a new [Organization].
	Register(ctx context.Context, args RegisterArguments) (Organization, error)
	// ModifyByID modifies an [Organization] by its unique identifier.
	ModifyByID(ctx context.Context, id string, opts ...UpdateOption) (Organization, error)
	// DeleteByID deletes an [Organization] by its unique identifier.
	DeleteByID(ctx context.Context, id string) error
}

// DEV-NOTE: Service arguments do not contain validation tags, this must be done in the transport layer (controller) or
// at domain layer (at entity level).
// Validations at domain layer shall only happen if the validation is part of the business rules (e.g.
// the name must match a certain format and length due government validations).

// RegisterArguments is the arguments required to register a new [Organization].
type RegisterArguments struct {
	ID   string
	Name string
}

// --- Implementation(s) ---

// LocalManager is a concrete implementation of the [Manager] interface that uses local resources (from the service
// perspective).
type LocalManager struct {
	repository Repository
	// DEV-NOTE: Consider the dual-write atomicity problem. If you need to publish events, and you need
	// strong consistency, consider using an event publisher that writes events into a table (outbox) in
	// the transactional database used to write entities. This publisher must retrieve the current transaction
	// and append the event-writing operations to the transaction.
	// This way, the events are only published if the transaction is committed.
	// If the transaction is rolled back, the events are not published.
	// Then use a log-trailing mechanism to propagate the events into the event infrastructure (e.g. Apache Kafka).
	eventPublisher event.Publisher
}

// compile-time assertion
var _ Manager = (*LocalManager)(nil)

// NewLocalManager creates a new [LocalManager] instance.
func NewLocalManager(r Repository, p event.Publisher) LocalManager {
	return LocalManager{repository: r, eventPublisher: p}
}

// Register creates a new [Organization].
func (l LocalManager) Register(ctx context.Context, args RegisterArguments) (Organization, error) {
	err := existByName(ctx, l.repository, args.Name)
	if err != nil {
		return Organization{}, err
	}

	org := New(ctx, args.ID, args.Name)
	if err = l.repository.Save(ctx, org); err != nil {
		return Organization{}, err
	}

	if err = l.eventPublisher.Publish(ctx, org.PullEvents()); err != nil {
		return Organization{}, err
	}
	return org, nil
}

// ModifyByID modifies an [Organization] by its unique identifier.
func (l LocalManager) ModifyByID(ctx context.Context, id string, opts ...UpdateOption) (Organization, error) {
	if len(opts) == 0 {
		return Organization{}, nil // no-op
	}
	org, err := getByID(ctx, l.repository, id)
	if err != nil {
		return Organization{}, err
	}

	org.Update(ctx, opts...)

	if err = existByName(ctx, l.repository, org.Name()); err != nil {
		return Organization{}, err
	}

	if err = l.repository.Save(ctx, org); err != nil {
		return Organization{}, err
	}
	if err = l.eventPublisher.Publish(ctx, org.PullEvents()); err != nil {
		return Organization{}, err
	}
	return org, nil
}

// DeleteByID deletes an [Organization] by its unique identifier.
func (l LocalManager) DeleteByID(ctx context.Context, id string) error {
	// DEV-NOTE: If you delete by ID, no events can be propagated as they are appended to
	// the entity.
	org, err := getByID(ctx, l.repository, id)
	if errors.Is(err, syserr.ErrResourceNotFound) {
		return nil // no-op
	} else if err != nil {
		return err
	}
	org.Delete(ctx)
	if err = l.repository.Delete(ctx, org); err != nil {
		return err
	}
	return l.eventPublisher.Publish(ctx, org.PullEvents())
}

// -- Fetcher --

// A Fetcher is the service that retrieves [Organization] information.
type Fetcher interface {
	// GetByID retrieves an [Organization] by its unique identifier.
	GetByID(ctx context.Context, id string) (Organization, error)
}

// --- Implementation(s) ---

// LocalFetcher is a concrete implementation of the [Fetcher] interface that uses local resources (from the service
// perspective).
type LocalFetcher struct {
	repository ReadRepository
}

// compile-time assertion
var _ Fetcher = (*LocalFetcher)(nil)

// NewLocalFetcher creates a new [LocalFetcher] instance.
func NewLocalFetcher(r ReadRepository) LocalFetcher {
	return LocalFetcher{repository: r}
}

// GetByID retrieves an [Organization] by its unique identifier.
func (l LocalFetcher) GetByID(ctx context.Context, id string) (Organization, error) {
	return getByID(ctx, l.repository, id)
}

// -- Lister --

// A Lister is the service that lists [Organization] information.
type Lister interface {
	// List retrieves a list of [Organization] entities.
	List(ctx context.Context, opts ...ListOption) (*paging.Page[Organization], error)
}

// --- Option(s) ---
type listOptions struct {
	pageOpts           paging.Options
	findNonDeletedOnly bool
}

// ListOption represents an option for listing [Organization] entities.
type ListOption func(*listOptions)

// WithListPageOptions sets the pagination options ([paging.Option]) for the list operation.
func WithListPageOptions(opts ...paging.Option) ListOption {
	return func(o *listOptions) {
		for _, opt := range opts {
			opt(&o.pageOpts)
		}
	}
}

// WithListNonDeletedOnly sets the option to find only non-deleted entities.
func WithListNonDeletedOnly() ListOption {
	return func(o *listOptions) {
		o.findNonDeletedOnly = true
	}
}

// --- Implementation(s) ---

// LocalLister is a concrete implementation of the [Lister] interface that uses local resources (from the service
// perspective).
type LocalLister struct {
	repository ReadRepository
}

// compile-time assertion
var _ Lister = (*LocalLister)(nil)

// NewLocalLister creates a new [LocalLister] instance.
func NewLocalLister(r ReadRepository) LocalLister {
	return LocalLister{repository: r}
}

// List retrieves a list of [Organization] entities.
func (l LocalLister) List(ctx context.Context, opts ...ListOption) (*paging.Page[Organization], error) {
	return l.repository.FindAll(ctx, opts...)
}
