// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/bojanz/currency"
)

type (
	OfferRequestClient interface {
		CreateOfferRequest(ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption) (*OfferRequest, error)
		GetOfferRequest(ctx context.Context, id string, opts ...RequestOption) (*OfferRequest, error)
		CreateOfferRequestItineraries(ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption) (*ItineraryOfferRequest, error)
		GetOfferRequestItineraries(ctx context.Context, id string, opts ...RequestOption) (*ItineraryOfferRequest, error)
		CreateBatchOfferRequest(ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption) (*BatchOfferRequest, error)
		GetBatchOfferRequest(ctx context.Context, id string, opts ...RequestOption) (*BatchOfferRequest, error)
		GetBatchOfferRequestItineraries(ctx context.Context, id string, opts ...RequestOption) (*ItineraryOfferRequest, error)
		CreatePartialOfferRequest(ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption) (*OfferRequest, error)
		GetFullPartialOfferRequest(ctx context.Context, requestInput PartialOfferRequestInput, opts ...RequestOption) (*OfferRequest, error)
		GetPartialOfferRequests(ctx context.Context, requestInput PartialOfferRequestInput, opts ...RequestOption) (*OfferRequest, error)
		ListOfferRequests(ctx context.Context, opts ...RequestOption) *Iter[OfferRequest]
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
		// AirlineCreditIDs are checked against the returned offers, see Offer.AvailableAirlineCreditIDs.
		AirlineCreditIDs []string `json:"airline_credit_ids,omitempty" url:"-"`
		// IncludeSplitTicket adds one-way searches per slice to find split-ticket combinations.
		IncludeSplitTicket bool `json:"include_split_ticket,omitempty" url:"-"`

		ReturnOffers bool `json:"-" url:"return_offers"`
		// The maximum amount of time in milliseconds to wait for each airline to respond, 2000 to 60000, default 20000.
		// Keep it below the client Timeout to get partial results instead of a client timeout. Zero leaves it unset.
		SupplierTimeout int `json:"-" url:"supplier_timeout,omitempty"`
	}

	OfferRequestSlice struct {
		DepartureDate Date        `json:"departure_date"`
		Destination   string      `json:"destination"`
		Origin        string      `json:"origin"`
		DepartureTime *TimeWindow `json:"departure_time,omitempty"`
		ArrivalTime   *TimeWindow `json:"arrival_time,omitempty"`
	}

	// TimeWindow is a local hh:mm range. Duffel rejects From == To with a 422.
	TimeWindow struct {
		From string `json:"from"`
		To   string `json:"to"`
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
		UserID                   string                    `json:"user_id,omitempty"`
	}

	// OfferRequest is the response from the OfferRequest endpoint, created using the OfferRequestInput.
	OfferRequest struct {
		ID               string                  `json:"id"`
		ClientKey        string                  `json:"client_key"` // For use with https://duffel.com/docs/guides/ancillaries-component
		LiveMode         bool                    `json:"live_mode"`
		CreatedAt        time.Time               `json:"created_at"`
		Slices           []BaseSlice             `json:"slices"`
		Passengers       []OfferRequestPassenger `json:"passengers"`
		CabinClass       CabinClass              `json:"cabin_class"`
		Offers           []Offer                 `json:"offers"`
		AirlineCreditIDs []string                `json:"airline_credit_ids,omitempty"`
	}

	// BatchOfferRequest is polled until RemainingBatches is 0, each poll returning only newly finished offers.
	BatchOfferRequest struct {
		ID               string    `json:"id"`
		ClientKey        string    `json:"client_key"`
		LiveMode         bool      `json:"live_mode"`
		CreatedAt        time.Time `json:"created_at"`
		TotalBatches     int       `json:"total_batches"`
		RemainingBatches int       `json:"remaining_batches"`
		Offers           []Offer   `json:"offers,omitempty"`
	}

	// ItineraryOfferRequest is the view=itineraries shape, with airlines, places and aircraft referenced by ID.
	ItineraryOfferRequest struct {
		ID               string                  `json:"id"`
		ClientKey        string                  `json:"client_key"`
		LiveMode         bool                    `json:"live_mode"`
		CreatedAt        time.Time               `json:"created_at"`
		CabinClass       CabinClass              `json:"cabin_class"`
		AirlineCreditIDs []string                `json:"airline_credit_ids"`
		Passengers       []OfferRequestPassenger `json:"passengers"`
		References       ItineraryReferences     `json:"references"`
		Slices           []ItinerarySlice        `json:"slices"`
		// Batch fields are only set on batch offer requests.
		TotalBatches     *int `json:"total_batches,omitempty"`
		RemainingBatches *int `json:"remaining_batches,omitempty"`
	}

	ItineraryReferences struct {
		Aircraft map[string]Aircraft `json:"aircraft"`
		Airlines map[string]Airline  `json:"airlines"`
		Places   map[string]Location `json:"places"`
	}

	// ItinerarySlice holds place IDs, resolved through ItineraryReferences.Places.
	ItinerarySlice struct {
		Origin      string      `json:"origin"`
		Destination string      `json:"destination"`
		Itineraries []Itinerary `json:"itineraries"`
	}

	Itinerary struct {
		Segments []ItinerarySegment `json:"segments"`
		Brands   []ItineraryBrand   `json:"brands"`
	}

	// ItinerarySegment holds place, airline and aircraft IDs, resolved through ItineraryReferences.
	ItinerarySegment struct {
		Origin                       string          `json:"origin"`
		OriginTerminal               string          `json:"origin_terminal"`
		Destination                  string          `json:"destination"`
		DestinationTerminal          string          `json:"destination_terminal"`
		RawDepartingAt               string          `json:"departing_at"`
		RawArrivingAt                string          `json:"arriving_at"`
		Duration                     Duration        `json:"duration"`
		Distance                     Distance        `json:"distance"`
		Aircraft                     string          `json:"aircraft"`
		MarketingCarrier             string          `json:"marketing_carrier"`
		MarketingCarrierFlightNumber string          `json:"marketing_carrier_flight_number"`
		OperatingCarrier             string          `json:"operating_carrier"`
		OperatingCarrierFlightNumber string          `json:"operating_carrier_flight_number"`
		Stops                        []ItineraryStop `json:"stops"`
	}

	ItineraryStop struct {
		ID          string   `json:"id"`
		Airport     string   `json:"airport"`
		ArrivingAt  string   `json:"arriving_at"`
		DepartingAt string   `json:"departing_at"`
		Duration    Duration `json:"duration"`
	}

	ItineraryBrand struct {
		FareBrandName string                    `json:"fare_brand_name"`
		NGSShelf      *int                      `json:"ngs_shelf,omitempty"`
		Conditions    *Conditions               `json:"conditions,omitempty"`
		Offers        []ItineraryOffer          `json:"offers"`
		Passengers    []ItineraryBrandPassenger `json:"passengers"`
	}

	ItineraryOffer struct {
		ID                                      string                   `json:"id"`
		Type                                    ItineraryOfferType       `json:"type"`
		Owner                                   string                   `json:"owner"`
		ExpiresAt                               *time.Time               `json:"expires_at,omitempty"`
		RawTotalAmount                          string                   `json:"total_amount"`
		RawTotalCurrency                        string                   `json:"total_currency"`
		RawBaseAmount                           string                   `json:"base_amount"`
		RawBaseCurrency                         string                   `json:"base_currency"`
		RawTaxAmount                            string                   `json:"tax_amount"`
		RawTaxCurrency                          string                   `json:"tax_currency"`
		TotalEmissionsKg                        any                      `json:"total_emissions_kg"`
		Conditions                              *Conditions              `json:"conditions,omitempty"`
		PaymentRequirements                     *OfferPaymentRequirement `json:"payment_requirements,omitempty"`
		PassengerIdentityDocumentsRequired      bool                     `json:"passenger_identity_documents_required"`
		SupportedPassengerIdentityDocumentTypes []string                 `json:"supported_passenger_identity_document_types"`
		AvailableAirlineCreditIDs               []string                 `json:"available_airline_credit_ids"`
		// The spec leaves these two untyped.
		SupportedLoyaltyProgrammes json.RawMessage `json:"supported_loyalty_programmes,omitempty"`
		PrivateFares               json.RawMessage `json:"private_fares,omitempty"`
	}

	ItineraryOfferType string

	ItineraryBrandPassenger struct {
		PassengerID string                      `json:"passenger_id"`
		Segments    []ItineraryPassengerSegment `json:"segments"`
	}

	ItineraryPassengerSegment struct {
		Baggages                []Baggage     `json:"baggages"`
		Cabin                   *SegmentCabin `json:"cabin,omitempty"`
		CabinClass              CabinClass    `json:"cabin_class"`
		CabinClassMarketingName string        `json:"cabin_class_marketing_name"`
		FareBasisCode           string        `json:"fare_basis_code"`
	}

	PartialOfferRequestInput struct {
		PartialOfferRequestID string
		SelectedPartialOffers []string
	}
)

func (a *API) CreateOfferRequest(ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption) (*OfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, OfferRequest](a).
		Post("/air/offer_requests", &requestInput).
		WithParams(requestInput).
		WithOptions(opts...).Single(ctx)
}

// CreatePartialOfferRequest sends SupplierTimeout but not ReturnOffers, which the partial endpoint does not take.
func (a *API) CreatePartialOfferRequest(
	ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption,
) (*OfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, OfferRequest](a).
		Post("/air/partial_offer_requests", &requestInput).
		WithParams(supplierTimeout(requestInput.SupplierTimeout)).
		WithOptions(opts...).Single(ctx)
}

func (a *API) GetPartialOfferRequests(ctx context.Context, requestInput PartialOfferRequestInput, opts ...RequestOption) (
	*OfferRequest, error,
) {
	return newRequestWithAPI[PartialOfferRequestInput, OfferRequest](a).
		Getf("/air/partial_offer_requests/%s", requestInput.PartialOfferRequestID).
		WithParams(requestInput).
		WithOptions(opts...).Single(ctx)
}

func (a *API) GetFullPartialOfferRequest(ctx context.Context, requestInput PartialOfferRequestInput, opts ...RequestOption) (
	*OfferRequest, error,
) {
	return newRequestWithAPI[PartialOfferRequestInput, OfferRequest](a).
		Getf("/air/partial_offer_requests/%s/fares", requestInput.PartialOfferRequestID).
		WithParams(requestInput).
		WithOptions(opts...).Single(ctx)
}

func (a *API) GetOfferRequest(ctx context.Context, id string, opts ...RequestOption) (*OfferRequest, error) {
	return newRequestWithAPI[EmptyPayload, OfferRequest](a).Getf("/air/offer_requests/%s", id).WithOptions(opts...).Single(ctx)
}

// CreateOfferRequestItineraries is CreateOfferRequest with view=itineraries.
func (a *API) CreateOfferRequestItineraries(
	ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption,
) (*ItineraryOfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, ItineraryOfferRequest](a).
		Post("/air/offer_requests", &requestInput).
		WithParams(requestInput).
		WithParam("view", "itineraries").
		WithOptions(opts...).Single(ctx)
}

// GetOfferRequestItineraries is GetOfferRequest with view=itineraries.
func (a *API) GetOfferRequestItineraries(ctx context.Context, id string, opts ...RequestOption) (*ItineraryOfferRequest, error) {
	return newRequestWithAPI[EmptyPayload, ItineraryOfferRequest](a).
		Getf("/air/offer_requests/%s", id).
		WithParam("view", "itineraries").
		WithOptions(opts...).Single(ctx)
}

// CreateBatchOfferRequest starts a search that returns offers in batches, see GetBatchOfferRequest.
// It sends SupplierTimeout but not ReturnOffers.
func (a *API) CreateBatchOfferRequest(
	ctx context.Context, requestInput OfferRequestInput, opts ...RequestOption,
) (*BatchOfferRequest, error) {
	return newRequestWithAPI[OfferRequestInput, BatchOfferRequest](a).
		Post("/air/batch_offer_requests", &requestInput).
		WithParams(supplierTimeout(requestInput.SupplierTimeout)).
		WithOptions(opts...).Single(ctx)
}

// GetBatchOfferRequest returns the offers of the batches finished since the previous call.
func (a *API) GetBatchOfferRequest(ctx context.Context, id string, opts ...RequestOption) (*BatchOfferRequest, error) {
	return newRequestWithAPI[EmptyPayload, BatchOfferRequest](a).
		Getf("/air/batch_offer_requests/%s", id).
		WithOptions(opts...).Single(ctx)
}

// GetBatchOfferRequestItineraries is GetBatchOfferRequest with view=itineraries.
func (a *API) GetBatchOfferRequestItineraries(ctx context.Context, id string, opts ...RequestOption) (*ItineraryOfferRequest, error) {
	return newRequestWithAPI[EmptyPayload, ItineraryOfferRequest](a).
		Getf("/air/batch_offer_requests/%s", id).
		WithParam("view", "itineraries").
		WithOptions(opts...).Single(ctx)
}

func (a *API) ListOfferRequests(ctx context.Context, opts ...RequestOption) *Iter[OfferRequest] {
	return newRequestWithAPI[EmptyPayload, OfferRequest](a).Get("/air/offer_requests").WithOptions(opts...).Iter(ctx)
}

func (o *ItineraryOffer) TotalAmount() currency.Amount {
	return amountOrZero(o.RawTotalAmount, o.RawTotalCurrency)
}

// Encode implements the ParamEncoder interface.
func (o OfferRequestInput) Encode(q url.Values) error {
	q.Set("return_offers", strconv.FormatBool(o.ReturnOffers))
	return supplierTimeout(o.SupplierTimeout).Encode(q)
}

const (
	ItineraryOfferTypeSingleTicket ItineraryOfferType = "single_ticket"
	ItineraryOfferTypeSplitTicket  ItineraryOfferType = "split_ticket"
)

// supplierTimeout encodes supplier_timeout alone, for endpoints that do not take return_offers.
type supplierTimeout int

func (s supplierTimeout) Encode(q url.Values) error {
	if s > 0 {
		q.Set("supplier_timeout", strconv.Itoa(int(s)))
	}
	return nil
}

func (o PartialOfferRequestInput) Encode(q url.Values) error {
	q["selected_partial_offer[]"] = o.SelectedPartialOffers
	return nil
}
