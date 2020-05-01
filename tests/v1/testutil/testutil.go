package testutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/moul/http2curl"
	"github.com/pquerna/otp/totp"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

// DetermineServiceURL returns the URL, if properly configured.
func DetermineServiceURL() string {
	ta := os.Getenv("TARGET_ADDRESS")
	if ta == "" {
		panic("must provide target address!")
	}

	u, err := url.Parse(ta)
	if err != nil {
		panic(err)
	}

	return u.String()
}

// EnsureServerIsUp checks that a server is up and doesn't return until it's certain one way or the other.
func EnsureServerIsUp(address string) {
	var (
		isDown           = true
		interval         = time.Second
		maxAttempts      = 50
		numberOfAttempts = 0
	)

	for isDown {
		if !IsUp(address) {
			log.Print("waiting before pinging again")
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
func IsUp(address string) bool {
	uri := fmt.Sprintf("%s/_meta_/ready", address)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
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

// CreateObligatoryUser creates a user for the sake of having an OAuth2 client.
func CreateObligatoryUser(address string, debug bool) (*models.User, error) {
	tu, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c, err := client.NewSimpleClient(context.Background(), tu, debug)
	if err != nil {
		return nil, err
	}

	// I had difficulty ensuring these values were unique, even when fake.Seed was called. Could've been fake's fault,
	// could've been docker's fault. In either case, it wasn't worth the time to investigate and determine the culprit
	username := fake.Username() + fake.HexColor() + fake.Country()
	in := &models.UserCreationInput{
		Username: username,
		Password: fake.Password(true, true, true, true, true, 64),
	}

	ucr, err := c.CreateUser(context.Background(), in)
	if err != nil {
		return nil, err
	} else if ucr == nil {
		return nil, errors.New("something happened")
	}

	u := &models.User{
		ID:       ucr.ID,
		Username: ucr.Username,
		// this is a dirty trick to reuse most of this model,
		HashedPassword:        in.Password,
		TwoFactorSecret:       ucr.TwoFactorSecret,
		PasswordLastChangedOn: ucr.PasswordLastChangedOn,
		CreatedOn:             ucr.CreatedOn,
		UpdatedOn:             ucr.UpdatedOn,
		ArchivedOn:            ucr.ArchivedOn,
	}

	return u, nil
}

func buildURL(address string, parts ...string) string {
	tu, err := url.Parse(address)
	if err != nil {
		panic(err)
	}

	u, err := url.Parse(strings.Join(parts, "/"))
	if err != nil {
		panic(err)
	}

	return tu.ResolveReference(u).String()
}

func getLoginCookie(serviceURL string, u *models.User) (*http.Cookie, error) {
	uri := buildURL(serviceURL, "users", "login")
	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("generating totp token: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		uri,
		strings.NewReader(
			fmt.Sprintf(
				`
					{
						"username": %q,
						"password": %q,
						"totp_token": %q
					}
				`,
				u.Username,
				u.HashedPassword,
				code,
			),
		),
	)
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

// CreateObligatoryClient creates the OAuth2 client we need for tests.
func CreateObligatoryClient(serviceURL string, u *models.User) (*models.OAuth2Client, error) {
	firstOAuth2ClientURI := buildURL(serviceURL, "oauth2", "client")

	code, err := totp.GenerateCode(
		strings.ToUpper(u.TwoFactorSecret),
		time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		firstOAuth2ClientURI,
		strings.NewReader(fmt.Sprintf(`
	{
		"username": %q,
		"password": %q,
		"totp_token": %q,

		"belongs_to_user": %d,
		"scopes": ["*"]
	}
		`, u.Username, u.HashedPassword, code, u.ID)),
	)
	if err != nil {
		return nil, err
	}

	cookie, err := getLoginCookie(serviceURL, u)
	if err != nil || cookie == nil {
		log.Fatalf("\ncookie problems!\n\tcookie == nil: %v\n\t\t\t  err: %v\n\t", cookie == nil, err)
	}
	req.AddCookie(cookie)
	var o models.OAuth2Client

	var command fmt.Stringer
	if command, err = http2curl.GetCurlCommand(req); err == nil {
		log.Println(command.String())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("bad status: %d", res.StatusCode)
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	bdump, err := httputil.DumpResponse(res, true)
	if err == nil && req.Method != http.MethodGet {
		log.Println(string(bdump))
	}

	return &o, json.NewDecoder(res.Body).Decode(&o)
}
