package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

type mockWebsocketConnection struct {
	mock.Mock
}

func (m *mockWebsocketConnection) SetWriteDeadline(t time.Time) error {
	return m.Called(t).Error(0)
}

func (m *mockWebsocketConnection) WriteMessage(messageType int, data []byte) error {
	return m.Called(messageType, data).Error(0)
}

func (m *mockWebsocketConnection) WriteControl(messageType int, data []byte, deadline time.Time) error {
	return m.Called(messageType, data, deadline).Error(0)
}

func Test_handleDataChange(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestHelper(t)

		msg := &types.DataChangeMessage{AttributableToUserID: s.exampleUser.ID}
		examplePayload, err := json.Marshal(msg)
		require.NoError(t, err)

		mc := &mockWebsocketConnection{}
		mc.On(
			"SetWriteDeadline",
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(nil)

		mc.On(
			"WriteMessage",
			websocket.TextMessage,
			examplePayload,
		).Return(nil)

		s.service.connections = map[string][]websocketConnection{
			s.exampleUser.ID: {mc},
		}

		err = s.service.handleDataChange(ctx, examplePayload)
		require.NoError(t, err)
	})

	T.Run("with invalid JSON", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestHelper(t)

		err := s.service.handleDataChange(ctx, []byte(`} not real JSON lol`))

		require.Error(t, err)
	})

	T.Run("with irrelevant message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestHelper(t)

		msg := &types.DataChangeMessage{}
		examplePayload, err := json.Marshal(msg)
		require.NoError(t, err)

		mc := &mockWebsocketConnection{}
		mc.On(
			"SetWriteDeadline",
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(nil)

		mc.On(
			"WriteMessage",
			websocket.TextMessage,
			examplePayload,
		).Return(nil)

		s.service.connections = map[string][]websocketConnection{
			s.exampleUser.ID: {mc},
		}

		err = s.service.handleDataChange(ctx, examplePayload)
		require.NoError(t, err)
	})

	T.Run("with error setting write deadline", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestHelper(t)

		msg := &types.DataChangeMessage{AttributableToUserID: s.exampleUser.ID}
		examplePayload, err := json.Marshal(msg)
		require.NoError(t, err)

		mc := &mockWebsocketConnection{}
		mc.On(
			"SetWriteDeadline",
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(errors.New("blah"))

		s.service.connections = map[string][]websocketConnection{
			s.exampleUser.ID: {mc},
		}

		err = s.service.handleDataChange(ctx, examplePayload)
		require.NoError(t, err)
	})

	T.Run("with error writing message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestHelper(t)

		msg := &types.DataChangeMessage{AttributableToUserID: s.exampleUser.ID}
		examplePayload, err := json.Marshal(msg)
		require.NoError(t, err)

		mc := &mockWebsocketConnection{}
		mc.On(
			"SetWriteDeadline",
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(nil)

		mc.On(
			"WriteMessage",
			websocket.TextMessage,
			examplePayload,
		).Return(errors.New("blah"))

		s.service.connections = map[string][]websocketConnection{
			s.exampleUser.ID: {mc},
		}

		err = s.service.handleDataChange(ctx, examplePayload)
		require.NoError(t, err)
	})
}

func Test_removeConnection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		toCreate := 5
		connections := []websocketConnection{}

		for i := 0; i < toCreate; i++ {
			connections = append(connections, &websocket.Conn{})
		}

		connections = removeConnection(connections, 1)

		require.Len(t, connections, toCreate-1)
	})
}

func Test_pingConnections(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.pollDuration = time.Second / 10

		mc := &mockWebsocketConnection{}
		mc.On(
			"WriteControl",
			websocket.PingMessage,
			[]byte("ping"),
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(nil)

		s.service.connections = map[string][]websocketConnection{
			s.exampleAccount.ID: {mc},
		}

		go s.service.pingConnections()

		<-time.After(time.Second)
	})

	T.Run("with error pinging", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.pollDuration = time.Second / 10

		mc := &mockWebsocketConnection{}
		mc.On(
			"WriteControl",
			websocket.PingMessage,
			[]byte("ping"),
			mock.MatchedBy(func(x time.Time) bool { return true }),
		).Return(errors.New("blah"))

		s.service.connections = map[string][]websocketConnection{
			s.exampleAccount.ID: {mc},
		}

		go s.service.pingConnections()

		<-time.After(time.Second)

		s.service.connectionsHat.RLock()
		assert.Empty(t, s.service.connections)
		s.service.connectionsHat.RUnlock()
	})
}
