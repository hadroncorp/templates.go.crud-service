package organization_test

import (
	"context"
	"testing"

	"github.com/hadroncorp/geck/eventmock"
	"github.com/hadroncorp/geck/persistence/paging"
	"github.com/hadroncorp/geck/security/identity"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/hadroncorp/service-template/organization"
	"github.com/hadroncorp/service-template/organizationmock"
)

type localManagerSuite struct {
	suite.Suite

	baseCtx context.Context
}

func TestLocalManagerSuite(t *testing.T) {
	suite.Run(t, new(localManagerSuite))
}

func (s *localManagerSuite) SetupSuite() {
	s.baseCtx = identity.WithPrincipal(context.Background(), identity.NewBasicPrincipal("some-user"))
}

func (s *localManagerSuite) TearDownSuite() {}

func (s *localManagerSuite) SetupTest() {}

func (s *localManagerSuite) TearDownTest() {}

func (s *localManagerSuite) TestLocalManager_Register_Valid() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		ExistsByName(s.baseCtx, "foo").
		Times(1).
		Return(false, error(nil))
	repository.EXPECT().
		Save(s.baseCtx, gomock.Any()).
		Times(1).
		Return(error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)
	eventPublisher.EXPECT().
		Publish(s.baseCtx, gomock.Any()).
		Times(1).
		Return(error(nil))

	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	out, err := manager.Register(s.baseCtx, organization.RegisterArguments{
		ID:   "1",
		Name: "foo",
	})

	// assert
	s.Assert().NoError(err)
	s.Assert().Equal("1", out.ID())
	s.Assert().Equal("foo", out.Name())
	s.Assert().NotZero(out.CreateTime())
	s.Assert().Equal("some-user", out.CreateBy())
}

func (s *localManagerSuite) TestLocalManager_Register_Already_Exists() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		ExistsByName(s.baseCtx, "foo").
		Times(1).
		Return(true, error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)
	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	out, err := manager.Register(s.baseCtx, organization.RegisterArguments{
		ID:   "1",
		Name: "foo",
	})

	// assert
	s.Assert().ErrorAs(err, &organization.ErrAlreadyExists)
	s.Assert().Zero(out.ID())
	s.Assert().Zero(out.Name())
	s.Assert().Zero(out.CreateTime())
	s.Assert().Zero(out.CreateBy())
}

func (s *localManagerSuite) TestLocalManager_ModifyByID_Noop() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	eventPublisher := eventmock.NewMockPublisher(ctrl)
	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	out, err := manager.ModifyByID(s.baseCtx, "1")

	// assert
	s.Assert().NoError(err)
	s.Assert().Equal("", out.ID())
}

func (s *localManagerSuite) TestLocalManager_ModifyByID_Modified() {
	// arrange
	ctx := identity.WithPrincipal(context.Background(), identity.NewBasicPrincipal("some-other-user"))
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		FindByKey(ctx, "1").
		Times(1).
		Return(lo.ToPtr(organization.New(s.baseCtx, "1", "foo")), error(nil))
	repository.EXPECT().
		ExistsByName(ctx, "bar").
		Times(1).
		Return(false, error(nil))
	repository.EXPECT().
		Save(ctx, gomock.Any()).
		Times(1).
		Return(error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)
	eventPublisher.EXPECT().
		Publish(ctx, gomock.Any()).
		Times(1).
		Return(error(nil))

	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	out, err := manager.ModifyByID(ctx, "1", organization.WithUpdatedName("bar"))

	// assert
	s.Assert().NoError(err)
	s.Assert().Equal("1", out.ID())
	s.Assert().Equal("bar", out.Name())
	s.Assert().NotZero(out.CreateTime())
	s.Assert().Equal("some-user", out.CreateBy())
	s.Assert().NotZero(out.LastUpdateTime())
	s.Assert().Equal("some-other-user", out.LastUpdateBy())
}

func (s *localManagerSuite) TestLocalManager_ModifyByID_Already_Exists() {
	// arrange
	ctx := identity.WithPrincipal(context.Background(), identity.NewBasicPrincipal("some-other-user"))
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		FindByKey(ctx, "1").
		Times(1).
		Return(lo.ToPtr(organization.New(s.baseCtx, "1", "foo")), error(nil))
	repository.EXPECT().
		ExistsByName(ctx, "bar").
		Times(1).
		Return(true, error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)

	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)
	_, err := manager.ModifyByID(ctx, "1", organization.WithUpdatedName("bar"))
	s.Assert().ErrorAs(err, &organization.ErrAlreadyExists)
}

func (s *localManagerSuite) TestLocalManager_DeleteByID_Found() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		FindByKey(s.baseCtx, "1").
		Times(1).
		Return(lo.ToPtr(organization.New(s.baseCtx, "1", "foo")), error(nil))
	repository.EXPECT().
		Delete(s.baseCtx, gomock.Any()).
		Times(1).
		Return(error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)
	eventPublisher.EXPECT().
		Publish(s.baseCtx, gomock.Any()).
		Times(1).
		Return(error(nil))

	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	err := manager.DeleteByID(s.baseCtx, "1")

	// assert
	s.Assert().NoError(err)
}

func (s *localManagerSuite) TestLocalManager_DeleteByID_Noop() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockRepository(ctrl)
	repository.EXPECT().
		FindByKey(s.baseCtx, "1").
		Times(1).
		Return((*organization.Organization)(nil), error(nil))
	eventPublisher := eventmock.NewMockPublisher(ctrl)

	var manager organization.Manager
	manager = organization.NewLocalManager(repository, eventPublisher)

	// act
	err := manager.DeleteByID(s.baseCtx, "1")

	// assert
	s.Assert().NoError(err)
}

type localFetcherSuite struct {
	suite.Suite
}

func TestLocalFetcherSuite(t *testing.T) {
	suite.Run(t, new(localFetcherSuite))
}

func (s *localFetcherSuite) TestLocalFetcher_FindByID_Found() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockReadRepository(ctrl)
	repository.EXPECT().
		FindByKey(gomock.Any(), "1").
		Times(1).
		Return(lo.ToPtr(organization.New(context.Background(), "1", "foo")), error(nil))

	var fetcher organization.Fetcher
	fetcher = organization.NewLocalFetcher(repository)

	// act
	out, err := fetcher.GetByID(context.Background(), "1")

	// assert
	s.Assert().NoError(err)
	s.Assert().Equal("1", out.ID())
	s.Assert().Equal("foo", out.Name())
}

type localListerSuite struct {
	suite.Suite
}

func TestLocalListerSuite(t *testing.T) {
	suite.Run(t, new(localListerSuite))
}

func (s *localListerSuite) TestLocalLister_List_Valid() {
	// arrange
	ctrl := gomock.NewController(s.T())
	repository := organizationmock.NewMockReadRepository(ctrl)
	repository.EXPECT().
		FindAll(gomock.Any(), gomock.Any()).
		Times(1).
		Return(&paging.Page[organization.Organization]{
			TotalItems: 1,
			Items: []organization.Organization{
				organization.New(context.Background(), "1", "foo"),
			},
		}, error(nil))

	var lister organization.Lister
	lister = organization.NewLocalLister(repository)

	// act
	out, err := lister.List(context.Background(),
		organization.WithListPageOptions(
			paging.WithLimit(10),
		),
		organization.WithListNonDeletedOnly(),
	)

	// assert
	s.Assert().NoError(err)
	s.Assert().Len(out.Items, 1)
	s.Assert().Equal("1", out.Items[0].ID())
	s.Assert().Equal("foo", out.Items[0].Name())
}
