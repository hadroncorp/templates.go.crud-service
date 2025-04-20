package organizationfx

import (
	"github.com/hadroncorp/enclave/kafka/kafkafx"
	"github.com/hadroncorp/geck/transportfx/httpfx"
	"go.uber.org/fx"

	"github.com/hadroncorp/service-template/organization"
)

var Module = fx.Module("hadron/iam/organization",
	fx.Provide(
		fx.Annotate(
			organization.NewPostgresRepository,
			fx.As(new(organization.Repository)),
		),
		fx.Annotate(
			organization.NewPostgresReadRepository,
			fx.As(new(organization.ReadRepository)),
		),
		fx.Annotate(
			organization.NewLocalManager,
			fx.As(new(organization.Manager)),
		),
		fx.Annotate(
			organization.NewLocalFetcher,
			fx.As(new(organization.Fetcher)),
		),
		fx.Annotate(
			organization.NewLocalLister,
			fx.As(new(organization.Lister)),
		),
		httpfx.AsController(organization.NewControllerHTTP),
		kafkafx.AsController(organization.NewControllerKafka),
	),
)
