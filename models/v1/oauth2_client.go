package models

import (
	"strconv"

	"gopkg.in/oauth2.v3"
)

const (
	// OAuth2ClientKey is a ContextKey for use with contexts involving OAuth2 clients
	OAuth2ClientKey ContextKey = "oauth2_client"
)

// OAuth2ClientHandler handles OAuth2 clients
type OAuth2ClientHandler interface {
	GetOAuth2Client(identifier string) (*OAuth2Client, error)
	GetOAuth2ClientCount(filter *QueryFilter) (uint64, error)
	GetOAuth2Clients(filter *QueryFilter) (*Oauth2ClientList, error)
	CreateOAuth2Client(input *Oauth2ClientCreationInput) (*OAuth2Client, error)
	UpdateOAuth2Client(updated *OAuth2Client) error
	DeleteOAuth2Client(identifier string) error
}

// OAuth2Client represents a user-authorized API client
type OAuth2Client struct {
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

var _ oauth2.ClientInfo = (*OAuth2Client)(nil)

// GetID returns the client ID. NOTE: I believe this is implemented for the above interface spec (oauth2.ClientInfo)
func (c *OAuth2Client) GetID() string {
	return c.ClientID
}

// GetSecret returns the ClientSecret
func (c *OAuth2Client) GetSecret() string {
	return c.ClientSecret
}

// GetDomain returns the client's domain
func (c *OAuth2Client) GetDomain() string {
	return c.RedirectURI
}

// GetUserID returns the client's UserID
func (c *OAuth2Client) GetUserID() string {
	return strconv.FormatUint(c.BelongsTo, 10)
}

// Oauth2ClientList is a response struct containing a list of OAuth2Clients
type Oauth2ClientList struct {
	Pagination
	Clients []OAuth2Client `json:"clients"`
}

// Oauth2ClientCreationInput is a struct for use when creating OAuth2 clients.
type Oauth2ClientCreationInput struct {
	UserLoginInput
	RedirectURI string   `json:"redirect_uri"`
	BelongsTo   uint64   `json:"belongs_to"`
	Scopes      []string `json:"scopes"`
}

// Oauth2ClientUpdateInput is a struct for use when updating OAuth2 clients
type Oauth2ClientUpdateInput struct {
	RedirectURI string   `json:"redirect_uri"`
	Scopes      []string `json:"scopes"`
}
