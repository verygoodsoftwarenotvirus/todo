package grpcclient

import (
	"context"
	"net/url"

	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func tokenEndpoint(baseURL *url.URL) oauth2.Endpoint {
	tu, au := *baseURL, *baseURL
	tu.Path, au.Path = "oauth2/token", "oauth2/authorize"

	return oauth2.Endpoint{
		TokenURL: tu.String(),
		AuthURL:  au.String(),
	}
}

func buildOAuthClient(uri *url.URL, clientID, clientSecret string) (credentials.PerRPCCredentials, error) {
	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"*"}, // SECUREME
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: tokenEndpoint(uri).TokenURL,
	}

	token, err := conf.TokenSource(context.Background()).Token()
	if err != nil {
		return nil, err
	}

	return oauth.NewOauthAccess(token), nil
}

// NewAuthorizedClient returns a gRPC client with OAuth2 enabled
func NewAuthorizedClient(address, clientID, clientSecret string) (todoproto.TodoClient, error) {
	uri, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	creds, err := buildOAuthClient(uri, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	grpcConn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithPerRPCCredentials(creds))
	if err != nil || grpcConn == nil {
		return nil, err
	}

	tc := todoproto.NewTodoClient(grpcConn)

	return tc, nil
}
