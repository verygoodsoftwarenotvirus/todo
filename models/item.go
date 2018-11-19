package models

type ItemHandler interface {
	GetItem(id uint) (*Item, error)
	GetItems(filter *QueryFilter) ([]Item, error)
	CreateItem(input *ItemInput) (*Item, error)
	UpdateItem(updated *Item) error
	DeleteItem(id uint) error
}

type Item struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Details     string  `json:"details"`
	CreatedOn   uint64  `json:"created_on"`
	UpdatedOn   uint64  `json:"updated_on"`
	CompletedOn *uint64 `json:"completed_on"`
}

func (i *Item) Update(input *ItemInput) {
	if input.Name != "" || input.Name != i.Name {
		i.Name = input.Name
	}

	if input.Details != "" || input.Details != i.Details {
		i.Details = input.Details
	}
}

const ItemInputCtxKey ContextKey = "item_input"

type ItemInput struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}
