package mock

import (
	"X_IM/pkg/wire/pkt"
	"github.com/stretchr/testify/mock"
)

type Container struct {
	mock.Mock
}

func (m *Container) Forward(serviceName string, packet *pkt.LogicPkt) error {
	args := m.Called(serviceName, packet)
	return args.Error(0)
}
