package server

import (
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2manage "gopkg.in/oauth2.v3/manage"
	oauth2models "gopkg.in/oauth2.v3/models"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

func (s *Server) initializeOauth2Routes() {
	manager := oauth2manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(oauth2store.NewMemoryTokenStore())

	// client memory store
	clientStore := oauth2store.NewClientStore()
	clientStore.Set("000000", &oauth2models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "https://localhost",
	})
	manager.MapClientStorage(clientStore)

	s.oauth2Handler = oauth2server.NewDefaultServer(manager)
	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)

	s.oauth2Handler.SetInternalErrorHandler(func(err error) (re *oauth2errors.Response) {
		s.logger.Errorf("Internal Error: %v", err.Error())
		return
	})

	s.oauth2Handler.SetResponseErrorHandler(func(re *oauth2errors.Response) {
		s.logger.Errorf("Response Error: %v", re.Error)
	})

}
