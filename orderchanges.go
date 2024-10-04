// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Order change flow:
// 1. Get an existing order by ID using client.GetOrder(...)
// 2. Create a new order change request using client.CreateOrderChangeRequest(...)
// 3. Get the order change offer using client.CreatePendingOrderChange(...)
package duffel

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bojanz/currency"
)

const (
	orderChangeRequestIDPrefix = "ocr_"
	orderChangeOfferIDPrefix   = "oco_"
	orderChangeIDPrefix        = "oce_"
)

type (

	// OrderChangeRequest is the input to the OrderChange API.
	// To change an order, you'll need to create an order change request.
	// An order change request describes the slices of an existing paid order
	// that you want to remove and search criteria for new slices you want to add.
	OrderChangeRequest struct {
		ID                string             `json:"id"`
		OrderID           string             `json:"order_id"`
		Slices            SliceChange        `json:"slices"`
		OrderChangeOffers []OrderChangeOffer `json:"order_change_offers"`
		CreatedAt         string             `json:"created_at"`
		UpdatedAt         string             `json:"updated_at"`
		LiveMode          bool               `json:"live_mode"`
	}

	OrderChangeOffer struct {
		ID string `json:"id"`
		// OrderChangeID is the ID for an order change if one has already been created from this order change offer
		OrderChangeID           string         `json:"order_change_id"`
		Slices                  SliceChangeset `json:"slices"`
		RefundTo                PaymentMethod  `json:"refund_to"`
		RawPenaltyTotalCurrency string         `json:"penalty_total_currency"`
		RawPenaltyTotalAmount   string         `json:"penalty_total_amount"`
		RawNewTotalCurrency     string         `json:"new_total_currency"`
		RawNewTotalAmount       string         `json:"new_total_amount"`
		RawChangeTotalCurrency  string         `json:"change_total_currency"`
		RawChangeTotalAmount    string         `json:"change_total_amount"`
		ExpiresAt               DateTime       `json:"expires_at"`
		CreatedAt               DateTime       `json:"created_at"`
		UpdatedAt               DateTime       `json:"updated_at"`
		LiveMode                bool           `json:"live_mode"`
		Conditions              Conditions     `json:"conditions"`
		PrivateFares            []PrivateFare  `json:"private_fares"`
	}

	OrderChange struct {
		ID                      string         `json:"id"`
		OrderID                 string         `json:"order_id"`
		Slices                  SliceChangeset `json:"slices"`
		RefundTo                PaymentMethod  `json:"refund_to"`
		RawPenaltyTotalCurrency string         `json:"penalty_total_currency"`
		RawPenaltyTotalAmount   string         `json:"penalty_total_amount"`
		RawNewTotalCurrency     string         `json:"new_total_currency"`
		RawNewTotalAmount       string         `json:"new_total_amount"`
		RawChangeTotalCurrency  string         `json:"change_total_currency"`
		RawChangeTotalAmount    string         `json:"change_total_amount"`
		ExpiresAt               string         `json:"expires_at"`
		CreatedAt               DateTime       `json:"created_at"`
		UpdatedAt               string         `json:"updated_at"`
		LiveMode                bool           `json:"live_mode"`
		ConfirmedAt             DateTime       `json:"confirmed_at"`
	}

	SliceChangeset struct {
		Add    []Slice `json:"add"`
		Remove []Slice `json:"remove"`
	}

	OrderChangeRequestParams struct {
		OrderID      string                   `json:"order_id"`
		PrivateFares map[string][]PrivateFare `json:"private_fares,omitempty"`
		Slices       SliceChange              `json:"slices,omitempty"`
	}

	SliceAdd struct {
		DepartureDate Date       `json:"departure_date"`
		Destination   string     `json:"destination"`
		Origin        string     `json:"origin"`
		CabinClass    CabinClass `json:"cabin_class"`
	}

	SliceRemove struct {
		SliceID string `json:"slice_id"`
	}

	SliceChange struct {
		Add    []SliceAdd    `json:"add,omitempty"`
		Remove []SliceRemove `json:"remove,omitempty"`
	}

	ListOrderChangeOffersParams struct {
		OrderChangeRequestID string                         `url:"order_change_request_id,omitempty"`
		Sort                 ListOrderChangeOffersSortParam `url:"sort,omitempty"`
		MaxConnections       int                            `url:"max_connections,omitempty"`
	}

	ListOrderChangeOffersSortParam string

	OrderChangeClient interface {
		CreateOrderChangeRequest(ctx context.Context, params OrderChangeRequestParams) (*OrderChangeRequest, error)
		GetOrderChangeRequest(ctx context.Context, id string) (*OrderChangeRequest, error)
		CreatePendingOrderChange(ctx context.Context, orderChangeRequestID string) (*OrderChange, error)
		ConfirmOrderChange(ctx context.Context, id string, payment PaymentCreateInput) (*OrderChange, error)
		GetOrderChange(ctx context.Context, id string) (*OrderChange, error)
		GetOrderChangeOffer(ctx context.Context, id string) (*OrderChangeOffer, error)
		ListOrderChangeOffers(ctx context.Context, params ...ListOrderChangeOffersParams) *Iter[OrderChangeOffer]
	}
)

const (
	SortParamChangeTotalAmount ListOrderChangeOffersSortParam = "change_total_amount"
	SortParamTotalDuration     ListOrderChangeOffersSortParam = "total_duration"
)

func (a *API) CreateOrderChangeRequest(ctx context.Context, params OrderChangeRequestParams) (
	*OrderChangeRequest, error,
) {
	return newRequestWithAPI[OrderChangeRequestParams, OrderChangeRequest](a).
		Post("/air/order_change_requests", &params).
		Single(ctx)
}

// GetOrderChangeRequest retrieves an order change request by its ID.
func (a *API) GetOrderChangeRequest(ctx context.Context, orderChangeRequestID string) (*OrderChangeRequest, error) {
	if err := validateID(orderChangeRequestID, orderChangeRequestIDPrefix); err != nil {
		return nil, err
	}

	return newRequestWithAPI[EmptyPayload, OrderChangeRequest](a).
		Getf("/air/order_change_requests/%s", orderChangeRequestID).
		Single(ctx)
}

// CreatePendingOrderChange creates a new pending order change.
func (a *API) CreatePendingOrderChange(ctx context.Context, offerID string) (*OrderChange, error) {
	if err := validateID(offerID, orderChangeOfferIDPrefix); err != nil {
		return nil, err
	}

	return newRequestWithAPI[map[string]string, OrderChange](a).
		Postf("/air/order_changes").
		Body(&map[string]string{"selected_order_change_offer": offerID}).
		Single(ctx)
}

// ConfirmOrderChange confirms a pending order change.
func (a *API) ConfirmOrderChange(
	ctx context.Context, orderChangeRequestID string, payment PaymentCreateInput,
) (*OrderChange, error) {
	if err := validateID(orderChangeRequestID, orderChangeRequestIDPrefix); err != nil {
		return nil, err
	}

	return newRequestWithAPI[PaymentCreateInput, OrderChange](a).
		Postf("/air/order_changes/%s/actions/confirm", orderChangeRequestID).
		Body(&payment).
		Single(ctx)
}

// GetOrderChange retrieves an order change by its ID.
func (a *API) GetOrderChange(ctx context.Context, id string) (*OrderChange, error) {
	if err := validateID(id, orderChangeIDPrefix); err != nil {
		return nil, err
	}

	return newRequestWithAPI[EmptyPayload, OrderChange](a).
		Getf("/air/order_changes/%s", id).
		Single(ctx)
}

// GetOrderChangeOffer retrieves an order change offer by its ID.
func (a *API) GetOrderChangeOffer(ctx context.Context, id string) (*OrderChangeOffer, error) {
	if err := validateID(id, orderChangeOfferIDPrefix); err != nil {
		return nil, err
	}

	return newRequestWithAPI[EmptyPayload, OrderChangeOffer](a).
		Getf("/air/order_change_offers/%s", id).
		Single(ctx)
}

// ListOrderChangeOffers retrieves a paginated list of order change offers.
func (a *API) ListOrderChangeOffers(
	ctx context.Context, params ...ListOrderChangeOffersParams,
) *Iter[OrderChangeOffer] {
	return newRequestWithAPI[ListOrderChangeOffersParams, OrderChangeOffer](a).
		Get("/air/order_change_offers").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

var _ OrderChangeClient = (*API)(nil)

func validateID(id, prefix string) error {
	if id == "" {
		return fmt.Errorf("id param is required")
	} else if !strings.HasPrefix(id, prefix) {
		return fmt.Errorf("id should begin with %s", prefix)
	}

	return nil
}

func (o *OrderChangeOffer) ChangeTotalAmount() currency.Amount {
	amount, err := currency.NewAmount(o.RawChangeTotalAmount, o.RawChangeTotalCurrency)
	if err != nil {
		return currency.Amount{}
	}

	return amount
}

func (o *OrderChangeOffer) NewTotalAmount() currency.Amount {
	amount, err := currency.NewAmount(o.RawNewTotalAmount, o.RawNewTotalCurrency)
	if err != nil {
		return currency.Amount{}
	}

	return amount
}

// PenaltyTotalAmount returns the penalty imposed by the airline for making this change.
func (o *OrderChangeOffer) PenaltyTotalAmount() currency.Amount {
	amount, err := currency.NewAmount(o.RawPenaltyTotalAmount, o.RawPenaltyTotalCurrency)
	if err != nil {
		return currency.Amount{}
	}

	return amount
}

func (l ListOrderChangeOffersParams) Encode(v url.Values) error {
	if l.OrderChangeRequestID != "" {
		v.Set("order_change_request_id", l.OrderChangeRequestID)
	}

	if l.Sort != "" {
		v.Set("sort", string(l.Sort))
	}

	if l.MaxConnections != 0 {
		v.Set("max_connections", strconv.Itoa(l.MaxConnections))
	}

	return nil
}
