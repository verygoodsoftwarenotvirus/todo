package models

type OauthClientHandler interface {
	GetOauthClient(identifier string) (*OauthClient, error)
	GetOauthClientCount() (uint64, error)
	GetOauthClients(filter *QueryFilter) (*OauthClientList, error)
	CreateOauthClient(input *OauthClientInput) (*OauthClient, error)
	UpdateOauthClient(updated *OauthClient) error
	DeleteOauthClient(id uint) error
}

type OauthClient struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	CreatedOn    uint64   `json:"created_on"`
	UpdatedOn    *uint64  `json:"updated_on"`
	ArchivedOn   *uint64  `json:"archived_on"`
}

type OauthClientList struct {
	Pagination
	Clients []OauthClient `json:"clients"`
}

type OauthClientInput struct {
	Scopes []string `json:"scopes"`
}
