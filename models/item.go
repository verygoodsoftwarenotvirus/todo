package models

type ItemHandler interface {
	GetItem(id uint) (*Item, error)
	GetItems(filter *QueryFilter) ([]Item, error)
	CreateItem(input *ItemInput) (*Item, error)
	UpdateItem(updated *Item) error
	DeleteItem(id uint) error
}

type Item struct {
	ID          uint   `db:"id"`
	Name        string `db:"name"`
	Details     string `db:"details"`
	CreatedOn   uint64 `db:"created_on"`
	CompletedOn uint64 `db:"completed_on"`
}

func (i *Item) Update(input *ItemInput) {
	if input.Name != "" || input.Name != i.Name {
		i.Name = input.Name
	}

	if input.Details != "" || input.Details != i.Details {
		i.Details = input.Details
	}
}

const ItemInputCtxKey = "item_input"

type ItemInput struct {
	Name    string
	Details string
}
