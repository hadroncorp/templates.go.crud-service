package notificationfx

import (
	"go.uber.org/fx"

	"github.com/hadroncorp/service-template/notification"
)

var Module = fx.Module("hadron/iam/notification",
	fx.Provide(
		fx.Annotate(
			notification.NewNoopSender,
			fx.As(new(notification.Sender)),
		),
	),
)
