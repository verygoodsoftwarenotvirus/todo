package models

import (
	"strconv"
)

const (
	Oauth2ClientKey ContextKey = "user"
)

type Oauth2ClientHandler interface {
	GetOauth2Client(identifier string) (*Oauth2Client, error)
	GetOauth2ClientCount(filter *QueryFilter) (uint64, error)
	GetOauth2Clients(filter *QueryFilter) (*Oauth2ClientList, error)
	CreateOauth2Client(input *Oauth2ClientInput) (*Oauth2Client, error)
	UpdateOauth2Client(updated *Oauth2Client) error
	DeleteOauth2Client(id uint) error
}

type Oauth2Client struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURI  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
	CreatedOn    uint64   `json:"created_on"`
	UpdatedOn    *uint64  `json:"updated_on"`
	ArchivedOn   *uint64  `json:"archived_on"`
	BelongsTo    uint64   `json:"belongs_to"`
}

func (c *Oauth2Client) GetID() string {
	return c.ClientID
}

func (c *Oauth2Client) GetSecret() string {
	return c.ClientSecret
}

func (c *Oauth2Client) GetDomain() string {
	return c.RedirectURI
}

func (c *Oauth2Client) GetUserID() string {
	return strconv.FormatUint(c.BelongsTo, 10)
}

type Oauth2ClientList struct {
	Pagination
	Clients []Oauth2Client `json:"clients"`
}

type Oauth2ClientInput struct {
	UserLoginInput
	RedirectURI string   `json:"redirect_uri"`
	BelongsTo   string   `json:"belongs_to"`
	Scopes      []string `json:"scopes"`
}
