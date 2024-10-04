package duffel

import (
	"context"
)

type (
	LoyaltyProgramme struct {
		Alliance       string `json:"alliance"`
		ID             string `json:"id"`
		Name           string `json:"name"`
		LogoURL        string `json:"logo_url"`
		OwnerAirlineID string `json:"owner_airline_id"`
	}

	LoyaltyProgrammeClient interface {
		ListLoyaltyProgramme(ctx context.Context) *Iter[LoyaltyProgramme]
		GetLoyaltyProgramme(ctx context.Context, id string) (*LoyaltyProgramme, error)
	}
)

// ListLoyaltyProgramme retrieves a paginated list of loyalty programmes.
func (a *API) ListLoyaltyProgramme(ctx context.Context) *Iter[LoyaltyProgramme] {
	return newRequestWithAPI[EmptyPayload, LoyaltyProgramme](a).
		Get("/air/loyalty_programmes").
		Iter(ctx)
}

// GetLoyaltyProgramme retrieves a loyalty programme by its ID.
func (a *API) GetLoyaltyProgramme(ctx context.Context, id string) (*LoyaltyProgramme, error) {
	return newRequestWithAPI[EmptyPayload, LoyaltyProgramme](a).
		Getf("/air/loyalty_programmes/%s", id).
		Single(ctx)
}

var _ LoyaltyProgrammeClient = (*API)(nil)
