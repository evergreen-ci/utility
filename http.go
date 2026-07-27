package utility

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const httpClientTimeout = 5 * time.Minute

var httpClientPool *sync.Pool

func init() {
	initHTTPPool()
}

func initHTTPPool() {
	httpClientPool = &sync.Pool{
		New: func() interface{} { return newBaseConfiguredHttpClient() },
	}
}

func newBaseConfiguredHttpClient() *http.Client {
	return DefaultHttpClient(DefaultTransport())
}

func DefaultHttpClient(rt http.RoundTripper) *http.Client {
	return &http.Client{
		Timeout:   httpClientTimeout,
		Transport: rt,
	}
}

func DefaultTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig:     &tls.Config{},
		Proxy:               http.ProxyFromEnvironment,
		DisableCompression:  false,
		DisableKeepAlives:   true,
		IdleConnTimeout:     20 * time.Second,
		MaxIdleConnsPerHost: 10,
		MaxIdleConns:        50,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

}

// GetHTTPClient produces default HTTP client from the pool,
// constructing a new client if needed. Always pair calls to
// GetHTTPClient with defered calls to PutHTTPClient.
func GetHTTPClient() *http.Client { return httpClientPool.Get().(*http.Client) }

// PutHTTPClient returns the client to the pool, automatically
// reconfiguring the transport.
func PutHTTPClient(c *http.Client) {
	c.Timeout = httpClientTimeout

	switch transport := c.Transport.(type) {
	case *http.Transport:
		transport.TLSClientConfig.InsecureSkipVerify = false
		c.Transport = transport
	case *rehttp.Transport:
		c.Transport = transport.RoundTripper
		PutHTTPClient(c)
		return
	case *oauth2.Transport:
		c.Transport = transport.Base
		PutHTTPClient(c)
		return
	default:
		c.Transport = DefaultTransport()
	}

	httpClientPool.Put(c)
}

// HTTPRetryConfiguration makes it possible to configure the retry
// semantics for retryable clients. In most cases, construct this
// object using the NewDefaultHttpRetryConf, which provides reasonable
// defaults.
type HTTPRetryConfiguration struct {
	MaxDelay        time.Duration
	BaseDelay       time.Duration
	MaxRetries      int
	TemporaryErrors bool
	Methods         []string
	Statuses        []int
	Errors          []error
	ErrorStrings    []string
}

// NewDefaultHTTPRetryConf constructs a HTTPRetryConfiguration object
// with reasonable defaults.
func NewDefaultHTTPRetryConf() HTTPRetryConfiguration {
	return HTTPRetryConfiguration{
		MaxRetries:      50,
		TemporaryErrors: true,
		MaxDelay:        5 * time.Second,
		BaseDelay:       50 * time.Millisecond,
		Methods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
		},
		Statuses: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
			http.StatusInsufficientStorage,
			http.StatusConflict,
			http.StatusRequestTimeout,
			http.StatusPreconditionFailed,
			http.StatusExpectationFailed,
		},
	}
}

// GetHTTPRetryableClient produces an HTTP client that automatically
// retries failed requests according to the configured
// parameters. Couple calls to GetHTTPRetryableClient with deferred
// calls to PutHTTPClient.
func GetHTTPRetryableClient(conf HTTPRetryConfiguration) *http.Client {
	client := GetHTTPClient()

	statusRetries := []rehttp.RetryFn{}
	if len(conf.Statuses) > 0 {
		statusRetries = append(statusRetries, rehttp.RetryStatuses(conf.Statuses...))
	} else {
		conf.TemporaryErrors = true
	}

	if conf.TemporaryErrors {
		statusRetries = append(statusRetries, rehttp.RetryTemporaryErr())
	}

	if len(conf.Errors) > 0 {
		statusRetries = append(statusRetries, rehttp.RetryIsErr(func(err error) bool {
			for _, errToCheck := range conf.Errors {
				if err == errToCheck {
					return true
				}
			}
			return false
		}))
	}

	if len(conf.ErrorStrings) > 0 {
		statusRetries = append(statusRetries, rehttp.RetryIsErr(func(err error) bool {
			for _, errToCheck := range conf.ErrorStrings {
				if err.Error() == errToCheck {
					return true
				}
			}
			return false
		}))
	}

	retryFns := []rehttp.RetryFn{rehttp.RetryAny(statusRetries...)}

	if len(conf.Methods) > 0 {
		retryFns = append(retryFns, rehttp.RetryHTTPMethods(conf.Methods...))
	}

	if conf.MaxRetries > 0 {
		retryFns = append(retryFns, rehttp.RetryMaxRetries(conf.MaxRetries))
	}

	client.Transport = rehttp.NewTransport(client.Transport,
		rehttp.RetryAll(retryFns...),
		rehttp.ExpJitterDelay(conf.BaseDelay, conf.MaxDelay))

	return client
}

// GetDefaultHTTPRetryableClient provides a retryable client with
// the default settings. Couple calls to GetHTTPRetryableClient, with defered
// calls to PutHTTPClient.
func GetDefaultHTTPRetryableClient() *http.Client {
	return GetHTTPRetryableClient(NewDefaultHTTPRetryConf())
}

// HTTPRetryFunction makes it possible to write customizable retry
// logic. Returning true if the request should be retried again and
// false otherwise.
type HTTPRetryFunction func(index int, req *http.Request, resp *http.Response, err error) bool

// HTTPDelayFunction makes it possible to write customizable retry
// backoff logic, by allowing you to evaluate the previous request and
// response and return the duration to wait before the next request.
type HTTPDelayFunction func(index int, req *http.Request, resp *http.Response, err error) time.Duration

func makeRetryFn(in HTTPRetryFunction) rehttp.RetryFn {
	return func(attempt rehttp.Attempt) bool {
		return in(attempt.Index, attempt.Request, attempt.Response, attempt.Error)
	}
}

func makeDelayFn(in HTTPDelayFunction) rehttp.DelayFn {
	return func(attempt rehttp.Attempt) time.Duration {
		return in(attempt.Index, attempt.Request, attempt.Response, attempt.Error)
	}
}

func setupCustomHTTPRetryableClient(retry HTTPRetryFunction, delay HTTPDelayFunction, client *http.Client) *http.Client {
	client.Transport = rehttp.NewTransport(client.Transport, makeRetryFn(retry), makeDelayFn(delay))
	return client
}

// GetCustomHTTPRetryableClient allows you to generate an HTTP client
// that automatically retries failed request based on the provided
// custom logic.
func GetCustomHTTPRetryableClient(retry HTTPRetryFunction, delay HTTPDelayFunction) *http.Client {
	client := GetHTTPClient()
	return setupCustomHTTPRetryableClient(retry, delay, client)
}

// GetCustomHTTPRetryableClientWithTransport allows you to generate an HTTP client
// that automatically retries failed request based on the provided
// custom logic and HTTP transport.
func GetCustomHTTPRetryableClientWithTransport(rt http.RoundTripper, retry HTTPRetryFunction, delay HTTPDelayFunction) *http.Client {
	client := GetHTTPClient()
	client.Transport = rehttp.NewTransport(rt, makeRetryFn(retry), makeDelayFn(delay))
	return client
}

// GetOAuth2HTTPClient produces an HTTP client that will supply OAuth2
// credentials with all requests. There is no validation of the
// token, and you should always call PutHTTPClient to return the
// client to the pool when you're done with it.
func GetOAuth2HTTPClient(oauthToken string) *http.Client {
	return SetupOauth2HTTPClient(oauthToken, GetHTTPClient())
}

// GetOauth2DefaultHTTPRetryableClient constructs an HTTP client that
// supplies OAuth2 credentials with all requests, retrying failed
// requests automatically according to the default retryable
// options. There is no validation of the token, and you should always
// call PutHTTPClient to return the client to the pool when you're
// done with it.
func GetOauth2DefaultHTTPRetryableClient(oauthToken string) *http.Client {
	return SetupOauth2HTTPClient(oauthToken, GetDefaultHTTPRetryableClient())
}

// GetOauth2HTTPRetryableClient constructs an HTTP client that
// supplies OAuth2 credentials with all requests, retrying failed
// requests automatically according to the configuration
// provided. There is no validation of the token, and you should
// always call PutHTTPClient to return the client to the pool when
// you're done with it.
func GetOauth2HTTPRetryableClient(oauthToken string, conf HTTPRetryConfiguration) *http.Client {
	return SetupOauth2HTTPClient(oauthToken, GetHTTPRetryableClient(conf))
}

// GetOauth2CustomHTTPRetryableClient constructs an HTTP client that
// supplies OAuth2 credentials with all requests, retrying failed
// requests automatically according to definitions of the provided
// functions. There is no validation of the token, and you should
// always call PutHTTPClient to return the client to the pool when
// you're done with it.
func GetOauth2CustomHTTPRetryableClient(token string, retry HTTPRetryFunction, delay HTTPDelayFunction) *http.Client {
	return SetupOauth2HTTPClient(token, GetCustomHTTPRetryableClient(retry, delay))
}

// SetupOauth2HTTPClient configures an HTTP client that
// supplies OAuth2 credentials with all requests. There is no validation
// of the token, and you should always call PutHTTPClient to return the
// client to the pool when you're done with it.
func SetupOauth2HTTPClient(token string, client *http.Client) *http.Client {
	client.Transport = &oauth2.Transport{
		Base: client.Transport,
		Source: oauth2.ReuseTokenSource(nil, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)),
	}
	return client
}

// SetupOauth2CustomHTTPRetryableClient configures an HTTP client that
// supplies OAuth2 credentials with all requests, retrying failed
// requests automatically according to definitions of the provided
// functions. There is no validation of the token, and you should
// always call PutHTTPClient to return the client to the pool when
// you're done with it.
func SetupOauth2CustomHTTPRetryableClient(token string, retry HTTPRetryFunction, delay HTTPDelayFunction, client *http.Client) *http.Client {
	return SetupOauth2HTTPClient(token, setupCustomHTTPRetryableClient(retry, delay, client))
}

// TemporayError defines an interface for use in retryable HTTP
// clients to identify certain errors as Temporary.
type TemporaryError interface {
	error
	Temporary() bool
}

// IsTemporaryError returns true if the error object is also a
// temporary error.
func IsTemporaryError(err error) bool {
	if terr, ok := err.(TemporaryError); ok {
		return terr.Temporary()
	}
	return false
}

// RetryRequestOptions specifically are the options when doing
// a retryable request. It wraps the RetryOptions and adds options
// on top that are specific to requests.
type RetryRequestOptions struct {
	RetryOptions

	// RetryOnInvalidBody is a flag that determines whether to retry
	// when reading the response body fails.
	RetryOnInvalidBody bool

	// RetryOn413 is a flag that determines whether to retry when
	// the server returns a 413 status code.
	RetryOn413 bool
}

// RetryRequest takes an http.Request and makes the request until it's successful,
// hits a max number of retries, or times out
func RetryRequest(ctx context.Context, r *http.Request, opts RetryRequestOptions) (*http.Response, error) {
	r = r.WithContext(ctx)

	// Save the entire request body so we can resend it on each attempt.
	var requestBody []byte
	if r.Body != nil {
		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			return nil, errors.Wrap(readErr, "failed to read request body")
		}
		r.Body.Close()
		requestBody = bodyBytes
	}

	client := GetDefaultHTTPRetryableClient()
	defer PutHTTPClient(client)

	attempt := 1
	var resp *http.Response
	var err error

	if err := Retry(ctx, func() (bool, error) {
		defer func() {
			attempt++
		}()

		// Ensure the same body is attached for each attempt
		if requestBody != nil {
			r.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		resp, err = client.Do(r)
		if err != nil {
			return true, err
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if opts.RetryOnInvalidBody {
				originalBody := resp.Body
				defer originalBody.Close()
				// Test if the body is valid by reading it.
				body := &bytes.Buffer{}
				if _, err := body.ReadFrom(resp.Body); err != nil {
					return true, err
				}

				// If it is valid, reset the body so the caller can read it.
				resp.Body = io.NopCloser(body)
			}
			return false, nil
		}

		if resp.StatusCode == http.StatusRequestEntityTooLarge && opts.RetryOn413 {
			return true, errors.Errorf("server returned status %d (request entity too large)", resp.StatusCode)
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return false, errors.Errorf("server returned status %d", resp.StatusCode)
		}

		// if we get here it should most likely be a 5xx status code

		return true, errors.Errorf("server returned status %d", resp.StatusCode)
	}, opts.RetryOptions); err != nil {
		return resp, err
	}

	return resp, nil
}

// RetryHTTPDelay returns the function that generates the exponential backoff
// delay between retried HTTP requests.
func RetryHTTPDelay(opts RetryOptions) HTTPDelayFunction {
	backoff := getBackoff(opts)
	return func(index int, req *http.Request, resp *http.Response, err error) time.Duration {
		return backoff.ForAttempt(float64(index))
	}
}

// MockHandler implements the http.Handler interface for mock HTTP servers. See
// the `https://pkg.go.dev/net/http/httptest` for more information on mocking
// HTTP servers.
type MockHandler struct {
	Mu         sync.Mutex
	Calls      []*url.URL
	Header     map[string][]string
	Body       []byte
	StatusCode int

	writeError error
}

// NewMockHandler returns a MockHandler object that can be used as an
// http.Handler.
func NewMockHandler() *MockHandler { return &MockHandler{} }

// ServeHTTP is a thread-safe handler for mocking HTTP responses. The request
// URLs are recorded and the customizable header, body, and status code are all
// written to the http.ResponseWriter.
// GetWriteError (see below) returns the most recent write error, if any.
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	h.Calls = append(h.Calls, r.URL)

	header := w.Header()
	for key, values := range h.Header {
		for _, val := range values {
			header.Add(key, val)
		}
	}

	if h.StatusCode > 0 {
		w.WriteHeader(h.StatusCode)
	}

	if h.Body != nil {
		if _, err := w.Write(h.Body); err != nil {
			h.writeError = err
		}
	}
}

// GetWriteError returns the most recent error from writing to the
// http.ResponseWriter.
func (h *MockHandler) GetWriteError() error { return h.writeError }
