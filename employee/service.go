package employee

import (
	"context"
	
	"github.com/hadroncorp/geck/persistence/criteria"
)

// --> Application Services <--

// -> Fetcher <-

type Fetcher interface {
	GetByKey(ctx context.Context, key string) (Employee, error)
	ListByKeys(ctx context.Context, keys []string) ([]Employee, error)
}

type LocalFetcher struct {
	repository ReadRepository
}

// compile-time assertion(s)
var _ Fetcher = (*LocalFetcher)(nil)

func (l LocalFetcher) GetByKey(ctx context.Context, key string) (Employee, error) {
	employee, err := l.repository.FindByKey(ctx, key)
	if err != nil {
		return Employee{}, err
	} else if employee == nil {
		return Employee{}, ErrNotFound
	}
	return *employee, nil
}

func (l LocalFetcher) ListByKeys(ctx context.Context, keys []string) ([]Employee, error) {
	page, err := l.repository.FindAll(ctx,
		criteria.WithFilter(_idField, criteria.In, keys),
	)
	if err != nil {
		return nil, err
	} else if page == nil {
		return nil, ErrNotFound
	}
	return page.Items, nil
}
