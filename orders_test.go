// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"
	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalOrderWithCurrency(t *testing.T) {
	a := assert.New(t)
	order := new(Order)

	err := json.Unmarshal(
		[]byte(`{"base_amount": "120.00", "base_currency": "USD", "cancelled_at":"2020-04-11T15:48:11.642Z"}`), order,
	)
	a.NoError(err)

	a.Equal("120.00 USD", order.BaseAmount().String())
	a.Equal("2020-04-11 15:48:11.642 +0000 UTC", order.CancelledAt.String())
}

func TestCreateOrder(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	// gock.Observe(gock.DumpRequest)

	gock.New("https://api.duffel.com").
		Post("/air/orders").
		File("fixtures/201-create-order-input.json").
		Reply(201).
		SetHeader(RequestIDHeader, "FvxRwfnMtKgc0EwCCoXE").
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/201-create-order.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.CreateOrder(
		ctx, CreateOrderInput{
			Type: OrderTypeInstant,
			Metadata: Metadata{
				"seat_preference": "isle", "meal_preference": "NLML", "payment_intent_id": "pit_00009htYpSCXrwaB9DnUm2",
			},
			Services: []ServiceCreateInput{
				{
					ID:       "ase_00009hj8USM7Ncg31cB123",
					Quantity: 1,
				},
			},
			SelectedOffers: []string{"off_00009htyDGjIfajdNBZRlw"},
			Payments: []PaymentCreateInput{
				{
					Type:     "balance",
					Currency: "GBP",
					Amount:   "30.20",
				},
			},
			Passengers: []OrderPassenger{
				{
					Type:              PassengerTypeAdult,
					ID:                "pas_00009hj8USM7Ncg31cBCLL",
					Title:             PassengerTitleMrs,
					FamilyName:        "Earhart",
					GivenName:         "Amelia",
					BornOn:            Date(time.Date(1987, time.July, 24, 0, 0, 0, 0, time.UTC)),
					Gender:            GenderFemale,
					InfantPassengerID: "pas_00009hj8USM8Ncg32aTGHL",
					PhoneNumber:       "+442080160509",
					Email:             "amelia@duffel.com",
					IdentityDocuments: []IdentityDocument{
						{
							UniqueIdentifier:   "19KL56147",
							ExpiresOn:          Date(time.Date(2025, time.April, 25, 0, 0, 0, 0, time.UTC)),
							IssuingCountryCode: "GB",
							Type:               "passport",
						},
					},
				},
			},
		},
	)
	a.NoError(err)
	reqID, _ := client.LastRequestID()
	a.Equal("FvxRwfnMtKgc0EwCCoXE", reqID)
	a.Equal("RZPNX8", order.BookingReference)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)
	a.Equal("100.00 GBP", order.Conditions.ChangeBeforeDeparture.PenaltyAmount().String())
	a.Equal("100.00 GBP", order.Conditions.RefundBeforeDeparture.PenaltyAmount().String())
	a.Len(order.Slices, 1)
	a.Len(order.Slices[0].Segments, 1)
	a.Len(order.Slices[0].Segments[0].Passengers, 1)
	a.Equal("passenger_0", order.Slices[0].Segments[0].Passengers[0].ID)
	a.Equal(CabinClassEconomy, order.Slices[0].Segments[0].Passengers[0].CabinClass)
	a.Equal("Economy Basic", order.Slices[0].Segments[0].Passengers[0].CabinClassMarketingName)
	a.Equal("14B", order.Slices[0].Segments[0].Passengers[0].Seat.Designator)
	a.Equal("Exit row seat", order.Slices[0].Segments[0].Passengers[0].Seat.Name)
	a.Equal(
		[]string{
			"Do not seat children in exit row seats", "Do not seat passengers with special needs in exit row seats",
		}, order.Slices[0].Segments[0].Passengers[0].Seat.Disclosures,
	)
}

func TestCreateOrderAwaitingCreation(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Post("/air/orders").
		Reply(200).
		SetHeader(RequestIDHeader, "AwaitingReq123").
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-create-order-awaiting.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.CreateOrder(
		ctx, CreateOrderInput{
			Type:           OrderTypeInstant,
			SelectedOffers: []string{"off_123"},
			Passengers:     []OrderPassenger{},
		},
	)
	a.ErrorIs(err, ErrOrderAwaitingCreation)
	a.True(order.IsAwaitingCreation)
	a.False(order.IsPendingConfirmation)
	a.Empty(order.ID)
	a.NotNil(order.Message)
	a.Equal("The booking has been confirmed. It will appear in the system soon.", *order.Message)
}

func TestCreateOrderPendingConfirmation(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Post("/air/orders").
		Reply(202).
		SetHeader(RequestIDHeader, "PendingReq456").
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/202-create-order-pending.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.CreateOrder(
		ctx, CreateOrderInput{
			Type:           OrderTypeInstant,
			SelectedOffers: []string{"off_456"},
			Passengers:     []OrderPassenger{},
		},
	)
	a.NoError(err)
	a.True(order.IsPendingConfirmation)
	a.False(order.IsAwaitingCreation)
	a.NotNil(order.Message)
}

func TestCreateOrderNormal(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Post("/air/orders").
		Reply(201).
		SetHeader(RequestIDHeader, "NormalReq789").
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/201-create-order.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.CreateOrder(
		ctx, CreateOrderInput{
			Type:           OrderTypeInstant,
			SelectedOffers: []string{"off_789"},
			Passengers:     []OrderPassenger{},
		},
	)
	a.NoError(err)
	a.False(order.IsAwaitingCreation)
	a.False(order.IsPendingConfirmation)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)
}

func TestListOrders(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Get("/air/orders").
		MatchParam("after", "g2wAAAACbQAAABBBZXJvbWlzdC1LaGFya2l2bQAAAB=").
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-list-orders-page2.json")

	gock.New("https://api.duffel.com").
		Get("/air/orders").
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-list-orders.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	iter := client.ListOrders(
		ctx, ListOrdersParams{
			BookingReference: "RZPNX8",
		},
	)

	iter.Next()
	a.NoError(iter.Err())
	order1 := iter.Current()

	a.Equal("RZPNX8", order1.BookingReference)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order1.ID)

	a.Equal(&ListMeta{After: "g2wAAAACbQAAABBBZXJvbWlzdC1LaGFya2l2bQAAAB=", Limit: 50}, iter.Meta())

	iter.Next()
	a.NoError(iter.Err())
	order2 := iter.Current()

	a.Len(iter.List().GetItems(), 1, "iterator has 1 item on this page")
	a.Equal(&ListMeta{Limit: 50}, iter.Meta())

	a.Equal("ABC123", order2.BookingReference)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order2.ID)
}

func TestGetOrderByID(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Get("/air/orders/ord_00009hthhsUZ8W4LxQgkjo").
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-get-order.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.GetOrder(ctx, "ord_00009hthhsUZ8W4LxQgkjo")
	a.NoError(err)

	a.Equal("RZPNX8", order.BookingReference)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)
}

func TestUpdateOrder(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)

	gock.New("https://api.duffel.com").
		Patch("/air/orders/ord_00009hthhsUZ8W4LxQgkjo").
		MatchType("json").
		JSON(
			Payload[OrderUpdateParams]{
				Data: OrderUpdateParams{
					Metadata: map[string]any{"seat_preference": "window"},
				},
			},
		).
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-update-order.json")

	ctx := context.TODO()
	client := New("duffel_test_123")
	order, err := client.UpdateOrder(
		ctx, "ord_00009hthhsUZ8W4LxQgkjo", OrderUpdateParams{
			Metadata: map[string]any{
				"seat_preference": "window",
			},
		},
	)
	a.NoError(err)

	a.Equal("RZPNX8", order.BookingReference)
	a.Equal(Metadata{"seat_preference": "window"}, order.Metadata)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", order.ID)
}

func TestAirlineInitiatedChangeActionTaken(t *testing.T) {
	a := assert.New(t)
	var aic AirlineInitiatedChanges
	a.NoError(json.Unmarshal([]byte(`{"action_taken":"accepted","action_taken_at":"2026-01-01T10:00:00Z","available_actions":["update"]}`), &aic))
	a.Equal(ActionTakenTypeAccepted, aic.ActionTaken)
	a.Equal(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), aic.ActionTakenAt)
	a.Equal([]AvailableActionType{AvailableActionTypeUpdate}, aic.AvailableActions)
}

func TestOrderAvailableActions(t *testing.T) {
	a := assert.New(t)
	var order Order
	a.NoError(json.Unmarshal([]byte(`{"available_actions":["cancel","change"]}`), &order))
	a.Equal([]AvailableActionType{AvailableActionTypeCancel, AvailableActionTypeChange}, order.AvailableActions)
}

func TestOrderUpdateParamsMetadataTag(t *testing.T) {
	b, err := json.Marshal(OrderUpdateParams{Metadata: map[string]any{"k": "v"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"metadata":{"k":"v"}}`, string(b))
}

func TestListOrdersParamsEncode(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	assert.NoError(t, err)
	after := time.Date(2026, 10, 14, 10, 0, 0, 0, paris)
	before := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	beforeFrac := before.Add(500 * time.Millisecond)

	tests := []struct {
		name   string
		params ListOrdersParams
		want   string
	}{
		{"empty", ListOrdersParams{}, ""},
		{"created_at local time", ListOrdersParams{CreatedAt: &TimeFilter{After: &after}}, "created_at[after]=2026-10-14T08:00:00"},
		{"departing_at both bounds", ListOrdersParams{DepartingAt: &TimeFilter{After: &after, Before: &before}}, "departing_at[after]=2026-10-14T10:00:00&departing_at[before]=2026-10-14T09:00:00"},
		{"arriving_at", ListOrdersParams{ArrivingAt: &TimeFilter{Before: &beforeFrac}}, "arriving_at[before]=2026-10-14T09:00:00.5"},
		{"passenger names", ListOrdersParams{PassengerNames: []string{"a", "b"}}, "passenger_name[]=a&passenger_name[]=b"},
		{"owner ids", ListOrdersParams{OwnerIDs: []string{"arl_1"}}, "owner_id[]=arl_1"},
		{"origin ids", ListOrdersParams{OriginIDs: []string{"arp_1"}}, "origin_id[]=arp_1"},
		{"destination ids", ListOrdersParams{DestinationIDs: []string{"arp_2"}}, "destination_id[]=arp_2"},
		{"scalars", ListOrdersParams{BookingReference: "RZPNX8", AwaitingPayment: true, Sort: ListOrdersSortPaymentRequiredByDesc}, "awaiting_payment=true&booking_reference=RZPNX8&sort=-payment_required_by"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := url.Values{}
			assert.NoError(t, tt.params.Encode(q))
			got, err := url.QueryUnescape(q.Encode())
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
