// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type (
	OfferRequestClient interface {
		CreateOfferRequest(ctx context.Context, requestInput OfferRequestInput) (*OfferRequest, error)
		GetOfferRequest(ctx context.Context, id string) (*OfferRequest, error)
		CreatePartialOfferRequest(ctx context.Context, requestInput OfferRequestInput) (*OfferRequest, error)
		GetFullPartialOfferRequest(ctx context.Context, requestInput PartialOfferRequestInput) (*OfferRequest, error)
		GetPartialOfferRequests(ctx context.Context, requestInput PartialOfferRequestInput) (*OfferRequest, error)
		ListOfferRequests(ctx context.Context) *Iter[OfferRequest]
	}

	OfferRequestInput struct {
		// The passengers who want to travel. If you specify an age for a passenger, the type may differ for the same passenger in different offers due to airline's different rules. e.g. one airline may treat a 14 year old as an adult, and another as a young adult. You may only specify an age or a type – not both.
		Passengers []OfferRequestPassenger `json:"passengers" url:"-"`
		// The slices that make up this offer request. One-way journeys can be expressed using one slice, whereas return trips will need two.
		Slices []OfferRequestSlice `json:"slices" url:"-"`
		// The cabin that the passengers want to travel in
		CabinClass CabinClass `json:"cabin_class,omitempty" url:"-"`
		// The maximum number of connections within any slice of the offer. For example 0 means a direct flight which will have a single segment within each slice and 1 means a maximum of two segments within each slice of the offer.
		MaxConnections *int `json:"max_connections,omitempty" url:"-"`
		// When set to true, the offer request resource returned will include all the offers returned by the airlines
		// The private fares codes of airline.  The key is the airline's IATA code that provided the private fare code.
		PrivateFares map[string][]PrivateFare `json:"private_fares,omitempty" url:"-"`

		ReturnOffers bool `json:"-" url:"return_offers"`
		// The maximum amount of time in milliseconds to wait for each airline to respond
		SupplierTimeout int `json:"-" url:"supplier_timeout,omitempty"`
	}

	OfferRequestSlice struct {
		DepartureDate Date   `json:"departure_date"`
		Destination   string `json:"destination"`
		Origin        string `json:"origin"`
	}

	// The corporate_code and tour_code are provided to you by the airline and the tracking_reference is to identify your business by the airlines.
	PrivateFare struct {
		CorporateCode     string          `json:"corporate_code,omitempty"`
		TrackingReference string          `json:"tracking_reference,omitempty"`
		TourCode          string          `json:"tour_code,omitempty"`
		Type              PrivateFareType `json:"type,omitempty"`
	}

	OfferRequestPassenger struct {
		ID                       string                    `json:"id,omitempty"`
		FamilyName               string                    `json:"family_name,omitempty"`
		GivenName                string                    `json:"given_name,omitempty"`
		Age                      int                       `json:"age,omitempty"`
		LoyaltyProgrammeAccounts []LoyaltyProgrammeAccount `json:"loyalty_programme_accounts,omitempty"`
		FareType                 FareType                  `json:"fare_type,omitempty"`
		Type                     PassengerType             `json:"type,omitempty"`
	}

	// OfferRequest is the response from the OfferRequest endpoint, created using the OfferRequestInput.
	OfferRequest struct {
		ID         string                  `json:"id"`
		ClientKey  string                  `json:"client_key"` // For use with https://duffel.com/docs/guides/ancillaries-component
		LiveMode   bool                    `json:"live_mode"`
		CreatedAt  time.Time               `json:"created_at"`
		Slices     []BaseSlice             `json:"slices"`
		Passengers []OfferRequestPassenger `json:"passengers"`
		CabinClass CabinClass              `json:"cabin_class"`
		Offers     []Offer                 `json:"offers"`
	}

	PartialOfferRequestInput struct {
		PartialOfferRequestID string
		SelectedPartialOffers []string
	}
)

func (a *API) CreateOfferRequest(ctx context.Context, requestInput OfferRequestInput) (*OfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, OfferRequest](a).
		Post("/air/offer_requests", &requestInput).
		WithParams(requestInput).
		Single(ctx)
}

func (a *API) CreatePartialOfferRequest(ctx context.Context, requestInput OfferRequestInput) (*OfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, OfferRequest](a).
		Post("/air/partial_offer_requests", &requestInput).
		Single(ctx)
}

func (a *API) GetPartialOfferRequests(ctx context.Context, requestInput PartialOfferRequestInput) (
	*OfferRequest, error,
) {
	return newRequestWithAPI[PartialOfferRequestInput, OfferRequest](a).
		Getf("/air/partial_offer_requests/%s", requestInput.PartialOfferRequestID).
		WithParams(requestInput).
		Single(ctx)
}

func (a *API) GetFullPartialOfferRequest(ctx context.Context, requestInput PartialOfferRequestInput) (
	*OfferRequest, error,
) {
	return newRequestWithAPI[PartialOfferRequestInput, OfferRequest](a).
		Getf("/air/partial_offer_requests/%s/fares", requestInput.PartialOfferRequestID).
		WithParams(requestInput).
		Single(ctx)
}

func (a *API) GetOfferRequest(ctx context.Context, id string) (*OfferRequest, error) {
	return newRequestWithAPI[EmptyPayload, OfferRequest](a).Getf("/air/offer_requests/%s", id).Single(ctx)
}

func (a *API) ListOfferRequests(ctx context.Context) *Iter[OfferRequest] {
	return newRequestWithAPI[EmptyPayload, OfferRequest](a).Get("/air/offer_requests").Iter(ctx)
}

// Encode implements the ParamEncoder interface.
func (o OfferRequestInput) Encode(q url.Values) error {
	q.Set("return_offers", strconv.FormatBool(o.ReturnOffers))
	return nil
}

func (o PartialOfferRequestInput) Encode(q url.Values) error {
	q["selected_partial_offer[]"] = o.SelectedPartialOffers
	return nil
}
