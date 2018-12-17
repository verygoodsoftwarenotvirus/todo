package main

import (
	"crypto/tls"
	"github.com/pquerna/otp/totp"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"

	"golang.org/x/oauth2"
)

const (
	defaultSecret = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func main() {

	cfg := &client.Config{
		Client: &http.Client{
			Transport: http.DefaultTransport,
			//Timeout:   5 * time.Second,
		},
		Debug:   false,
		Address: "https://localhost",
	}

	cfg.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	todoClient, err := client.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	code, err := totp.GenerateCode(strings.ToUpper(defaultSecret), time.Now())
	if err != nil {
		panic(err)
	}

	cookie, err := todoClient.Login("username", "password", code)
	if err != nil || cookie == nil {
		panic(err)
	}

	conf := &oauth2.Config{
		ClientID:     defaultSecret,
		ClientSecret: defaultSecret,
		Scopes:       []string{"*"},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://localhost/oauth2/token",
			AuthURL:  "https://localhost/oauth2/authorize",
		},
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{
		Timeout:   2 * time.Second,
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	aurl := conf.AuthCodeURL(
		"xyz",
		oauth2.SetAuthURLParam("client_id", defaultSecret),
		oauth2.SetAuthURLParam("client_secret", defaultSecret),
		oauth2.SetAuthURLParam("redirect_uri", "https://yourredirecturl.com"),
	)
	req, err := http.NewRequest(http.MethodPost, aurl, nil)
	if err != nil || req == nil {
		panic(err)
	}
	req.AddCookie(cookie)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("error trying to get authorized", err)
	}
	u, _ := url.Parse(res.Header.Get("Location"))
	actualCode := u.Query().Get("code")

	log.Println(actualCode)

	/*
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			u := config.AuthCodeURL("xyz")
			http.Redirect(w, r, u, http.StatusFound)
		})

		http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			state := r.Form.Get("state")
			if state != "xyz" {
				http.Error(w, "State invalid", http.StatusBadRequest)
				return
			}
			code := r.Form.Get("code")
			if code == "" {
				http.Error(w, "Code not found", http.StatusBadRequest)
				return
			}
			token, err := config.Exchange(context.Background(), code)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			e := json.NewEncoder(w)
			e.SetIndent("", "  ")
			e.Encode(*token)
		})

		log.Println("Client is running at 9094 port.")
		log.Fatal(http.ListenAndServe(":9094", nil))
	*/
}
