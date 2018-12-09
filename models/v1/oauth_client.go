package models

type OauthClientHandler interface {
	GetOauthClient(identifier string) (*OauthClient, error)
	GetOauthClients(filter *QueryFilter) ([]OauthClient, error)
	CreateOauthClient(input *OauthClientInput) (*OauthClient, error)
	UpdateOauthClient(updated *OauthClient) error
	DeleteOauthClient(id uint) error
}

type OauthClient struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	CreatedOn    uint64   `json:"created_on"`
	UpdatedOn    *uint64  `json:"updated_on"`
	ArchivedOn   *uint64  `json:"archived_on"`
}

type OauthClientInput struct {
	Scopes []string `json:"scopes"`
}
