package frontend

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

func Test_buildItemsTableDashboardPage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItemList := &types.ItemList{
			Items: []*types.Item{
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 1",
					Details:          "Stuff 1",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 2",
					Details:          "Stuff 2",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 3",
					Details:          "Stuff 3",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
			},
			Pagination: types.Pagination{},
		}

		actual, err := buildItemsTableDashboardPage(exampleItemList)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func Test_buildViewerForItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItem := &types.Item{
			ArchivedOn:       nil,
			LastUpdatedOn:    nil,
			ExternalID:       "blah-blah-blah",
			Name:             "Things",
			Details:          "Stuff",
			CreatedOn:        uint64(time.Now().Unix()),
			ID:               12345,
			BelongsToAccount: 54321,
		}

		expected := `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
        <h1 class="h2">Item #12345</h1>
    </div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="Name">Name</label>
                <div class="input-group">
                    <input class="form-control" type="text" id="Name" placeholder="Name"required="" value="Things" />
                    <div class="invalid-feedback" style="width: 100%;">Name is required.</div>
                </div>
            </div>
            <div class="mb3">
                <label for="Details">Details</label>
                <div class="input-group">
                    <input class="form-control" type="text" id="Details" placeholder="Details" value="Stuff" />
                    
                </div>
            </div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

		actual, err := buildViewerForItem(exampleItem)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildItemsTableDashboardView(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItemList := &types.ItemList{
			Items: []*types.Item{
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 1",
					Details:          "Stuff 1",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 2",
					Details:          "Stuff 2",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
				{
					ArchivedOn:       nil,
					LastUpdatedOn:    nil,
					ExternalID:       "blah-blah-blah",
					Name:             "Things 3",
					Details:          "Stuff 3",
					CreatedOn:        uint64(time.Now().Unix()),
					ID:               12345,
					BelongsToAccount: 54321,
				},
			},
			Pagination: types.Pagination{},
		}

		actual, err := buildItemsTableDashboardView(exampleItemList)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}
