package bleve

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

var _ suite.AfterTest = (*bleveIndexManagerTestSuite)(nil)
var _ suite.BeforeTest = (*bleveIndexManagerTestSuite)(nil)

type bleveIndexManagerTestSuite struct {
	suite.Suite

	ctx              context.Context
	indexPath        string
	exampleAccountID uint64
}

func createTmpIndexPath(t *testing.T) string {
	t.Helper()

	tmpIndexPath, err := ioutil.TempDir("", fmt.Sprintf("bleve-testidx-%d", time.Now().Unix()))
	require.NoError(t, err)

	return tmpIndexPath
}

func (s *bleveIndexManagerTestSuite) BeforeTest(_, _ string) {
	t := s.T()

	s.indexPath = createTmpIndexPath(t)

	err := os.MkdirAll(s.indexPath, 0700)
	require.NoError(t, err)

	s.ctx = context.Background()
	s.exampleAccountID = fakes.BuildFakeAccount().ID
}

func (s *bleveIndexManagerTestSuite) AfterTest(_, _ string) {
	s.Require().NoError(os.RemoveAll(s.indexPath))
}

func TestNewBleveIndexManager(T *testing.T) {
	T.Parallel()

	suite.Run(T, new(bleveIndexManagerTestSuite))
}

func (s *bleveIndexManagerTestSuite) TestNewBleveIndexManagerWithTestIndex() {
	t := s.T()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "constructor_test_happy_path_test.bleve"))

	_, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
}

func (s *bleveIndexManagerTestSuite) TestNewBleveIndexManagerWithItemsIndex() {
	t := s.T()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "constructor_test_happy_path_items.bleve"))

	_, err := NewBleveIndexManager(exampleIndexPath, types.ItemsSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
}

func (s *bleveIndexManagerTestSuite) TestNewBleveIndexManagerWithInvalidName() {
	t := s.T()

	exampleIndexPath := search.IndexPath("constructor_test_invalid_name.bleve")

	_, err := NewBleveIndexManager(exampleIndexPath, "invalid", logging.NewNonOperationalLogger())
	assert.Error(t, err)
}

func (s *bleveIndexManagerTestSuite) TestIndex() {
	t := s.T()

	const exampleQuery = "_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "_test_obligatory.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            123,
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}

	assert.NoError(t, im.Index(s.ctx, x.ID, x))
}

func (s *bleveIndexManagerTestSuite) TestSearch() {
	t := s.T()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_obligatory.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := exampleType{
		ID:            123,
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, &x))

	results, err := im.Search(s.ctx, x.Name, s.exampleAccountID)
	assert.NotEmpty(t, results)
	assert.NoError(t, err)
}

func (s *bleveIndexManagerTestSuite) TestSearchWithInvalidQuery() {
	t := s.T()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_invalid_query.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	results, err := im.Search(s.ctx, "", s.exampleAccountID)
	assert.Empty(t, results)
	assert.Error(t, err)
}

func (s *bleveIndexManagerTestSuite) TestSearchWithEmptyIndexAndSearch() {
	t := s.T()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_empty_index.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	results, err := im.Search(s.ctx, "example", s.exampleAccountID)
	assert.Empty(t, results)
	assert.NoError(t, err)
}

func (s *bleveIndexManagerTestSuite) TestSearchWithClosedIndex() {
	t := s.T()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_closed_index.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            123,
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, x))

	assert.NoError(t, im.(*bleveIndexManager).index.Close())

	results, err := im.Search(s.ctx, x.Name, s.exampleAccountID)
	assert.Empty(t, results)
	assert.Error(t, err)
}

func (s *bleveIndexManagerTestSuite) TestSearchWithInvalidID() {
	t := s.T()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_invalid_id.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleTypeWithStringID{
		ID:            "whatever",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.(*bleveIndexManager).index.Index(x.ID, x))

	results, err := im.Search(s.ctx, x.Name, s.exampleAccountID)
	assert.Empty(t, results)
	assert.Error(t, err)
}

func (s *bleveIndexManagerTestSuite) TestSearchForAdmin() {
	t := s.T()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_obligatory.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := exampleType{
		ID:            123,
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, &x))

	results, err := im.SearchForAdmin(s.ctx, x.Name)
	assert.NotEmpty(t, results)
	assert.NoError(t, err)
}

func (s *bleveIndexManagerTestSuite) TestDelete() {
	t := s.T()

	const exampleQuery = "delete_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "delete_test.bleve"))

	im, err := NewBleveIndexManager(exampleIndexPath, testingSearchIndexName, logging.NewNonOperationalLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            123,
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}

	assert.NoError(t, im.Index(s.ctx, x.ID, x))
	assert.NoError(t, im.Delete(s.ctx, x.ID))
}
