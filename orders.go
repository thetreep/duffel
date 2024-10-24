// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"net/url"
	"time"

	"github.com/bojanz/currency"
	"github.com/gorilla/schema"
)

const orderIDPrefix = "ord_"

type (
	ListOrdersSort string

	Order struct {
		ID                      string                      `json:"id"`
		LiveMode                bool                        `json:"live_mode"`
		Metadata                Metadata                    `json:"metadata"`
		RawBaseAmount           *string                     `json:"base_amount,omitempty"`
		RawBaseCurrency         *string                     `json:"base_currency,omitempty"`
		BookingReference        string                      `json:"booking_reference"`
		CancelledAt             *time.Time                  `json:"cancelled_at,omitempty"`
		CreatedAt               time.Time                   `json:"created_at"`
		Conditions              Conditions                  `json:"conditions,omitempty"`
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
		Metadata map[string]any
	}

	ListOrdersParams struct {
		// Filters orders by their booking reference.
		// The filter requires an exact match but is case insensitive.
		BookingReference string `url:"booking_reference,omitempty"`

		// Whether to filter orders that are awaiting payment or not.
		// If not specified, all orders regardless of their payment state will be returned.
		AwaitingPayment bool `url:"awaiting_payment,omitempty"`

		// By default, orders aren't returned in any specific order.
		// This parameter allows you to sort the list of orders by the payment_required_by date
		Sort ListOrdersSort `url:"sort,omitempty"`

		// Filters the returned orders by owner.id. Values must be valid airline.ids.
		// Check the Airline schema for details.
		OwnerIDs []string `url:"owner_id,omitempty"`

		// Filters the returned orders by origin. Values must be valid origin identifiers.
		// Check the Order schema for details.
		OriginIDs []string `url:"origin_id,omitempty"`

		// Filters the returned orders by destination. Values must be valid destination identifiers.
		// Check the Order schema for details.
		DestinationIDs []string `url:"destination_id,omitempty"`

		// Filters the returned orders by departure datetime.
		// Orders will be included if any of their segments matches the given criteria
		DepartingAt *TimeFilter `url:"departing_at,omitempty"`

		// Filters the returned orders by arrival datetime.
		// Orders will be included if any of their segments matches the given criteria.
		ArrivingAt *TimeFilter `url:"arriving_at,omitempty"`

		// Filters the returned orders by creation datetime.
		CreatedAt *TimeFilter `url:"created_at,omitempty"`

		// Orders will be included if any of their passengers matches any of the given names.
		// Matches are case-insensitive, and include partial matches.
		PassengerNames []string `url:"passenger_name,omitempty"`
	}

	Metadata map[string]any

	TimeFilter struct {
		Before *time.Time `url:"before,omitempty"`
		After  *time.Time `url:"after,omitempty"`
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
		ActionTaken      ActionTakenType       `json:"actions_taken"`
		ActionTakenAt    time.Time             `json:"actions_taken_at"`
		Added            []Slice               `json:"added"`
		AvailableActions []AvailableActionType `json:"available_actions"`
		CreatedAt        time.Time             `json:"created_at"`
		OrderID          string                `json:"order_id"`
		Removed          []Slice               `json:"removed"`
		// TODO: TravelAgentTicket // preview
		UpdatedAt time.Time `json:"updated_at"`
	}

	ActionTakenType string

	AvailableActionType string

	UpdateAirlineInitiatedChangeInput struct {
		ActionTaken ActionTakenType `json:"action_taken"`
	}

	ListAirlineInitiatedChangesParams struct {
		OrderID string `url:"order_id,omitempty"`
	}

	OrderClient interface {
		// GetOrder Get a single order by ID.
		GetOrder(ctx context.Context, id string) (*Order, error)

		// UpdateOrder Update a single order by ID.
		UpdateOrder(ctx context.Context, id string, params OrderUpdateParams) (*Order, error)

		// ListOrders List orders.
		ListOrders(ctx context.Context, params ...ListOrdersParams) *Iter[Order]

		// CreateOrder Create an order.
		CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, error)

		// ListOrderServices List available services for an order.
		ListOrderServices(ctx context.Context, id string) ([]*AvailableService, error)

		// AddOrderService Add a service to an order.
		AddOrderService(ctx context.Context, id string, input AddOrderServiceInput) (*Order, error)

		// UpdateAirlineInitiatedChange Update an airline-initiated change.
		UpdateAirlineInitiatedChange(ctx context.Context, id string, input UpdateAirlineInitiatedChangeInput) (
			*Order, error,
		)

		// AcceptAirlineInitiatedChange Accept an airline-initiated change.
		AcceptAirlineInitiatedChange(ctx context.Context, id string) (*Order, error)

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

	IssuedDocumentTypeElectronicTicket                 = IssuedDocumentType("electronic_ticket")
	IssuedDocumentTypeElectronicMiscDocumentAssociated = IssuedDocumentType("electronic_miscellaneous_document_associated")
	IssuedDocumentTypeElectronicMiscDocumentStandalone = IssuedDocumentType("electronic_miscellaneous_document_standalone")

	OrderContentManaged     = OrderContent("managed")
	OrderContentSelfManaged = OrderContent("self_managed")
)

// CreateOrder creates a new order.
func (a *API) CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, error) {
	return newRequestWithAPI[CreateOrderInput, Order](a).Post("/air/orders", &input).Single(ctx)
}

// UpdateOrder updates an existing order with update-able fields (mostly metadata).
func (a *API) UpdateOrder(ctx context.Context, id string, params OrderUpdateParams) (*Order, error) {
	return newRequestWithAPI[OrderUpdateParams, Order](a).Patch("/air/orders/"+id, &params).Single(ctx)
}

// GetOrder returns a single order by ID.
func (a *API) GetOrder(ctx context.Context, id string) (*Order, error) {
	return newRequestWithAPI[EmptyPayload, Order](a).Get("/air/orders/" + id).Single(ctx)
}

// ListOrders returns a list of orders.
func (a *API) ListOrders(ctx context.Context, params ...ListOrdersParams) *Iter[Order] {
	return newRequestWithAPI[ListOrdersParams, Order](a).
		Get("/air/orders").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

// ListOrderServices returns a list of available services for an order.
func (a *API) ListOrderServices(ctx context.Context, id string) ([]*AvailableService, error) {
	return newRequestWithAPI[EmptyPayload, AvailableService](a).
		Get("/air/orders/" + id + "/available_services").Slice(ctx)
}

// AddOrderService adds a service to an order.
func (a *API) AddOrderService(ctx context.Context, id string, input AddOrderServiceInput) (*Order, error) {
	return newRequestWithAPI[AddOrderServiceInput, Order](a).
		Post("/air/orders/"+id+"/services", &input).
		Single(ctx)
}

// UpdateAirlineInitiatedChange updates an airline-initiated change.
func (a *API) UpdateAirlineInitiatedChange(
	ctx context.Context, id string, input UpdateAirlineInitiatedChangeInput,
) (*Order, error) {
	return newRequestWithAPI[UpdateAirlineInitiatedChangeInput, Order](a).
		Patch("/air/airline_initiated_changes/"+id, &input).
		Single(ctx)
}

// AcceptAirlineInitiatedChange accepts an airline-initiated change.
func (a *API) AcceptAirlineInitiatedChange(ctx context.Context, id string) (*Order, error) {
	return newRequestWithAPI[EmptyPayload, Order](a).
		Post("/air/airline_initiated_changes/"+id+"/actions/accept", nil).
		Single(ctx)
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

func (o ListOrdersParams) Encode(q url.Values) error {
	enc := schema.NewEncoder()
	enc.SetAliasTag("url")
	return enc.Encode(o, q)
}

func (l ListAirlineInitiatedChangesParams) Encode(q url.Values) error {
	if l.OrderID != "" {
		q.Add("order_id", l.OrderID)
	}

	return nil
}
