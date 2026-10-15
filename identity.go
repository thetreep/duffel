package duffel

import (
	"context"
	"net/url"
	"time"
)

type (
	CustomerUserClient interface {
		CreateCustomerUser(ctx context.Context, input CustomerUserInput, opts ...RequestOption) (*CustomerUser, error)
		GetCustomerUser(ctx context.Context, id string, opts ...RequestOption) (*CustomerUser, error)
		UpdateCustomerUser(ctx context.Context, id string, input CustomerUserInput, opts ...RequestOption) (*CustomerUser, error)
		ListCustomerUsers(ctx context.Context, params ...ListCustomerUsersParams) *Iter[CustomerUser]
		CreateCustomerUserGroup(ctx context.Context, input CustomerUserGroupInput, opts ...RequestOption) (*CustomerUserGroup, error)
		GetCustomerUserGroup(ctx context.Context, id string, opts ...RequestOption) (*CustomerUserGroup, error)
		UpdateCustomerUserGroup(ctx context.Context, id string, input CustomerUserGroupInput, opts ...RequestOption) (*CustomerUserGroup, error)
		ListCustomerUserGroups(ctx context.Context, opts ...RequestOption) *Iter[CustomerUserGroup]
		DeleteCustomerUserGroup(ctx context.Context, id string, opts ...RequestOption) error
		CreateComponentClientKey(ctx context.Context, input ComponentClientKeyInput, opts ...RequestOption) (*ComponentClientKey, error)
	}

	CustomerUser struct {
		ID                string             `json:"id"`
		Email             string             `json:"email"`
		GivenName         string             `json:"given_name"`
		FamilyName        string             `json:"family_name"`
		PhoneNumber       string             `json:"phone_number"`
		PreferredLanguage string             `json:"preferred_language"`
		Group             *CustomerUserGroup `json:"group,omitempty"`
		LiveMode          bool               `json:"live_mode"`
		CreatedAt         time.Time          `json:"created_at"`
	}

	// CustomerUserInput is sent whole by UpdateCustomerUser, which is a PUT.
	CustomerUserInput struct {
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		// PhoneNumber is in E.164 format.
		PhoneNumber string `json:"phone_number,omitempty"`
		// PreferredLanguage is an ISO 639-1 code such as "en".
		PreferredLanguage string `json:"preferred_language,omitempty"`
		GroupID           string `json:"group_id,omitempty"`
	}

	ListCustomerUsersParams struct {
		Email string
	}

	// CustomerUserGroup has no UserIDs when nested in CustomerUser.
	CustomerUserGroup struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		UserIDs []string `json:"user_ids,omitempty"`
	}

	// CustomerUserGroupInput requires Name on create. On update, unset fields are left unchanged.
	CustomerUserGroupInput struct {
		Name    string   `json:"name,omitempty"`
		UserIDs []string `json:"user_ids,omitempty"`
	}

	// ComponentClientKeyInput signs UserID alone or with OrderID or BookingID. Empty gives an anonymous key.
	ComponentClientKeyInput struct {
		UserID    string `json:"user_id,omitempty"`
		OrderID   string `json:"order_id,omitempty"`
		BookingID string `json:"booking_id,omitempty"`
	}

	ComponentClientKey struct {
		ComponentClientKey string `json:"component_client_key"`
	}
)

func (a *API) CreateCustomerUser(ctx context.Context, input CustomerUserInput, opts ...RequestOption) (*CustomerUser, error) {
	return newRequestWithAPI[CustomerUserInput, CustomerUser](a).
		Post("/identity/customer/users", &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) GetCustomerUser(ctx context.Context, id string, opts ...RequestOption) (*CustomerUser, error) {
	return newRequestWithAPI[EmptyPayload, CustomerUser](a).
		Getf("/identity/customer/users/%s", id).
		WithOptions(opts...).Single(ctx)
}

func (a *API) UpdateCustomerUser(ctx context.Context, id string, input CustomerUserInput, opts ...RequestOption) (*CustomerUser, error) {
	return newRequestWithAPI[CustomerUserInput, CustomerUser](a).
		Put("/identity/customer/users/"+id, &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) ListCustomerUsers(ctx context.Context, params ...ListCustomerUsersParams) *Iter[CustomerUser] {
	return newRequestWithAPI[ListCustomerUsersParams, CustomerUser](a).
		Get("/identity/customer/users").
		WithParams(normalizeParams(params)...).
		Iter(ctx)
}

func (a *API) CreateCustomerUserGroup(
	ctx context.Context, input CustomerUserGroupInput, opts ...RequestOption,
) (*CustomerUserGroup, error) {
	return newRequestWithAPI[CustomerUserGroupInput, CustomerUserGroup](a).
		Post("/identity/customer/user_groups", &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) GetCustomerUserGroup(ctx context.Context, id string, opts ...RequestOption) (*CustomerUserGroup, error) {
	return newRequestWithAPI[EmptyPayload, CustomerUserGroup](a).
		Getf("/identity/customer/user_groups/%s", id).
		WithOptions(opts...).Single(ctx)
}

func (a *API) UpdateCustomerUserGroup(
	ctx context.Context, id string, input CustomerUserGroupInput, opts ...RequestOption,
) (*CustomerUserGroup, error) {
	return newRequestWithAPI[CustomerUserGroupInput, CustomerUserGroup](a).
		Patch("/identity/customer/user_groups/"+id, &input).
		WithOptions(opts...).Single(ctx)
}

func (a *API) ListCustomerUserGroups(ctx context.Context, opts ...RequestOption) *Iter[CustomerUserGroup] {
	return newRequestWithAPI[EmptyPayload, CustomerUserGroup](a).
		Get("/identity/customer/user_groups").
		WithOptions(opts...).Iter(ctx)
}

func (a *API) DeleteCustomerUserGroup(ctx context.Context, id string, opts ...RequestOption) error {
	return newRequestWithAPI[EmptyPayload, EmptyPayload](a).
		Deletef("/identity/customer/user_groups/%s", id).
		WithOptions(opts...).Empty(ctx)
}

// CreateComponentClientKey creates a key for Duffel UI components.
func (a *API) CreateComponentClientKey(
	ctx context.Context, input ComponentClientKeyInput, opts ...RequestOption,
) (*ComponentClientKey, error) {
	return newRequestWithAPI[ComponentClientKeyInput, ComponentClientKey](a).
		Post("/identity/component_client_keys", &input).
		WithOptions(opts...).Single(ctx)
}

func (p ListCustomerUsersParams) Encode(q url.Values) error {
	if p.Email != "" {
		q.Set("email", p.Email)
	}
	return nil
}

var _ CustomerUserClient = (*API)(nil)
