package duffel

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAirlineCredits(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-airline-credits.json")

	credits, err := Collect(client.ListAirlineCredits(context.TODO(), ListAirlineCreditsParams{UserID: "icu_00009hj8USM7Ncg31cBCLL"}))
	require.NoError(t, err)
	a.Equal("/air/airline_credits", rec.Path)
	a.Equal("icu_00009hj8USM7Ncg31cBCLL", rec.Query.Get("user_id"))
	require.Len(t, credits, 1)
	a.Equal("acd_00009UhD4ongolulWd91Ky", credits[0].ID)
}

func TestGetAirlineCredit(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-airline-credit.json")

	credit, err := client.GetAirlineCredit(context.TODO(), "acd_00009UhD4ongolulWd91Ky")
	require.NoError(t, err)
	a.Equal("/air/airline_credits/acd_00009UhD4ongolulWd91Ky", rec.Path)
	a.Equal(AirlineCreditTypeETicket, credit.Type)
	a.Equal("30.20 GBP", credit.Amount().String())
	a.Equal("10.20 GBP", credit.PenaltyAmount().String())
	a.Equal("12345678", credit.Code)
	a.Equal("icu_00009hj8USM7Ncg31cBCLL", credit.UserID)
	a.NotNil(credit.ExpiresAt)
	a.NotNil(credit.SpentAt)
}

func TestCreateAirlineCredit(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-airline-credit.json")

	credit, err := client.CreateAirlineCredit(
		context.TODO(), CreateAirlineCreditInput{
			AirlineIATACode: "AA",
			Amount:          "30.20",
			AmountCurrency:  "GBP",
			Code:            "12345678",
			Type:            AirlineCreditTypeETicket,
			IssuedOn:        Date(time.Date(2020, 4, 11, 0, 0, 0, 0, time.UTC)),
			ExpiresAt:       time.Date(2021, 4, 11, 15, 48, 11, 0, time.UTC),
			UserID:          "icu_00009hj8USM7Ncg31cBCLL",
		},
	)
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.JSONEq(
		`{"data":{"airline_iata_code":"AA","amount":"30.20","amount_currency":"GBP","code":"12345678","type":"eticket",
		"issued_on":"2020-04-11","expires_at":"2021-04-11T15:48:11Z","user_id":"icu_00009hj8USM7Ncg31cBCLL"}}`, rec.Body,
	)
	a.Equal("acd_00009UhD4ongolulWd91Ky", credit.ID)
}

func TestCreateWebhook(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-webhook.json")

	webhook, err := client.CreateWebhook(
		context.TODO(), CreateWebhookInput{
			URL:    "https://www.example.com:4000/webhooks",
			Events: []WebhookEventType{WebhookEventOrderCreated, WebhookEventOrderAirlineInitiatedChangeDetected},
		},
	)
	require.NoError(t, err)
	a.Equal("/air/webhooks", rec.Path)
	a.JSONEq(
		`{"data":{"url":"https://www.example.com:4000/webhooks","events":["order.created","order.airline_initiated_change_detected"]}}`,
		rec.Body,
	)
	a.Equal("end_0000A3tQSmKyqOrcySrGbo", webhook.ID)
	a.Equal("QKfUULLQh+8SegYmIsF6kA==", webhook.Secret)
	a.Equal(2020, time.Time(webhook.CreatedAt).Year())
}

func TestListWebhooks(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-webhooks.json")

	webhooks, err := Collect(client.ListWebhooks(context.TODO()))
	require.NoError(t, err)
	assert.Equal(t, "/air/webhooks", rec.Path)
	require.Len(t, webhooks, 1)
	assert.Empty(t, webhooks[0].Secret)
}

func TestUpdateWebhook(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-update-webhook.json")

	active := false
	_, err := client.UpdateWebhook(context.TODO(), "end_0000A3tQSmKyqOrcySrGbo", UpdateWebhookInput{Active: &active})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, rec.Method)
	assert.Equal(t, "/air/webhooks/end_0000A3tQSmKyqOrcySrGbo", rec.Path)
	assert.JSONEq(t, `{"data":{"active":false}}`, rec.Body)
}

func TestListWebhookEvents(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-webhook-events.json")

	failed := false
	after := time.Date(2026, 10, 1, 12, 0, 0, 0, time.FixedZone("CEST", 2*3600))
	events, err := Collect(
		client.ListWebhookEvents(
			context.TODO(), ListWebhookEventsParams{
				Type: WebhookEventOrderCreated, DeliverySuccess: &failed, CreatedAt: &TimeFilter{After: &after},
			},
		),
	)
	require.NoError(t, err)
	a.Equal("/air/webhooks/events", rec.Path)
	a.Equal("order.created", rec.Query.Get("type"))
	a.Equal("false", rec.Query.Get("delivery_success"))
	a.Equal("2026-10-01T10:00:00", rec.Query.Get("created_at[after]"))
	require.Len(t, events, 1)
	a.Equal("pay_00009hthhsUZ8W4LxQgkjo", events[0].Data.Object.PaymentID)
}

func TestGetWebhookEvent(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-webhook-event.json")

	event, err := client.GetWebhookEvent(context.TODO(), "wev_0000A4tQSmKyqOrcySrGbo")
	require.NoError(t, err)
	a.Equal("/air/webhooks/events/wev_0000A4tQSmKyqOrcySrGbo", rec.Path)
	a.Equal(WebhookEventOrderCreated, event.Type)
	a.Equal("ord_0000ABd6wggSct7BoraU1o", event.Data.Object.ID)
	a.Equal("off_00009htYpSCXrwaB9DnUm0", event.Data.Object.OfferID)
	a.Equal("org_0000APMFjhs4X2rJ6k7UIE", event.IdentityOrganisationID)
}

func TestListWebhookDeliveries(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-webhook-deliveries.json")

	deliveries, err := Collect(
		client.ListWebhookDeliveries(context.TODO(), ListWebhookDeliveriesParams{EndpointID: "end_0000A3tQSmKyqOrcySrGbo"}),
	)
	require.NoError(t, err)
	assert.Equal(t, "/air/webhooks/deliveries", rec.Path)
	assert.Equal(t, "end_0000A3tQSmKyqOrcySrGbo", rec.Query.Get("endpoint_id"))
	require.Len(t, deliveries, 1)
	assert.Equal(t, 200, deliveries[0].ResponseStatusCode)
}

func TestGetWebhookDelivery(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-webhook-delivery.json")

	delivery, err := client.GetWebhookDelivery(context.TODO(), "del_0000A3tQSmKyqOrcySrGbo")
	require.NoError(t, err)
	assert.Equal(t, "/air/webhooks/deliveries/del_0000A3tQSmKyqOrcySrGbo", rec.Path)
	assert.Equal(t, "wev_0000A3tQSmKyqOrcySrGbo", delivery.EventID)
}

func TestVerifyWebhookSignature(t *testing.T) {
	const secret = "QKfUULLQh+8SegYmIsF6kA=="
	body := []byte(`{"id":"wev_0000A3tQSmKyqOrcySrGbo","type":"ping.triggered"}`)
	// Computed independently with Python hmac, following the Duffel webhook guide.
	const sig = "c84f097277b18c89ce5891f20ca64ff930388476f05b52e2cfc7643d610d6f66"

	assert.NoError(t, VerifyWebhookSignature(secret, body, "t=1616202842,v1="+sig))
	assert.NoError(t, VerifyWebhookSignature(secret, body, "t=1616202842,v1=00,v1="+sig))
	for name, tc := range map[string]struct {
		secret, header string
		body           []byte
	}{
		"tampered body":    {secret, "t=1616202842,v1=" + sig, []byte(`{"id":"wev_other"}`)},
		"other timestamp":  {secret, "t=1616202843,v1=" + sig, body},
		"wrong secret":     {"other", "t=1616202842,v1=" + sig, body},
		"missing v1":       {secret, "t=1616202842", body},
		"missing t":        {secret, "v1=" + sig, body},
		"malformed header": {secret, "garbage", body},
	} {
		assert.ErrorIs(t, VerifyWebhookSignature(tc.secret, tc.body, tc.header), ErrInvalidWebhookSignature, name)
	}
}

func TestWebhookEventDecodesDeliveryBody(t *testing.T) {
	var event WebhookEvent
	body := `{"id":"wev_0000A4tQSmKyqOrcySrGbo","api_version":"v2","type":"order.creation_failed",
	"data":{"object":{"offer_id":"off_00009htYpSCXrwaB9DnUm0","order_creation_attempted_at":"2026-10-15T10:00:00Z"}},
	"live_mode":true,"idempotency_key":"off_00009htYpSCXrwaB9DnUm0","created_at":"2026-10-15T10:00:01.123456Z",
	"identity_organisation_id":"org_0000APMFjhs4X2rJ6k7UIE"}`
	require.NoError(t, json.Unmarshal([]byte(body), &event))
	assert.Equal(t, WebhookEventOrderCreationFailed, event.Type)
	assert.Equal(t, "off_00009htYpSCXrwaB9DnUm0", event.Data.Object.OfferID)
	assert.Equal(t, 15, time.Time(*event.Data.Object.OrderCreationAttemptedAt).Day())
}

func TestBodylessActions(t *testing.T) {
	for name, tc := range map[string]struct {
		status       int
		call         func(Duffel) error
		method, path string
	}{
		"DeleteWebhook": {
			http.StatusNoContent, func(c Duffel) error { return c.DeleteWebhook(context.TODO(), "end_1") },
			http.MethodDelete, "/air/webhooks/end_1",
		},
		"PingWebhook": {
			http.StatusNoContent, func(c Duffel) error { return c.PingWebhook(context.TODO(), "end_1") },
			http.MethodPost, "/air/webhooks/end_1/actions/ping",
		},
		"RedeliverWebhookEvent": {
			http.StatusNoContent, func(c Duffel) error { return c.RedeliverWebhookEvent(context.TODO(), "wev_1") },
			http.MethodPost, "/air/webhooks/events/wev_1/actions/redeliver",
		},
		"DeleteCustomerUserGroup": {
			http.StatusNoContent, func(c Duffel) error { return c.DeleteCustomerUserGroup(context.TODO(), "usg_1") },
			http.MethodDelete, "/identity/customer/user_groups/usg_1",
		},
		"DeleteCard": {
			http.StatusOK, func(c Duffel) error { return c.DeleteCard(context.TODO(), "tcd_1") },
			http.MethodDelete, "/payments/cards/tcd_1",
		},
	} {
		t.Run(name, func(t *testing.T) {
			client, rec := newRecordingClient(t, tc.status, "")
			require.NoError(t, tc.call(client))
			assert.Equal(t, tc.method, rec.Method)
			assert.Equal(t, tc.path, rec.Path)
		})
	}
}
