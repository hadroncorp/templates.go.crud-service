//go:build integration

package organization_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hadroncorp/geck/persistence/paging"
	"github.com/hadroncorp/geck/persistence/postgres/postgrestest"
	gecksql "github.com/hadroncorp/geck/persistence/sql"
	"github.com/stretchr/testify/suite"

	"github.com/hadroncorp/service-template/internal/postgresgen"
	"github.com/hadroncorp/service-template/organization"
)

type postgresRepositoryIntegrationSuite struct {
	suite.Suite

	baseCtx           context.Context
	baseCtxCancelFunc context.CancelFunc
	dbContainer       *postgrestest.Container
	db                gecksql.DB
	queryer           *postgresgen.Queries
	repository        organization.Repository
	readRepository    organization.ReadRepository
}

func TestLocalManagerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(postgresRepositoryIntegrationSuite))
}

func (s *postgresRepositoryIntegrationSuite) SetupSuite() {
	// setup context
	const testSuiteTimeout = time.Minute
	s.baseCtx, s.baseCtxCancelFunc = context.WithTimeout(context.Background(), testSuiteTimeout)

	// setup container
	var err error
	s.dbContainer, err = postgrestest.NewContainer(s.baseCtx, s.T())
	s.Require().NoError(err)
	s.Require().NotNil(s.dbContainer)

	// bootstrap container
	var db *sql.DB
	db, err = postgrestest.StartContainer(s.baseCtx, s.T(), s.dbContainer, "./thirdparty/postgres/migrations")
	s.Require().NoError(err)
	s.Require().NotNil(db)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s.db = gecksql.NewDB(db,
		gecksql.WithInterceptor(
			gecksql.NewDatabaseLogger(logger),
		),
		gecksql.WithInterceptor(
			gecksql.NewDatabaseTxPropagator(gecksql.WithAutoCreateTx()),
		),
	)

	// setup seeds
	s.queryer = postgresgen.New(s.db)
	s.Require().NoError(s.execSeed())

	// setup repository
	s.repository = organization.NewPostgresRepository(s.db)
	tokenConfig, err := paging.NewTokenConfig()
	s.Require().NoError(err)
	s.readRepository = organization.NewPostgresReadRepository(s.db, tokenConfig)
}

func (s *postgresRepositoryIntegrationSuite) TearDownSuite() {
	shutdownCtx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	s.Assert().NoError(s.dbContainer.Instance.Terminate(shutdownCtx))
}

func (s *postgresRepositoryIntegrationSuite) execSeed() error {
	now := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	createCommands := []postgresgen.CreateOrganizationParams{
		{
			OrganizationID: "1",
			Name:           "foo",
			CreateTime:     now.Add(time.Minute),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
		{
			OrganizationID: "2",
			Name:           "baz",
			CreateTime:     now.Add(time.Minute * 2),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute * 2),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
		{
			OrganizationID: "3",
			Name:           "to-delete-by-key",
			CreateTime:     now.Add(time.Minute * 3),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute * 3),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
		{
			OrganizationID: "4",
			Name:           "to-delete",
			CreateTime:     now.Add(time.Minute * 4),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute * 5),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
	}

	tx, err := s.db.BeginTx(s.baseCtx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	defer func() {
		s.Assert().True(errors.Is(tx.Rollback(), sql.ErrTxDone))
	}()
	qtx := s.queryer.WithTx(tx)
	for _, cmd := range createCommands {
		if err = qtx.CreateOrganization(s.baseCtx, cmd); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_ExistsByName_Exists() {
	// arrange
	// act
	exists, err := s.repository.ExistsByName(s.baseCtx, "foo")
	// assert
	s.Assert().NoError(err)
	s.Assert().True(exists)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_ExistsByName_NotExists() {
	// arrange
	// act
	exists, err := s.repository.ExistsByName(s.baseCtx, strconv.Itoa(rand.Int()))
	// assert
	s.Assert().NoError(err)
	s.Assert().False(exists)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_Save_Create() {
	// arrange
	entity := organization.New(s.baseCtx, strconv.Itoa(rand.Int()), "bar")
	// act
	err := s.repository.Save(s.baseCtx, entity)
	// assert
	s.Assert().NoError(err)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_Save_Update() {
	// arrange
	entity := organization.New(s.baseCtx, "2", "baz")
	_ = entity.Update(s.baseCtx, organization.WithUpdatedName("baz-updated"))
	// act
	err := s.repository.Save(s.baseCtx, entity)
	// assert

	s.Assert().NoError(err)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_DeleteByKey() {
	// arrange
	// act
	err := s.repository.DeleteByKey(s.baseCtx, "3")
	// assert
	s.Assert().NoError(err)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_Delete() {
	// arrange
	entity := organization.New(s.baseCtx, "4", "to-delete")
	// act
	err := s.repository.Delete(s.baseCtx, entity)
	// assert
	s.Assert().NoError(err)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_FindByKey_Exists() {
	// arrange
	// act
	entity, err := s.repository.FindByKey(s.baseCtx, "1")
	// assert
	s.Assert().NoError(err)
	s.Assert().NotNil(entity)
}

func (s *postgresRepositoryIntegrationSuite) TestReadPostgresRepository_FindByKey_Exists() {
	// arrange
	// act
	entity, err := s.readRepository.FindByKey(s.baseCtx, "1")
	// assert
	s.Assert().NoError(err)
	s.Assert().NotNil(entity)
}

func (s *postgresRepositoryIntegrationSuite) TestPostgresRepository_FindByKey_NotExists() {
	// arrange
	// act
	entity, err := s.repository.FindByKey(s.baseCtx, strconv.Itoa(rand.Int()))
	// assert
	s.Assert().NoError(err)
	s.Assert().Nil(entity)
}

func (s *postgresRepositoryIntegrationSuite) TestReadPostgresRepository_FindAll() {
	// arrange
	const pageSize = 1
	// act
	page, err := s.readRepository.FindAll(s.baseCtx,
		organization.WithListPageOptions(
			paging.WithLimit(pageSize),
		),
		organization.WithListNonDeletedOnly(),
	)
	// assert
	s.Assert().NoError(err)
	s.Assert().NotEmpty(page)

	firstPageHead := page.Items[0]

	page, err = s.readRepository.FindAll(s.baseCtx,
		organization.WithListPageOptions(
			paging.WithPageToken(page.NextPageToken),
		),
	)
	s.Assert().NoError(err)
	s.Assert().NotEmpty(page)
	s.Assert().Equal("2", page.Items[0].ID())
	s.Assert().True(page.Items[0].CreateTime().After(firstPageHead.CreateTime()))

	page, err = s.readRepository.FindAll(s.baseCtx,
		organization.WithListPageOptions(
			paging.WithPageToken(page.PreviousPageToken),
		),
	)
	s.Assert().NoError(err)
	s.Assert().NotEmpty(page)
	s.Assert().Zero(page.PreviousPageToken)
	s.Assert().NotZero(page.NextPageToken)
	s.Assert().Equal(firstPageHead.CreateTime(), page.Items[0].CreateTime())
}
