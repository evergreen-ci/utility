package utility

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestRetryRequestOn413(t *testing.T) {
	var callCount int32

	handler := NewMockHandler()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decide the status code dynamically for the first vs. subsequent calls
		callCount++
		if callCount == 1 {
			handler.StatusCode = http.StatusRequestEntityTooLarge // 413
		} else {
			handler.StatusCode = http.StatusOK
		}
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	opts := RetryRequestOptions{
		RetryOptions: RetryOptions{
			MaxAttempts: 3,
		},
		RetryOn413: true,
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err, "failed to create request")

	resp, err := RetryRequest(context.Background(), req, opts)
	require.NoError(t, err, "request failed")

	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 status code")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&callCount), int32(2), "expected multiple attempts")
}

func TestMockHandler(t *testing.T) {
	handler := NewMockHandler()
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()

	t.Run("Serve", func(t *testing.T) {
		urlString := server.URL + "/example"
		handler.Header = map[string][]string{
			"header1": {"header1"},
			"header2": {"header1", "header2"},
		}
		handler.Body = []byte("some body")
		handler.StatusCode = http.StatusOK

		req, err := http.NewRequest(http.MethodGet, urlString, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Len(t, handler.Calls, 1)
		for key, values := range handler.Header {
			assert.Equal(t, values, resp.Header.Values(key))
		}
		assert.Equal(t, handler.StatusCode, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, handler.Body, body)
		assert.NoError(t, handler.GetWriteError())

		urlString = server.URL + "/example2"
		handler.Header = map[string][]string{
			"header1": {"header1"},
		}
		handler.Body = []byte("example not found")
		handler.StatusCode = http.StatusNotFound

		req, err = http.NewRequest(http.MethodGet, urlString, nil)
		require.NoError(t, err)
		resp, err = client.Do(req)
		require.NoError(t, err)

		assert.Len(t, handler.Calls, 2)
		for key, values := range handler.Header {
			assert.Equal(t, values, resp.Header.Values(key))
		}
		assert.Equal(t, handler.StatusCode, resp.StatusCode)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, handler.Body, body)
		assert.NoError(t, handler.GetWriteError())
	})
}

// 	assert.Equal(t, int32(2), callCount, "expected exactly two total attempts")
// }

func TestRetryRequestBodyRemainsValidOnSecondAttempt(t *testing.T) {
	var callCount int32

	// On the first attempt, the server reads the body, then returns 413.
	// On the second attempt, it again reads the body and returns 200.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		reqBody, err := io.ReadAll(r.Body)
		require.NoError(t, err, "failed to read request body")

		if atomic.AddInt32(&callCount, 1) == 1 {
			w.WriteHeader(http.StatusRequestEntityTooLarge) // 413
			_, _ = w.Write([]byte("failing body"))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("valid body: " + string(reqBody)))
		}
	}))
	defer server.Close()

	opts := RetryRequestOptions{
		RetryOptions: RetryOptions{
			MaxAttempts: 2,
		},
		RetryOn413: true,
	}

	// We include some request body content to confirm it remains valid across retries.
	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("request body data"))
	require.NoError(t, err, "failed to create request")

	resp, err := RetryRequest(context.Background(), req, opts)
	require.NoError(t, err, "request failed")
	defer resp.Body.Close()

	assert.Equal(t, int32(2), callCount, "expected exactly two total attempts")

	// Read the final response body (which should be from the second request).
	data, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr, "failed to read body on second attempt")

	// The second attempt should contain the original request body text:
	assert.Contains(t, string(data), "request body data", "expected second attempt to see same request body")

}
