// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type (
	RateLimit struct {
		Limit     int
		Remaining int
		ResetAt   time.Time
		Period    time.Duration
	}
)

var headerTimeFormats = []string{
	time.RFC1123,
	"Mon, 2 Jan 2006 15:04:05 MST",
}

func parseRateLimit(resp *http.Response) (*RateLimit, error) {
	rl := &RateLimit{}

	limit, err := strconv.Atoi(resp.Header.Get("Ratelimit-Limit"))
	if err != nil {
		return nil, err
	}
	rl.Limit = limit

	remaining, err := strconv.Atoi(resp.Header.Get("Ratelimit-Remaining"))
	if err != nil {
		return nil, err
	}
	rl.Remaining = remaining

	resetHeader := resp.Header.Get("Ratelimit-Reset")
	for _, format := range headerTimeFormats {
		if resetAt, err := time.Parse(format, resetHeader); err == nil {
			rl.ResetAt = resetAt
			break
		}
	}

	if rl.ResetAt.IsZero() {
		return nil, fmt.Errorf("failed to parse Ratelimit-Reset header: %s, no known date formats match", resetHeader)
	}

	date, err := time.Parse(time.RFC1123, resp.Header.Get("Date"))
	if err != nil {
		return nil, err
	}

	rl.Period = rl.ResetAt.Sub(date)

	return rl, nil
}
