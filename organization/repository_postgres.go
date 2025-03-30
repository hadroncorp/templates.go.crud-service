package organization

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/hadroncorp/geck/persistence/audit"
	"github.com/hadroncorp/geck/persistence/paging"
	gecksql "github.com/hadroncorp/geck/persistence/sql"
	"github.com/samber/lo"

	"github.com/hadroncorp/service-template/internal/postgresgen"
)

// DEV-NOTE: Repositories are separated into write and read repositories as they might have different
// requirements and implementations. For example, a write repository might have a transactional
// implementation, while a read repository might not.
//
// Also, a read repository might not be strong consistent, while a write repository might be.
//
// Finally, even though they might be using the same database engine, a read repository might use a different
// connection pool configuration than a write repository, even connecting to a different set of
// instances (e.g. read replicas).

// - Write Repository(s) -

// PostgresRepository is the concrete implementation of the [Repository] interface for Postgres.
type PostgresRepository struct {
	db *postgresgen.Queries
}

// compile-time assertion(s)
var (
	_ Repository = (*PostgresRepository)(nil)
)

// NewPostgresRepository creates a new [PostgresRepository] instance.
func NewPostgresRepository(db gecksql.DB) PostgresRepository {
	return PostgresRepository{
		db: postgresgen.New(db),
	}
}

func (p PostgresRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	return p.db.ExistOrganizationByName(ctx, name)
}

func (p PostgresRepository) Save(ctx context.Context, entity Organization) error {
	if entity.IsNew() {
		return p.db.CreateOrganization(ctx, postgresgen.CreateOrganizationParams{
			OrganizationID: entity.id,
			Name:           entity.name,
			CreateTime:     entity.CreateTime(),
			CreateBy:       entity.CreateBy(),
			LastUpdateTime: entity.LastUpdateTime(),
			LastUpdateBy:   entity.LastUpdateBy(),
			RowVersion:     int64(entity.Version()),
			IsDeleted:      entity.IsDeleted(),
		})
	}
	return p.db.UpdateOrganization(ctx, postgresgen.UpdateOrganizationParams{
		OrganizationID: entity.id,
		Name:           entity.name,
		LastUpdateTime: entity.LastUpdateTime(),
		LastUpdateBy:   entity.LastUpdateBy(),
		RowVersion:     int64(entity.Version()),
		IsDeleted:      entity.IsDeleted(),
	})
}

func (p PostgresRepository) DeleteByKey(ctx context.Context, key string) error {
	return p.db.DeleteOrganization(ctx, key)
}

func (p PostgresRepository) Delete(ctx context.Context, entity Organization) error {
	return p.DeleteByKey(ctx, entity.id)
}

func (p PostgresRepository) FindByKey(ctx context.Context, key string) (*Organization, error) {
	model, err := p.db.GetOrganizationByID(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &Organization{
		id:   model.OrganizationID,
		name: model.Name,
		Auditable: audit.New(audit.NewArgs{
			CreateTime:     model.CreateTime,
			CreateBy:       model.CreateBy,
			LastUpdateTime: model.LastUpdateTime,
			LastUpdateBy:   model.LastUpdateBy,
			Version:        uint64(model.RowVersion),
			IsDeleted:      model.IsDeleted,
		}),
	}, nil
}

// - Read Repository(s) -

// PostgresReadRepository is the concrete implementation of the [ReadRepository] interface for Postgres.
type PostgresReadRepository struct {
	db                 *postgresgen.Queries
	pageTokenCipherKey []byte
}

// compile-time assertion(s)
var (
	_ ReadRepository = (*PostgresReadRepository)(nil)
)

// NewPostgresReadRepository creates a new [PostgresReadRepository] instance.
func NewPostgresReadRepository(db gecksql.DB, tokenConfig paging.TokenConfig) PostgresReadRepository {
	return PostgresReadRepository{
		db:                 postgresgen.New(db),
		pageTokenCipherKey: tokenConfig.CipherKeyBytes,
	}
}

func (p PostgresReadRepository) FindAll(ctx context.Context, opts ...ListOption) (*paging.Page[Organization], error) {
	listOpts := listOptions{}
	for _, opt := range opts {
		opt(&listOpts)
	}

	var queryParams postgresgen.ListOrganizationsParams
	if listOpts.pageOpts.HasPageToken() {
		if err := paging.ParseToken(p.pageTokenCipherKey, listOpts.pageOpts.PageToken(), &queryParams); err != nil {
			return nil, err
		}
	} else {
		queryParams.IsDeleted = sql.NullBool{
			Bool:  false,
			Valid: listOpts.findNonDeletedOnly,
		}
		queryParams.PageSize = sql.NullInt32{
			Int32: int32(listOpts.pageOpts.Limit()),
			Valid: listOpts.pageOpts.Limit() > 0,
		}
	}
	models, err := p.db.ListOrganizations(ctx, queryParams)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if queryParams.IsCursorForward.Valid && !queryParams.IsCursorForward.Bool {
		slices.Reverse(models)
	}

	hasPages, err := p.db.HasMorePagesOrganizationList(ctx, postgresgen.HasMorePagesOrganizationListParams{
		CursorNext: sql.NullTime{
			Time:  models[len(models)-1].CreateTime,
			Valid: true,
		},
		CursorPrev: sql.NullTime{
			Time:  models[0].CreateTime,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	var (
		prevToken string
		nextToken string
	)
	if hasPages.HasPrev {
		prevToken, err = paging.NewToken(p.pageTokenCipherKey, postgresgen.ListOrganizationsParams{
			IsDeleted: queryParams.IsDeleted,
			CursorValue: sql.NullTime{
				Time:  models[0].CreateTime,
				Valid: true,
			},
			IsCursorForward: sql.NullBool{
				Bool:  false,
				Valid: true,
			},
			PageSize: queryParams.PageSize,
		})
		if err != nil {
			return nil, err
		}
	}

	if hasPages.HasNext {
		nextToken, err = paging.NewToken(p.pageTokenCipherKey, postgresgen.ListOrganizationsParams{
			IsDeleted: queryParams.IsDeleted,
			CursorValue: sql.NullTime{
				Time:  models[len(models)-1].CreateTime,
				Valid: true,
			},
			IsCursorForward: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
			PageSize: queryParams.PageSize,
		})
		if err != nil {
			return nil, err
		}
	}

	return &paging.Page[Organization]{
		TotalItems:        len(models),
		PreviousPageToken: prevToken,
		NextPageToken:     nextToken,
		Items: lo.Map(models, func(item postgresgen.Organization, _ int) Organization {
			return Organization{
				id:   item.OrganizationID,
				name: item.Name,
				Auditable: audit.New(audit.NewArgs{
					CreateTime:     item.CreateTime,
					CreateBy:       item.CreateBy,
					LastUpdateTime: item.LastUpdateTime,
					LastUpdateBy:   item.LastUpdateBy,
					Version:        uint64(item.RowVersion),
					IsDeleted:      item.IsDeleted,
				}),
			}
		}),
	}, nil
}

func (p PostgresReadRepository) FindByKey(ctx context.Context, key string) (*Organization, error) {
	model, err := p.db.GetOrganizationByID(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &Organization{
		id:   model.OrganizationID,
		name: model.Name,
		Auditable: audit.New(audit.NewArgs{
			CreateTime:     model.CreateTime,
			CreateBy:       model.CreateBy,
			LastUpdateTime: model.LastUpdateTime,
			LastUpdateBy:   model.LastUpdateBy,
			Version:        uint64(model.RowVersion),
			IsDeleted:      model.IsDeleted,
		}),
	}, nil
}
