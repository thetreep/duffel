// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/schema"
	"github.com/segmentio/encoding/json"
)

type RequestBuilder[Req any, Resp any] struct {
	client         *client[Req, Resp]
	method         string
	resourcePath   string
	requestOptions []RequestOption
	body           *Req
}

type RequestMiddleware func(r *http.Request) error
type ResponseMiddleware func(r *http.Response) error

type ParamEncoder[T any] interface {
	Encode(v url.Values) error
}

func WithRequestPagination(meta *ListMeta) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()

		if meta != nil {
			enc := schema.NewEncoder()
			enc.SetAliasTag("url")
			err := enc.Encode(meta, q)
			if err != nil {
				return (err)
			}
			req.URL.RawQuery = q.Encode()
		}
		return nil
	}
}

func WithEncodableParams[T any](params ...ParamEncoder[T]) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()

		for _, param := range params {
			err := param.Encode(q)
			if err != nil {
				return err
			}
		}

		req.URL.RawQuery = q.Encode()
		return nil
	}
}

func WithURLParams[T any](params ...T) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()
		enc := schema.NewEncoder()
		enc.SetAliasTag("url")

		for _, param := range params {
			err := enc.Encode(param, q)
			if err != nil {
				return err
			}
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// WithHeader sets a header on a single request, for example x-duffel-device-ip on CreateOrder.
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}

func WithURLParam(key, value string) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()
		q.Add(key, value)
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// newRequestWithAPI returns a new fluent requst builder for the given API.
// The input types are the request payload and response payload.
// For Get requests, the request payload is used to type the URL params.
//
// The response payload is used to type the parsed response body, and is the type returned by the finalizers.
func newRequestWithAPI[ReqT any, ResponseT any](a *API) *RequestBuilder[ReqT, ResponseT] {
	return &RequestBuilder[ReqT, ResponseT]{
		client: newInternalClient[ReqT, ResponseT](a),
		method: http.MethodGet,
	}
}

// WithParam adds a single query param to the URL.
// These operations will be applied in defined order after the request is initialized.
func (r *RequestBuilder[Req, Resp]) WithParam(key, value string) *RequestBuilder[Req, Resp] {
	r.requestOptions = append(r.requestOptions, WithURLParam(key, value))
	return r
}

// WithParams sets the URL query params for the request.
// These operations will be applied in defined order after the request is initialized.
func (r *RequestBuilder[Req, Resp]) WithParams(obj ...ParamEncoder[Req]) *RequestBuilder[Req, Resp] {
	r.requestOptions = append(r.requestOptions, WithEncodableParams(obj...))
	return r
}

func (r *RequestBuilder[Req, Resp]) WithOptions(opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.requestOptions = append(r.requestOptions, opts...)
	return r
}

// Get sets the request method to GET and the request path to the given path. Global request options are applied.
func (r *RequestBuilder[Req, Resp]) Get(path string, opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.method = http.MethodGet
	r.resourcePath = path
	r.requestOptions = append(r.requestOptions, opts...)
	return r
}

// Getf is like Get but accepts a format string and args.
func (r *RequestBuilder[Req, Resp]) Getf(path string, a ...any) *RequestBuilder[Req, Resp] {
	return r.Get(fmt.Sprintf(path, a...))
}

func (r *RequestBuilder[Req, Resp]) Delete(path string, opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.method = http.MethodDelete
	r.resourcePath = path
	r.requestOptions = append(r.requestOptions, opts...)
	return r
}

func (r *RequestBuilder[Req, Resp]) Deletef(path string, a ...any) *RequestBuilder[Req, Resp] {
	return r.Delete(fmt.Sprintf(path, a...))
}

// Post sets the request method to POST, the request path to the given path, and the request payload to body. Global request options are applied.
func (r *RequestBuilder[Req, Resp]) Post(path string, body *Req, opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.method = http.MethodPost
	r.resourcePath = path
	r.body = body
	r.requestOptions = append(r.requestOptions, opts...)
	return r
}

func (r *RequestBuilder[Req, Resp]) Postf(path string, a ...any) *RequestBuilder[Req, Resp] {
	return r.Post(fmt.Sprintf(path, a...), r.body)
}

func (r *RequestBuilder[Req, Resp]) Body(body *Req) *RequestBuilder[Req, Resp] {
	r.body = body
	return r
}

// Patch sets the request method to PATCH, the request path to the given path, and the request payload to body. Global request options are applied.
func (r *RequestBuilder[Req, Resp]) Patch(path string, body *Req, opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.method = http.MethodPatch
	r.resourcePath = path
	r.requestOptions = append(r.requestOptions, opts...)
	r.body = body
	return r
}

// Put sets the request method to PUT, the request path to the given path, and the request payload to body.
func (r *RequestBuilder[Req, Resp]) Put(path string, body *Req, opts ...RequestOption) *RequestBuilder[Req, Resp] {
	r.method = http.MethodPut
	r.resourcePath = path
	r.requestOptions = append(r.requestOptions, opts...)
	r.body = body
	return r
}

// Iter finalizes the request and returns an iterator over the response.
func (r *RequestBuilder[Req, Resp]) Iter(ctx context.Context) *Iter[Resp] {
	return GetIter(
		func(lastMeta *ListMeta) (*List[Resp], error) {
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(r.client.api.options.Timeout))
			defer cancel()

			list := new(List[Resp])
			response, err := r.makeRequest(ctx, WithRequestPagination(lastMeta))
			if err != nil {
				return nil, errors.Wrap(err, "failed to make request")
			}

			container := new(ResponsePayload[[]*Resp])
			err = decodeResponse(response, &container)
			if err != nil {
				return nil, err
			}

			list.SetListMeta(container.Meta)
			list.SetItems(container.Data)
			list.setRequestID(response.Header.Get(RequestIDHeader))
			return list, nil
		},
	)
}

// waitRateLimit blocks for the wait advised by a 429 and reports whether the request should be retried.
func waitRateLimit(ctx context.Context, err error) bool {
	derr, ok := asDuffelError(err)
	if !ok || derr.StatusCode != http.StatusTooManyRequests {
		return false
	}
	timer := time.NewTimer(max(derr.RetryAfter, time.Second))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Slice finalizes the request and returns the first page of items as a slice along with the error.
// This is only needed for endpoints without pagination, such as place suggestions.
func (r *RequestBuilder[Req, Resp]) Slice(ctx context.Context) ([]*Resp, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(r.client.api.options.Timeout))
	defer cancel()

	response, err := r.makeRequest(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	container := new(ResponsePayload[[]*Resp])
	err = decodeResponse(response, &container)
	if err != nil {
		return nil, err
	}

	return container.Data, nil
}

// Single finalizes the request and returns a single item response.
func (r *RequestBuilder[Req, Resp]) Single(ctx context.Context) (*Resp, error) {
	data, _, err := r.SingleWithResponse(ctx)
	return data, err
}

// SingleWithResponse finalizes the request and returns the decoded response along with the HTTP status code.
func (r *RequestBuilder[Req, Resp]) SingleWithResponse(ctx context.Context) (*Resp, int, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(r.client.api.options.Timeout))
	defer cancel()

	response, err := r.makeRequest(ctx)
	if err != nil {
		return nil, 0, err
	}
	container := new(ResponsePayload[*Resp])
	err = decodeResponse(response, &container)
	if err == nil && container.Data == nil {
		err = &DecodeError{StatusCode: response.StatusCode, RequestID: response.Header.Get(RequestIDHeader), Err: errors.New("response has no data")}
	}
	if err != nil {
		return nil, response.StatusCode, err
	}

	return container.Data, response.StatusCode, nil
}

// Empty finalizes the request and returns an error if the request fails.
// It is to be used for requests that are expected to return no data.
func (r *RequestBuilder[Req, Resp]) Empty(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(r.client.api.options.Timeout))
	defer cancel()

	response, err := r.makeRequest(ctx)
	if err != nil {
		return err
	}

	_, _ = io.Copy(io.Discard, response.Body)
	return response.Body.Close()
}

// makeRequest retries a 429 on GET only: a replayed POST could book or charge twice.
func (r *RequestBuilder[Req, Resp]) makeRequest(ctx context.Context, opts ...RequestOption) (*http.Response, error) {
	opts = append(r.requestOptions, opts...)
	for attempt := 0; ; attempt++ {
		resp, err := r.client.makeRequest(ctx, r.resourcePath, r.method, r.body, opts...)
		if r.method != http.MethodGet || attempt == maxRateLimitRetries || !waitRateLimit(ctx, err) {
			return resp, err
		}
	}
}

const maxRateLimitRetries = 3

// decodeResponse returns *DecodeError on failure because the request did reach Duffel.
func decodeResponse[T any](resp *http.Response, v T) error {
	defer resp.Body.Close()

	reader, err := gzipResponseReader(resp)
	if err == nil {
		err = json.NewDecoder(reader).Decode(v)
	}
	if err != nil {
		return &DecodeError{StatusCode: resp.StatusCode, RequestID: resp.Header.Get(RequestIDHeader), Err: err}
	}
	return nil
}

// normalizeParams returns a slice of interfaces from the given params.
// This is only neeeded because Go doesn't allow slice conversion of slice spreads.
// See: https://github.com/golang/go/wiki/InterfaceSlice
func normalizeParams[T ParamEncoder[T]](params []T) []ParamEncoder[T] {
	p := make([]ParamEncoder[T], len(params))
	for i, param := range params {
		p[i] = ParamEncoder[T](param)
	}
	return p
}
