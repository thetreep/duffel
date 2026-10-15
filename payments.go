// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
)

type (
	PaymentType string

	PaymentStatusType string

	Payment struct {
		Amount    string      `json:"amount"`
		CreatedAt DateTime    `json:"created_at"`
		Currency  string      `json:"currency"`
		ID        string      `json:"id"`
		LiveMode  bool        `json:"live_mode"`
		Type      PaymentType `json:"type"`
		OrderID   string      `json:"order_id,omitempty"`
		// Status moves from pending to failed or succeeded.
		Status          PaymentStatusType `json:"status,omitempty"`
		FailureReason   string            `json:"failure_reason,omitempty"`
		AirlineCreditID string            `json:"airline_credit_id,omitempty"`
	}

	CreatePaymentRequest struct {
		OrderID string        `json:"order_id"`
		Payment CreatePayment `json:"payment"`
	}

	CreatePayment struct {
		Amount                string      `json:"amount"`
		Currency              string      `json:"currency"`
		Type                  PaymentType `json:"type"`
		CardID                string      `json:"card_id,omitempty"`
		ThreeDSecureSessionID string      `json:"three_d_secure_session_id,omitempty"`
	}

	OrderPaymentClient interface {
		CreatePayment(ctx context.Context, req CreatePaymentRequest, opts ...RequestOption) (*Payment, error)
		GetPayment(ctx context.Context, id string, opts ...RequestOption) (*Payment, error)
		ListPayments(ctx context.Context, orderID string, opts ...RequestOption) *Iter[Payment]
	}
)

const (
	PaymentTypeBalance       = PaymentType("balance")
	PaymentTypeCash          = PaymentType("arc_bsp_cash")
	PaymentTypeCard          = PaymentType("card")
	PaymentTypeAirlineCredit = PaymentType("airline_credit")

	PaymentStatusSucceeded = PaymentStatusType("succeeded")
	PaymentStatusFailed    = PaymentStatusType("failed")
	PaymentStatusPending   = PaymentStatusType("pending")
	PaymentStatusCancelled = PaymentStatusType("cancelled")
)

func (a *API) CreatePayment(ctx context.Context, req CreatePaymentRequest, opts ...RequestOption) (*Payment, error) {
	return newRequestWithAPI[CreatePaymentRequest, Payment](a).Post("/air/payments", &req).WithOptions(opts...).Single(ctx)
}

func (a *API) GetPayment(ctx context.Context, id string, opts ...RequestOption) (*Payment, error) {
	return newRequestWithAPI[EmptyPayload, Payment](a).Getf("/air/payments/%s", id).WithOptions(opts...).Single(ctx)
}

// ListPayments lists the payments of an order, which Duffel requires.
func (a *API) ListPayments(ctx context.Context, orderID string, opts ...RequestOption) *Iter[Payment] {
	if err := validateID(orderID, orderIDPrefix); err != nil {
		return ErrIter[Payment](err)
	}
	return newRequestWithAPI[EmptyPayload, Payment](a).
		Get("/air/payments").
		WithParam("order_id", orderID).
		WithOptions(opts...).Iter(ctx)
}

var _ OrderPaymentClient = (*API)(nil)
