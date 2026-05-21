// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestRateLimitPreemptionReturnsDuffelError(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Get("/air/orders/ord_123").
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "0").
		SetHeader("Ratelimit-Reset", time.Now().Add(30*time.Second).Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-get-order.json")

	client := New("duffel_test_123")
	_, err := client.GetOrder(context.TODO(), "ord_123")

	a.Error(err)
	a.True(IsErrorType(err, RateLimitError))
	a.True(IsErrorCode(err, RateLimitExceeded))
	a.True(ErrIsRetryable(err))

	derr, ok := err.(*DuffelError)
	a.True(ok)
	a.Equal(http.StatusTooManyRequests, derr.StatusCode)
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
