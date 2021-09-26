package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// APIClient represents a user-authorized API client.
	APIClient struct {
		_ struct{}

		LastUpdatedOn *uint64 `json:"lastUpdatedOn"`
		ArchivedOn    *uint64 `json:"archivedOn"`
		Name          string  `json:"name"`
		ClientID      string  `json:"clientID"`
		ID            string  `json:"id"`
		BelongsToUser string  `json:"belongsToUser"`
		ClientSecret  []byte  `json:"-"`
		CreatedOn     uint64  `json:"createdOn"`
	}

	// APIClientList is a response struct containing a list of API clients.
	APIClientList struct {
		_ struct{}

		Clients []*APIClient `json:"clients"`
		Pagination
	}

	// APIClientCreationInput is a struct for use when creating API clients.
	APIClientCreationInput struct {
		_ struct{}

		UserLoginInput
		ID            string `json:"-"`
		Name          string `json:"clientName"`
		ClientID      string `json:"-"`
		BelongsToUser string `json:"-"`
		ClientSecret  []byte `json:"-"`
	}

	// APIClientCreationResponse is a struct for informing users of what their API client's secret key is.
	APIClientCreationResponse struct {
		_ struct{}

		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}

	// APIClientDataManager handles API clients.
	APIClientDataManager interface {
		GetAPIClientByClientID(ctx context.Context, clientID string) (*APIClient, error)
		GetAPIClientByDatabaseID(ctx context.Context, clientID, ownerUserID string) (*APIClient, error)
		GetTotalAPIClientCount(ctx context.Context) (uint64, error)
		GetAPIClients(ctx context.Context, owneruserID string, filter *QueryFilter) (*APIClientList, error)
		CreateAPIClient(ctx context.Context, input *APIClientCreationInput) (*APIClient, error)
		ArchiveAPIClient(ctx context.Context, clientID, ownerUserID string) error
	}

	// APIClientDataService describes a structure capable of serving traffic related to API clients.
	APIClientDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// ValidateWithContext validates an APICreationInput.
func (x *APIClientCreationInput) ValidateWithContext(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	if err := x.UserLoginInput.ValidateWithContext(ctx, minUsernameLength, minPasswordLength); err != nil {
		return err
	}

	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
