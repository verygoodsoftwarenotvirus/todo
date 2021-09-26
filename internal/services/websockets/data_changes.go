package websockets

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

func (s *service) handleDataChange(ctx context.Context, payload []byte) error {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	s.logger.Debug("data change message received")

	var msg *types.DataChangeMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		observability.AcknowledgeError(err, s.logger, span, "unmarshalling data change message")
		return err
	}

	s.logger.WithValue("msg", msg).Debug("handling data change")

	s.connectionsHat.RLock()
	defer s.connectionsHat.RUnlock()
	for userID, connections := range s.connections {
		if msg.AttributableToUserID != userID {
			continue
		}
		logger := s.logger.WithValue(keys.UserIDKey, userID).WithValue("connection_count", len(connections))
		for i, conn := range connections {
			logger = logger.WithValue("connection_index", i)

			if err := conn.SetWriteDeadline(time.Now().Add(s.websocketDeadline)); err != nil {
				observability.AcknowledgeError(err, logger, span, "setting write deadline")
			}

			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				logger.Error(err, "writing message to websocket")
				s.logger.Error(err, "writing message to websocket")
			}
		}
	}

	return nil
}

func removeConnection(s []*websocket.Conn, index int) []*websocket.Conn {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

func (s *service) pingConnections() {
	const pollDuration = 10 * time.Second
	ticker := time.NewTicker(pollDuration)

	s.logger.Debug("pinging websocket connections")

	for range ticker.C {
		s.connectionsHat.Lock()
		for userID, connections := range s.connections {
			l := s.logger.WithValue(keys.UserIDKey, userID).WithValue("connection_count", len(connections))
			for i, conn := range connections {
				l = l.WithValue("connection_index", i)
				if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(pollDuration/2)); err != nil {
					l.Error(err, "pinging websocket connection")
					s.connections[userID] = removeConnection(s.connections[userID], i)
				}
			}
		}
		s.connectionsHat.Unlock()
	}
}
