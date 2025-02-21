package user

import (
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
)

const (
	_idField = "user_id"
)

type ReadRepository interface {
	persistence.ReadRepository[string, User]
	paging.Repository[User]
}
