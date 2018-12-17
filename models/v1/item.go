package models

type ItemHandler interface {
	GetItem(id uint64) (*Item, error)
	GetItemCount(filter *QueryFilter) (uint64, error)
	GetItems(filter *QueryFilter) (*ItemList, error)
	CreateItem(input *ItemInput) (*Item, error)
	UpdateItem(updated *Item) error
	DeleteItem(id uint64) error
}

type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Details     string  `json:"details"`
	CreatedOn   uint64  `json:"created_on"`
	UpdatedOn   *uint64 `json:"updated_on"`
	CompletedOn *uint64 `json:"completed_on"`
	BelongsTo   *uint64 `json:"belongs_to"`
}

func (i *Item) Update(input *ItemInput) {
	if input.Name != "" || input.Name != i.Name {
		i.Name = input.Name
	}

	if input.Details != "" || input.Details != i.Details {
		i.Details = input.Details
	}
}

type ItemList struct {
	Pagination
	Items []Item `json:"items"`
}

type ItemInput struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}
