package place

import (
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
)

const (
	_idField = "place_id"
)

type ReadRepository interface {
	persistence.ReadRepository[string, Place]
	paging.Repository[Place]
}
