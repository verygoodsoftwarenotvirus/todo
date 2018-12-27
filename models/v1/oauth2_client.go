package models

import (
	"strconv"

	"gopkg.in/oauth2.v3"
)

const (
	Oauth2ClientKey ContextKey = "oauth2_client"
)

type Oauth2ClientHandler interface {
	GetOauth2Client(identifier string) (*Oauth2Client, error)
	GetOauth2ClientCount(filter *QueryFilter) (uint64, error)
	GetOauth2Clients(filter *QueryFilter) (*Oauth2ClientList, error)
	CreateOauth2Client(input *Oauth2ClientCreationInput) (*Oauth2Client, error)
	UpdateOauth2Client(updated *Oauth2Client) error
	DeleteOauth2Client(identifier string) error
}

type Oauth2Client struct {
	ID              string   `json:"id"`
	ClientID        string   `json:"client_id"`
	ClientSecret    string   `json:"client_secret"`
	RedirectURI     string   `json:"redirect_uri"`
	Scopes          []string `json:"scopes"`
	ImplicitAllowed bool     `json:"implicit_allowed"`
	CreatedOn       uint64   `json:"created_on"`
	UpdatedOn       *uint64  `json:"updated_on"`
	ArchivedOn      *uint64  `json:"archived_on"`
	BelongsTo       uint64   `json:"belongs_to"`
}

var _ oauth2.ClientInfo = (*Oauth2Client)(nil)

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

type Oauth2ClientCreationInput struct {
	UserLoginInput
	RedirectURI string   `json:"redirect_uri"`
	BelongsTo   uint64   `json:"belongs_to"`
	Scopes      []string `json:"scopes"`
}

type Oauth2ClientUpdateInput struct {
	RedirectURI string   `json:"redirect_uri"`
	Scopes      []string `json:"scopes"`
}
