//go:build !integration

package organization_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type controllerKafkaSuite struct {
	suite.Suite
}

func TestControllerKafkaIntegrationTest(t *testing.T) {
	suite.Run(t, new(controllerKafkaSuite))
}

func (s *controllerKafkaSuite) SetupSuite() {

}

func (s *controllerKafkaSuite) TearDownSuite() {

}

func (s *controllerKafkaSuite) TestController_() {}
