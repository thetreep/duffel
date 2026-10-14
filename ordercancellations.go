// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bojanz/currency"
)

const orderCancellationIDPrefix = "ore_"

type (
	// OrderCancellation times are RFC 3339 strings, parsed by the *AtTime accessors.
	OrderCancellation struct {
		ID                string          `json:"id"`
		OrderID           string          `json:"order_id"`
		RefundTo          PaymentMethod   `json:"refund_to"`
		RawRefundCurrency string          `json:"refund_currency"`
		RawRefundAmount   string          `json:"refund_amount"`
		ExpiresAt         string          `json:"expires_at"`
		CreatedAt         string          `json:"created_at"`
		ConfirmedAt       string          `json:"confirmed_at"`
		LiveMode          bool            `json:"live_mode"`
		AirlineCredits    []AirlineCredit `json:"airline_credits"`
	}

	AirlineCredit struct {
		RawCreditAmount   string `json:"credit_amount"`
		RawCreditCurrency string `json:"credit_currency"`
		CreditCode        string `json:"credit_code"`
		CreditName        string `json:"credit_name"`
		ID                string `json:"id"`
		IssuedOn          Date   `json:"issued_on"`
		PassengerID       string `json:"passenger_id"`
	}

	// OrderCancellationRequest is response from the OrderCancellation API.
	//
	// Once you've created a pending order cancellation, you'll know
	// the `refund_amount` you're due to get back.
	//
	// To actually cancel the order, you'll need to confirm the cancellation.
	// The booking with the airline will be cancelled, and the `refund_amount` will be
	// returned to the original payment method (i.e. your Duffel balance).
	// You'll then need to refund your customer (e.g. back to their credit/debit card).
	OrderCancellationRequest struct {
		OrderID string `json:"order_id"`
	}

	ListOrderCancellationParams struct {
		OrderID string `url:"order_id,omitempty"`
	}

	OrderCancellationClient interface {
		CreateOrderCancellation(ctx context.Context, orderID string, opts ...RequestOption) (*OrderCancellation, error)
		ConfirmOrderCancellation(ctx context.Context, orderCancellationID string, opts ...RequestOption) (*OrderCancellation, error)
		GetOrderCancellation(ctx context.Context, orderCancellationID string, opts ...RequestOption) (*OrderCancellation, error)
		ListOrderCancellations(ctx context.Context, params ...ListOrderCancellationParams) *Iter[OrderCancellation]
	}
)

// CreateOrderCancellation creates a new pending order cancellation.
func (a *API) CreateOrderCancellation(ctx context.Context, orderID string, opts ...RequestOption) (*OrderCancellation, error) {
	return newRequestWithAPI[OrderCancellationRequest, OrderCancellation](a).
		Post(
			"/air/order_cancellations", &OrderCancellationRequest{
				OrderID: orderID,
			},
		).
		WithOptions(opts...).Single(ctx)
}

// ConfirmOrderCancellation confirms a pending order cancellation.
func (a *API) ConfirmOrderCancellation(ctx context.Context, orderCancellationID string, opts ...RequestOption) (*OrderCancellation, error) {
	if !strings.HasPrefix(orderCancellationID, orderCancellationIDPrefix) {
		return nil, fmt.Errorf(
			"orderCancellationID should have prefix %s, got %q", orderCancellationIDPrefix, orderCancellationID,
		)
	}

	return newRequestWithAPI[EmptyPayload, OrderCancellation](a).
		Post(fmt.Sprintf("/air/order_cancellations/%s/actions/confirm", orderCancellationID), nil).
		WithOptions(opts...).Single(ctx)
}

// GetOrderCancellation retrieves an order cancellation by its ID.
func (a *API) GetOrderCancellation(ctx context.Context, orderCancellationID string, opts ...RequestOption) (*OrderCancellation, error) {
	if !strings.HasPrefix(orderCancellationID, orderCancellationIDPrefix) {
		return nil, fmt.Errorf(
			"orderCancellationID should have prefix %s, got %q", orderCancellationIDPrefix, orderCancellationID,
		)
	}

	return newRequestWithAPI[EmptyPayload, OrderCancellation](a).
		Getf("/air/order_cancellations/%s", orderCancellationID).
		WithOptions(opts...).Single(ctx)
}

// ListOrderCancellations Retrieves a paginated list of order cancellations.
func (a *API) ListOrderCancellations(
	ctx context.Context, params ...ListOrderCancellationParams,
) *Iter[OrderCancellation] {
	return newRequestWithAPI[ListOrderCancellationParams, OrderCancellation](a).
		Get("/air/order_cancellations").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

func (o *OrderCancellation) RefundAmount() currency.Amount {
	amount, err := currency.NewAmount(o.RawRefundAmount, o.RawRefundCurrency)
	if err != nil {
		return currency.Amount{}
	}
	return amount
}

func (o *OrderCancellation) ExpiresAtTime() (time.Time, bool) { return parseRFC3339(o.ExpiresAt) }

func (o *OrderCancellation) CreatedAtTime() (time.Time, bool) { return parseRFC3339(o.CreatedAt) }

// ConfirmedAtTime returns false while the cancellation is pending (confirmed_at is null).
func (o *OrderCancellation) ConfirmedAtTime() (time.Time, bool) { return parseRFC3339(o.ConfirmedAt) }

func parseRFC3339(v string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, v)
	return t, err == nil
}

func (l ListOrderCancellationParams) Encode(v url.Values) error {
	if l.OrderID != "" {
		v.Set("order_id", l.OrderID)
	}

	return nil
}

var _ OrderCancellationClient = (*API)(nil)
