package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/pquerna/otp/totp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
	httpclient "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http"
)

// CreateServiceUser creates a user.
func CreateServiceUser(ctx context.Context, address, username string) (*types.User, error) {
	if username == "" {
		username = gofakeit.Password(true, true, true, false, false, 32)
	}

	if address == "" {
		return nil, errors.New("empty address not allowed")
	}

	parsedAddress, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c, err := httpclient.NewClient(parsedAddress)
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	in := &types.NewUserCreationInput{
		Username: username,
		Password: gofakeit.Password(true, true, true, true, true, 64),
	}

	ucr, userCreationErr := c.CreateUser(ctx, in)
	if userCreationErr != nil {
		return nil, userCreationErr
	} else if ucr == nil {
		return nil, errors.New("something strange happened")
	}

	twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(ucr.TwoFactorQRCode)
	if err != nil {
		return nil, fmt.Errorf("parsing TOTP QR code: %w", err)
	}

	token, tokenErr := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
	if tokenErr != nil {
		return nil, fmt.Errorf("generating totp code: %w", tokenErr)
	}

	if validationErr := c.VerifyTOTPSecret(ctx, ucr.ID, token); validationErr != nil {
		return nil, fmt.Errorf("verifying totp code: %w", validationErr)
	}

	u := &types.User{
		ID:       ucr.ID,
		Username: ucr.Username,
		// this is a dirty trick to reuse most of this model,
		HashedPassword:  in.Password,
		TwoFactorSecret: twoFactorSecret,
		CreatedOn:       ucr.CreatedOn,
	}

	return u, nil
}

// GetLoginCookie fetches a login cookie for a given user.
func GetLoginCookie(ctx context.Context, serviceURL string, u *types.User) (*http.Cookie, error) {
	tu, err := url.Parse(serviceURL)
	if err != nil {
		panic(err)
	}

	lu, err := url.Parse(strings.Join([]string{"users", "login"}, "/"))
	if err != nil {
		panic(err)
	}

	uri := tu.ResolveReference(lu).String()

	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("generating totp token: %w", err)
	}

	body, err := json.Marshal(&types.UserLoginInput{
		Username:  u.Username,
		Password:  u.HashedPassword,
		TOTPToken: code,
	})
	if err != nil {
		return nil, fmt.Errorf("generating login request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if err = res.Body.Close(); err != nil {
		log.Println("error closing body")
	}

	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, errors.New("no cookie found :(")
}
