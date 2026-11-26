package wmf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newErrorTestClient returns a client that talks to an httptest server driven by hdl.
func newErrorTestClient(t *testing.T, hdl http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(hdl)
	t.Cleanup(srv.Close)

	return &Client{
		DefaultURL:      srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}, srv
}

// statusHandler answers every request with sts, the given headers and body.
func statusHandler(sts int, body string, hdr map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range hdr {
			w.Header().Set(k, v)
		}

		w.WriteHeader(sts)
		_, _ = w.Write([]byte(body))
	}
}

func TestHTTPErrorStatus(t *testing.T) {
	sentinels := []error{ErrNotFound, ErrForbidden, ErrTooManyRequests, ErrServerError}

	for _, tc := range []struct {
		sts  int
		want error
	}{
		{http.StatusNotFound, ErrNotFound},
		{http.StatusForbidden, ErrForbidden},
		{http.StatusTooManyRequests, ErrTooManyRequests},
		{http.StatusInternalServerError, ErrServerError},
		{http.StatusBadGateway, ErrServerError},
		{http.StatusServiceUnavailable, ErrServerError},
		{http.StatusGatewayTimeout, ErrServerError},
		{http.StatusBadRequest, nil},
	} {
		t.Run(fmt.Sprint(tc.sts), func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(tc.sts, "{}", nil))

			_, err := clt.GetPageHTML(context.Background(), "enwiki", "Earth")

			var he *HTTPError
			if assert.True(t, errors.As(err, &he)) {
				assert.Equal(t, tc.sts, he.StatusCode)
			}

			for _, s := range sentinels {
				assert.Equal(t, s == tc.want, errors.Is(err, s), "sentinel %q", s)
			}
		})
	}
}

func TestHTTPErrorRetryAfter(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		want time.Duration
	}{
		{"seconds", "7", 7 * time.Second},
		{"http date", "Wed, 21 Oct 2026 07:28:00 GMT", 0},
		{"invalid", "abc", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(http.StatusTooManyRequests, "{}", map[string]string{"Retry-After": tc.val}))

			_, err := clt.GetPageHTML(context.Background(), "enwiki", "Earth")

			var he *HTTPError
			if assert.True(t, errors.As(err, &he)) {
				assert.Equal(t, tc.want, he.RetryAfter)
			}
		})
	}
}

func TestHTTPErrorText(t *testing.T) {
	var url string

	clt, srv := newErrorTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		url = r.URL.RequestURI()
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "x"}`))
	})

	_, err := clt.GetPageHTML(context.Background(), "enwiki", "Earth")

	assert.EqualError(t, err, "wmf api call returned status not 200, 206 or 302 for url "+srv.URL+url+" with error 404 Not Found:{\"error\":\"x\"}")
}

func TestHTTPErrorWrapped(t *testing.T) {
	clt, _ := newErrorTestClient(t, statusHandler(http.StatusNotFound, "{}", nil))

	_, err := clt.GetPageHTML(context.Background(), "enwiki", "Earth")
	err = fmt.Errorf("fetch: %w", err)

	var he *HTTPError
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.True(t, errors.As(err, &he))
}

func TestHTTPErrorTimeout(t *testing.T) {
	clt, _ := newErrorTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	clt.HTTPClient = &http.Client{Timeout: 50 * time.Millisecond}

	_, err := clt.GetPageHTML(context.Background(), "enwiki", "Earth")

	var he *HTTPError
	var ne net.Error
	assert.False(t, errors.As(err, &he))
	assert.True(t, errors.As(err, &ne) && ne.Timeout())
}

// apiErrorBody is an Action API answer (HTTP 200) carrying an error object.
func apiErrorBody(code, info string) string {
	return `{"error":{"code":"` + code + `","info":"` + info + `"}}`
}

// apiErrorJSON is the text the client produced for that body before *Error became an error type.
func apiErrorJSON(t *testing.T, code, info string) string {
	t.Helper()

	dta, err := json.Marshal(struct {
		Code string `json:"code,omitempty"`
		Info string `json:"info,omitempty"`
	}{code, info})
	if err != nil {
		t.Fatal(err)
	}

	return string(dta)
}

func TestAPIErrorGetPages(t *testing.T) {
	clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody("maxlag", "Waiting for 10.0.0.1: 5 seconds lagged"), nil))

	_, err := clt.GetPages(context.Background(), "enwiki", []string{"Earth"})

	var ae *Error
	assert.True(t, errors.Is(err, ErrMaxLag))
	if assert.True(t, errors.As(err, &ae)) {
		assert.Equal(t, "maxlag", ae.Code)
	}
	assert.EqualError(t, err, "OK:"+apiErrorJSON(t, "maxlag", "Waiting for 10.0.0.1: 5 seconds lagged"))
}

func TestAPIErrorGetUsers(t *testing.T) {
	clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody("maxlag", "lagged"), nil))

	_, err := clt.GetUsers(context.Background(), "enwiki", []int{1})

	var ae *Error
	assert.True(t, errors.Is(err, ErrMaxLag))
	if assert.True(t, errors.As(err, &ae)) {
		assert.Equal(t, "maxlag", ae.Code)
	}
	assert.EqualError(t, err, "OK:"+apiErrorJSON(t, "maxlag", "lagged"))
}

func TestAPIErrorCodes(t *testing.T) {
	sentinels := []error{ErrMaxLag, ErrRateLimited, ErrReadOnly}

	for _, tc := range []struct {
		code string
		want error
	}{
		{"maxlag", ErrMaxLag},
		{"ratelimited", ErrRateLimited},
		{"readonly", ErrReadOnly},
		{"badtitle", nil},
	} {
		t.Run(tc.code, func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody(tc.code, "info"), nil))

			_, err := clt.GetPages(context.Background(), "enwiki", []string{"Earth"})

			var ae *Error
			if assert.True(t, errors.As(err, &ae)) {
				assert.Equal(t, tc.code, ae.Code)
			}

			for _, s := range sentinels {
				assert.Equal(t, s == tc.want, errors.Is(err, s), "sentinel %q", s)
			}
		})
	}
}

func TestAPIErrorRandomPages(t *testing.T) {
	clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody("maxlag", "lagged"), nil))

	_, err := clt.GetRandomPages(context.Background(), "enwiki")

	var ae *Error
	assert.EqualError(t, err, "OK:lagged")
	assert.True(t, errors.As(err, &ae))
	assert.True(t, errors.Is(err, ErrMaxLag))
}

func TestAPIErrorWrapped(t *testing.T) {
	clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody("maxlag", "lagged"), nil))

	_, err := clt.GetPages(context.Background(), "enwiki", []string{"Earth"})

	assert.True(t, errors.Is(fmt.Errorf("x: %w", err), ErrMaxLag))
}

func TestGetWikibaseEntityErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sts     int
		body    string
		wantErr error
		notErr  error
		wantTxt string
		wantRev int
	}{
		{name: "redirected", sts: http.StatusOK, body: `{"entities":{"Q200":{"id":"Q200","type":"item"}}}`, wantErr: ErrEntityRedirected},
		{name: "missing", sts: http.StatusOK, body: `{"entities":{"Q100":{"id":"Q100","missing":true}}}`, wantErr: ErrEntityNotFound},
		{name: "no entities", sts: http.StatusOK, body: `{"entities":{}}`, wantErr: ErrEntityNotFound},
		{name: "several other entities", sts: http.StatusOK, body: `{"entities":{"Q200":{"id":"Q200"},"Q300":{"id":"Q300"}}}`, wantErr: ErrEntityNotFound},
		{name: "http 404", sts: http.StatusNotFound, body: `{}`, wantErr: ErrEntityNotFound},
		{name: "http 503", sts: http.StatusServiceUnavailable, body: `Not Found upstream`, wantErr: ErrServerError, notErr: ErrEntityNotFound},
		{name: "api error", sts: http.StatusOK, body: apiErrorBody("maxlag", "lag"), wantErr: ErrMaxLag, wantTxt: "wikibase entity API error: maxlag - lag"},
		{name: "found", sts: http.StatusOK, body: `{"entities":{"Q100":{"id":"Q100","lastrevid":5}}}`, wantRev: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(tc.sts, tc.body, nil))

			rsp, err := clt.GetWikibaseEntity(context.Background(), "enwiki", "Q100")

			if tc.wantErr == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, tc.wantRev, rsp.Entities["Q100"].LastRevID)
				}
				return
			}

			assert.Nil(t, rsp)
			assert.True(t, errors.Is(err, tc.wantErr), "want %v, got %v", tc.wantErr, err)
			if tc.notErr != nil {
				assert.False(t, errors.Is(err, tc.notErr))
			}
			if tc.wantTxt != "" {
				assert.EqualError(t, err, tc.wantTxt)
			}
		})
	}
}

func TestAPIErrorTextAllSites(t *testing.T) {
	const j = `{"code":"maxlag","info":"lagged"}`
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		call func(*Client) error
		want string
	}{
		{"GetAllPages", func(c *Client) error { return c.GetAllPages(ctx, "enwiki", func([]*Page) {}) }, "OK:" + j},
		{"GetPageByRevision", func(c *Client) error { _, err := c.GetPageByRevision(ctx, "enwiki", 1); return err }, "error returned by WMF - OK:" + j},
		{"GetAllRevisions", func(c *Client) error { _, err := c.GetAllRevisions(ctx, "enwiki", 1); return err }, "error returned by WMF - OK:" + j},
		{"GetLanguages", func(c *Client) error { _, err := c.GetLanguages(ctx, "enwiki"); return err }, "error fetching languages, OK:" + j},
		{"GetNamespaces", func(c *Client) error { _, err := c.GetNamespaces(ctx, "enwiki"); return err }, "OK:" + j},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(http.StatusOK, apiErrorBody("maxlag", "lagged"), nil))

			err := tc.call(clt)

			assert.EqualError(t, err, tc.want)
			assert.True(t, errors.Is(err, ErrMaxLag))
		})
	}
}

func TestHTTPErrorThroughRequestWrappers(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		call func(*Client) error
	}{
		{"GetAllPages", func(c *Client) error { return c.GetAllPages(ctx, "enwiki", func([]*Page) {}) }},
		{"GetPageByRevision (sendRequest)", func(c *Client) error { _, err := c.GetPageByRevision(ctx, "enwiki", 1); return err }},
		{"GetAllRevisions", func(c *Client) error { _, err := c.GetAllRevisions(ctx, "enwiki", 1); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clt, _ := newErrorTestClient(t, statusHandler(http.StatusServiceUnavailable, "down", nil))

			err := tc.call(clt)

			var he *HTTPError
			assert.True(t, errors.Is(err, ErrServerError))
			if assert.True(t, errors.As(err, &he)) {
				assert.Equal(t, http.StatusServiceUnavailable, he.StatusCode)
				// Text unchanged: the request details, then the original error text.
				assert.True(t, strings.HasPrefix(err.Error(), "URL: "))
				assert.True(t, strings.HasSuffix(err.Error(), "\nError: "+he.Error()))
			}
		})
	}
}
