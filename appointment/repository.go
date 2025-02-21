package appointment

import (
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
)

const (
	// criteria field names
	_scheduledByField  = "user_id"
	_placeIDField      = "place_id"
	_scheduleTimeField = "schedule_time"
)

type Repository interface {
	persistence.WriteRepository[string, Appointment]
	persistence.ReadRepository[string, Appointment]
}

type ReadRepository interface {
	persistence.ReadRepository[string, ReadModel]
}

type ListUserRepository interface {
	paging.Repository[ListUserReadModel]
}

type ListPlaceRepository interface {
	paging.Repository[ListPlaceReadModel]
}
