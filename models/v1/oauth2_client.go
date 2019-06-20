package models

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/oauth2.v3"
)

const (
	// OAuth2ClientKey is a ContextKey for use with contexts involving OAuth2 clients
	OAuth2ClientKey ContextKey = "oauth2_client"
)

type (
	// OAuth2ClientDataManager handles OAuth2 clients
	OAuth2ClientDataManager interface {
		GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*OAuth2Client, error)
		GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*OAuth2Client, error)
		GetAllOAuth2ClientCount(ctx context.Context) (uint64, error)
		GetOAuth2ClientCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
		GetOAuth2Clients(ctx context.Context, filter *QueryFilter, userID uint64) (*OAuth2ClientList, error)
		GetAllOAuth2Clients(ctx context.Context) ([]*OAuth2Client, error)
		GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*OAuth2Client, error)
		CreateOAuth2Client(ctx context.Context, input *OAuth2ClientCreationInput) (*OAuth2Client, error)
		UpdateOAuth2Client(ctx context.Context, updated *OAuth2Client) error
		ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error
	}

	// OAuth2ClientDataServer describes a structure capable of serving traffic related to oauth2 clients
	OAuth2ClientDataServer interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		// There is deliberately no update function
		ArchiveHandler(res http.ResponseWriter, req *http.Request)

		CreationInputMiddleware(next http.Handler) http.Handler
		OAuth2ClientInfoMiddleware(next http.Handler) http.Handler
		ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*OAuth2Client, error)

		// wrappers for our implementation library
		HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error
		HandleTokenRequest(res http.ResponseWriter, req *http.Request) error
	}

	// OAuth2Client represents a user-authorized API client
	OAuth2Client struct {
		ID              uint64   `json:"id"`
		Name            string   `json:"name"`
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

	// OAuth2ClientList is a response struct containing a list of OAuth2Clients
	OAuth2ClientList struct {
		Pagination
		Clients []OAuth2Client `json:"clients"`
	}

	// OAuth2ClientCreationInput is a struct for use when creating OAuth2 clients.
	OAuth2ClientCreationInput struct {
		UserLoginInput
		Name         string   `json:"name"`
		ClientID     string   `json:"-"`
		ClientSecret string   `json:"-"`
		RedirectURI  string   `json:"redirect_uri"`
		BelongsTo    uint64   `json:"belongs_to"`
		Scopes       []string `json:"scopes"`
	}

	// OAuth2ClientUpdateInput is a struct for use when updating OAuth2 clients
	OAuth2ClientUpdateInput struct {
		RedirectURI string   `json:"redirect_uri"`
		Scopes      []string `json:"scopes"`
	}
)

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

// HasScope returns whether or not the provided scope is included in the scope list
func (c *OAuth2Client) HasScope(scope string) (found bool) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return false
	}
	if c != nil && c.Scopes != nil {
		for _, s := range c.Scopes {
			if strings.TrimSpace(strings.ToLower(s)) == strings.TrimSpace(strings.ToLower(scope)) ||
				strings.TrimSpace(s) == "*" {
				return true
			}
		}
	}
	return false
}
