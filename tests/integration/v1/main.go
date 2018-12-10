package integration

import (
	"crypto/tls"
	"fmt"
	"github.com/pquerna/otp/totp"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"

	"github.com/bxcodec/faker"
)

const (
	debug                  = false
	nonexistentID          = "999999999"
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://demo-server"

	defaultTestInstanceClientID     = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientSecret = "YLBAILERTSETANNAWIESUACEBPUEDAMEVIHCIHWTERCESASIEREH"
	defaultTestInstanceAuthToken    = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

var (
	todoClient *client.V1Client
	fake       = faker.GetLorem()
)

func sp(s string) *string { return &s }

func fetchCookieForOauthTesting(username, password, totpSecret string) *http.Cookie {
	loginURL := fmt.Sprintf("%s://%s/users/login", todoClient.URL.Scheme, todoClient.URL.Hostname())

	code, err := totp.GenerateCode(strings.ToUpper(totpSecret), time.Now())

	body := strings.NewReader(fmt.Sprintf(`
		{
			"username": %q,
			"password": %q,
			"totp_token": %q
		}
	`, username, password, code))
	req, _ := http.NewRequest(http.MethodPost, loginURL, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	cookies := resp.Cookies()
	if len(cookies) == 1 {
		return resp.Cookies()[0]
	}
	return nil
}

const (
	exampleUsername   = "username"
	examplePassword   = "password"
	exampleTOTPSecret = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func testOAuth() {
	var tempServer *http.Server
	go func() {
		router := http.NewServeMux()
		router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
			x, _ := url.ParseQuery(req.RequestURI)
			log.Printf("req:\n\t %+v", x)
		})
		tempServer = &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  10 * time.Second,
			Handler:      router,
			Addr:         ":4321",
		}
		defer tempServer.Close()
		log.Fatal(tempServer.ListenAndServe())
	}()

	loggedInCookie := fetchCookieForOauthTesting(exampleUsername, examplePassword, exampleTOTPSecret)
	if loggedInCookie == nil {
		panic("wtf")
	}

	conf := &oauth2.Config{
		ClientID:     defaultTestInstanceClientID,
		ClientSecret: defaultTestInstanceClientSecret,
		RedirectURL:  tempServer.Addr,
		Scopes:       []string{"*"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/authorize", localTestInstanceURL),
			TokenURL: fmt.Sprintf("%s/token", localTestInstanceURL),
		},
	}
	//tok, err := conf.Exchange(
	//	oauth2.NoContext,
	//	defaultTestInstanceAuthToken,
	//	oauth2.SetAuthURLParam("client_id", conf.ClientID),
	//	oauth2.SetAuthURLParam("client_secret", conf.ClientSecret),
	//	oauth2.SetAuthURLParam("redirect_url", "YOUR_REDIRECT_URL"),
	//)
	//
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	//oaClient := conf.Client(oauth2.NoContext, tok)

	oaClient := *http.DefaultClient

	oaClient.Transport = http.DefaultTransport
	oaClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	u, _ := url.Parse(fmt.Sprintf("%s/authorize", localTestInstanceURL))
	values := u.Query()
	values.Set("client_id", conf.ClientID)
	values.Set("client_secret", conf.ClientSecret)
	values.Set("response_type", "code")
	values.Set("scope", "*")
	values.Set("redirect_uri", fmt.Sprintf("http://localhost%s", tempServer.Addr))
	u.RawQuery = values.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	req.AddCookie(loggedInCookie)
	res, err := oaClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Printf(`


		token check response status was %d


		`, res.StatusCode)
	}
}

func initializeClient() {
	cfg := &client.Config{
		Client: &http.Client{
			Transport: http.DefaultTransport,
			//Timeout:   5 * time.Second,
		},
		Debug:     debug,
		Address:   localTestInstanceURL,
		AuthToken: sp(defaultTestInstanceAuthToken),
	}

	cfg.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	todoClient = c
}

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !todoClient.IsUp() {
			log.Printf("waiting half a second before pinging again")
			time.Sleep(500 * time.Millisecond)
			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatalf("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}

}

func init() {
	initializeClient()
	ensureServerIsUp()
	testOAuth()
}
