package user

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/criteria"
	"github.com/hadroncorp/geck/persistence/paging"
	gecksql "github.com/hadroncorp/geck/persistence/sql"
	"github.com/samber/lo"
)

const (
	_idField = "user_id"
)

type ReadRepository interface {
	persistence.ReadRepository[string, User]
	paging.Repository[User]
	FindAllByKeys(ctx context.Context, keys []string) ([]User, error)
}

// -- Postgres --

type PostgresReadRepository struct {
	db              *goqu.Database
	tableName       string
	fieldTranslator persistence.FieldTranslator
}

// compile-time assertion(s)
var _ ReadRepository = (*PostgresReadRepository)(nil)

func NewPostgresReadRepository(db gecksql.DB) PostgresReadRepository {
	return PostgresReadRepository{
		db:        goqu.New("postgres", db),
		tableName: "platform_users",
		fieldTranslator: persistence.FieldTranslator{
			Source: map[string]string{
				_idField: "user_id",
			},
		},
	}
}

func (p PostgresReadRepository) FindByKey(ctx context.Context, key string) (*User, error) {
	model := postgresUser{}
	found, err := p.db.From(p.tableName).
		Where(goqu.C("user_id").Eq(key)).
		ScanStructContext(ctx, &model)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, nil
	}
	entity := model.toEntity()
	return &entity, nil
}

func (p PostgresReadRepository) FindAllByKeys(ctx context.Context, keys []string) ([]User, error) {
	models := make([]postgresUser, 0, len(keys))
	cte := p.db.From(goqu.L("unnest(?) AS keys", keys))
	err := p.db.From(p.tableName).
		With("temp_keys", cte).
		Where(goqu.C("user_id").In(
			goqu.From("temp_keys").Select("keys"),
		)).
		ScanStructsContext(ctx, &models)
	if err != nil {
		return nil, err
	}
	return lo.Map(models, func(item postgresUser, _ int) User {
		return item.toEntity()
	}), nil
}

func (p PostgresReadRepository) FindAll(ctx context.Context, opts ...criteria.Option) (*paging.Page[User], error) {
	criteriaOptions := criteria.Criteria{}
	for _, opt := range opts {
		opt(&criteriaOptions)
	}

	items, err := gecksql.ExecCriteria[postgresUser](ctx, gecksql.ExecCriteriaParams{
		DB:              p.db,
		Table:           p.tableName,
		Criteria:        criteriaOptions,
		FieldTranslator: &p.fieldTranslator,
	})
	if err != nil {
		return nil, err
	} else if len(items) == 0 {
		return nil, nil
	}

	hasNext, hasPrev, err := gecksql.HasMorePages(ctx, gecksql.HasMorePagesParams{
		DB:              p.db,
		Table:           p.tableName,
		CursorName:      criteriaOptions.Sorting.Field,
		StartCursor:     items[0],
		EndCursor:       items[len(items)-1],
		Criteria:        criteriaOptions,
		FieldTranslator: &p.fieldTranslator,
	})
	if err != nil {
		return nil, err
	} else if !hasNext && !hasPrev {
		return &paging.Page[User]{
			TotalItems: len(items),
			Items: lo.Map(items, func(item postgresUser, _ int) User {
				return item.toEntity()
			}),
		}, nil
	}

	tokens, err := gecksql.NewPageTokens(ctx, gecksql.NewPageTokensParams[postgresUser]{
		DB:              p.db,
		Table:           p.tableName,
		CursorName:      criteriaOptions.Sorting.Field,
		ResultSet:       items,
		FieldTranslator: &p.fieldTranslator,
		Criteria:        criteriaOptions,
		InitialSort:     0,
	})
	if err != nil {
		return nil, err
	}
	return &paging.Page[User]{
		TotalItems:        len(items),
		PreviousPageToken: tokens.Previous,
		NextPageToken:     tokens.Next,
		Items: lo.Map(items, func(item postgresUser, _ int) User {
			return item.toEntity()
		}),
	}, nil
}
