// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/bojanz/currency"
)

const orderIDPrefix = "ord_"

type (
	ListOrdersSort string

	Order struct {
		ID                      string                      `json:"id"`
		LiveMode                bool                        `json:"live_mode"`
		Message                 *string                     `json:"message,omitempty"`
		Metadata                Metadata                    `json:"metadata"`
		RawBaseAmount           *string                     `json:"base_amount,omitempty"`
		RawBaseCurrency         *string                     `json:"base_currency,omitempty"`
		BookingReference        string                      `json:"booking_reference"`
		CancelledAt             *time.Time                  `json:"cancelled_at,omitempty"`
		CreatedAt               time.Time                   `json:"created_at"`
		Conditions              Conditions                  `json:"conditions"`
		Documents               []IssuedDocument            `json:"documents,omitempty"`
		Owner                   Airline                     `json:"owner"`
		Passengers              []OrderPassenger            `json:"passengers,omitempty"`
		PaymentStatus           PaymentStatus               `json:"payment_status"`
		Services                []Service                   `json:"services,omitempty"`
		Slices                  []Slice                     `json:"slices,omitempty"`
		SyncedAt                time.Time                   `json:"synced_at"`
		RawTaxAmount            *string                     `json:"tax_amount,omitempty"`
		RawTaxCurrency          *string                     `json:"tax_currency,omitempty"`
		RawTotalAmount          string                      `json:"total_amount"`
		RawTotalCurrency        string                      `json:"total_currency"`
		AirlineInitiatedChanges []AirlineInitiatedChanges   `json:"airline_initiated_changes"`
		Cancellation            *OrderCancellation          `json:"cancellation,omitempty"`
		Changes                 []PassengerInitiatedChanges `json:"changes"`
		Content                 OrderContent                `json:"content"`
		OfferID                 string                      `json:"offer_id"`
		Type                    OrderType                   `json:"type"`
		// AvailableActions lists what can be done on the order: cancel, change or update.
		AvailableActions      []AvailableActionType `json:"available_actions,omitempty"`
		IsPendingConfirmation bool                  `json:"-"`
		IsAwaitingCreation    bool                  `json:"-"`
		// TODO: Users // preview - slice of string ids representing users allowed to manage this order
	}

	SliceConditions struct {
		ChangeBeforeDeparture *ChangeCondition `json:"change_before_departure,omitempty"`
	}

	Conditions struct {
		RefundBeforeDeparture *ChangeCondition `json:"refund_before_departure,omitempty"`
		ChangeBeforeDeparture *ChangeCondition `json:"change_before_departure,omitempty"`
	}

	ChangeCondition struct {
		Allowed            bool    `json:"allowed"`
		RawPenaltyAmount   *string `json:"penalty_amount,omitempty"`
		RawPenaltyCurrency *string `json:"penalty_currency,omitempty"`
	}

	IssuedDocument struct {
		PassengerIDs     []string           `json:"passenger_ids"`
		Type             IssuedDocumentType `json:"type"`
		UniqueIdentifier string             `json:"unique_identifier"`
	}

	// NOTE: If you receive a 500 Internal Server Error when trying to create an order,
	// it may have still been created on the airline’s side.
	// Please contact Duffel support before trying the request again.

	OrderType string

	IssuedDocumentType string

	OrderContent string

	CreateOrderInput struct {
		Type OrderType `json:"type"`

		// Metadata contains a set of key-value pairs that you can attach to an object.
		// It can be useful for storing additional information about the object, in a
		// structured format. Duffel does not use this information.
		//
		// You should not store sensitive information in this field.
		Metadata Metadata `json:"metadata,omitempty"`

		// Passengers The personal details of the passengers, expanding on
		// the information initially provided when creating the offer request.
		Passengers []OrderPassenger `json:"passengers"`

		Payments []PaymentCreateInput `json:"payments,omitempty"`

		// SelectedOffers The ids of the offers you want to book. You must specify an array containing exactly one selected offer.
		SelectedOffers []string `json:"selected_offers"`

		Services []ServiceCreateInput `json:"services,omitempty"`
	}

	AddOrderServiceInput struct {
		// AddServices The services you want to add to the order.
		AddServices []ServiceCreateInput `json:"add_services"`

		// Payment The payment details to pay for the services.
		Payment PaymentCreateInput `json:"payment"`
	}

	// ServiceCreateInput The services you want to book along with the first selected offer.
	// This key should be omitted when the order’s type is hold, as we do not support services for hold orders yet.
	ServiceCreateInput struct {
		// ID The id of the service from the offer's available_services that you want to book
		ID string `json:"id"`

		// Quantity The quantity of the service to book. This will always be 1 for seat services.
		Quantity int `json:"quantity"`
	}

	Service struct {
		// Duffel's unique identifier for the booked service
		ID string `json:"id"`

		// Metadata The metadata varies by the type of service.
		// It includes further data about the service. For example, for
		// baggages, it may have data about size and weight restrictions.
		Metadata Metadata `json:"metadata"`

		// List of passenger ids the service applies to.
		// The service applies to all the passengers in this list.
		PassengerIDs []string `json:"passenger_ids"`

		// Quantity The quantity of the service that was booked
		Quantity int `json:"quantity"`

		// List of segment ids the service applies to. The service applies to all the segments in this list.
		SegmentIDs []string `json:"segment_ids"`

		// RawTotalAmount The total price of the service for all passengers and segments it applies to,
		// accounting for quantity and including taxes
		RawTotalAmount   string `json:"total_amount,omitempty"`
		RawTotalCurrency string `json:"total_currency,omitempty"`

		// Type Possible values: "baggage" or "seat"
		Type string `json:"type"`
	}

	// OrderUpdateParams is used as the input to UpdateOrder.
	// Only certain order fields are updateable.
	// Each field that can be updated is detailed in the `OrderUpdateParams` object.
	OrderUpdateParams struct {
		// Metadata replaces the order metadata. An empty non-nil map clears it, a nil map is sent as null.
		Metadata map[string]any `json:"metadata"`
	}

	ListOrdersParams struct {
		// Filters orders by their booking reference.
		// The filter requires an exact match but is case insensitive.
		BookingReference string

		// Whether to filter orders that are awaiting payment or not.
		// If not specified, all orders regardless of their payment state will be returned.
		AwaitingPayment bool

		// By default, orders aren't returned in any specific order.
		// This parameter allows you to sort the list of orders by the payment_required_by date
		Sort ListOrdersSort

		// Filters the returned orders by owner.id. Values must be valid airline.ids.
		// Check the Airline schema for details.
		OwnerIDs []string

		// Filters the returned orders by origin. Values must be valid origin identifiers.
		// Check the Order schema for details.
		OriginIDs []string

		// Filters the returned orders by destination. Values must be valid destination identifiers.
		// Check the Order schema for details.
		DestinationIDs []string

		// Filters the returned orders by departure datetime.
		// Orders will be included if any of their segments matches the given criteria.
		// Segment times are local: the wall clock is sent and the zone ignored.
		DepartingAt *TimeFilter

		// Filters the returned orders by arrival datetime.
		// Orders will be included if any of their segments matches the given criteria.
		// Segment times are local: the wall clock is sent and the zone ignored.
		ArrivingAt *TimeFilter

		// Filters the returned orders by creation datetime. Any zone works, the value is sent in UTC.
		CreatedAt *TimeFilter

		// Orders will be included if any of their passengers matches any of the given names.
		// Matches are case-insensitive, and include partial matches.
		PassengerNames []string
	}

	Metadata map[string]any

	TimeFilter struct {
		Before *time.Time
		After  *time.Time
	}

	PassengerInitiatedChanges struct {
		ID                     string              `json:"id"`
		RawChangeTotalAmount   string              `json:"change_total_amount"`
		RawChangeTotalCurrency string              `json:"change_total_currency"`
		ConfirmedAt            time.Time           `json:"confirmed_at"`
		CreatedAt              time.Time           `json:"created_at"`
		ExpiresAt              time.Time           `json:"expires_at"`
		LiveMode               bool                `json:"live_mode"`
		RawNewTotalAmount      string              `json:"new_total_amount"`
		RawNewTotalCurrency    string              `json:"new_total_currency"`
		OrderID                string              `json:"order_id"`
		RawPenaltyAmount       string              `json:"penalty_total_amount"`
		RawPenaltyCurrency     string              `json:"penalty_total_currency"`
		RefundTo               RefundPaymentMethod `json:"refund_to"`
		Slices                 struct {
			Add    []Slice `json:"add"`
			Remove []Slice `json:"remove"`
		} `json:"slices"`
	}

	AirlineInitiatedChanges struct {
		ID               string                `json:"id"`
		ActionTaken      ActionTakenType       `json:"action_taken"`
		ActionTakenAt    time.Time             `json:"action_taken_at"`
		Added            []Slice               `json:"added"`
		AvailableActions []AvailableActionType `json:"available_actions"`
		CreatedAt        time.Time             `json:"created_at"`
		OrderID          string                `json:"order_id"`
		Removed          []Slice               `json:"removed"`
		// TODO: TravelAgentTicket // preview
		UpdatedAt time.Time `json:"updated_at"`
	}

	ActionTakenType string

	// AvailableActionType is shared by orders (cancel, change, update) and airline-initiated changes (adds accept).
	AvailableActionType string

	UpdateAirlineInitiatedChangeInput struct {
		ActionTaken ActionTakenType `json:"action_taken"`
	}

	ListAirlineInitiatedChangesParams struct {
		OrderID string `url:"order_id,omitempty"`
	}

	OrderClient interface {
		// GetOrder Get a single order by ID.
		GetOrder(ctx context.Context, id string, opts ...RequestOption) (*Order, error)

		// UpdateOrder Update a single order by ID.
		UpdateOrder(ctx context.Context, id string, params OrderUpdateParams, opts ...RequestOption) (*Order, error)

		// ListOrders List orders.
		ListOrders(ctx context.Context, params ...ListOrdersParams) *Iter[Order]

		// CreateOrder Create an order.
		CreateOrder(ctx context.Context, input CreateOrderInput, opts ...RequestOption) (*Order, error)

		// ListOrderServices List available services for an order.
		ListOrderServices(ctx context.Context, id string, opts ...RequestOption) ([]*AvailableService, error)

		// AddOrderService Add a service to an order.
		AddOrderService(ctx context.Context, id string, input AddOrderServiceInput, opts ...RequestOption) (
			*Order, error,
		)

		// UpdateAirlineInitiatedChange Update an airline-initiated change.
		UpdateAirlineInitiatedChange(
			ctx context.Context, id string, input UpdateAirlineInitiatedChangeInput, opts ...RequestOption,
		) (
			*Order, error,
		)

		// AcceptAirlineInitiatedChange Accept an airline-initiated change.
		AcceptAirlineInitiatedChange(ctx context.Context, id string, opts ...RequestOption) (*Order, error)

		// ListAirlineInitiatedChanges List airline-initiated changes.
		ListAirlineInitiatedChanges(
			ctx context.Context, params ...ListAirlineInitiatedChangesParams,
		) ([]*AirlineInitiatedChanges, error)
	}
)

const (
	ListOrdersSortPaymentRequiredByAsc  ListOrdersSort = "payment_required_by"
	ListOrdersSortPaymentRequiredByDesc ListOrdersSort = "-payment_required_by"

	OrderTypeHold    OrderType = "hold"
	OrderTypeInstant OrderType = "instant"

	ActionTakenTypeAccepted  = ActionTakenType("accepted")
	ActionTakenTypeCancelled = ActionTakenType("cancelled")
	ActionTakenTypeChanged   = ActionTakenType("changed")

	AvailableActionTypeAccept = AvailableActionType("accept")
	AvailableActionTypeCancel = AvailableActionType("cancel")
	AvailableActionTypeChange = AvailableActionType("change")
	AvailableActionTypeUpdate = AvailableActionType("update")

	IssuedDocumentTypeElectronicTicket                 = IssuedDocumentType("electronic_ticket")
	IssuedDocumentTypeElectronicMiscDocumentAssociated = IssuedDocumentType("electronic_miscellaneous_document_associated")
	IssuedDocumentTypeElectronicMiscDocumentStandalone = IssuedDocumentType("electronic_miscellaneous_document_standalone")

	OrderContentManaged     = OrderContent("managed")
	OrderContentSelfManaged = OrderContent("self_managed")
)

// CreateOrder creates a new order. Outcomes:
//   - 201: the order is created. err is nil.
//   - 202: the order exists but the airline has not confirmed it. err is nil and IsPendingConfirmation is set.
//   - 200: the booking is accepted but the order does not exist yet (card payment flow). The returned Order has
//     only Message set and IsAwaitingCreation true, and err is ErrOrderAwaitingCreation. Do not retry.
//   - *DecodeError: Duffel answered 2xx but the body was unreadable. The order may exist and be paid. Do not retry.
func (a *API) CreateOrder(ctx context.Context, input CreateOrderInput, opts ...RequestOption) (*Order, error) {
	order, statusCode, err := newRequestWithAPI[CreateOrderInput, Order](a).
		Post("/air/orders", &input).
		WithOptions(opts...).SingleWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusAccepted:
		order.IsPendingConfirmation = true
	case http.StatusOK:
		order.IsAwaitingCreation = true
		return order, ErrOrderAwaitingCreation
	}
	return order, nil
}

// UpdateOrder updates an existing order with update-able fields (mostly metadata).
func (a *API) UpdateOrder(ctx context.Context, id string, params OrderUpdateParams, opts ...RequestOption) (
	*Order, error,
) {
	return newRequestWithAPI[OrderUpdateParams, Order](a).Patch(
		"/air/orders/"+id, &params,
	).WithOptions(opts...).Single(ctx)
}

// GetOrder returns a single order by ID.
func (a *API) GetOrder(ctx context.Context, id string, opts ...RequestOption) (*Order, error) {
	return newRequestWithAPI[EmptyPayload, Order](a).Get("/air/orders/" + id).WithOptions(opts...).Single(ctx)
}

// ListOrders returns a list of orders.
func (a *API) ListOrders(ctx context.Context, params ...ListOrdersParams) *Iter[Order] {
	return newRequestWithAPI[ListOrdersParams, Order](a).
		Get("/air/orders").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

// ListOrderServices returns a list of available services for an order.
func (a *API) ListOrderServices(ctx context.Context, id string, opts ...RequestOption) ([]*AvailableService, error) {
	return newRequestWithAPI[EmptyPayload, AvailableService](a).
		Get("/air/orders/" + id + "/available_services").WithOptions(opts...).Slice(ctx)
}

// AddOrderService adds a service to an order.
func (a *API) AddOrderService(
	ctx context.Context, id string, input AddOrderServiceInput, opts ...RequestOption,
) (*Order, error) {
	return newRequestWithAPI[AddOrderServiceInput, Order](a).
		Post("/air/orders/"+id+"/services", &input).
		WithOptions(opts...).Single(ctx)
}

// UpdateAirlineInitiatedChange updates an airline-initiated change.
func (a *API) UpdateAirlineInitiatedChange(
	ctx context.Context, id string, input UpdateAirlineInitiatedChangeInput,
	opts ...RequestOption,
) (*Order, error) {
	return newRequestWithAPI[UpdateAirlineInitiatedChangeInput, Order](a).
		Patch("/air/airline_initiated_changes/"+id, &input).
		WithOptions(opts...).Single(ctx)
}

// AcceptAirlineInitiatedChange accepts an airline-initiated change.
func (a *API) AcceptAirlineInitiatedChange(ctx context.Context, id string, opts ...RequestOption) (*Order, error) {
	return newRequestWithAPI[EmptyPayload, Order](a).
		Post("/air/airline_initiated_changes/"+id+"/actions/accept", nil).
		WithOptions(opts...).Single(ctx)
}

// ListAirlineInitiatedChanges returns a list of airline-initiated changes.
func (a *API) ListAirlineInitiatedChanges(
	ctx context.Context, params ...ListAirlineInitiatedChangesParams,
) ([]*AirlineInitiatedChanges, error) {
	return newRequestWithAPI[ListAirlineInitiatedChangesParams, AirlineInitiatedChanges](a).
		Get("/air/airline_initiated_changes").
		WithParams(normalizeParams(params)...).Slice(ctx)
}

func (o *Order) BaseAmount() *currency.Amount {
	if o.RawBaseAmount != nil && o.RawBaseCurrency != nil {
		amount, err := currency.NewAmount(*o.RawBaseAmount, *o.RawBaseCurrency)
		if err != nil {
			return nil
		}
		return &amount
	}
	return nil
}

func (o *Order) TaxAmount() *currency.Amount {
	if o.RawTaxAmount != nil && o.RawTaxCurrency != nil {
		amount, err := currency.NewAmount(*o.RawTaxAmount, *o.RawTaxCurrency)
		if err != nil {
			return nil
		}
		return &amount
	}
	return nil
}

func (o *Order) TotalAmount() currency.Amount {
	amount, err := currency.NewAmount(o.RawTotalAmount, o.RawTotalCurrency)
	if err != nil {
		return currency.Amount{}
	}
	return amount
}

func (c *ChangeCondition) PenaltyAmount() *currency.Amount {
	if c.RawPenaltyAmount != nil && c.RawPenaltyCurrency != nil {
		amount, err := currency.NewAmount(*c.RawPenaltyAmount, *c.RawPenaltyCurrency)
		if err != nil {
			return nil
		}
		return &amount
	}

	return nil
}

func (s *Service) TotalAmount() currency.Amount {
	amount, err := currency.NewAmount(s.RawTotalAmount, s.RawTotalCurrency)
	if err != nil {
		return currency.Amount{}
	}
	return amount
}

// Encode writes the filters by hand: gorilla/schema cannot express created_at[after] style keys.
func (o ListOrdersParams) Encode(q url.Values) error {
	if o.BookingReference != "" {
		q.Set("booking_reference", o.BookingReference)
	}
	if o.AwaitingPayment {
		q.Set("awaiting_payment", "true")
	}
	if o.Sort != "" {
		q.Set("sort", string(o.Sort))
	}
	for key, values := range map[string][]string{
		"owner_id[]":       o.OwnerIDs,
		"origin_id[]":      o.OriginIDs,
		"destination_id[]": o.DestinationIDs,
		"passenger_name[]": o.PassengerNames,
	} {
		for _, v := range values {
			q.Add(key, v)
		}
	}
	o.DepartingAt.encode(q, "departing_at", false)
	o.ArrivingAt.encode(q, "arriving_at", false)
	o.CreatedAt.encode(q, "created_at", true)
	return nil
}

// Duffel rejects "Z" and offsets in these filters (checked on the sandbox).
const naiveDateTime = "2006-01-02T15:04:05.999999"

func (f *TimeFilter) encode(q url.Values, key string, toUTC bool) {
	if f == nil {
		return
	}
	format := func(t time.Time) string {
		if toUTC {
			t = t.UTC()
		}
		return t.Format(naiveDateTime)
	}
	if f.Before != nil {
		q.Set(key+"[before]", format(*f.Before))
	}
	if f.After != nil {
		q.Set(key+"[after]", format(*f.After))
	}
}

func (l ListAirlineInitiatedChangesParams) Encode(q url.Values) error {
	if l.OrderID != "" {
		q.Add("order_id", l.OrderID)
	}

	return nil
}
