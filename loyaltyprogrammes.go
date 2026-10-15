package duffel

import (
	"context"
)

type (
	LoyaltyProgramme struct {
		Alliance string `json:"alliance"`
		ID       string `json:"id"`
		Name     string `json:"name"`
		LogoURL  string `json:"logo_url"`
		// OwnerAirlineID is deprecated by Duffel in favour of OwnerAirline.
		OwnerAirlineID string   `json:"owner_airline_id"`
		OwnerAirline   *Airline `json:"owner_airline,omitempty"`
	}

	LoyaltyProgrammeClient interface {
		ListLoyaltyProgramme(ctx context.Context, opts ...RequestOption) *Iter[LoyaltyProgramme]
		GetLoyaltyProgramme(ctx context.Context, id string, opts ...RequestOption) (*LoyaltyProgramme, error)
	}
)

// ListLoyaltyProgramme retrieves a paginated list of loyalty programmes.
func (a *API) ListLoyaltyProgramme(ctx context.Context, opts ...RequestOption) *Iter[LoyaltyProgramme] {
	return newRequestWithAPI[EmptyPayload, LoyaltyProgramme](a).
		Get("/air/loyalty_programmes").
		WithOptions(opts...).Iter(ctx)
}

// GetLoyaltyProgramme retrieves a loyalty programme by its ID.
func (a *API) GetLoyaltyProgramme(ctx context.Context, id string, opts ...RequestOption) (*LoyaltyProgramme, error) {
	return newRequestWithAPI[EmptyPayload, LoyaltyProgramme](a).
		Getf("/air/loyalty_programmes/%s", id).
		WithOptions(opts...).Single(ctx)
}

var _ LoyaltyProgrammeClient = (*API)(nil)
