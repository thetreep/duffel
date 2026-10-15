// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestClientError(t *testing.T) {
	ctx := context.TODO()
	a := assert.New(t)
	gock.New("https://api.duffel.com/air/offer_requests").
		MatchParam("return_offers", "false").
		Reply(400).
		File("fixtures/400-bad-request.json")
	defer gock.Off()

	client := New("duffel_test_123")
	data, err := client.CreateOfferRequest(
		ctx, OfferRequestInput{
			ReturnOffers: false,
		},
	)
	a.Error(err)
	a.Nil(data)

	a.Equal("duffel: The airline responded with an unexpected error, please contact support", err.Error())

	derr := err.(*DuffelError)
	a.True(derr.IsType(AirlineError))
	a.True(derr.IsCode(AirlineUnknown))
	a.True(IsErrorType(err, AirlineError))
	a.True(IsErrorCode(err, AirlineUnknown))
	a.False(ErrIsRetryable(err))

	reqId, ok := RequestIDFromError(err)
	a.True(ok)
	a.Equal("FZW0H3HdJwKk5HMAAKxB", reqId)
}

func TestClientErrorBadGateway(t *testing.T) {
	ctx := context.TODO()
	a := assert.New(t)
	gock.New("https://api.duffel.com/air/offer_requests").
		Reply(502).
		AddHeader("Content-Type", "text/html").
		File("fixtures/502-bad-gateway.html")
	defer gock.Off()

	client := New("duffel_test_123")
	data, err := client.CreateOfferRequest(
		ctx, OfferRequestInput{
			ReturnOffers: true,
		},
	)
	a.Error(err)
	a.Nil(data)
	a.Equal("duffel: An internal server error occurred. Please try again later.", err.Error())
	a.False(ErrIsRetryable(err))
}

func TestClientError500NotRetryable(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)
	gock.New("https://api.duffel.com/air/offer_requests").
		MatchParam("return_offers", "false").
		Reply(500).
		File("fixtures/400-bad-request.json")

	client := New("duffel_test_123")
	_, err := client.CreateOfferRequest(
		context.TODO(), OfferRequestInput{
			ReturnOffers: false,
		},
	)
	a.Error(err)
	a.False(ErrIsRetryable(err))
}

func TestClientError503Retryable(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)
	gock.New("https://api.duffel.com/air/offer_requests").
		MatchParam("return_offers", "false").
		Reply(503).
		File("fixtures/503-service-unavailable.json")

	client := New("duffel_test_123")
	_, err := client.CreateOfferRequest(
		context.TODO(), OfferRequestInput{
			ReturnOffers: false,
		},
	)
	a.Error(err)
	a.True(ErrIsRetryable(err))
}

func TestClientErrorWithSource(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)
	gock.New("https://api.duffel.com/air/offer_requests").
		MatchParam("return_offers", "false").
		Reply(422).
		File("fixtures/422-validation-error.json")

	client := New("duffel_test_123")
	_, err := client.CreateOfferRequest(
		context.TODO(), OfferRequestInput{
			ReturnOffers: false,
		},
	)
	a.Error(err)

	derr := err.(*DuffelError)
	a.True(derr.IsType(ValidationError))
	a.True(derr.IsCode(ValidationRequired))
	a.NotNil(derr.Errors[0].Source)
	a.Equal("origin", derr.Errors[0].Source.Field)
	a.Equal("/slices/0/origin", derr.Errors[0].Source.Pointer)
}

// newTestClient points a real client at an httptest server, which unlike gock is safe for concurrent calls.
func newTestClient(t *testing.T, h http.HandlerFunc) Duffel {
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New("duffel_test_123", WithHost(srv.URL))
}

func serveFixture(t *testing.T, status int, fixture string, headers map[string]string) http.HandlerFunc {
	body, err := os.ReadFile(fixture)
	require.NoError(t, err)
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}
}

// recordedRequest is what newRecordingClient received, so tests can check the method, path, query and body sent.
type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   string
}

// newRecordingClient records each request and serves the fixture, or an empty body when fixture is "".
func newRecordingClient(t *testing.T, status int, fixture string) (Duffel, *recordedRequest) {
	serve := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }
	if fixture != "" {
		serve = serveFixture(t, status, fixture, nil)
	}
	rec := new(recordedRequest)
	return newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			sent, _ := io.ReadAll(r.Body)
			*rec = recordedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: string(sent)}
			serve(w, r)
		},
	), rec
}

func TestRateLimitRemainingZeroKeepsSuccessfulResponse(t *testing.T) {
	a := assert.New(t)
	client := newTestClient(
		t, serveFixture(
			t, http.StatusCreated, "fixtures/201-create-order.json", map[string]string{
				"Ratelimit-Limit":     "30",
				"Ratelimit-Remaining": "0",
				"Ratelimit-Reset":     time.Now().Add(30 * time.Second).UTC().Format(time.RFC1123),
			},
		),
	)

	order, err := client.CreateOrder(context.TODO(), CreateOrderInput{Type: OrderTypeInstant})
	a.NoError(err)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)

	rl, ok := client.LastRateLimit()
	a.True(ok)
	a.Equal(30, rl.Limit)
	a.Equal(0, rl.Remaining)
}

func TestMissingRateLimitHeadersKeepSuccessfulResponse(t *testing.T) {
	a := assert.New(t)
	client := newTestClient(t, serveFixture(t, http.StatusCreated, "fixtures/201-create-order.json", nil))

	order, err := client.CreateOrder(context.TODO(), CreateOrderInput{Type: OrderTypeInstant})
	a.NoError(err)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)

	_, ok := client.LastRateLimit()
	a.False(ok)
}

func TestRealRateLimitIsRetryableButPostIsNotReplayed(t *testing.T) {
	a := assert.New(t)
	var calls atomic.Int32
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.Header().Set("Ratelimit-Limit", "30")
			w.Header().Set("Ratelimit-Remaining", "0")
			w.Header().Set("Ratelimit-Reset", time.Now().Add(20*time.Second).UTC().Format(time.RFC1123))
			w.Header().Set(RequestIDHeader, "req_429")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errors":[{"type":"rate_limit_error","code":"rate_limit_exceeded","message":"Too many requests"}]}`))
		},
	)

	_, err := client.CreateOrder(context.TODO(), CreateOrderInput{})
	a.EqualValues(1, calls.Load())
	a.True(IsErrorCode(err, RateLimitExceeded))
	a.True(ErrIsRetryable(err))

	var derr *DuffelError
	a.ErrorAs(err, &derr)
	a.Equal(http.StatusTooManyRequests, derr.StatusCode)
	a.Greater(derr.RetryAfter, 10*time.Second)
	a.Equal("req_429", derr.Meta.RequestID)
}

func TestNonJSONErrorBodyKeepsStatus(t *testing.T) {
	a := assert.New(t)
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("slow down"))
		},
	)

	_, err := client.CreateOrder(context.TODO(), CreateOrderInput{})
	var derr *DuffelError
	a.ErrorAs(err, &derr)
	a.Equal(http.StatusTooManyRequests, derr.StatusCode)
	a.True(derr.Retryable)
	a.True(derr.IsCode(RateLimitExceeded))
}

func TestDecodeErrorAfterSuccess(t *testing.T) {
	a := assert.New(t)
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(RequestIDHeader, "req_bad_body")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{not json"))
		},
	)

	_, err := client.CreateOrder(context.TODO(), CreateOrderInput{Type: OrderTypeInstant})
	var decErr *DecodeError
	a.ErrorAs(err, &decErr)
	a.Equal(http.StatusCreated, decErr.StatusCode)
	a.Equal("req_bad_body", decErr.RequestID)
	a.False(ErrIsRetryable(err))

	id, ok := RequestIDFromError(err)
	a.True(ok)
	a.Equal("req_bad_body", id)

	// Iter wraps the error, errors.As must still find it.
	iter := client.ListOrders(context.TODO())
	a.False(iter.Next())
	a.ErrorAs(iter.Err(), &decErr)
}

func TestErrorHelpersSeeThroughWrapping(t *testing.T) {
	a := assert.New(t)
	base := &DuffelError{
		Meta:      ErrorMeta{RequestID: "req_1"},
		Errors:    []Error{{Type: InvalidStateError, Code: OfferNoLongerAvailable}},
		Retryable: true,
	}

	for _, err := range []error{base, errors.Wrap(base, "adapter"), fmt.Errorf("adapter: %w", base)} {
		a.True(IsErrorCode(err, OfferNoLongerAvailable))
		a.True(IsErrorType(err, InvalidStateError))
		a.True(ErrIsRetryable(err))
		id, ok := RequestIDFromError(err)
		a.True(ok)
		a.Equal("req_1", id)
	}

	a.False(IsErrorCode(fmt.Errorf("plain"), OfferNoLongerAvailable))
	_, ok := RequestIDFromError(fmt.Errorf("plain"))
	a.False(ok)
}

func TestConcurrentCallsShareOneClient(t *testing.T) {
	var calls atomic.Int32
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(RequestIDHeader, fmt.Sprintf("req_%d", calls.Add(1)))
			w.Header().Set("Ratelimit-Limit", "120")
			w.Header().Set("Ratelimit-Remaining", "60")
			w.Header().Set("Ratelimit-Reset", time.Now().Add(time.Minute).UTC().Format(time.RFC1123))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"id":"ord_1"}}`))
		},
	)

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(
			func() {
				order, err := client.GetOrder(context.TODO(), "ord_1")
				assert.NoError(t, err)
				assert.Equal(t, "ord_1", order.ID)
				client.LastRequestID()
				client.LastRateLimit()
			},
		)
	}
	wg.Wait()

	assert.EqualValues(t, 50, calls.Load())
	_, ok := client.LastRequestID()
	assert.True(t, ok)
}

func TestWithHeaderOnCreateOrder(t *testing.T) {
	a := assert.New(t)
	var got string
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Get("x-duffel-device-ip")
			serveFixture(t, http.StatusCreated, "fixtures/201-create-order.json", nil)(w, r)
		},
	)

	_, err := client.CreateOrder(
		context.TODO(), CreateOrderInput{Type: OrderTypeInstant}, WithHeader("x-duffel-device-ip", "203.0.113.5"),
	)
	a.NoError(err)
	a.Equal("203.0.113.5", got)
}

func TestDefaultTimeoutIs130Seconds(t *testing.T) {
	a := assert.New(t)
	client := New("duffel_test_123")
	api := client.(*API)
	a.Equal(130*time.Second, api.options.Timeout)
}

func TestCustomTimeout(t *testing.T) {
	a := assert.New(t)
	client := New("duffel_test_123", WithTimeout(60*time.Second))
	api := client.(*API)
	a.Equal(60*time.Second, api.options.Timeout)
}

func TestCreateOrderAppliesOptionsOnce(t *testing.T) {
	a := assert.New(t)
	var query string
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.RawQuery
			serveFixture(t, http.StatusCreated, "fixtures/201-create-order.json", nil)(w, r)
		},
	)

	_, err := client.CreateOrder(context.TODO(), CreateOrderInput{}, WithURLParam("x", "1"))
	a.NoError(err)
	a.Equal("x=1", query)
}

func TestCreateOrderNullDataIsDecodeError(t *testing.T) {
	a := assert.New(t)
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":null}`))
		},
	)

	order, err := client.CreateOrder(context.TODO(), CreateOrderInput{})
	a.Nil(order)
	var decErr *DecodeError
	a.ErrorAs(err, &decErr)
	a.Equal(http.StatusOK, decErr.StatusCode)
}

func TestProxyErrorBodiesAreNotServerErrors(t *testing.T) {
	for _, body := range []string{"forbidden", `{"message":"Not Found"}`} {
		client := newTestClient(
			t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(body))
			},
		)

		_, err := client.GetOrder(context.TODO(), "ord_1")
		var derr *DuffelError
		assert.ErrorAs(t, err, &derr)
		assert.Equal(t, http.StatusForbidden, derr.StatusCode)
		assert.False(t, derr.IsCode(InternalServerError))
		assert.Equal(t, "duffel: HTTP 403 without a Duffel error body.", err.Error())
	}
}

func TestIterRetriesRateLimitedPage(t *testing.T) {
	a := assert.New(t)
	var calls atomic.Int32
	client := newTestClient(
		t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"errors":[{"type":"rate_limit_error","code":"rate_limit_exceeded"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"ord_1"}],"meta":{}}`))
		},
	)

	orders, err := Collect(client.ListOrders(context.TODO()))
	a.NoError(err)
	a.Len(orders, 1)
	a.EqualValues(2, calls.Load())
}
