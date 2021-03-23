package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	httpclient "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http/requests"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/pquerna/otp/totp"
)

const (
	debug = false
)

type action struct {
	name   string
	weight uint
}

func randomIDFromMap(m map[uint64]struct{}) uint64 {
	keys := reflect.ValueOf(m).MapKeys()

	return keys[rand.Intn(len(keys))].Uint()
}

func selectAction(actions ...action) *action {
	var totalWeight uint = 0
	for _, rb := range actions {
		totalWeight += rb.weight
	}

	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(int(totalWeight))

	for _, rb := range actions {
		r -= int(rb.weight)
		if r <= 0 {
			return &rb
		}
	}

	return nil
}

func createAttacker(ctx context.Context, name string) (*vegeta.Attacker, *httpclient.Client, *requests.Builder, error) {
	c, b, err := createClientForTest(ctx, name)
	if err != nil {
		return nil, nil, nil, err
	}

	attacker := vegeta.NewAttacker(
		vegeta.Client(c.AuthenticatedClient()),
		vegeta.MaxConnections(100),
	)

	return attacker, c, b, nil
}

func initializeTargetFromRequest(req *http.Request, target *vegeta.Target) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	target.Method = req.Method
	target.URL = req.URL.String()
	target.Body = body
	target.Header = req.Header

	return nil
}

func createClientForTest(ctx context.Context, name string) (*httpclient.Client, *requests.Builder, error) {
	user, err := utils.CreateServiceUser(ctx, urlToUse, "")
	if err != nil {
		return nil, nil, err
	}

	cookie, err := utils.GetLoginCookie(ctx, urlToUse, user)
	if err != nil {
		return nil, nil, err
	}

	cookieClient, err := initializeCookiePoweredClient(cookie)
	if err != nil {
		return nil, nil, err
	}

	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}

	apiClient, err := cookieClient.CreateAPIClient(ctx, cookie, &types.APICientCreationInput{
		Name: name,
		UserLoginInput: types.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TOTPToken: token,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	secretKey, err := base64.RawURLEncoding.DecodeString(apiClient.ClientSecret)
	if err != nil {
		return nil, nil, err
	}

	pasetoClient, err := initializePASETOPoweredClient(apiClient.ClientID, secretKey)
	if err != nil {
		return nil, nil, err
	}

	logger := logging.NewNonOperationalLogger()

	builder, err := requests.NewBuilder(parsedURLToUse, logger, encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON))
	if err != nil {
		return nil, nil, err
	}

	return pasetoClient, builder, nil
}

func initializeCookiePoweredClient(cookie *http.Cookie) (*httpclient.Client, error) {
	if urlToUse == "" {
		panic("url not set!")
	}

	c, err := httpclient.NewClient(
		parsedURLToUse,
		httpclient.UsingLogger(logging.NewNonOperationalLogger()),
		httpclient.UsingCookie(cookie),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.UsingDebug(true)); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}
func initializePASETOPoweredClient(clientID string, secretKey []byte) (*httpclient.Client, error) {
	c, err := httpclient.NewClient(
		parsedURLToUse,
		httpclient.UsingLogger(logging.NewNonOperationalLogger()),
		httpclient.UsingPASETO(clientID, secretKey),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.UsingDebug(true)); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}
