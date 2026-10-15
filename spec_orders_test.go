package duffel

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrderDecodesSpecFields(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-order-full.json")

	order, err := client.GetOrder(context.TODO(), "ord_00009hthhsUZ8W4LxQgkjo")
	require.NoError(t, err)
	a.Equal("/air/orders/ord_00009hthhsUZ8W4LxQgkjo", rec.Path)

	a.Equal([]string{"icu_00009htyDGjIfajdNBZRlw"}, order.Users)
	a.Equal(time.Date(2020, 4, 11, 15, 48, 11, 0, time.UTC), order.VoidWindowEndsAt.UTC())
	require.Len(t, order.BookingReferences, 1)
	a.Equal("RZPNX8", order.BookingReferences[0].BookingReference)
	a.Equal("BA", order.BookingReferences[0].Carrier.IATACode)
	a.Len(order.IntendedPaymentMethods, 2)
	a.Equal("GB", order.TaxBreakdown[0].Code)
	a.NotNil(order.AirlineInitiatedChanges[0].TravelAgentTicket)

	credits := order.Cancellation.AirlineCredits
	require.Len(t, credits, 2)
	a.Equal(AirlineCreditTypeETicket, credits[0].Type)
	a.Equal("90.80 GBP", credits[0].Amount().String())
	a.Nil(credits[0].CreatedAt)
	a.Equal("AA", credits[1].AirlineIATACode)
	a.NotNil(credits[1].CreatedAt)
}

func TestPriceOrder(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-price-order.json")

	order, err := client.PriceOrder(
		context.TODO(), "ord_00009hthhsUZ8W4LxQgkjo", PriceOrderInput{
			IntendedPaymentMethods: []IntendedPaymentMethodInput{
				{Type: IntendedPaymentMethodAirlineCredit, AirlineCreditID: "acd_00009htYpSCXrwaB9DnUm1"},
			},
		},
	)
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.Equal("/air/orders/ord_00009hthhsUZ8W4LxQgkjo/actions/price", rec.Path)
	a.JSONEq(
		`{"data":{"intended_payment_methods":[{"type":"airline_credit","airline_credit_id":"acd_00009htYpSCXrwaB9DnUm1"}]}}`,
		rec.Body,
	)
	a.Len(order.IntendedPaymentMethods, 2)
}

func TestCreateOrderSendsUsersAndThreeDSecureSession(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-order.json")

	_, err := client.CreateOrder(
		context.TODO(), CreateOrderInput{
			Type:           OrderTypeInstant,
			SelectedOffers: []string{"off_00009htyDGjIfajdNBZRlw"},
			Users:          []string{"icu_0000AgZitpOnQtd3NQxjwO"},
			Passengers:     []OrderPassenger{{ID: "pas_00009hj8USM7Ncg31cBCLL", Title: PassengerTitleDr, UserID: "icu_0000AgZitpOnQtd3NQxjwO"}},
			Payments: []PaymentCreateInput{
				{Type: PaymentMethodCard, Amount: "30.20", Currency: "GBP", ThreeDSecureSessionID: "3ds_0000AWr2XsTRIF1Vp34gh5"},
			},
		},
	)
	require.NoError(t, err)
	a.Contains(rec.Body, `"users":["icu_0000AgZitpOnQtd3NQxjwO"]`)
	a.Contains(rec.Body, `"user_id":"icu_0000AgZitpOnQtd3NQxjwO"`)
	a.Contains(rec.Body, `"title":"dr"`)
	a.Contains(rec.Body, `"three_d_secure_session_id":"3ds_0000AWr2XsTRIF1Vp34gh5"`)
}

func TestUpdateOrderUsers(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-update-order.json")

	_, err := client.UpdateOrder(
		context.TODO(), "ord_00009hthhsUZ8W4LxQgkjo", OrderUpdateParams{
			Metadata: map[string]any{}, Users: []string{"icu_0000AgZitpOnQtd3NQxjwO"},
		},
	)
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"metadata":{},"users":["icu_0000AgZitpOnQtd3NQxjwO"]}}`, rec.Body)
}

func TestListAirlineInitiatedChangesTravelAgentTicket(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-airline-initiated-changes.json")

	changes, err := client.ListAirlineInitiatedChanges(
		context.TODO(), ListAirlineInitiatedChangesParams{OrderID: "ord_00009hthhsUZ8W4LxQgkjo"},
	)
	require.NoError(t, err)
	a.Equal("/air/airline_initiated_changes", rec.Path)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", rec.Query.Get("order_id"))
	require.Len(t, changes, 1)
	a.Equal(ActionTakenTypeAccepted, changes[0].ActionTaken)
	a.Equal(AvailableActionTypeAccept, changes[0].AvailableActions[0])
	a.NotNil(changes[0].TravelAgentTicket)
	a.NotEmpty(changes[0].Added)
	a.NotEmpty(changes[0].Removed)
}

func TestConfirmOrderChangeWrapsPayment(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-confirm-order-change.json")

	_, err := client.ConfirmOrderChange(
		context.TODO(), "oce_0000A3tQSmKyqOrcySrGbo", PaymentCreateInput{
			Type: PaymentMethodBalance, Amount: "90.80", Currency: "GBP",
		},
	)
	require.NoError(t, err)
	a.Equal("/air/order_changes/oce_0000A3tQSmKyqOrcySrGbo/actions/confirm", rec.Path)
	a.JSONEq(`{"data":{"payment":{"type":"balance","amount":"90.80","currency":"GBP"}}}`, rec.Body)
}

func TestConfirmOrderChangeWithoutPayment(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-confirm-order-change.json")

	_, err := client.ConfirmOrderChange(context.TODO(), "oce_0000A3tQSmKyqOrcySrGbo", PaymentCreateInput{})
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{}}`, rec.Body)

	_, err = client.ConfirmOrderChange(context.TODO(), "ocr_0000A3tQSmKyqOrcySrGbo", PaymentCreateInput{})
	assert.Error(t, err)
}

func TestGetPayment(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-payment.json")

	payment, err := client.GetPayment(context.TODO(), "pay_00009hthhsUZ8W4LxQgkjo")
	require.NoError(t, err)
	a.Equal("/air/payments/pay_00009hthhsUZ8W4LxQgkjo", rec.Path)
	a.Equal(PaymentStatusSucceeded, payment.Status)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", payment.OrderID)
	a.Empty(payment.FailureReason)
}

func TestListPayments(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-payments.json")

	payments, err := Collect(client.ListPayments(context.TODO(), "ord_00009hthhsUZ8W4LxQgkjo"))
	require.NoError(t, err)
	a.Equal("/air/payments", rec.Path)
	a.Equal("ord_00009hthhsUZ8W4LxQgkjo", rec.Query.Get("order_id"))
	a.Len(payments, 1)

	_, err = Collect(client.ListPayments(context.TODO(), ""))
	a.Error(err)
}

func TestCreatePaymentThreeDSecureSession(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/200-create-payment.json")

	_, err := client.CreatePayment(
		context.TODO(), CreatePaymentRequest{
			OrderID: "ord_00003x8pVDGcS8y2AWCoWv",
			Payment: CreatePayment{
				Amount: "30.20", Currency: "GBP", Type: PaymentTypeCard, ThreeDSecureSessionID: "3ds_0000AWr2XsTRIF1Vp34gh5",
			},
		},
	)
	require.NoError(t, err)
	assert.JSONEq(
		t, `{"data":{"order_id":"ord_00003x8pVDGcS8y2AWCoWv","payment":{"amount":"30.20","currency":"GBP","type":"card",
		"three_d_secure_session_id":"3ds_0000AWr2XsTRIF1Vp34gh5"}}}`, rec.Body,
	)
}
