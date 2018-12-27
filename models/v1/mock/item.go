package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.ItemHandler = (*MockItemHandler)(nil)

type MockItemHandler struct {
	mock.Mock
}

func (m *MockItemHandler) GetItem(itemID, userID uint64) (*models.Item, error) {
	args := m.Called(itemID, userID)
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemHandler) GetItemCount(filter *models.QueryFilter) (uint64, error) {
	args := m.Called(filter)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockItemHandler) GetItems(filter *models.QueryFilter) (*models.ItemList, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.ItemList), args.Error(1)
}

func (m *MockItemHandler) CreateItem(input *models.ItemInput) (*models.Item, error) {
	args := m.Called(input)
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemHandler) UpdateItem(updated *models.Item) error {
	return m.Called(updated).Error(0)
}

func (m *MockItemHandler) DeleteItem(id uint64) error {
	return m.Called(id).Error(0)
}
