package wmf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockTransport) Return429(req *http.Request) *mock.Call {
	return m.On("RoundTrip", req).Return(&http.Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}, nil)
}

func (m *MockTransport) Return200(req *http.Request) *mock.Call {
	return m.On("RoundTrip", req).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("success")),
	}, nil)
}

type mockTimer struct {
	delays []time.Duration
}

func (m *mockTimer) After(d time.Duration) <-chan time.Time {
	m.delays = append(m.delays, d)
	c := make(chan time.Time, 1)
	// Retry immediately
	c <- time.Now()
	return c
}

type clientTestSuite struct {
	suite.Suite
	client        *Client
	mockTransport *MockTransport
	timer         *mockTimer
	req           *http.Request
}

func (c *clientTestSuite) SetupTest() {
	c.mockTransport = new(MockTransport)
	c.timer = &mockTimer{}
	c.req, _ = http.NewRequest(http.MethodGet, "https://example.com", nil)

	c.client = &Client{
		HTTPClient:         &http.Client{Transport: c.mockTransport},
		EnableRetryAfter:   true,
		MaxAttempts:        3,
		DefaultRetryAfter:  10 * time.Millisecond,
		ExponentialBackOff: false,
		Tracer: func(ctx context.Context, attr map[string]string) (func(error, string), context.Context) {
			return func(error, string) {}, ctx
		},
		retryHooks: []retry.Option{retry.WithTimer(c.timer)},
	}
}

func (c *clientTestSuite) TearDownTest() {
	c.mockTransport.AssertExpectations(c.T())
}

func (c *clientTestSuite) TestSuccessOnFirstAttempt() {
	c.mockTransport.Return200(c.req).Once()

	res, err := c.client.do(c.client.HTTPClient, c.req)

	c.NoError(err)
	c.Equal(http.StatusOK, res.StatusCode)
	c.Empty(c.timer.delays)
}

func (c *clientTestSuite) TestRetriesDisabled() {
	c.client.EnableRetryAfter = false

	c.mockTransport.Return429(c.req).Once()

	res, err := c.client.do(c.client.HTTPClient, c.req)

	c.Error(err)
	c.Nil(res)
	c.Empty(c.timer.delays)
}

func (c *clientTestSuite) TestRecoversOnThirdAttempt() {
	c.mockTransport.Return429(c.req).Twice()
	c.mockTransport.Return200(c.req).Once()

	res, err := c.client.do(c.client.HTTPClient, c.req)

	c.NoError(err)
	c.Equal(http.StatusOK, res.StatusCode)

	expectedDelays := []time.Duration{10 * time.Millisecond, 10 * time.Millisecond}
	c.Equal(expectedDelays, c.timer.delays)
}

func (c *clientTestSuite) TestFailsAfterMaxAttempts() {
	c.client.MaxAttempts = 2

	c.mockTransport.Return429(c.req).Twice()

	res, err := c.client.do(c.client.HTTPClient, c.req)

	c.Error(err)
	c.Nil(res)

	expectedDelays := []time.Duration{10 * time.Millisecond}
	c.Equal(expectedDelays, c.timer.delays)
}

func (c *clientTestSuite) TestRespectsRetryAfterHeader() {
	c.mockTransport.On("RoundTrip", c.req).Return(&http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Retry-After": []string{"5"}},
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}, nil).Once()

	c.mockTransport.Return200(c.req).Once()

	res, err := c.client.do(c.client.HTTPClient, c.req)

	c.NoError(err)
	c.Equal(http.StatusOK, res.StatusCode)

	expectedDelays := []time.Duration{5 * time.Second}
	c.Equal(expectedDelays, c.timer.delays)
}

func (c *clientTestSuite) TestExponentialBackoff() {
	c.client.MaxAttempts = 5
	c.client.ExponentialBackOff = true

	c.mockTransport.Return429(c.req).Times(3)
	c.mockTransport.Return200(c.req).Once()

	_, err := c.client.do(c.client.HTTPClient, c.req)

	c.NoError(err)

	expectedDelays := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond}
	c.Equal(expectedDelays, c.timer.delays)
}

func (c *clientTestSuite) TestMaxRetryDelayCap() {
	c.client.MaxAttempts = 5
	c.client.ExponentialBackOff = true
	c.client.DefaultRetryAfter = 10 * time.Millisecond
	c.client.MaxRetryAfter = 15 * time.Millisecond

	c.mockTransport.Return429(c.req).Times(3)
	c.mockTransport.Return200(c.req).Once()

	_, err := c.client.do(c.client.HTTPClient, c.req)

	c.NoError(err)

	expectedDelays := []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 15 * time.Millisecond}
	c.Equal(expectedDelays, c.timer.delays)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}
