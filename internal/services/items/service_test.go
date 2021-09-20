package items

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService() *service {
	return &service{
		logger:          logging.NewNoopLogger(),
		itemDataManager: &mocktypes.ItemDataManager{},
		itemIDFetcher:   func(req *http.Request) string { return "" },
		encoderDecoder:  mockencoding.NewMockEncoderDecoder(),
		search:          &mocksearch.IndexManager{},
		tracer:          tracing.NewTracer("test"),
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamStringIDFetcher",
			ItemIDURIParamKey,
		).Return(func(*http.Request) string { return "" })

		cfg := Config{
			SearchIndexPath:      "example/path",
			PreWritesTopicName:   "pre-writes",
			PreUpdatesTopicName:  "pre-updates",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreUpdatesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreArchivesTopicName).Return(&publishers.MockProducer{}, nil)

		s, err := ProvideService(
			logging.NewNoopLogger(),
			&cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			rpm,
			pp,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, rpm, pp)
	})

	T.Run("with error providing pre-writes producer", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SearchIndexPath:      "example/path",
			PreWritesTopicName:   "pre-writes",
			PreUpdatesTopicName:  "pre-updates",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return((*publishers.MockProducer)(nil), errors.New("blah"))

		s, err := ProvideService(
			logging.NewNoopLogger(),
			&cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			nil,
			pp,
		)

		assert.Nil(t, s)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, pp)
	})

	T.Run("with error providing pre-updates producer", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SearchIndexPath:      "example/path",
			PreWritesTopicName:   "pre-writes",
			PreUpdatesTopicName:  "pre-updates",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreUpdatesTopicName).Return((*publishers.MockProducer)(nil), errors.New("blah"))

		s, err := ProvideService(
			logging.NewNoopLogger(),
			&cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			nil,
			pp,
		)

		assert.Nil(t, s)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, pp)
	})

	T.Run("with error providing pre-archives producer", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SearchIndexPath:      "example/path",
			PreWritesTopicName:   "pre-writes",
			PreUpdatesTopicName:  "pre-updates",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreUpdatesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreArchivesTopicName).Return((*publishers.MockProducer)(nil), errors.New("blah"))

		s, err := ProvideService(
			logging.NewNoopLogger(),
			&cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			nil,
			pp,
		)

		assert.Nil(t, s)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, pp)
	})

	T.Run("with error providing index", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SearchIndexPath:      "example/path",
			PreWritesTopicName:   "pre-writes",
			PreUpdatesTopicName:  "pre-updates",
			PreArchivesTopicName: "pre-archives",
		}

		s, err := ProvideService(
			logging.NewNoopLogger(),
			&cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return nil, errors.New("blah")
			},
			mockrouting.NewRouteParamManager(),
			nil,
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
