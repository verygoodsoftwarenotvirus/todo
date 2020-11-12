package testutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
)

const (
	base64ImagePrefix = `data:image/jpeg;base64,`
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

	svcAddr := u.String()

	log.Printf("using target address: %q\n", svcAddr)
	return svcAddr
}

// DetermineDatabaseURL returns the DB connection URL, if properly configured.
func DetermineDatabaseURL() (address, vendor string) {
	dbv := os.Getenv("DB_VENDOR")
	if dbv == "" {
		panic("must provide DB vendor!")
	}

	dba := os.Getenv("DB_ADDRESS")
	if dba == "" && dbv != "sqlite" {
		panic("must provide target address!")
	}

	u, err := url.Parse(dba)
	if err != nil {
		panic(err)
	}

	return u.String(), dbv
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
			log.Printf("waiting %s before pinging again", interval)
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

// CreateObligatoryUser creates a user for the sake of having an OAuth2 client.
func CreateObligatoryUser(address string, debug bool) (*models.User, error) {
	ctx := context.Background()

	tu, parseErr := url.Parse(address)
	if parseErr != nil {
		return nil, parseErr
	}

	c, clientInitErr := client.NewSimpleClient(ctx, tu, debug)
	if clientInitErr != nil {
		return nil, clientInitErr
	}

	username := fake.Password(true, true, true, false, false, 32)
	in := &models.UserCreationInput{
		Username: username,
		Password: fake.Password(true, true, true, true, true, 64),
	}

	ucr, userCreationErr := c.CreateUser(ctx, in)
	if userCreationErr != nil {
		return nil, userCreationErr
	} else if ucr == nil {
		return nil, errors.New("something happened")
	}

	twoFactorSecret, err := ParseTwoFactorSecretFromBase64EncodedQRCode(ucr.TwoFactorQRCode)
	if err != nil {
		return nil, err
	}

	token, tokenErr := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
	if tokenErr != nil {
		return nil, fmt.Errorf("generating totp code: %w", tokenErr)
	}

	if validationErr := c.VerifyTOTPSecret(ctx, ucr.ID, token); validationErr != nil {
		return nil, fmt.Errorf("verifying totp code: %w", validationErr)
	}

	u := &models.User{
		ID:       ucr.ID,
		Username: ucr.Username,
		// this is a dirty trick to reuse most of this model,
		HashedPassword:        in.Password,
		TwoFactorSecret:       twoFactorSecret,
		PasswordLastChangedOn: ucr.PasswordLastChangedOn,
		CreatedOn:             ucr.CreatedOn,
		LastUpdatedOn:         ucr.LastUpdatedOn,
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

func getLoginCookie(ctx context.Context, serviceURL string, u *models.User) (*http.Cookie, error) {
	uri := buildURL(serviceURL, "users", "login")

	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("generating totp token: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		uri,
		strings.NewReader(
			fmt.Sprintf(
				`
					{
						"username": %q,
						"password": %q,
						"totpToken": %q
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
func CreateObligatoryClient(ctx context.Context, serviceURL string, u *models.User) (*models.OAuth2Client, error) {
	if u == nil {
		return nil, errors.New("user is nil")
	}

	firstOAuth2ClientURI := buildURL(serviceURL, "oauth2", "client")

	code, err := totp.GenerateCode(
		strings.ToUpper(u.TwoFactorSecret),
		time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		firstOAuth2ClientURI,
		strings.NewReader(fmt.Sprintf(`
	{
		"username": %q,
		"password": %q,
		"totpToken": %q,
		"belongsToUser": %d,
		"scopes": ["*"]
	}
		`, u.Username, u.HashedPassword, code, u.ID)),
		// remember we use u.HashedPassword as a temp container for the plain password
	)
	if err != nil {
		return nil, err
	}

	cookie, err := getLoginCookie(ctx, serviceURL, u)
	if err != nil || cookie == nil {
		log.Fatalf("\ncookie problems!\n\tcookie == nil: %v\n\terr: %v\n\t", cookie == nil, err)
	}

	req.AddCookie(cookie)

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

	var o models.OAuth2Client
	err = json.NewDecoder(res.Body).Decode(&o)
	return &o, err
}

// ParseTwoFactorSecretFromBase64EncodedQRCode accepts a base64-encoded QR code representing an otpauth:// URI,
// parses the QR code and extracts the 2FA secret from the URI. It can also return an error.
func ParseTwoFactorSecretFromBase64EncodedQRCode(qrCode string) (string, error) {
	qrCode = strings.TrimPrefix(qrCode, base64ImagePrefix)

	unbased, err := base64.StdEncoding.DecodeString(qrCode)
	if err != nil {
		return "", fmt.Errorf("Cannot decode b64: %w", err)
	}

	im, err := png.Decode(bytes.NewReader(unbased))
	if err != nil {
		return "", fmt.Errorf("Bad png: %w", err)
	}

	bb, err := gozxing.NewBinaryBitmapFromImage(im)
	if err != nil {
		return "", fmt.Errorf("Bad binary bitmap: %w", err)
	}

	res, err := qrcode.NewQRCodeReader().DecodeWithoutHints(bb)
	if err != nil {
		return "", fmt.Errorf("error decoding: %w", err)
	}

	totpDetails := res.String()

	u, err := url.Parse(totpDetails)
	if err != nil {
		return "", fmt.Errorf("error parsing URI: %w", err)
	}

	return u.Query().Get("secret"), nil
}
