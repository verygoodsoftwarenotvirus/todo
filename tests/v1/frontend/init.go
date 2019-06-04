package frontend

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"github.com/icrowley/fake"
)

var urlToUse string

const (
	seleniumHubAddr = "http://selenium-hub:4444/wd/hub"
	// These paths will be different on your system.
	geckoDriverPath = "/usr/local/bin/chromedriver"
	port            = 4444

	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		ta := os.Getenv("TARGET_ADDRESS")
		if ta == "" {
			urlToUse = defaultTestInstanceURL
		} else {
			u, err := url.Parse(ta)
			if err != nil {
				panic(err)
			}
			urlToUse = u.String()
		}
	} else {
		urlToUse = localTestInstanceURL
	}

	logger := zerolog.NewZeroLogger()
	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(urlToUse)

	fake.Seed(time.Now().UnixNano())

	//for {
	//	res, err := http.Head(seleniumHubAddr + "/status")
	//	if err != nil {
	//		time.Sleep(time.Second / 2)
	//	} else if res != nil {
	//		if res.StatusCode != http.StatusOK {
	//			log.Println(res.StatusCode)
	//		} else {
	//			break
	//		}
	//	}
	//}
	// NOTE: this is sad, but also the only thing that consistently works
	// see above for my vain attempts at a real solution to this problem
	time.Sleep(10 * time.Second)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
