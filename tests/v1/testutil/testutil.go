package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	http2 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

// EnsureServerIsUp checks that a server is up and doesn't return until it's certain one way or the other
func EnsureServerIsUp(address string) {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !IsUp(address) {
			log.Print("waiting before pinging again")
			time.Sleep(500 * time.Millisecond)
			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatal("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}
}

// IsUp can check if an instance of our server is alive
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

	return res.StatusCode == http.StatusOK
}

// CreateObligatoryUser creates a user for the sake of having an OAuth2 client
func CreateObligatoryUser(address string, debug bool) (*models.User, error) {
	tu, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c, err := http2.NewSimpleClient(tu, debug)
	if err != nil {
		return nil, err
	}

	// I had difficulty ensuring these values were unique, even when fake.Seed was called. Could've been fake's fault,
	// could've been docker's fault. In either case, it wasn't worth the time to investigate and determine the culprit.
	username := fake.UserName() + fake.HexColor() + fake.Country()
	in := &models.UserInput{
		Username: username,
		Password: fake.Password(64, 128, true, true, true),
	}

	ucr, err := c.CreateNewUser(context.Background(), in)
	if err != nil {
		return nil, err
	} else if ucr == nil {
		return nil, errors.New("something happened")
	}

	u := &models.User{
		ID:                    ucr.ID,
		Username:              ucr.Username,
		HashedPassword:        in.Password, // this is a dirty trick to reuse most of this model
		TwoFactorSecret:       ucr.TwoFactorSecret,
		IsAdmin:               ucr.IsAdmin,
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

func getLoginCookie(serviceURL string, u models.User) (*http.Cookie, error) {
	uri := buildURL(serviceURL, "users", "login")

	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return nil, errors.Wrap(err, "generating totp token")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		uri,
		strings.NewReader(
			fmt.Sprintf(`
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
		return nil, errors.Wrap(err, "building request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, errors.New("no cookie found :(")
}

// CreateObligatoryClient creates the OAuth2 client we need for tests
func CreateObligatoryClient(serviceURL string, u models.User) (clientID, clientSecret string, err error) {
	firstOAuth2ClientURI := buildURL(serviceURL, "oauth2", "client")

	code, err := totp.GenerateCode(
		strings.ToUpper(u.TwoFactorSecret),
		time.Now().UTC(),
	)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		firstOAuth2ClientURI,
		strings.NewReader(fmt.Sprintf(`
	{
		"username": %q,
		"password": %q,
		"totp_token": %q,

		"belongs_to": %d,
		"scopes": ["*"]
	}
		`, u.Username, u.HashedPassword, code, u.ID),
		),
	)
	if err != nil {
		return "", "", err
	}

	cookie, err := getLoginCookie(serviceURL, u)
	if err != nil || cookie == nil {
		log.Fatalf(`
cookie problems!
	cookie == nil: %v
			  err: %v
	`, cookie == nil, err)
	}
	req.AddCookie(cookie)

	var command fmt.Stringer
	if command, err = http2curl.GetCurlCommand(req); err == nil {
		log.Println(command.String())
	}

	var o models.OAuth2Client

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	} else if res.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("bad status: %d", res.StatusCode)
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

	err = json.NewDecoder(res.Body).Decode(&o)

	return o.ClientID, o.ClientSecret, nil
}
