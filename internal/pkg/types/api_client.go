package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// APIClientKey is a ContextKey for use with contexts involving API clients.
	APIClientKey ContextKey = "api_client"
)

type (
	// APIClient represents a user-authorized API client.
	APIClient struct {
		LastUpdatedOn *uint64 `json:"lastUpdatedOn"`
		ArchivedOn    *uint64 `json:"archivedOn"`
		Name          string  `json:"name"`
		ClientID      string  `json:"clientID"`
		ExternalID    string  `json:"externalID"`
		ClientSecret  []byte  `json:"-"`
		CreatedOn     uint64  `json:"createdOn"`
		ID            uint64  `json:"id"`
		BelongsToUser uint64  `json:"belongsToUser"`
	}

	// APIClientList is a response struct containing a list of API clients.
	APIClientList struct {
		Clients []*APIClient `json:"clients"`
		Pagination
	}

	// APICientCreationInput is a struct for use when creating API clients.
	APICientCreationInput struct {
		UserLoginInput
		Name          string `json:"name"`
		ClientID      string `json:"-"`
		ClientSecret  []byte `json:"-"`
		BelongsToUser uint64 `json:"-"`
	}

	// APIClientCreationResponse is a struct for informing users of what their API client's secret key is.
	APIClientCreationResponse struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		ID           uint64 `json:"id"`
	}

	// APIClientSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	APIClientSQLQueryBuilder interface {
		BuildGetBatchOfAPIClientsQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetAPIClientByClientIDQuery(clientID string) (query string, args []interface{})
		BuildGetAPIClientByDatabaseIDQuery(clientID, userID uint64) (query string, args []interface{})
		BuildGetAllAPIClientsCountQuery() string
		BuildGetAPIClientsQuery(accountID uint64, filter *QueryFilter) (query string, args []interface{})
		BuildCreateAPIClientQuery(input *APICientCreationInput) (query string, args []interface{})
		BuildUpdateAPIClientQuery(input *APIClient) (query string, args []interface{})
		BuildArchiveAPIClientQuery(clientID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAPIClientQuery(userID uint64) (query string, args []interface{})
	}

	// APIClientDataManager handles API clients.
	APIClientDataManager interface {
		GetAPIClientByClientID(ctx context.Context, clientID string) (*APIClient, error)
		GetAPIClientByDatabaseID(ctx context.Context, clientID, ownerUserID uint64) (*APIClient, error)
		GetAllAPIClients(ctx context.Context, resultChannel chan []*APIClient, bucketSize uint16) error
		GetTotalAPIClientCount(ctx context.Context) (uint64, error)
		GetAPIClients(ctx context.Context, ownerUserID uint64, filter *QueryFilter) (*APIClientList, error)
		CreateAPIClient(ctx context.Context, input *APICientCreationInput, createdByUser uint64) (*APIClient, error)
		ArchiveAPIClient(ctx context.Context, clientID, ownerUserID, archivedByUser uint64) error
		GetAuditLogEntriesForAPIClient(ctx context.Context, clientID uint64) ([]*AuditLogEntry, error)
	}

	// APIClientDataService describes a structure capable of serving traffic related to API clients.
	APIClientDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)

		CreationInputMiddleware(next http.Handler) http.Handler
	}
)

// Validate validates a ItemCreationInput.
func (x *APICientCreationInput) Validate(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	if err := x.UserLoginInput.Validate(ctx, minUsernameLength, minPasswordLength); err != nil {
		return err
	}

	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
