package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gopkg.in/oauth2.v3"
)

type clientStore struct {
	dataManager models.OAuth2ClientDataManager
}

func newClientStore(db models.OAuth2ClientDataManager) oauth2.ClientStore {
	cs := &clientStore{
		dataManager: db,
	}
	return cs
}

// GetByID implements oauth2.ClientStorage
func (s *clientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	client, err := s.dataManager.GetOAuth2ClientByClientID(context.Background(), id)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid client")
	} else if err != nil {
		return nil, fmt.Errorf("querying for client: %w", err)
	}

	return client, nil
}
