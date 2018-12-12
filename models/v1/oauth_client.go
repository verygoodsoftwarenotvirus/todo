package models

type Oauth2ClientHandler interface {
	GetOauth2Client(identifier string) (*Oauth2Client, error)
	GetOauth2ClientCount() (uint64, error)
	GetOauth2Clients(filter *QueryFilter) (*Oauth2ClientList, error)
	CreateOauth2Client(input *Oauth2ClientInput) (*Oauth2Client, error)
	UpdateOauth2Client(updated *Oauth2Client) error
	DeleteOauth2Client(id uint) error
}

type Oauth2Client struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Domain       string   `json:"domain"`
	Scopes       []string `json:"scopes"`
	CreatedOn    uint64   `json:"created_on"`
	UpdatedOn    *uint64  `json:"updated_on"`
	ArchivedOn   *uint64  `json:"archived_on"`
}

type Oauth2ClientList struct {
	Pagination
	Clients []Oauth2Client `json:"clients"`
}

type Oauth2ClientInput struct {
	UserLoginInput
	Domain string   `json:"domain"`
	Scopes []string `json:"scopes"`
}
