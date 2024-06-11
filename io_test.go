package utility

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNopWriteCloser(t *testing.T) {
	wc := &myWriteCloser{}
	nop := NopWriteCloser(wc)
	nop.Close()
	assert.False(t, wc.closed)

	message := []byte("message")
	_, err := nop.Write(message)
	assert.Equal(t, message, wc.contents)
	assert.NoError(t, err)
}

type myWriteCloser struct {
	closed   bool
	contents []byte
}

func (w *myWriteCloser) Close() {
	w.closed = true
}
func (w *myWriteCloser) Write(p []byte) (n int, err error) {
	w.contents = append(w.contents, p...)
	return 0, nil
}

func TestPaginatedReadCloser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("PaginatedRoute", func(t *testing.T) {
		handler := &mockHandler{pages: 3}
		server := httptest.NewServer(handler)
		handler.baseURL = server.URL

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var r io.ReadCloser
		r = &paginatedReadCloser{
			ctx:        ctx,
			client:     http.DefaultClient,
			respHeader: resp.Header,
			ReadCloser: resp.Body,
		}

		data, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, "PAGINATED BODY PAGE 1\nPAGINATED BODY PAGE 2\nPAGINATED BODY PAGE 3\n", string(data))
		assert.NoError(t, r.Close())
	})
	t.Run("PaginatedRouteWithReqHeader", func(t *testing.T) {
		handler := &mockHandler{headerKeys: []string{"key1", "key2"}, pages: 3}
		server := httptest.NewServer(handler)
		handler.baseURL = server.URL

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		req.Header.Add("key1", "val1")
		req.Header.Add("key2", "val2")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var r io.ReadCloser
		r = &paginatedReadCloser{
			ctx:        ctx,
			client:     http.DefaultClient,
			reqHeader:  req.Header,
			respHeader: resp.Header,
			ReadCloser: resp.Body,
		}

		data, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, "HEADER: key1:val1,key2:val2\nPAGINATED BODY PAGE 1\nHEADER: key1:val1,key2:val2\nPAGINATED BODY PAGE 2\nHEADER: key1:val1,key2:val2\nPAGINATED BODY PAGE 3\n", string(data))
		assert.NoError(t, r.Close())
	})
	t.Run("NonPaginatedRoute", func(t *testing.T) {
		handler := &mockHandler{}
		server := httptest.NewServer(handler)
		handler.baseURL = server.URL

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var r io.ReadCloser
		r = &paginatedReadCloser{
			ctx:        ctx,
			client:     http.DefaultClient,
			respHeader: resp.Header,
			ReadCloser: resp.Body,
		}

		data, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, "NON-PAGINATED BODY PAGE", string(data))
		assert.NoError(t, r.Close())
	})
	t.Run("SplitPageByteSlice", func(t *testing.T) {
		handler := &mockHandler{pages: 2}
		server := httptest.NewServer(handler)
		handler.baseURL = server.URL

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var r io.ReadCloser
		r = &paginatedReadCloser{
			ctx:        ctx,
			client:     http.DefaultClient,
			respHeader: resp.Header,
			ReadCloser: resp.Body,
		}

		p := make([]byte, 33) // 1.5X len of each page
		n, err := r.Read(p)
		require.NoError(t, err)
		assert.Equal(t, len(p), n)
		assert.Equal(t, "PAGINATED BODY PAGE 1\nPAGINATED B", string(p))
		p = make([]byte, 33)
		n, err = r.Read(p)
		require.Equal(t, io.EOF, err)
		assert.Equal(t, 11, n)
		assert.Equal(t, "ODY PAGE 2\n", string(p[:11]))
		assert.NoError(t, r.Close())
	})
}

type mockHandler struct {
	baseURL    string
	headerKeys []string
	pages      int
	count      int
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pages > 0 {
		if h.count <= h.pages-1 {
			w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"%s\"", h.baseURL, "next"))

			if len(h.headerKeys) > 0 {
				var headerPairs []string
				for _, headerKey := range h.headerKeys {
					headerPairs = append(headerPairs, fmt.Sprintf("%s:%s", headerKey, r.Header.Get(headerKey)))
				}
				_, _ = w.Write([]byte(fmt.Sprintf("HEADER: %s\n", strings.Join(headerPairs, ","))))
			}

			_, _ = w.Write([]byte(fmt.Sprintf("PAGINATED BODY PAGE %d\n", h.count+1)))
		}
		h.count++
	} else {
		_, _ = w.Write([]byte("NON-PAGINATED BODY PAGE"))
	}
}
