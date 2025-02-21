package employee

import (
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
)

const (
	_idField = "employee_id"
)

type ReadRepository interface {
	persistence.ReadRepository[string, Employee]
	paging.Repository[Employee]
}
