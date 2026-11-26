package wmf

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Sentinels for HTTP failures
var (
	// ErrNotFound matches an HTTP 404 answer.
	ErrNotFound = errors.New("not found")
	// ErrForbidden matches an HTTP 403 answer.
	ErrForbidden = errors.New("forbidden")
	// ErrTooManyRequests matches an HTTP 429 answer.
	ErrTooManyRequests = errors.New("too many requests")
	// ErrServerError matches an HTTP 5xx answer.
	ErrServerError = errors.New("server error")
)

// HTTPError is an error from Wikimedia API with a status other than 200, 206 or 302.
//
// Exception: with Client.EnableRetryAfter set, 429 and 502-504 answers are retried
// inside the client and, when they finally fail, come back as *RetryAfterError or a
// plain error, not as *HTTPError, so the HTTP sentinels do not match them.
type HTTPError struct {
	StatusCode int
	// RetryAfter is the Retry-After header given in seconds, set for any status.
	// It is 0 when the header is absent, an HTTP date, or not a number.
	RetryAfter time.Duration
	URL        string
	// Body is "<status>:<response body without whitespace>".
	Body string
}

// Error returns the same text the client returned before HTTPError existed.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("wmf api call returned status not 200, 206 or 302 for url %s with error %s", e.URL, e.Body)
}

// Is matches the HTTP sentinels by status code.
func (e *HTTPError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrTooManyRequests:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrServerError:
		return e.StatusCode >= 500 && e.StatusCode <= 599
	}

	return false
}

// ErrEntityRedirected is returned by GetWikibaseEntity when the entity was merged
// into another one: the answer holds only the target entity, under its own ID.
var ErrEntityRedirected = errors.New("entity redirected")

// Sentinels for MediaWiki Action API error codes
var (
	// ErrMaxLag matches the API error code "maxlag" (replicas are lagging).
	ErrMaxLag = errors.New("maxlag")
	// ErrRateLimited matches the API error code "ratelimited".
	ErrRateLimited = errors.New("ratelimited")
	// ErrReadOnly matches the API error code "readonly".
	ErrReadOnly = errors.New("readonly")
)

// Error returns the error object as JSON, the text the client returned before Error was an error type.
func (e *Error) Error() string {
	dta, err := json.Marshal(e)
	if err != nil {
		return e.Info
	}

	return string(dta)
}

// Is matches the API sentinels by error code.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrMaxLag:
		return e.Code == "maxlag"
	case ErrRateLimited:
		return e.Code == "ratelimited"
	case ErrReadOnly:
		return e.Code == "readonly"
	}

	return false
}

// textError keeps an existing error text while still unwrapping to the underlying error.
type textError struct {
	msg string
	err error
}

func (e *textError) Error() string { return e.msg }
func (e *textError) Unwrap() error { return e.err }
