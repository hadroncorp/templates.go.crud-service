package main

import (
	"github.com/hadroncorp/enclave"
	enclavekafka "github.com/hadroncorp/enclave/kafka"

	"github.com/hadroncorp/service-template/notificationfx"
	"github.com/hadroncorp/service-template/organizationfx"
)

func main() {
	enclave.RunApplication(
		enclave.WithSQL(),
		enclave.WithObservabilitySQL(),
		enclave.WithPostgres(),
		enclavekafka.WithKafkaEvents(),
		enclave.WithServerHTTP(),
		enclave.WithFxOptions(
			organizationfx.Module,
			notificationfx.Module,
		),
	)
}
