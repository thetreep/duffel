package duffel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// WebhookSignatureHeader carries t=<unix ts>,v1=<hex HMAC-SHA256> on every webhook delivery.
const WebhookSignatureHeader = "X-Duffel-Signature"

var ErrInvalidWebhookSignature = errors.New("duffel: invalid webhook signature")

type (
	WebhookClient interface {
		CreateWebhook(ctx context.Context, input CreateWebhookInput, opts ...RequestOption) (*Webhook, error)
		ListWebhooks(ctx context.Context, opts ...RequestOption) *Iter[Webhook]
		UpdateWebhook(ctx context.Context, id string, input UpdateWebhookInput, opts ...RequestOption) (*Webhook, error)
		DeleteWebhook(ctx context.Context, id string, opts ...RequestOption) error
		PingWebhook(ctx context.Context, id string, opts ...RequestOption) error
		ListWebhookEvents(ctx context.Context, params ...ListWebhookEventsParams) *Iter[WebhookEvent]
		GetWebhookEvent(ctx context.Context, id string, opts ...RequestOption) (*WebhookEvent, error)
		RedeliverWebhookEvent(ctx context.Context, id string, opts ...RequestOption) error
		ListWebhookDeliveries(ctx context.Context, params ...ListWebhookDeliveriesParams) *Iter[WebhookDelivery]
		GetWebhookDelivery(ctx context.Context, id string, opts ...RequestOption) (*WebhookDelivery, error)
	}

	Webhook struct {
		ID        string             `json:"id"`
		URL       string             `json:"url"`
		Events    []WebhookEventType `json:"events"`
		Active    bool               `json:"active"`
		LiveMode  bool               `json:"live_mode"`
		CreatedAt DateTime           `json:"created_at"`
		UpdatedAt DateTime           `json:"updated_at"`
		// Secret signs deliveries. Duffel returns it only from CreateWebhook.
		Secret string `json:"secret,omitempty"`
	}

	CreateWebhookInput struct {
		URL    string             `json:"url"`
		Events []WebhookEventType `json:"events"`
	}

	// UpdateWebhookInput sends only the fields that are set.
	UpdateWebhookInput struct {
		URL    string             `json:"url,omitempty"`
		Events []WebhookEventType `json:"events,omitempty"`
		Active *bool              `json:"active,omitempty"`
	}

	// WebhookEvent is both the body Duffel posts to a webhook and the resource of the events endpoints.
	WebhookEvent struct {
		ID         string            `json:"id"`
		Type       WebhookEventType  `json:"type"`
		APIVersion string            `json:"api_version"`
		Data       *WebhookEventData `json:"data"`
		LiveMode   bool              `json:"live_mode"`
		// IdempotencyKey is the ID of the related object, for deduplication.
		IdempotencyKey string   `json:"idempotency_key"`
		CreatedAt      DateTime `json:"created_at"`
		// IdentityOrganisationID is documented on deliveries only, not in the events schema.
		IdentityOrganisationID string `json:"identity_organisation_id,omitempty"`
	}

	WebhookEventData struct {
		Object WebhookEventObject `json:"object"`
	}

	// WebhookEventObject merges the object shapes of all flight events. Which fields are set depends on the event type.
	WebhookEventObject struct {
		ID                       string    `json:"id,omitempty"`
		OfferID                  string    `json:"offer_id,omitempty"`
		OrderID                  string    `json:"order_id,omitempty"`
		PaymentID                string    `json:"payment_id,omitempty"`
		AirlineInitiatedChangeID string    `json:"aic_id,omitempty"`
		OrderCreationAttemptedAt *DateTime `json:"order_creation_attempted_at,omitempty"`
	}

	WebhookDelivery struct {
		ID                 string           `json:"id"`
		EndpointID         string           `json:"endpoint_id"`
		EventID            string           `json:"event_id"`
		Type               WebhookEventType `json:"type"`
		URL                string           `json:"url"`
		ResponseStatusCode int              `json:"response_status_code"`
		ResponseBody       string           `json:"response_body"`
		LiveMode           bool             `json:"live_mode"`
		CreatedAt          DateTime         `json:"created_at"`
	}

	ListWebhookEventsParams struct {
		Type            WebhookEventType
		DeliverySuccess *bool
		CreatedAt       *TimeFilter
	}

	ListWebhookDeliveriesParams struct {
		Type            WebhookEventType
		DeliverySuccess *bool
		EndpointID      string
		CreatedAt       *TimeFilter
	}

	WebhookEventType string
)

const (
	WebhookEventAirlineCreditCreated                WebhookEventType = "air.airline_credit.created"
	WebhookEventAirlineCreditSpent                  WebhookEventType = "air.airline_credit.spent"
	WebhookEventAirlineCreditInvalidated            WebhookEventType = "air.airline_credit.invalidated"
	WebhookEventPaymentFailed                       WebhookEventType = "air.payment.failed"
	WebhookEventPaymentSucceeded                    WebhookEventType = "air.payment.succeeded"
	WebhookEventPaymentCancelled                    WebhookEventType = "air.payment.cancelled"
	WebhookEventPaymentPending                      WebhookEventType = "air.payment.pending"
	WebhookEventOrderAirlineInitiatedChangeDetected WebhookEventType = "order.airline_initiated_change_detected"
	WebhookEventOrderCreated                        WebhookEventType = "order.created"
	WebhookEventOrderCreationFailed                 WebhookEventType = "order.creation_failed"
	WebhookEventOrderChanged                        WebhookEventType = "air.order.changed"
	WebhookEventOrderCancellationCreated            WebhookEventType = "order_cancellation.created"
	WebhookEventOrderCancellationConfirmed          WebhookEventType = "order_cancellation.confirmed"
	WebhookEventCarsBookingCreated                  WebhookEventType = "cars.booking.created"
	WebhookEventCarsBookingCancelled                WebhookEventType = "cars.booking.cancelled"
	WebhookEventStaysBookingCreationFailed          WebhookEventType = "stays.booking_creation_failed"
	WebhookEventStaysBookingCreated                 WebhookEventType = "stays.booking.created"
	WebhookEventStaysBookingCancelled               WebhookEventType = "stays.booking.cancelled"
	WebhookEventAssistantConversationUpdated        WebhookEventType = "assistant.conversation.updated"
	WebhookEventPingTriggered                       WebhookEventType = "ping.triggered"
)

func (a *API) CreateWebhook(ctx context.Context, input CreateWebhookInput, opts ...RequestOption) (*Webhook, error) {
	return newRequestWithAPI[CreateWebhookInput, Webhook](a).
		Post("/air/webhooks", &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) ListWebhooks(ctx context.Context, opts ...RequestOption) *Iter[Webhook] {
	return newRequestWithAPI[EmptyPayload, Webhook](a).
		Get("/air/webhooks").
		WithOptions(opts...).Iter(ctx)
}

func (a *API) UpdateWebhook(ctx context.Context, id string, input UpdateWebhookInput, opts ...RequestOption) (*Webhook, error) {
	return newRequestWithAPI[UpdateWebhookInput, Webhook](a).
		Patch("/air/webhooks/"+id, &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) DeleteWebhook(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Deletef("/air/webhooks/%s", id).
		WithOptions(opts...).Empty(ctx)
}

// PingWebhook makes Duffel send a ping.triggered event to the webhook.
func (a *API) PingWebhook(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Postf("/air/webhooks/%s/actions/ping", id).
		WithOptions(opts...).Empty(ctx)
}

func (a *API) ListWebhookEvents(ctx context.Context, params ...ListWebhookEventsParams) *Iter[WebhookEvent] {
	return newRequestWithAPI[ListWebhookEventsParams, WebhookEvent](a).
		Get("/air/webhooks/events").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

func (a *API) GetWebhookEvent(ctx context.Context, id string, opts ...RequestOption) (*WebhookEvent, error) {
	return newRequestWithAPI[EmptyPayload, WebhookEvent](a).
		Getf("/air/webhooks/events/%s", id).
		WithOptions(opts...).Single(ctx)
}

func (a *API) RedeliverWebhookEvent(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Postf("/air/webhooks/events/%s/actions/redeliver", id).
		WithOptions(opts...).Empty(ctx)
}

func (a *API) ListWebhookDeliveries(ctx context.Context, params ...ListWebhookDeliveriesParams) *Iter[WebhookDelivery] {
	return newRequestWithAPI[ListWebhookDeliveriesParams, WebhookDelivery](a).
		Get("/air/webhooks/deliveries").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

func (a *API) GetWebhookDelivery(ctx context.Context, id string, opts ...RequestOption) (*WebhookDelivery, error) {
	return newRequestWithAPI[EmptyPayload, WebhookDelivery](a).
		Getf("/air/webhooks/deliveries/%s", id).
		WithOptions(opts...).Single(ctx)
}

func (p ListWebhookEventsParams) Encode(q url.Values) error {
	encodeWebhookFilters(q, p.Type, p.DeliverySuccess, p.CreatedAt)
	return nil
}

func (p ListWebhookDeliveriesParams) Encode(q url.Values) error {
	encodeWebhookFilters(q, p.Type, p.DeliverySuccess, p.CreatedAt)
	if p.EndpointID != "" {
		q.Set("endpoint_id", p.EndpointID)
	}
	return nil
}

// encodeWebhookFilters sends created_at as a naive UTC time, like ListOrders and the spec examples.
func encodeWebhookFilters(q url.Values, typ WebhookEventType, success *bool, createdAt *TimeFilter) {
	if typ != "" {
		q.Set("type", string(typ))
	}
	if success != nil {
		q.Set("delivery_success", strconv.FormatBool(*success))
	}
	createdAt.encode(q, "created_at", true)
}

// VerifyWebhookSignature checks the X-Duffel-Signature header of a delivery against the raw request body and the
// secret returned by CreateWebhook. It does not check the timestamp age: reject old deliveries yourself if needed.
func VerifyWebhookSignature(secret string, body []byte, signatureHeader string) error {
	var timestamp string
	var signatures []string
	for _, part := range strings.Split(signatureHeader, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return ErrInvalidWebhookSignature
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	expected := []byte(hex.EncodeToString(mac.Sum(nil)))
	for _, sig := range signatures {
		if hmac.Equal(expected, []byte(strings.ToLower(sig))) {
			return nil
		}
	}
	return ErrInvalidWebhookSignature
}

var _ WebhookClient = (*API)(nil)
