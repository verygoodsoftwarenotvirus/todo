package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/pquerna/otp/totp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	errEmptyAddressUnallowed = errors.New("empty address not allowed")
)

// CreateServiceUser creates a user.
func CreateServiceUser(ctx context.Context, address, username string) (*types.User, error) {
	if username == "" {
		username = gofakeit.Password(true, true, true, false, false, 32)
	}

	if address == "" {
		return nil, errEmptyAddressUnallowed
	}

	parsedAddress, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c, err := httpclient.NewClient(parsedAddress)
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	in := &types.UserRegistrationInput{
		Username: username,
		Password: gofakeit.Password(true, true, true, true, true, 64),
	}

	ucr, err := c.CreateUser(ctx, in)
	if err != nil {
		return nil, err
	}

	token, tokenErr := totp.GenerateCode(ucr.TwoFactorSecret, time.Now().UTC())
	if tokenErr != nil {
		return nil, fmt.Errorf("generating totp code: %w", tokenErr)
	}

	if validationErr := c.VerifyTOTPSecret(ctx, ucr.CreatedUserID, token); validationErr != nil {
		return nil, fmt.Errorf("verifying totp code: %w", validationErr)
	}

	u := &types.User{
		ID:       ucr.CreatedUserID,
		Username: ucr.Username,
		// this is a dirty trick to reuse most of this model,
		HashedPassword:  in.Password,
		TwoFactorSecret: ucr.TwoFactorSecret,
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

	return nil, http.ErrNoCookie
}

// DetermineServiceURL returns the url, if properly configured.
func DetermineServiceURL() *url.URL {
	ta := os.Getenv("TARGET_ADDRESS")
	if ta == "" {
		panic("must provide target address!")
	}

	u, err := url.Parse(ta)
	if err != nil {
		panic(err)
	}

	return u
}

// EnsureServerIsUp checks that a server is up and doesn't return until it's certain one way or the other.
func EnsureServerIsUp(ctx context.Context, address string) {
	var (
		isDown           = true
		interval         = time.Second
		maxAttempts      = 50
		numberOfAttempts = 0
	)

	for isDown {
		if !IsUp(ctx, address) {
			log.Printf("waiting %s before pinging %q again", interval, address)
			time.Sleep(interval)

			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatal("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}
}

// IsUp can check if an instance of our server is alive.
func IsUp(ctx context.Context, address string) bool {
	uri := fmt.Sprintf("%s/_meta_/ready", address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	if err = res.Body.Close(); err != nil {
		log.Println("error closing body")
	}

	return res.StatusCode == http.StatusOK
}
