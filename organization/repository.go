package organization

import (
	"context"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/paging"
)

// DEV-NOTE: In Repository, do not define routines with advanced filtering as the repository
// might be an OLTP (Online Transaction Processing) database and not a OLAP (Online Analytical Processing) database.
// If you need advanced filtering, consider creating a new repository that is specific for that purpose.
//
// If you do otherwise, you might end up with a repository that is not efficient for the purpose it was designed for.

// Repository offers a set of routines to manage [Organization] persistence store operations.
type Repository interface {
	persistence.WriteRepository[string, Organization]
	persistence.ReadRepository[string, Organization]
	// ExistsByName checks if an [Organization] exists by its name.
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// ReadRepository offers a set of routines to manage [Organization] read operations.
type ReadRepository interface {
	persistence.ReadRepository[string, Organization]
	FindAll(ctx context.Context, opts ...ListOption) (*paging.Page[Organization], error)
}
