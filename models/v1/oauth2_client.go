package models

import (
	"context"
	"strconv"

	"gopkg.in/oauth2.v3"
)

const (
	// OAuth2ClientKey is a ContextKey for use with contexts involving OAuth2 clients
	OAuth2ClientKey ContextKey = "oauth2_client"
)

// OAuth2ClientHandler handles OAuth2 clients
type OAuth2ClientHandler interface {
	GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*OAuth2Client, error)
	GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*OAuth2Client, error)
	GetOAuth2ClientCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
	GetOAuth2Clients(ctx context.Context, filter *QueryFilter, userID uint64) (*OAuth2ClientList, error)
	GetAllOAuth2Clients(ctx context.Context) ([]OAuth2Client, error)
	CreateOAuth2Client(ctx context.Context, input *OAuth2ClientCreationInput) (*OAuth2Client, error)
	UpdateOAuth2Client(ctx context.Context, updated *OAuth2Client) error
	DeleteOAuth2Client(ctx context.Context, clientID, userID uint64) error
}

// OAuth2Client represents a user-authorized API client
type OAuth2Client struct {
	ID              uint64   `json:"id"`
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

// OAuth2ClientList is a response struct containing a list of OAuth2Clients
type OAuth2ClientList struct {
	Pagination
	Clients []OAuth2Client `json:"clients"`
}

// OAuth2ClientCreationInput is a struct for use when creating OAuth2 clients.
type OAuth2ClientCreationInput struct {
	UserLoginInput
	ClientID     string   `json:"-"`
	ClientSecret string   `json:"-"`
	RedirectURI  string   `json:"redirect_uri"`
	BelongsTo    uint64   `json:"belongs_to"`
	Scopes       []string `json:"scopes"`
}

// OAuth2ClientUpdateInput is a struct for use when updating OAuth2 clients
type OAuth2ClientUpdateInput struct {
	RedirectURI string   `json:"redirect_uri"`
	Scopes      []string `json:"scopes"`
}
