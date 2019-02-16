package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

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

	ensureServerIsUp(urlToUse)

	u, err := createObligatoryUser()
	if err != nil {
		logger.Fatal(err)
	}

	clientID, clientSecret, err := createObligatoryClient(*u)
	if err != nil {
		logger.Fatal(err)
	}

	initializeClient(clientID, clientSecret)

	fmt.Println("Running tests")
}

// mostly duplicated code from the client

func buildURL(parts ...string) string {
	tu, _ := url.Parse(urlToUse)
	u, _ := url.Parse(strings.Join(parts, "/"))
	return tu.ResolveReference(u).String()
}

func createObligatoryUser() (*models.User, error) {
	uri := buildURL("/users")

	username, password := fake.UserName(), fake.Password(64, 128, true, true, true)

	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(fmt.Sprintf(`
	{
		"username": %q,
		"password": %q
	}
	`, username, password)))

	var r models.UserCreationResponse
	httpc := buildHTTPClient()
	res, err := httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&r)

	u := &models.User{
		Username:        username,
		HashedPassword:  password, // we're misusing this field, but it's ok, we can just keep it a secret between friends
		TwoFactorSecret: r.TwoFactorSecret,
	}

	return u, err
}

func createObligatoryClient(u models.User) (clientID, clientSecret string, err error) {
	firstOAuth2ClientURI := buildURL("oauth2", "init_client")

	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now())
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, firstOAuth2ClientURI, strings.NewReader(fmt.Sprintf(`
	{

		"username": %q,
		"password": %q,
		"totp_token": %q,

		"belongs_to": %d,
		"scopes": ["*"]
	}
	`, u.Username, u.HashedPassword, code, u.ID)))
	if err != nil {
		return "", "", err
	}

	if command, err := http2curl.GetCurlCommand(req); err == nil {
		log.Println(command.String())
	}

	var o models.OAuth2Client

	httpc := buildHTTPClient()
	res, err := httpc.Do(req)
	if err != nil {
		return "", "", err
	} else if res.StatusCode >= http.StatusCreated {
		return "", "", errors.New("bad status")
	}
	defer res.Body.Close()

	if bdump, err := httputil.DumpResponse(res, true); err == nil && req.Method != http.MethodGet {
		log.Println(string(bdump))
	}

	err = json.NewDecoder(res.Body).Decode(&o)

	return o.ClientID, o.ClientSecret, nil
}
