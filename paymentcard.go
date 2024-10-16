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
		ID            string           `json:"id"`
		LiveMode      bool             `json:"live_mode"`
		Last4Digits   string           `json:"last_4_digits"`
		MultiUse      bool             `json:"multi_use"`
		Brand         PaymentCardBrand `json:"brand"`
		UnavailableAt DateTime         `json:"unavailable_at,omitempty"`
	}

	PaymentCardBrand string

	PaymentCardClient interface {
		CreatePaymentCardRecord(
			ctx context.Context, payload *CreatePaymentCardRecordRequest,
		) (*PaymentCard, error)
		CreateTemporaryPaymentCardRecordFromSavedPaymentCardRecord(
			ctx context.Context, payload *CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest,
		) (*PaymentCard, error)
		DeleteSavedPaymentCardRecord(ctx context.Context, id string) error
	}
)

const (
	CardBrandVisa            PaymentCardBrand = "visa"
	CardBrandAirplus         PaymentCardBrand = "uatp"
	CardBrandMastercard      PaymentCardBrand = "mastercard"
	CardBrandAmericanExpress PaymentCardBrand = "american_express"
	CardBrandDinersClub      PaymentCardBrand = "diners_club"
	CardBrandJCB             PaymentCardBrand = "jcb"
)

func (a *API) CreatePaymentCardRecord(
	ctx context.Context, payload *CreatePaymentCardRecordRequest,
) (*PaymentCard, error) {
	return newRequestWithAPI[CreatePaymentCardRecordRequest, PaymentCard](a).
		Post("/vault/cards", payload).
		Single(ctx)
}

func (a *API) CreateTemporaryPaymentCardRecordFromSavedPaymentCardRecord(
	ctx context.Context, payload *CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest,
) (*PaymentCard, error) {
	return newRequestWithAPI[CreateTemporaryPaymentCardRecordFromSavedPaymentCardRequest, PaymentCard](a).
		Post("/vault/cards", payload).
		Single(ctx)
}

func (a *API) DeleteSavedPaymentCardRecord(ctx context.Context, id string) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Deletef("/vault/cards/%s", id).
		Empty(ctx)
}

var _ AircraftClient = (*API)(nil)
