// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
)

type (
	AircraftClient interface {
		ListAircraft(ctx context.Context, opts ...RequestOption) *Iter[Aircraft]
		GetAircraft(ctx context.Context, id string, opts ...RequestOption) (*Aircraft, error)
	}
)

func (a *API) ListAircraft(ctx context.Context, opts ...RequestOption) *Iter[Aircraft] {
	return newRequestWithAPI[EmptyPayload, Aircraft](a).
		Get("/air/aircraft").
		WithOptions(opts...).Iter(ctx)
}

func (a *API) GetAircraft(ctx context.Context, id string, opts ...RequestOption) (*Aircraft, error) {
	return newRequestWithAPI[EmptyPayload, Aircraft](a).
		Getf("/air/aircraft/%s", id).
		WithOptions(opts...).Single(ctx)
}

var _ AircraftClient = (*API)(nil)
