//go:build integration

package organization_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hadroncorp/geck/event"
	"github.com/hadroncorp/geck/persistence/identifier"
	"github.com/hadroncorp/geck/persistence/postgres/postgrestest"
	gecksql "github.com/hadroncorp/geck/persistence/sql"
	"github.com/hadroncorp/geck/transport/stream/kafka"
	"github.com/hadroncorp/geck/transport/stream/kafka/kafkatest"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"

	"event-schema-registry/iampb"
	"github.com/hadroncorp/service-template/internal/postgresgen"
	"github.com/hadroncorp/service-template/organization"
)

// DEV-NOTE: These suites use actual infrastructure components (e.g. databases, message brokers, etc., NO MOCKS).

type managerIntegrationSuite struct {
	suite.Suite

	psqlContainer       *postgrestest.Container
	dbClient            gecksql.DB
	queryer             *postgresgen.Queries
	kafkaContainer      *kafkatest.Container
	kafkaProducerClient *kgo.Client

	manager organization.LocalManager
}

func TestManagerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(managerIntegrationSuite))
}

func (s *managerIntegrationSuite) SetupSuite() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	// Start postgres
	var err error
	s.psqlContainer, err = postgrestest.NewContainer(ctx, s.T())
	s.Require().NoError(err)
	db, err := postgrestest.StartContainer(ctx, s.T(), s.psqlContainer, "./thirdparty/postgres/migrations")
	s.Require().NoError(err)
	s.dbClient = gecksql.NewDB(db,
		gecksql.WithInterceptor(gecksql.NewDatabaseTxPropagator()),
	)
	// setup seeds
	s.queryer = postgresgen.New(s.dbClient)
	s.Require().NoError(s.execSeed(ctx))

	// Start Kafka
	s.kafkaContainer, err = kafkatest.NewContainer(ctx, s.T())
	s.Require().NoError(err)
	s.Require().NotZero(s.kafkaContainer.SeedBrokerAddrs)
	s.T().Setenv("KAFKA_SEED_BROKERS", strings.Join(s.kafkaContainer.SeedBrokerAddrs, ","))
	s.Require().NoError(s.kafkaContainer.Instance.Start(ctx))
	s.Require().NoError(err)
	s.kafkaProducerClient, err = kgo.NewClient(
		kgo.SeedBrokers(s.kafkaContainer.SeedBrokerAddrs...),
		kgo.AllowAutoTopicCreation(),
	)
	s.Require().NoError(err)
	streamWriter := kafka.NewSyncWriter(s.kafkaProducerClient)

	// Start manager
	eventPublisher := event.NewStreamPublisher(streamWriter, identifier.FactoryKSUID{})
	repo := organization.NewPostgresRepository(s.dbClient)
	s.manager = organization.NewLocalManager(repo, eventPublisher)
}

func (s *managerIntegrationSuite) execSeed(ctx context.Context) error {
	now := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	createCommands := []postgresgen.CreateOrganizationParams{
		{
			OrganizationID: "1",
			Name:           "to-update",
			CreateTime:     now.Add(time.Minute),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
		{
			OrganizationID: "2",
			Name:           "to-delete",
			CreateTime:     now.Add(time.Minute * 4),
			CreateBy:       "some-user",
			LastUpdateTime: now.Add(time.Minute * 5),
			LastUpdateBy:   "some-user",
			RowVersion:     0,
			IsDeleted:      false,
		},
	}

	tx, err := s.dbClient.BeginTx(ctx, &sql.TxOptions{
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
		if err = qtx.CreateOrganization(ctx, cmd); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *managerIntegrationSuite) TearDownSuite() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	s.kafkaProducerClient.Close()
	s.Assert().NoError(s.psqlContainer.Instance.Terminate(ctx))
	s.Assert().NoError(s.kafkaContainer.Instance.Terminate(ctx))
}

func (s *managerIntegrationSuite) setupReader(topic event.Topic, handlerFunc kafka.ReaderHandlerFunc) func(ctx context.Context) error {
	var err error
	readerManager, err := kafka.NewChannelReaderManager(
		kafka.WithReaderManagerClientOpts(
			kgo.SeedBrokers(s.kafkaContainer.SeedBrokerAddrs...),
		),
		kafka.WithReaderManagerGroupID(fmt.Sprintf("org-test-%d", rand.Int())),
	)
	s.Require().NoError(err)
	s.Require().NoError(readerManager.Register(topic.String(), handlerFunc))
	go func() {
		_ = readerManager.Start()
	}()
	return readerManager.Close
}

func (s *managerIntegrationSuite) TestCreateOrganization() {
	// Arrange
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	wasReceived := &atomic.Bool{}
	wasReceived.Store(false)
	closeReaderFunc := s.setupReader(organization.TopicCreated, func(scopedCtx context.Context, message *kgo.Record) error {
		createdEvent := &iampb.OrganizationCreatedEvent{}
		s.Require().NoError(proto.Unmarshal(message.Value, createdEvent))
		s.Assert().Equal("3", createdEvent.GetOrganizationId())
		s.Assert().Equal("foo", createdEvent.GetName())
		wasReceived.Store(true)
		return nil
	})
	defer func() {
		s.Assert().NoError(closeReaderFunc(ctx))
	}()

	// Act
	org, err := s.manager.Register(ctx, organization.RegisterArguments{
		ID:   "3",
		Name: "foo",
	})

	// Assert
	s.Assert().NoError(err)
	s.Assert().Equal("3", org.ID())
	s.Assert().Equal("foo", org.Name())

	var exists bool
	err = s.dbClient.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM organizations WHERE organization_id = $1 LIMIT 1)", org.ID()).
		Scan(&exists)
	s.Assert().NoError(err)
	s.Assert().True(exists)

	select {
	case <-ctx.Done():
		s.T().Fatal(ctx.Err())
	case <-time.After(time.Second * 5):
		// Wait for the message to be consumed
		s.Assert().True(wasReceived.Load(), "message received")
	}
}

func (s *managerIntegrationSuite) TestUpdateOrganization() {
	// Arrange
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	wasReceived := &atomic.Bool{}
	wasReceived.Store(false)
	closeReaderFunc := s.setupReader(organization.TopicUpdated, func(scopedCtx context.Context, message *kgo.Record) error {
		wasReceived.Store(true)
		ev := &iampb.OrganizationUpdatedEvent{}
		s.Require().NoError(proto.Unmarshal(message.Value, ev))
		s.Assert().Equal("1", ev.GetOrganizationId())
		s.Assert().Equal("updated_org", ev.GetName())
		s.Assert().NotEmpty(ev.GetUpdateBy())
		s.Require().NotNil(ev.GetUpdateTime())
		s.Assert().NotZero(ev.GetUpdateTime().AsTime())
		return nil
	})
	defer func() {
		s.Assert().NoError(closeReaderFunc(ctx))
	}()

	// Act
	org, err := s.manager.ModifyByID(ctx, "1", organization.WithUpdatedName(lo.ToPtr("updated_org")))

	// Assert
	s.Assert().NoError(err)
	s.Assert().Equal("1", org.ID())
	s.Assert().Equal("updated_org", org.Name())

	var name string
	err = s.dbClient.QueryRowContext(ctx, "SELECT name FROM organizations WHERE organization_id = $1 LIMIT 1", org.ID()).
		Scan(&name)
	s.Assert().NoError(err)
	s.Assert().Equal("updated_org", name)

	select {
	case <-ctx.Done():
		s.T().Fatal(ctx.Err())
	case <-time.After(time.Second * 5):
		// Wait for the message to be consumed
		s.Assert().True(wasReceived.Load(), "message received")
	}
}

func (s *managerIntegrationSuite) TestDeleteOrganization() {
	// Arrange
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	wasReceived := &atomic.Bool{}
	wasReceived.Store(false)
	closeReaderFunc := s.setupReader(organization.TopicDeleted, func(scopedCtx context.Context, message *kgo.Record) error {
		wasReceived.Store(true)
		ev := &iampb.OrganizationDeletedEvent{}
		s.Require().NoError(proto.Unmarshal(message.Value, ev))
		s.Assert().Equal("2", ev.GetOrganizationId())
		s.Assert().NotEmpty(ev.GetDeleteBy())
		s.Require().NotNil(ev.GetDeleteTime())
		s.Assert().NotZero(ev.GetDeleteTime().AsTime())
		return nil
	})
	defer func() {
		s.Assert().NoError(closeReaderFunc(ctx))
	}()

	// Act
	err := s.manager.DeleteByID(ctx, "2")

	// Assert
	s.Assert().NoError(err)

	var exists bool
	err = s.dbClient.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM organizations WHERE organization_id = $1 LIMIT 1)", "2").
		Scan(&exists)
	s.Assert().NoError(err)
	s.Assert().False(exists)

	select {
	case <-ctx.Done():
		s.T().Fatal(ctx.Err())
	case <-time.After(time.Second * 5):
		// Wait for the message to be consumed
		s.Assert().True(wasReceived.Load(), "message received")
	}
}
