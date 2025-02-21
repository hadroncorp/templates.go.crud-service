package place

import (
	"context"

	"github.com/hadroncorp/geck/persistence/criteria"
)

// --> Application Services <--

// -> Fetcher <-

type Fetcher interface {
	GetByKey(ctx context.Context, key string) (Place, error)
	ListByKeys(ctx context.Context, keys []string) ([]Place, error)
}

type LocalFetcher struct {
	repository ReadRepository
}

// compile-time assertion(s)
var _ Fetcher = (*LocalFetcher)(nil)

func (l LocalFetcher) GetByKey(ctx context.Context, key string) (Place, error) {
	employee, err := l.repository.FindByKey(ctx, key)
	if err != nil {
		return Place{}, err
	} else if employee == nil {
		return Place{}, ErrNotFound
	}
	return *employee, nil
}

func (l LocalFetcher) ListByKeys(ctx context.Context, keys []string) ([]Place, error) {
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
