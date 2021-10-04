package bleve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

type (
	exampleType struct {
		Name          string
		ID            string
		BelongsToUser string
	}
)

var (
	_ suite.AfterTest  = (*indexManagerTestSuite)(nil)
	_ suite.BeforeTest = (*indexManagerTestSuite)(nil)
)

type indexManagerTestSuite struct {
	suite.Suite

	ctx              context.Context
	indexPath        string
	exampleAccountID string
}

func createTmpIndexPath(t *testing.T) string {
	t.Helper()

	tmpIndexPath, err := os.MkdirTemp("", fmt.Sprintf("search-testidx-%d", time.Now().Unix()))
	require.NoError(t, err)

	return tmpIndexPath
}

func (s *indexManagerTestSuite) BeforeTest(_, _ string) {
	t := s.T()

	s.indexPath = createTmpIndexPath(t)

	err := os.MkdirAll(s.indexPath, 0700)
	require.NoError(t, err)

	s.ctx = context.Background()
	s.exampleAccountID = fakes.BuildFakeAccount().ID
}

func (s *indexManagerTestSuite) AfterTest(_, _ string) {
	s.Require().NoError(os.RemoveAll(s.indexPath))
}

func TestIndexManager(T *testing.T) {
	T.Parallel()

	suite.Run(T, new(indexManagerTestSuite))
}

func (s *indexManagerTestSuite) TestNewIndexManagerWithTestIndex() {
	t := s.T()

	ctx := context.Background()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "constructor_test_happy_path_test.search"))

	_, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
}

func (s *indexManagerTestSuite) TestNewIndexManagerWithItemsIndex() {
	t := s.T()

	ctx := context.Background()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "constructor_test_happy_path_items.search"))

	_, err := NewIndexManager(ctx, exampleIndexPath, "items", logging.NewNoopLogger())
	assert.NoError(t, err)
}

func (s *indexManagerTestSuite) TestNewIndexManagerWithInvalidName() {
	t := s.T()

	ctx := context.Background()

	exampleIndexPath := search.IndexPath("constructor_test_invalid_name.search")

	_, err := NewIndexManager(ctx, exampleIndexPath, "invalid", logging.NewNoopLogger())
	assert.Error(t, err)
}

func (s *indexManagerTestSuite) TestIndex() {
	t := s.T()

	ctx := context.Background()

	const exampleQuery = "_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "_test_obligatory.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            "123",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}

	assert.NoError(t, im.Index(s.ctx, x.ID, x))
}

func (s *indexManagerTestSuite) TestSearch() {
	t := s.T()

	ctx := context.Background()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_obligatory.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := exampleType{
		ID:            "123",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, &x))

	results, err := im.Search(s.ctx, x.Name, s.exampleAccountID)
	assert.NotEmpty(t, results)
	assert.NoError(t, err)
}

func (s *indexManagerTestSuite) TestSearchWithInvalidQuery() {
	t := s.T()

	ctx := context.Background()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_invalid_query.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	results, err := im.Search(s.ctx, "", s.exampleAccountID)
	assert.Empty(t, results)
	assert.Error(t, err)
}

func (s *indexManagerTestSuite) TestSearchWithEmptyIndexAndSearch() {
	t := s.T()

	ctx := context.Background()

	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_empty_index.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	results, err := im.Search(s.ctx, "example", s.exampleAccountID)
	assert.Empty(t, results)
	assert.NoError(t, err)
}

func (s *indexManagerTestSuite) TestSearchWithClosedIndex() {
	t := s.T()

	ctx := context.Background()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_closed_index.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            "123",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, x))

	assert.NoError(t, im.(*indexManager).index.Close())

	results, err := im.Search(s.ctx, x.Name, s.exampleAccountID)
	assert.Empty(t, results)
	assert.Error(t, err)
}

func (s *indexManagerTestSuite) TestSearchForAdmin() {
	t := s.T()

	ctx := context.Background()

	const exampleQuery = "search_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "search_test_obligatory.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := exampleType{
		ID:            "123",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}
	assert.NoError(t, im.Index(s.ctx, x.ID, &x))

	results, err := im.SearchForAdmin(s.ctx, x.Name)
	assert.NotEmpty(t, results)
	assert.NoError(t, err)
}

func (s *indexManagerTestSuite) TestDelete() {
	t := s.T()

	ctx := context.Background()

	const exampleQuery = "delete_test"
	exampleIndexPath := search.IndexPath(filepath.Join(s.indexPath, "delete_test.search"))

	im, err := NewIndexManager(ctx, exampleIndexPath, testingSearchIndexName, logging.NewNoopLogger())
	assert.NoError(t, err)
	require.NotNil(t, im)

	x := &exampleType{
		ID:            "123",
		Name:          exampleQuery,
		BelongsToUser: s.exampleAccountID,
	}

	assert.NoError(t, im.Index(s.ctx, x.ID, x))
	assert.NoError(t, im.Delete(s.ctx, x.ID))
}
