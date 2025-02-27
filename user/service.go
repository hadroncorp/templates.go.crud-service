package user

import (
	"context"

	"github.com/hadroncorp/geck/persistence/criteria"
)

// --> Application Services <--

// -> Fetcher <-

type Fetcher interface {
	GetByKey(ctx context.Context, key string) (User, error)
	ListByKeys(ctx context.Context, keys []string) ([]User, error)
}

type LocalFetcher struct {
	repository ReadRepository
}

// compile-time assertion(s)
var _ Fetcher = (*LocalFetcher)(nil)

func (l LocalFetcher) GetByKey(ctx context.Context, key string) (User, error) {
	employee, err := l.repository.FindByKey(ctx, key)
	if err != nil {
		return User{}, err
	} else if employee == nil {
		return User{}, ErrNotFound
	}
	return *employee, nil
}

func (l LocalFetcher) ListByKeys(ctx context.Context, keys []string) ([]User, error) {
	users, err := l.repository.FindAllByKeys(ctx, keys)
	if err != nil {
		return nil, err
	} else if len(users) == 0 {
		return nil, ErrNotFound
	}
	return users, nil
}
