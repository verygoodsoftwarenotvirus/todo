package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// BuildAPIClientAuthTokenRequest builds a request.
func (c *Client) BuildAPIClientAuthTokenRequest(ctx context.Context, input *types.PASETOCreationInput, secretKey []byte) (*http.Request, error) {
	return c.requestBuilder.BuildAPIClientAuthTokenRequest(ctx, input, secretKey)
}

func (c *Client) fetchAuthTokenForAPIClient(ctx context.Context, httpclient *http.Client, clientID string, secretKey []byte) (string, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if httpclient == nil {
		httpclient = http.DefaultClient
	}

	if httpclient.Timeout == 0 {
		httpclient.Timeout = defaultTimeout
	}

	if secretKey == nil {
		return "", ErrNilInputProvided
	}

	input := &types.PASETOCreationInput{
		ClientID:    clientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return "", fmt.Errorf("validating input: %w", validationErr)
	}

	if c.accountID != 0 {
		input.AccountID = c.accountID
	}

	req, err := c.requestBuilder.BuildAPIClientAuthTokenRequest(ctx, input, secretKey)
	if err != nil {
		return "", err
	}

	// use the default client here because we want a transport that doesn't worry about cookies or tokens.
	res, err := c.executeRawRequest(ctx, httpclient, req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}

	if resErr := errorFromResponse(res); resErr != nil {
		return "", resErr
	}

	var tokenRes types.PASETOResponse

	if unmarshalErr := c.unmarshalBody(ctx, res, &tokenRes); unmarshalErr != nil {
		return "", unmarshalErr
	}

	return tokenRes.Token, nil
}
