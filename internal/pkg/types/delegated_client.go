package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
)

const (
	// DelegatedClientKey is a ContextKey for use with contexts involving delegated clients.
	DelegatedClientKey ContextKey = "delegated_client"
)

type (
	// DelegatedClient represents a User-authorized API client.
	DelegatedClient struct {
		ID                      uint64                          `json:"id"`
		ExternalID              string                          `json:"externalID"`
		Name                    string                          `json:"name"`
		ClientID                string                          `json:"clientID"`
		ClientSecret            string                          `json:"clientSecret"`
		AccountUserPermissions  bitmask.ServiceUserPermissions  `json:"siteUserPermissions"`
		ServiceAdminPermissions bitmask.ServiceAdminPermissions `json:"ServiceAdminPermissions"`
		CreatedOn               uint64                          `json:"createdOn"`
		LastUpdatedOn           *uint64                         `json:"lastUpdatedOn"`
		ArchivedOn              *uint64                         `json:"archivedOn"`
		BelongsToUser           uint64                          `json:"belongsToUser"`
	}

	// DelegatedClientList is a response struct containing a list of DelegatedClients.
	DelegatedClientList struct {
		Pagination
		Clients []*DelegatedClient `json:"clients"`
	}

	// DelegatedClientCreationInput is a struct for use when creating delegated clients.
	DelegatedClientCreationInput struct {
		UserLoginInput
		Name          string `json:"name"`
		ClientID      string `json:"-"`
		ClientSecret  string `json:"-"`
		BelongsToUser uint64 `json:"-"`
	}

	// DelegatedClientSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	DelegatedClientSQLQueryBuilder interface {
		BuildGetBatchOfDelegatedClientsQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{})
		BuildGetAllDelegatedClientsCountQuery() string
		BuildGetDelegatedClientsQuery(userID uint64, filter *QueryFilter) (query string, args []interface{})
		BuildCreateDelegatedClientQuery(input *DelegatedClientCreationInput) (query string, args []interface{})
		BuildUpdateDelegatedClientQuery(input *DelegatedClient) (query string, args []interface{})
		BuildArchiveDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForDelegatedClientQuery(clientID uint64) (query string, args []interface{})
	}

	// DelegatedClientDataManager handles delegated clients.
	DelegatedClientDataManager interface {
		GetDelegatedClient(ctx context.Context, clientID, userID uint64) (*DelegatedClient, error)
		GetAllDelegatedClients(ctx context.Context, resultChannel chan []*DelegatedClient, bucketSize uint16) error
		GetTotalDelegatedClientCount(ctx context.Context) (uint64, error)
		GetDelegatedClients(ctx context.Context, userID uint64, filter *QueryFilter) (*DelegatedClientList, error)
		CreateDelegatedClient(ctx context.Context, input *DelegatedClientCreationInput) (*DelegatedClient, error)
		UpdateDelegatedClient(ctx context.Context, updated *DelegatedClient, changes []FieldChangeSummary) error
		ArchiveDelegatedClient(ctx context.Context, clientID, userID uint64) error
		GetAuditLogEntriesForDelegatedClient(ctx context.Context, clientID uint64) ([]*AuditLogEntry, error)
	}

	// DelegatedClientDataService describes a structure capable of serving traffic related to delegated clients.
	DelegatedClientDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		// There is deliberately no update function.
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)

		CreationInputMiddleware(next http.Handler) http.Handler
	}
)

// Validate validates a ItemCreationInput.
func (x *DelegatedClientCreationInput) Validate(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	if err := x.UserLoginInput.Validate(ctx, minUsernameLength, minPasswordLength); err != nil {
		return err
	}

	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
