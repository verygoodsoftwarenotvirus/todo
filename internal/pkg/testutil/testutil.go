package testutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

const (
	base64ImagePrefix = `data:image/jpeg;base64,`
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

// BuildMaxAdminPerms builds a helpful AdminPermissionChecker.
func BuildMaxAdminPerms() permissions.AdminPermissionChecker {
	return bitmask.NewPermissionBitmask(math.MaxUint32)
}

// BuildNoAdminPerms builds a helpful AdminPermissionChecker.
func BuildNoAdminPerms() permissions.AdminPermissionChecker {
	return bitmask.NewPermissionBitmask(0)
}

// DetermineServiceURL returns the url, if properly configured.
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

// CreateServiceUser creates a user.
func CreateServiceUser(ctx context.Context, address, username string, debug bool) (*types.User, error) {
	if username == "" {
		username = fake.Password(true, true, true, false, false, 32)
	}

	tu := httpclient.MustParseURL(address)
	c := httpclient.NewClient(
		httpclient.WithURL(tu),
	)

	in := &types.UserCreationInput{
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

func getLoginCookie(ctx context.Context, serviceURL string, u *types.User) (*http.Cookie, error) {
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

// CreateObligatoryOAuth2Client creates the OAuth2 httpclient we need for tests.
func CreateObligatoryOAuth2Client(ctx context.Context, serviceURL string, u *types.User) (*types.OAuth2Client, error) {
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
		"scopes": ["*"]
	}
		`, u.Username, u.HashedPassword, code)),
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

	var o types.OAuth2Client
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

// CreateBodyFromStruct takes any value in and returns an io.ReadCloser for an http.Request's body.
func CreateBodyFromStruct(t *testing.T, in interface{}) io.ReadCloser {
	t.Helper()

	out, err := json.Marshal(in)
	require.NoError(t, err)

	return ioutil.NopCloser(bytes.NewReader(out))
}

// BuildArbitraryImage builds an image with a bunch of colors in it.
func BuildArbitraryImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % math.MaxUint8), G: uint8(y % math.MaxUint8), B: uint8(x + y%math.MaxUint8), A: math.MaxUint8})
		}
	}

	return img
}
