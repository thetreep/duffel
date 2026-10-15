package duffel

import (
	"context"
	"net/url"
	"time"

	"github.com/bojanz/currency"
)

type (
	AirlineCreditClient interface {
		ListAirlineCredits(ctx context.Context, params ...ListAirlineCreditsParams) *Iter[AirlineCredit]
		GetAirlineCredit(ctx context.Context, id string, opts ...RequestOption) (*AirlineCredit, error)
		CreateAirlineCredit(ctx context.Context, input CreateAirlineCreditInput, opts ...RequestOption) (*AirlineCredit, error)
	}

	ListAirlineCreditsParams struct {
		UserID string
	}

	// CreateAirlineCreditInput records a credit issued outside Duffel. Set UserID, or GivenName and FamilyName.
	CreateAirlineCreditInput struct {
		AirlineIATACode string            `json:"airline_iata_code"`
		Amount          string            `json:"amount"`
		AmountCurrency  string            `json:"amount_currency"`
		Code            string            `json:"code"`
		Type            AirlineCreditType `json:"type"`
		IssuedOn        Date              `json:"issued_on"`
		ExpiresAt       time.Time         `json:"expires_at"`
		UserID          string            `json:"user_id,omitempty"`
		GivenName       string            `json:"given_name,omitempty"`
		FamilyName      string            `json:"family_name,omitempty"`
	}
)

const (
	AirlineCreditTypeETicket AirlineCreditType = "eticket"
	AirlineCreditTypeMCO     AirlineCreditType = "mco"
	AirlineCreditTypeEMD     AirlineCreditType = "emd"
)

func (a *API) ListAirlineCredits(ctx context.Context, params ...ListAirlineCreditsParams) *Iter[AirlineCredit] {
	return newRequestWithAPI[ListAirlineCreditsParams, AirlineCredit](a).
		Get("/air/airline_credits").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

func (a *API) GetAirlineCredit(ctx context.Context, id string, opts ...RequestOption) (*AirlineCredit, error) {
	return newRequestWithAPI[EmptyPayload, AirlineCredit](a).
		Getf("/air/airline_credits/%s", id).
		WithOptions(opts...).Single(ctx)
}

func (a *API) CreateAirlineCredit(ctx context.Context, input CreateAirlineCreditInput, opts ...RequestOption) (*AirlineCredit, error) {
	return newRequestWithAPI[CreateAirlineCreditInput, AirlineCredit](a).
		Post("/air/airline_credits", &input).
		WithOptions(opts...).Single(ctx)
}

func (p ListAirlineCreditsParams) Encode(q url.Values) error {
	if p.UserID != "" {
		q.Set("user_id", p.UserID)
	}
	return nil
}

// Amount falls back to the older credit_amount and credit_currency fields.
func (c *AirlineCredit) Amount() currency.Amount {
	value, code := c.RawAmount, c.RawAmountCurrency
	if value == "" {
		value, code = c.RawCreditAmount, c.RawCreditCurrency
	}
	return amountOrZero(value, code)
}

func (c *AirlineCredit) PenaltyAmount() *currency.Amount {
	return optionalAmount(c.RawPenaltyAmount, c.RawPenaltyCurrency)
}

var _ AirlineCreditClient = (*API)(nil)
