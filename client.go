// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/segmentio/encoding/json"
)

const RequestIDHeader = "x-request-id"

type Payload[T any] struct {
	Data T `json:"data"`
}

type ResponsePayload[T any] struct {
	Meta *ListMeta `json:"meta"`
	Data T         `json:"data"`
}

type RequestOption func(req *http.Request) error

type EmptyPayload struct{}

func buildRequestPayload[T any](data T) *Payload[T] {
	return &Payload[T]{
		Data: data,
	}
}

func encodePayload[T any](requestInput T) (io.ReadCloser, error) {
	payload := bytes.NewBuffer(nil)
	err := json.NewEncoder(payload).Encode(buildRequestPayload(requestInput))
	if err != nil {
		return nil, err
	}

	return io.NopCloser(payload), nil
}

func newInternalClient[Req any, Resp any](a *API) *client[Req, Resp] {
	return &client[Req, Resp]{api: a}
}

func (c *client[R, T]) makeRequest(
	ctx context.Context, resourceName string, method string, input *R, opts ...RequestOption,
) (*http.Response, error) {
	body, err := encodePayload(input)
	if err != nil {
		return nil, err
	}

	if c.api.APIToken == "" {
		return nil, fmt.Errorf("duffel: missing API token")
	}

	u, err := c.buildRequestURL(resourceName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if method != http.MethodGet {
		req.Body = body
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if !c.api.options.Debug {
		req.Header.Add("Accept-Encoding", "gzip")
	}
	req.Header.Add("User-Agent", c.api.options.UserAgent)
	req.Header.Add("Duffel-Version", c.api.options.Version)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.api.APIToken))

	// Apply request options
	for _, o := range opts {
		if o != nil {
			err := o(req)
			if err != nil {
				return nil, err
			}
		}
	}

	if c.api.options.Debug {
		b, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Printf("REQUEST:\n%s\n", string(b))
	}

	resp, err := c.api.httpDoer.Do(req)
	if err != nil {
		return nil, err
	}
	rateLimit := c.api.recordResponse(resp)

	if c.api.options.Debug {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		fmt.Printf("RESPONSE:\n%s\n", string(b))
	}

	if resp.StatusCode > 399 {
		defer resp.Body.Close()
		return nil, decodeError(resp, rateLimit)
	}

	return resp, nil
}

func (c *client[R, T]) buildRequestURL(resourceName string) (*url.URL, error) {
	u, err := url.Parse(c.api.options.Host)

	if err != nil {
		return nil, err
	}
	u.Path = resourceName

	return u, nil
}

func gzipResponseReader(response *http.Response) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
	} else {
		reader = response.Body
	}
	return reader, nil
}

// decodeError always returns a *DuffelError with the status so an HTTP error never looks like a transport failure.
func decodeError(response *http.Response, rateLimit *RateLimit) error {
	derr := &DuffelError{
		StatusCode: response.StatusCode,
		Retryable: response.StatusCode == http.StatusServiceUnavailable ||
			response.StatusCode == http.StatusGatewayTimeout ||
			response.StatusCode == http.StatusTooManyRequests,
	}
	if response.StatusCode == http.StatusTooManyRequests && rateLimit != nil {
		derr.RetryAfter = rateLimit.Period
	}
	defer derr.fillRequestID(response)

	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") {
		if reader, err := gzipResponseReader(response); err == nil {
			if json.NewDecoder(reader).Decode(derr) == nil && len(derr.Errors) > 0 {
				return derr
			}
		}
	}

	fallback := Error{Type: ApiError, Title: http.StatusText(response.StatusCode)}
	switch {
	case response.StatusCode == http.StatusTooManyRequests:
		fallback.Type, fallback.Code, fallback.Message = RateLimitError, RateLimitExceeded, "Rate limit exceeded."
	case response.StatusCode < 500:
		fallback.Message = fmt.Sprintf("HTTP %d without a Duffel error body.", response.StatusCode)
	default:
		fallback.Code, fallback.Message = InternalServerError, "An internal server error occurred. Please try again later."
	}
	derr.Errors = []Error{fallback}
	return derr
}
