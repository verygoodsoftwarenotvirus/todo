package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gopkg.in/oauth2.v3"
)

type clientStore struct {
	dataManager types.OAuth2ClientDataManager
}

func newClientStore(db types.OAuth2ClientDataManager) oauth2.ClientStore {
	cs := &clientStore{
		dataManager: db,
	}
	return cs
}

var errInvalidClient = errors.New("invalid client")

// GetByID implements oauth2.ClientStorage.
func (s *clientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	client, err := s.dataManager.GetOAuth2ClientByClientID(context.Background(), id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errInvalidClient
	} else if err != nil {
		return nil, fmt.Errorf("querying for client: %w", err)
	}

	return client, nil
}
