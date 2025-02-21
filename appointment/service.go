package appointment

import (
	"context"
	"errors"
	"time"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/criteria"
	"github.com/hadroncorp/geck/persistence/paging"
	"github.com/samber/lo"

	"github.com/hadroncorp/service-template/employee"
	"github.com/hadroncorp/service-template/place"
	"github.com/hadroncorp/service-template/user"
)

// --> Domain Services <--

// NOTE: Remember every service name must end in 'er'. For example: Fetcher, Manager, Querier, Scanner.
// This is recommended by Google's Go best practices, you can see it implemented by yourself by getting a dive into
// Go's stdlib APIs.

// NOTE: Place/write here routines (or structures) shared to more than one application services (e.g. AdminManager, Scheduler).
// For example, [fetchByKey] routine ensures that if no entity was found, a [ErrNotFound] is returned. It also
// removes the reference from the pointer the repository issued. This is required for more than one application service.

func fetchByKey(repository persistence.ReadRepository[string, Appointment], ctx context.Context, key string) (Appointment, error) {
	appointment, err := repository.FindByKey(ctx, key)
	if err != nil {
		return Appointment{}, err
	} else if appointment == nil {
		return Appointment{}, ErrNotFound
	}
	return *appointment, nil
}

// --> Application Services <--

// NOTE: Place/write here routines (or structures) exposing features of the package's domain. Remember to enrich the entity with
// domain rules. If you have a domain rule involving several entities, then place that logic into these services.
//
// Remember to:
// 	- Use clear and concise names for your services.
//	- Use ubiquitous language (aka. domain language) (i.e. a language every stakeholder understand).
// 	- Design your struct and routine names like a library API. Use Go's stdlib as inspiration.
//	- Avoid generic names like Create, Update, Delete unless neccesary. Leave those for Management or Persistence APIs.

// -> Scheduler <-

type Scheduler interface {
	Schedule(ctx context.Context, args NewArgs, opts ...NewOption) error
	Cancel(ctx context.Context, key, reason string) error
	Reschedule(ctx context.Context, key, reason string, newTime time.Time) (Appointment, error)
}

// -- Local --

// NOTE: Using the term 'Local' as prefix is key. This tells the service implementation is local to the current bounded
// context (microservice).

type LocalScheduler struct {
	repository Repository
}

// compile-time assertion(s)
var _ Scheduler = (*LocalScheduler)(nil)

func (l LocalScheduler) Schedule(ctx context.Context, args NewArgs, opts ...NewOption) error {
	appointment, err := New(args, opts...)
	if err != nil {
		return err
	}
	return l.repository.Save(ctx, appointment)
}

func (l LocalScheduler) Cancel(ctx context.Context, key, reason string) error {
	appointment, err := fetchByKey(l.repository, ctx, key)
	if err != nil {
		return err
	}
	if err = appointment.Cancel(ctx, reason); err != nil {
		return err
	}
	return l.repository.Save(ctx, appointment)
}

func (l LocalScheduler) Reschedule(ctx context.Context, key, reason string, newTime time.Time) (Appointment, error) {
	appointment, err := fetchByKey(l.repository, ctx, key)
	if err != nil {
		return Appointment{}, err
	}
	if err = appointment.Reschedule(ctx, reason, newTime); err != nil {
		return Appointment{}, err
	}
	if err = l.repository.Save(ctx, appointment); err != nil {
		return Appointment{}, err
	}
	return appointment, nil
}

// -> Admin Manager <-

// NOTE: As you can see, by using ubiquitous language and defining a clear API design, we can tell developers what operations
// are for administrators (AdminManager) and what operations are available for most users (Scheduler, Fetcher). No Security APIs required
// to understand that.

type AdminManager interface {
	UpdateByKey(ctx context.Context, key string, opts ...UpdateOption) (Appointment, error)
	DeleteByKey(ctx context.Context, key string) error
}

// -- Local --

type LocalAdminManager struct {
	repository Repository
}

// compile-time assertion(s)
var _ AdminManager = (*LocalAdminManager)(nil)

func (l LocalAdminManager) UpdateByKey(ctx context.Context, key string, opts ...UpdateOption) (Appointment, error) {
	appointment, err := fetchByKey(l.repository, ctx, key)
	if err != nil {
		return Appointment{}, err
	}
	if err = appointment.Update(ctx, opts...); err != nil {
		return Appointment{}, err
	}
	if err = l.repository.Save(ctx, appointment); err != nil {
		return Appointment{}, err
	}
	return appointment, nil
}

func (l LocalAdminManager) DeleteByKey(ctx context.Context, key string) error {
	appointment, err := fetchByKey(l.repository, ctx, key)
	if errors.Is(err, ErrNotFound) {
		return nil // no-op
	} else if err != nil {
		return err
	}
	appointment.Delete(ctx)
	return l.repository.Delete(ctx, appointment)
}

// -> Fetcher <-

// NOTE: Separating read from write operations is key. Read operations can behave very differently from transactional operations
// as they might use another storage to fetch data. Even if its the same "database", there might be read nodes specially
// provisioned by your team to reduce overhead from the master (write) nodes, improving resiliency and reducing the need to
// scale for just read operations. By defining a fetcher (or similar) APIs, you already telling API clients
// this is a service for read operations. They dont need to see internal implementations.

// NOTE: Transactional services (e.g. AdminManager, Scheduler) must NOT use Fetcher APIs as they do not guarantee data consistency.
// Exceptions might be data stored in external providers/storages, nevertheless, consider there might be data incosistencies. This
// can be solved with distributed transactions or 2-phase commit procesess, both of them are hard to implement and also to maintain.

// NOTE: When dealing with read operations, you might have different options to retrieve dependency entities:
//
// A. Denormalized data: All required data is stored at row/document level. For each row/document in your data space (i.e. table, collection),
// all of the data from entity dependencies are stored as well.
//
// If those dependencies (with their denormalized fields) are immutable or not frequently updated, this might be the best option as is faster than others in simpler terms.
// If not (updated frequently), there will be a lot of heavy (and complex) write operations to update just one dependency entity (as it might be duplicated
// for N records for M data spaces); these write ops might even provision a lock to avoid concurrent operations read old data.
// Nevertheless, if using a modern persistence storage engine, you can use views (or even better, materialized views) to let the engine handle this for you;
// you might need to manually trigger updates (for materialized views).
//
// Finally, if dependency entities are external to the current microservice, apply lazy-loading patterns (calling services synchronously, then store the result locally) or adopt an Event-Driven
// Architecture (EDA) for your platform (asynchronous and easier to scale, harder to debug and handle failure scenarios).
//
// B. Aggregation at service level: Referenced entities are not guaranteed to be stored on the same persistence store or data space (i.e. table, collection) as the required entity.
// Referenced entites might even come from external sources like another microservice or even a third-party provider/service.
//
// This option will fetch dependency entites synchronously for each original entity, populating its fields with fetched dependencies. All done at application service-level.
//
// You can improve performance by fetching these with concurrency and by using Batch APIs. Just be sure to define a concurrent worker count constraint (with semaphore
// pattern and/or with connection pooling) to avoid congesting your own service.
//
// Recommended when persistence storages are local to the current microservice (bounded context). If storages are external, you should consider another option.
//
// C. Reference identifiers-only: No aggregations needed. Just read from your read database (can be the same as the transactional one).
// Dependency entities will be shown by their identifiers only.
//
// D. Aggregation at gateway level: Referenced entities are not guaranteed to be stored on the same persistence store or data space (i.e. table, collection) as the required entity.
// Referenced entites might even come from external sources like another microservice or even a third-party provider/service.
//
// This option will fetch dependency entites synchronously for each original entity, populating its fields with fetched dependencies. All done at gateway service-level.
// Gateway might be a separate microservice or just a local-defined service.
//
// A separated microservice might be the best option if you look to encapsulate several operations for a certain client (e.g. mobile, web). This is known as the gateway
// aggregation cloud pattern.
//
// If choosing the latter, you might define a GatewayFetcher service which is basically implementing the facade design pattern.
//
// Just be careful as this can be a bottleneck for your service and can lead to cascading failures as well. Rely heavily on concurrency and resiliency patterns.
//
// More information about the gateway aggregation pattern here: https://learn.microsoft.com/en-us/azure/architecture/patterns/gateway-aggregation.

type Fetcher interface {
	GetByKey(ctx context.Context, key string) (ReadModel, error)
	ListByUser(ctx context.Context, args ListByUserArgs) (*paging.Page[ListUserReadModel], error)
	ListByPlace(ctx context.Context, args ListByPlaceArgs) (*paging.Page[ListPlaceReadModel], error)
}

type ListByUserArgs struct {
	criteria.ArgumentTemplate
	UserID string
}

type ListByPlaceArgs struct {
	criteria.ArgumentTemplate
	PlaceID string
}

// -- Local --

type LocalGatewayFetcher struct {
	placeFetcher    place.Fetcher
	employeeFetcher employee.Fetcher
	userFetcher     user.Fetcher

	repository          ReadRepository
	listUserRepository  ListUserRepository
	listPlaceRepository ListPlaceRepository
}

// compile-time assertion(s)
var _ Fetcher = (*LocalGatewayFetcher)(nil)

func (l LocalGatewayFetcher) GetByKey(ctx context.Context, key string) (ReadModel, error) {
	appointment, err := l.repository.FindByKey(ctx, key)
	if err != nil {
		return ReadModel{}, err
	} else if appointment == nil {
		return ReadModel{}, ErrNotFound
	}

	appointment.Place, err = l.placeFetcher.GetByKey(ctx, appointment.Place.ID())
	if err != nil {
		return ReadModel{}, err
	}
	appointment.ScheduledBy, err = l.userFetcher.GetByKey(ctx, appointment.ScheduledBy.ID())
	if err != nil {
		return ReadModel{}, err
	}
	appointment.TargetedTo, err = l.employeeFetcher.GetByKey(ctx, appointment.TargetedTo.ID())
	if err != nil {
		return ReadModel{}, err
	}
	return *appointment, nil
}

func (l LocalGatewayFetcher) ListByUser(ctx context.Context, args ListByUserArgs) (*paging.Page[ListUserReadModel], error) {
	page, err := l.listUserRepository.FindAll(ctx,
		criteria.WithFilter(_scheduledByField, criteria.Equal, args.UserID),
		criteria.WithPageSize(args.PageSize),
		criteria.WithPageToken(args.PageToken),
		criteria.WithSorting(_scheduleTimeField, criteria.SortDescending),
	)
	if err != nil {
		return nil, err
	} else if page == nil {
		return nil, ErrNotFound
	}

	placeIDs := make([]string, 0, len(page.Items))
	employeeIDs := make([]string, 0, len(page.Items))
	for i := range page.Items {
		placeIDs = append(placeIDs, page.Items[i].Place.ID())
		employeeIDs = append(employeeIDs, page.Items[i].TargetedTo.ID())
	}

	// using batching APIs
	places, err := l.placeFetcher.ListByKeys(ctx, placeIDs)
	if err != nil {
		return nil, err
	}
	placesMap := lo.Associate(places, func(item place.Place) (string, place.Place) {
		return item.ID(), item
	})
	employees, err := l.employeeFetcher.ListByKeys(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}
	employeesMap := lo.Associate(employees, func(item employee.Employee) (string, employee.Employee) {
		return item.ID(), item
	})

	for i := range page.Items {
		page.Items[i].Place = placesMap[page.Items[i].Place.ID()]
		page.Items[i].TargetedTo = employeesMap[page.Items[i].TargetedTo.ID()]
	}
	return page, nil
}

func (l LocalGatewayFetcher) ListByPlace(ctx context.Context, args ListByPlaceArgs) (*paging.Page[ListPlaceReadModel], error) {
	page, err := l.listPlaceRepository.FindAll(ctx,
		criteria.WithFilter(_placeIDField, criteria.Equal, args.PlaceID),
		criteria.WithPageSize(args.PageSize),
		criteria.WithPageToken(args.PageToken),
		criteria.WithSorting(_scheduleTimeField, criteria.SortDescending),
	)
	if err != nil {
		return nil, err
	} else if page == nil {
		return nil, ErrNotFound
	}

	userIDs := make([]string, 0, len(page.Items))
	employeeIDs := make([]string, 0, len(page.Items))
	for i := range page.Items {
		userIDs = append(userIDs, page.Items[i].ScheduledBy.ID())
		employeeIDs = append(employeeIDs, page.Items[i].TargetedTo.ID())
	}

	// using batching APIs
	users, err := l.userFetcher.ListByKeys(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	usersMap := lo.Associate(users, func(item user.User) (string, user.User) {
		return item.ID(), item
	})
	employees, err := l.employeeFetcher.ListByKeys(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}
	employeesMap := lo.Associate(employees, func(item employee.Employee) (string, employee.Employee) {
		return item.ID(), item
	})

	for i := range page.Items {
		page.Items[i].ScheduledBy = usersMap[page.Items[i].ScheduledBy.ID()]
		page.Items[i].TargetedTo = employeesMap[page.Items[i].TargetedTo.ID()]
	}
	return page, nil
}
