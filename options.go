// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"net/http"
	"time"
)

// WithAPIToken sets the API host to the default Duffel production host.
func WithDefaultAPI() Option {
	return func(c *Options) {
		c.Host = "https://api.duffel.com"
	}
}

// WithHost allows you to specify the Duffel API host to use for making requests.
func WithHost(host string) Option {
	return func(c *Options) {
		c.Host = host
	}
}

// WithVersion allows you to specify "Duffel-Version" header for the API version that you are targeting.
func WithAPIVersion(version string) Option {
	return func(c *Options) {
		c.Version = version
	}
}

// WithUserAgent allows you to specify a custom user agent string to use for making requests.
func WithUserAgent(ua string) Option {
	return func(c *Options) {
		c.UserAgent = ua
	}
}

// WithHTTPClient allows you to specify a custom http.Client to use for making requests.
// This is useful if you want to use a custom transport or proxy.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Options) {
		c.HttpDoer = client
	}
}

// WithDebug enables debug logging of requests and responses.
// DO NOT USE IN PRODUCTION.
func WithDebug() Option {
	return func(c *Options) {
		c.Debug = true
	}
}

// WithTimeout sets the deadline timeout.
// The Duffel API recommends at least 130 seconds for order creation endpoints.
// Default is 130 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *Options) {
		c.Timeout = d
	}
}
