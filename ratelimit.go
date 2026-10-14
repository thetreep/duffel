// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"net/http"
	"strconv"
	"time"
)

type (
	RateLimit struct {
		Limit     int
		Remaining int
		ResetAt   time.Time
		// Period is the time left until ResetAt, measured against the response Date header.
		Period time.Duration
	}
)

var headerTimeFormats = []string{
	time.RFC1123,
	"Mon, 2 Jan 2006 15:04:05 MST",
}

// parseRateLimit returns nil on missing or bad headers: they are advisory and must never fail a response.
func parseRateLimit(resp *http.Response) *RateLimit {
	limit, err := strconv.Atoi(resp.Header.Get("Ratelimit-Limit"))
	if err != nil {
		return nil
	}
	remaining, err := strconv.Atoi(resp.Header.Get("Ratelimit-Remaining"))
	if err != nil {
		return nil
	}
	resetAt, ok := parseHeaderTime(resp.Header.Get("Ratelimit-Reset"))
	if !ok {
		return nil
	}
	// Measure against the server clock when possible so local clock skew does not distort the period.
	now, ok := parseHeaderTime(resp.Header.Get("Date"))
	if !ok {
		now = time.Now()
	}

	return &RateLimit{
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
		Period:    max(resetAt.Sub(now), 0),
	}
}

func parseHeaderTime(v string) (time.Time, bool) {
	for _, format := range headerTimeFormats {
		if t, err := time.Parse(format, v); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
