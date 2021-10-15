package websockets

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

type nopWriter struct{}

func (n *nopWriter) Write([]byte) (int, error) {
	return 0, nil
}

type arbitraryHijacker struct {
	http.ResponseWriter
}

type nopConn struct{}

func (n *nopConn) Read([]byte) (int, error) {
	return 0, nil
}

func (n *nopConn) Write([]byte) (int, error) {
	return 0, nil
}

func (n *nopConn) Close() error {
	return nil
}

type nopAddr struct{}

func (n *nopAddr) Network() string {
	return ""
}

func (n *nopAddr) String() string {
	return ""
}

func (n *nopConn) LocalAddr() net.Addr {
	return &nopAddr{}
}

func (n *nopConn) RemoteAddr() net.Addr {
	return &nopAddr{}
}

func (n *nopConn) SetDeadline(time.Time) error {
	return nil
}

func (n *nopConn) SetReadDeadline(time.Time) error {
	return nil
}

func (n *nopConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (a *arbitraryHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return &nopConn{}, bufio.NewReadWriter(bufio.NewReader(strings.NewReader("")), bufio.NewWriter(&nopWriter{})), nil
}

func TestWebsocketsService_SubscribeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		cookie := &http.Cookie{
			Name: helper.service.cookieName,
		}
		helper.req.AddCookie(cookie)

		helper.req.Method = http.MethodGet
		helper.req.Header.Set("Sec-Websocket-Key", t.Name())
		helper.req.Header.Set("Connection", "upgrade")
		helper.req.Header.Set("Upgrade", "websocket")
		helper.req.Header.Set("Sec-Websocket-Version", "13")

		helper.service.SubscribeHandler(&arbitraryHijacker{helper.res}, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
	})

	T.Run("adds to connection slice", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		cookie := &http.Cookie{
			Name: helper.service.cookieName,
		}
		helper.req.AddCookie(cookie)

		helper.req.Method = http.MethodGet
		helper.req.Header.Set("Sec-Websocket-Key", t.Name())
		helper.req.Header.Set("Connection", "upgrade")
		helper.req.Header.Set("Upgrade", "websocket")
		helper.req.Header.Set("Sec-Websocket-Version", "13")

		helper.service.SubscribeHandler(&arbitraryHijacker{helper.res}, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code)

		helper.service.SubscribeHandler(&arbitraryHijacker{helper.res}, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code)
	})

	T.Run("with error fetching session context", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		helper.service.SubscribeHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with missing cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		helper.req.Method = http.MethodGet
		helper.req.Header.Set("Sec-Websocket-Key", t.Name())
		helper.req.Header.Set("Connection", "upgrade")
		helper.req.Header.Set("Upgrade", "websocket")
		helper.req.Header.Set("Sec-Websocket-Version", "13")

		helper.service.SubscribeHandler(&arbitraryHijacker{helper.res}, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error upgrading connection", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		cookie := &http.Cookie{
			Name: helper.service.cookieName,
		}
		helper.req.AddCookie(cookie)

		helper.req.Method = http.MethodGet
		helper.req.Header.Set("Sec-Websocket-Key", t.Name())
		helper.req.Header.Set("Connection", "upgrade")
		helper.req.Header.Set("Upgrade", "websocket")
		helper.req.Header.Set("Sec-Websocket-Version", "13")

		helper.service.SubscribeHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})
}
