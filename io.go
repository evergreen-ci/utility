package utility

import (
	"context"
	"io"
	"net/http"

	"github.com/peterhellberg/link"
	"github.com/pkg/errors"
)

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// the provided writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{w}
}

type paginatedReadCloser struct {
	ctx        context.Context
	client     *http.Client
	reqHeader  http.Header
	respHeader http.Header

	io.ReadCloser
}

// NewPaginatedReadCloser returns an io.ReadCloser implementation for paginated
// HTTP responses. It is safe to pass in a non-paginated response, thus the
// caller need not check for the appropriate header keys. Optionally pass in
// a header for subsequent page requests, useful if user information such as
// API keys are required.
func NewPaginatedReadCloser(ctx context.Context, client *http.Client, resp *http.Response, reqHeader http.Header) *paginatedReadCloser {
	return &paginatedReadCloser{
		ctx:        ctx,
		client:     client,
		reqHeader:  reqHeader,
		respHeader: resp.Header,
		ReadCloser: resp.Body,
	}
}

// Read reads the underlying HTTP response body. Once the body is read, the
// HTTP response header is used to request the next page, if any. If a
// subsequent page is successfully requested, the previous body is closed and
// both the body and header are replaced with the new response. An io.EOF error
// is returned when the current response body is exhausted and either the
// response header does not contain a "next" key or the "next" key's URL
// returns no data.
func (r *paginatedReadCloser) Read(p []byte) (int, error) {
	if r.ReadCloser == nil {
		return 0, io.EOF
	}

	var (
		n      int
		offset int
		err    error
	)
	for offset < len(p) {
		n, err = r.ReadCloser.Read(p[offset:])
		offset += n
		if err == io.EOF {
			err = r.getNextPage()
		}
		if err != nil {
			break
		}
	}

	return offset, err
}

func (r *paginatedReadCloser) getNextPage() error {
	group, ok := link.ParseHeader(r.respHeader)["next"]
	if ok {
		req, err := http.NewRequestWithContext(r.ctx, http.MethodGet, group.URI, nil)
		if err != nil {
			return errors.Wrap(err, "creating http request for next page")
		}
		if r.reqHeader != nil {
			req.Header = r.reqHeader
		}

		resp, err := r.client.Do(req)
		if err != nil {
			return errors.Wrap(err, "requesting next page")
		}
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("failed to get next page with resp '%s'", resp.Status)
		}

		if err = r.Close(); err != nil {
			return errors.Wrap(err, "closing last response reader")
		}

		r.respHeader = resp.Header
		r.ReadCloser = resp.Body
	} else {
		return io.EOF
	}

	return nil
}
