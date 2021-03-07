package bleve

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	exampleType struct {
		Name          string `json:"name"`
		ID            uint64 `json:"id"`
		BelongsToUser uint64 `json:"belongsToUser"`
	}

	exampleTypeWithStringID struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		BelongsToUser uint64 `json:"belongsToUser"`
	}
)

func TestNewBleveIndexManager(T *testing.T) {
	T.Parallel()

	temp := os.TempDir()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "constructor_test_happy_path.bleve"))

		_, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})

	T.Run("invalid path", func(t *testing.T) {
		t.Parallel()
		exampleIndexPath := search.IndexPath("")

		_, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.Error(t, err)
	})

	T.Run("invalid name", func(t *testing.T) {
		t.Parallel()
		exampleIndexPath := search.IndexPath("constructor_test_invalid_name.bleve")

		_, err := NewBleveIndexManager(exampleIndexPath, "invalid", logging.NewNonOperationalLogger())
		assert.Error(t, err)
	})
}

func TestBleveIndexManager_Index(T *testing.T) {
	T.Parallel()

	temp := os.TempDir()
	exampleUserID := fakes.BuildFakeUser().ID

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		const exampleQuery = "_test"
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "_test_obligatory.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		x := &exampleType{
			ID:            123,
			Name:          exampleQuery,
			BelongsToUser: exampleUserID,
		}
		assert.NoError(t, im.Index(ctx, x.ID, x))

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})
}

func TestBleveIndexManager_Search(T *testing.T) {
	T.Parallel()

	temp := os.TempDir()
	exampleUserID := fakes.BuildFakeUser().ID

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		const exampleQuery = "search_test"
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "search_test_obligatory.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		x := exampleType{
			ID:            123,
			Name:          exampleQuery,
			BelongsToUser: exampleUserID,
		}
		assert.NoError(t, im.Index(ctx, x.ID, &x))

		results, err := im.Search(ctx, x.Name, exampleUserID)
		assert.NotEmpty(t, results)
		assert.NoError(t, err)

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})

	T.Run("with empty index and search", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleIndexPath := search.IndexPath(filepath.Join(temp, "search_test_empty_index.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		results, err := im.Search(ctx, "", exampleUserID)
		assert.Empty(t, results)
		assert.NoError(t, err)

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})

	T.Run("with closed index", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		const exampleQuery = "search_test"
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "search_test_closed_index.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		x := &exampleType{
			ID:            123,
			Name:          exampleQuery,
			BelongsToUser: exampleUserID,
		}
		assert.NoError(t, im.Index(ctx, x.ID, x))

		assert.NoError(t, im.(*bleveIndexManager).index.Close())

		results, err := im.Search(ctx, x.Name, exampleUserID)
		assert.Empty(t, results)
		assert.Error(t, err)

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		const exampleQuery = "search_test"
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "search_test_invalid_id.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		x := &exampleTypeWithStringID{
			ID:            "whatever",
			Name:          exampleQuery,
			BelongsToUser: exampleUserID,
		}
		assert.NoError(t, im.(*bleveIndexManager).index.Index(x.ID, x))

		results, err := im.Search(ctx, x.Name, exampleUserID)
		assert.Empty(t, results)
		assert.Error(t, err)

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})
}

func TestBleveIndexManager_Delete(T *testing.T) {
	T.Parallel()

	temp := os.TempDir()
	exampleUserID := fakes.BuildFakeUser().ID

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		const exampleQuery = "delete_test"
		exampleIndexPath := search.IndexPath(filepath.Join(temp, "delete_test.bleve"))

		im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
		assert.NoError(t, err)
		require.NotNil(t, im)

		x := &exampleType{
			ID:            123,
			Name:          exampleQuery,
			BelongsToUser: exampleUserID,
		}
		assert.NoError(t, im.Index(ctx, x.ID, x))

		assert.NoError(t, im.Delete(ctx, x.ID))

		assert.NoError(t, os.RemoveAll(string(exampleIndexPath)))
	})
}
