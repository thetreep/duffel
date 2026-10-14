// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import "context"

type (
	PlacesClient interface {
		PlaceSuggestions(ctx context.Context, query string, opts ...RequestOption) ([]*Place, error)
		Cities(ctx context.Context, opts ...RequestOption) *Iter[City]
		City(ctx context.Context, id string, opts ...RequestOption) (*City, error)
	}

	Place struct {
		ID              string     `json:"id"`
		Airports        []*Airport `json:"airports"`
		City            *City      `json:"city"`
		CityName        string     `json:"city_name"`
		CountryName     string     `json:"country_name"`
		IATACityCode    string     `json:"iata_city_code"`
		IATACode        string     `json:"iata_code"`
		IATACountryCode string     `json:"iata_country_code"`
		ICAOCode        string     `json:"icao_code"`
		Latitude        float64    `json:"latitude"`
		Longitude       float64    `json:"longitude"`
		Name            string     `json:"name"`
		TimeZone        string     `json:"time_zone"`
		Type            PlaceType  `json:"type"`
	}

	PlaceType string
)

const PlaceTypeAirport = "airport"
const PlaceTypeCity = "city"

func (a *API) PlaceSuggestions(ctx context.Context, query string, opts ...RequestOption) ([]*Place, error) {
	return newRequestWithAPI[EmptyPayload, Place](a).
		Get("/places/suggestions").WithParam("query", query).
		WithOptions(opts...).Slice(ctx)
}

func (a *API) Cities(ctx context.Context, opts ...RequestOption) *Iter[City] {
	return newRequestWithAPI[EmptyPayload, City](a).Get("/air/cities").WithOptions(opts...).Iter(ctx)
}

func (a *API) City(ctx context.Context, id string, opts ...RequestOption) (*City, error) {
	return newRequestWithAPI[EmptyPayload, City](a).Getf("/air/cities/%s", id).WithOptions(opts...).Single(ctx)
}
