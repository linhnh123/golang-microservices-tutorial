package dbclient

import (
	"github.com/linhnh123/golang-microservices-tutorial/accountservice/model"
	"github.com/stretchr/testify/mock"
)

type MockBoltClient struct {
	mock.Mock
}

func (m *MockBoltClient) QueryAccount(accountId string) (model.Account, error) {
	args := m.Mock.Called(accountId)
	return args.Get(0).(model.Account), args.Error(1)
}

func (m *MockBoltClient) OpenBoltDb() {

}

func (m *MockBoltClient) CloseBoltDb() {

}

func (m *MockBoltClient) Seed() {

}
