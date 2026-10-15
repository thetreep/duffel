package duffel

import (
	"context"
)

type (
	CreatePaymentCardRecordRequest struct {
		AddressCity        string `json:"address_city"`
		AddressCountryCode string `json:"address_country_code"`
		AddressLine1       string `json:"address_line_1"`
		AddressLine2       string `json:"address_line_2"`
		AddressPostalCode  string `json:"address_postal_code"`
		AddressRegion      string `json:"address_region"`
		// Deprecated: Brand is no longer required
		Brand        PaymentCardBrand `json:"brand,omitempty"`
		ExpiryMonth  string           `json:"expiry_month"`
		ExpiryYear   string           `json:"expiry_year"`
		Name         string           `json:"name"`
		Number       string           `json:"number"`
		SecurityCode string           `json:"cvc"`
		// MultiUse controls whether the card should be saved for future use. If false, the card will be saved temporarily.
		MultiUse bool `json:"multi_use"`
	}

	CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest struct {
		CardID       string `json:"card_id"`
		SecurityCode string `json:"cvc"`
	}

	PaymentCard struct {
		UnavailableAt DateTime         `json:"unavailable_at"`
		ID            string           `json:"id"`
		Last4Digits   string           `json:"last_4_digits"`
		Brand         PaymentCardBrand `json:"brand"`
		LiveMode      bool             `json:"live_mode"`
		MultiUse      bool             `json:"multi_use"`
		Name          string           `json:"name,omitempty"`
	}

	PaymentCardBrand string

	// CreateCardInput creates a card on /payments/cards. To get a single-use card from a multi-use one, set CardID
	// and SecurityCode only.
	CreateCardInput struct {
		Name               string `json:"name,omitempty"`
		Number             string `json:"number,omitempty"`
		ExpiryMonth        string `json:"expiry_month,omitempty"`
		ExpiryYear         string `json:"expiry_year,omitempty"`
		SecurityCode       string `json:"cvc,omitempty"`
		AddressLine1       string `json:"address_line_1,omitempty"`
		AddressLine2       string `json:"address_line_2,omitempty"`
		AddressCity        string `json:"address_city,omitempty"`
		AddressRegion      string `json:"address_region,omitempty"`
		AddressPostalCode  string `json:"address_postal_code,omitempty"`
		AddressCountryCode string `json:"address_country_code,omitempty"`
		MultiUse           bool   `json:"multi_use,omitempty"`
		CardID             string `json:"card_id,omitempty"`
	}

	ThreeDSecureSessionInput struct {
		CardID string `json:"card_id"`
		// ResourceID is the offer, order or booking ID that will be paid.
		ResourceID string               `json:"resource_id"`
		Services   []ServiceCreateInput `json:"services,omitempty"`
		Exception  string               `json:"exception,omitempty"`
	}

	ThreeDSecureSession struct {
		ID         string                    `json:"id"`
		CardID     string                    `json:"card_id"`
		ResourceID string                    `json:"resource_id"`
		Status     ThreeDSecureSessionStatus `json:"status"`
		// ClientID initialises the 3DS UI component when Status is challenge_required.
		ClientID string `json:"client_id"`
		LiveMode bool   `json:"live_mode"`
	}

	ThreeDSecureSessionStatus string

	PaymentCardClient interface {
		CreatePaymentCardRecord(
			ctx context.Context, payload *CreatePaymentCardRecordRequest,
			opts ...RequestOption,
		) (*PaymentCard, error)
		CreateTemporaryPaymentCardRecordFromSavedPaymentCardRecord(
			ctx context.Context, payload *CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest,
			opts ...RequestOption,
		) (*PaymentCard, error)
		DeleteSavedPaymentCardRecord(ctx context.Context, id string, opts ...RequestOption) error
		CreateCard(ctx context.Context, input CreateCardInput, opts ...RequestOption) (*PaymentCard, error)
		DeleteCard(ctx context.Context, id string, opts ...RequestOption) error
		CreateThreeDSecureSession(ctx context.Context, input ThreeDSecureSessionInput, opts ...RequestOption) (*ThreeDSecureSession, error)
	}
)

const (
	CardBrandVisa            PaymentCardBrand = "visa"
	CardBrandAirplus         PaymentCardBrand = "uatp"
	CardBrandMastercard      PaymentCardBrand = "mastercard"
	CardBrandAmericanExpress PaymentCardBrand = "american_express"
	CardBrandDinersClub      PaymentCardBrand = "diners_club"
	CardBrandJCB             PaymentCardBrand = "jcb"
	CardBrandDiscover        PaymentCardBrand = "discover"

	ThreeDSecureSessionStatusChallengeRequired ThreeDSecureSessionStatus = "challenge_required"
	ThreeDSecureSessionStatusReadyForPayment   ThreeDSecureSessionStatus = "ready_for_payment"
	ThreeDSecureSessionStatusFailed            ThreeDSecureSessionStatus = "failed"
	ThreeDSecureSessionStatusExpired           ThreeDSecureSessionStatus = "expired"

	// ThreeDSecureExceptionSecureCorporatePayment opts a corporate card out of issuer authentication.
	ThreeDSecureExceptionSecureCorporatePayment = "secure_corporate_payment"
)

func (a *API) CreatePaymentCardRecord(
	ctx context.Context, payload *CreatePaymentCardRecordRequest,
	opts ...RequestOption,
) (*PaymentCard, error) {
	return newRequestWithAPI[CreatePaymentCardRecordRequest, PaymentCard](a).
		Post("/vault/cards", payload).
		WithOptions(opts...).Single(ctx)
}

func (a *API) CreateTemporaryPaymentCardRecordFromSavedPaymentCardRecord(
	ctx context.Context, payload *CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest,
	opts ...RequestOption,
) (*PaymentCard, error) {
	return newRequestWithAPI[CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest, PaymentCard](a).
		Post("/vault/cards", payload).
		WithOptions(opts...).Single(ctx)
}

func (a *API) DeleteSavedPaymentCardRecord(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Deletef("/vault/cards/%s", id).
		WithOptions(opts...).Empty(ctx)
}

// CreateCard uses /payments/cards, which replaces /vault/cards in the current spec.
func (a *API) CreateCard(ctx context.Context, input CreateCardInput, opts ...RequestOption) (*PaymentCard, error) {
	return newRequestWithAPI[CreateCardInput, PaymentCard](a).
		Post("/payments/cards", &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) DeleteCard(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Deletef("/payments/cards/%s", id).
		WithOptions(opts...).Empty(ctx)
}

// CreateThreeDSecureSession authenticates a card before paying, see PaymentCreateInput.ThreeDSecureSessionID.
func (a *API) CreateThreeDSecureSession(
	ctx context.Context, input ThreeDSecureSessionInput, opts ...RequestOption,
) (*ThreeDSecureSession, error) {
	return newRequestWithAPI[ThreeDSecureSessionInput, ThreeDSecureSession](a).
		Post("/payments/three_d_secure_sessions", &input).
		WithOptions(opts...).Single(ctx)
}

var _ PaymentCardClient = (*API)(nil)
