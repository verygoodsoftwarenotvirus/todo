package httpclient

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

	"gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

const (
	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

func buildURL(parts ...string) string {
	tu, _ := url.Parse(urlToUse)
	u, _ := url.Parse(strings.Join(parts, "/"))
	return tu.ResolveReference(u).String()
}

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	logger := zerolog.ProvideLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	ensureServerIsUp()

	u, err := createObligatoryUser()
	if err != nil {
		logger.Fatal(err)
	}

	fake.Seed(time.Now().UnixNano())

	clientID, clientSecret, err := createObligatoryClient(*u)
	if err != nil {
		logger.Fatal(err)
	}

	todoClient = initializeClient(clientID, clientSecret)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

// mostly duplicated code from the http_client

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !isUp() {
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

func isUp() bool {
	uri := fmt.Sprintf("%s/_meta_/health", urlToUse)

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	httpc := buildHTTPClient()

	res, err := httpc.Do(req)
	if err != nil {
		return false
	}

	return res.StatusCode == http.StatusOK
}

func createObligatoryUser() (*models.User, error) {
	tu, _ := url.Parse(urlToUse)

	c, err := client.NewSimpleClient(tu, debug)
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
