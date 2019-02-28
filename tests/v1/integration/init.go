package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	initializeTracer()
	logger := zerolog.ProvideLogger(zerolog.ProvideZerologger())

	logger.WithValue("url", urlToUse).Info("checking server")
	ensureServerIsUp()

	u, err := createObligatoryUser()
	if err != nil {
		logger.Fatal(err)
	}

	clientID, clientSecret, err := createObligatoryClient(*u)
	if err != nil {
		logger.Fatal(err)
	}

	uri, err := url.Parse(urlToUse)
	if err != nil {
		logger.Fatal(err)
	}
	initializeClient(uri, clientID, clientSecret)

	spaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", spaces, spaces)
}

// mostly duplicated code from the client

func buildURL(parts ...string) string {
	tu, _ := url.Parse(urlToUse)
	u, _ := url.Parse(strings.Join(parts, "/"))
	return tu.ResolveReference(u).String()
}

func createObligatoryUser() (*models.User, error) {
	tu, _ := url.Parse(urlToUse)

	c, err := client.NewSimpleClient(tu, debug)
	if err != nil {
		return nil, err
	}

	in := &models.UserInput{
		Username: fake.UserName(),
		Password: fake.Password(64, 128, true, true, true),
	}

	ucr, err := c.CreateNewUser(context.Background(), in)

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

	return u, err
}

func getLoginCookie(u models.User) (*http.Cookie, error) {
	uri := buildURL("users", "login")

	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now())
	if err != nil {
		return nil, err
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
		return nil, err
	}

	res, err := buildHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}

	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, errors.New("no cookie found :(")
}

func createObligatoryClient(u models.User) (clientID, clientSecret string, err error) {
	firstOAuth2ClientURI := buildURL("oauth2", "client")

	code, err := totp.GenerateCode(
		strings.ToUpper(u.TwoFactorSecret),
		time.Now(),
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

	cookie, err := getLoginCookie(u)
	if err != nil || cookie == nil {
		log.Fatalf(`
cookie problems!
	cookie == nil: %v
			  err: %v
	`, cookie == nil, err)
	}
	req.AddCookie(cookie)

	if command, err := http2curl.GetCurlCommand(req); err == nil {
		log.Println(command.String())
	}

	var o models.OAuth2Client

	res, err := buildHTTPClient().Do(req)
	if err != nil {
		return "", "", err
	} else if res.StatusCode >= http.StatusCreated {
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
