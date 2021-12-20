package utility

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/rehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestPooledHTTPClient(t *testing.T) {
	t.Run("Initialized", func(t *testing.T) {
		require.NotNil(t, httpClientPool)
		cl := GetHTTPClient()
		require.NotNil(t, cl)
		require.NotPanics(t, func() { PutHTTPClient(cl) })
	})
	t.Run("ResettingSkipCert", func(t *testing.T) {
		cl := GetHTTPClient()
		cl.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
		PutHTTPClient(cl)
		cl2 := GetHTTPClient()
		assert.Equal(t, cl, cl2)
		assert.False(t, cl.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
		assert.False(t, cl2.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
	})
	t.Run("RehttpPool", func(t *testing.T) {
		initHTTPPool()
		cl := GetHTTPClient()
		clt := cl.Transport
		PutHTTPClient(cl)
		rcl := GetDefaultHTTPRetryableClient()
		assert.Equal(t, cl, rcl)
		assert.NotEqual(t, clt, rcl.Transport)
		assert.Equal(t, clt, rcl.Transport.(*rehttp.Transport).RoundTripper)
	})
}

type mockTransport struct {
	count         int
	expectedToken string
	status        int
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.count++

	resp := http.Response{
		StatusCode: t.status,
		Header:     http.Header{},
		Body:       ioutil.NopCloser(strings.NewReader("hi")),
	}

	token := req.Header.Get("Authorization")
	split := strings.Split(token, " ")
	if len(split) != 2 || split[0] != "Bearer" || split[1] != t.expectedToken {
		resp.StatusCode = http.StatusForbidden
	}
	return &resp, nil
}

func TestRetryableOauthClient(t *testing.T) {
	t.Run("Passing", func(t *testing.T) {
		c := GetOauth2DefaultHTTPRetryableClient("hi")
		defer PutHTTPClient(c)

		transport := &mockTransport{expectedToken: "hi", status: http.StatusOK}
		oldTransport := c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper
		defer func() {
			c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper = oldTransport
		}()
		c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper = transport

		resp, err := c.Get("https://example.com")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, transport.count)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Failed", func(t *testing.T) {
		conf := NewDefaultHTTPRetryConf()
		conf.MaxRetries = 5
		c := GetOauth2HTTPRetryableClient("hi", conf)
		defer PutHTTPClient(c)

		transport := &mockTransport{expectedToken: "hi", status: http.StatusInternalServerError}
		oldTransport := c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper
		defer func() {
			c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper = oldTransport
		}()
		c.Transport.(*oauth2.Transport).Base.(*rehttp.Transport).RoundTripper = transport

		resp, err := c.Get("https://example.com")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 6, transport.count)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestMockHandler(t *testing.T) {
	handler := NewMockHandler()
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()

	t.Run("Serve", func(t *testing.T) {
		url := server.URL + "/example"
		handler.Header = map[string][]string{
			"header1": []string{"header1"},
			"header2": []string{"header1", "header2"},
		}
		handler.Body = []byte("some body")
		handler.StatusCode = http.StatusOK

		req, err := http.NewRequest(http.MethodGet, url, nil)
		resp, err := client.Do(req)
		require.NoError(t, err)

		for key, values := range handler.Header {
			assert.Equal(t, values, resp.Header.Values(key))
		}
		assert.Equal(t, handler.StatusCode, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, handler.Body, body)
		assert.NoError(t, handler.WriteError())

		url = server.URL + "/example2"
		handler.Header = map[string][]string{
			"header1": []string{"header1"},
		}
		handler.Body = []byte("example not found")
		handler.StatusCode = http.StatusNotFound

		req, err = http.NewRequest(http.MethodGet, url, nil)
		resp, err = client.Do(req)
		require.NoError(t, err)

		for key, values := range handler.Header {
			assert.Equal(t, values, resp.Header.Values(key))
		}
		assert.Equal(t, handler.StatusCode, resp.StatusCode)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, handler.Body, body)
		assert.NoError(t, handler.WriteError())
	})
}
