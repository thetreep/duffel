package duffel

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var customerUserInput = CustomerUserInput{
	Email:             "amelia@duffel.com",
	GivenName:         "Amelia",
	FamilyName:        "Earhart",
	PhoneNumber:       "+442080160509",
	PreferredLanguage: "en",
	GroupID:           "usg_0000AgZitpOnQtd3NQxjwO",
}

const customerUserInputJSON = `{"data":{"email":"amelia@duffel.com","given_name":"Amelia","family_name":"Earhart",
"phone_number":"+442080160509","preferred_language":"en","group_id":"usg_0000AgZitpOnQtd3NQxjwO"}}`

func TestCreateCustomerUser(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-customer-user.json")

	user, err := client.CreateCustomerUser(context.TODO(), customerUserInput)
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.Equal("/identity/customer/users", rec.Path)
	a.JSONEq(customerUserInputJSON, rec.Body)
	a.Equal("icu_0000AgZitpOnQtd3NQxjwO", user.ID)
	a.Equal("Northwind Solutions", user.Group.Name)
}

func TestGetCustomerUser(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-customer-user.json")

	user, err := client.GetCustomerUser(context.TODO(), "icu_0000AgZitpOnQtd3NQxjwO")
	require.NoError(t, err)
	assert.Equal(t, "/identity/customer/users/icu_0000AgZitpOnQtd3NQxjwO", rec.Path)
	assert.Equal(t, "amelia@duffel.com", user.Email)
}

func TestUpdateCustomerUserIsPut(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-customer-user.json")

	_, err := client.UpdateCustomerUser(context.TODO(), "icu_0000AgZitpOnQtd3NQxjwO", customerUserInput)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, rec.Method)
	assert.Equal(t, "/identity/customer/users/icu_0000AgZitpOnQtd3NQxjwO", rec.Path)
	assert.JSONEq(t, customerUserInputJSON, rec.Body)
}

func TestListCustomerUsers(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-customer-users.json")

	users, err := Collect(client.ListCustomerUsers(context.TODO(), ListCustomerUsersParams{Email: "amelia@duffel.com"}))
	require.NoError(t, err)
	assert.Equal(t, "/identity/customer/users", rec.Path)
	assert.Equal(t, "amelia@duffel.com", rec.Query.Get("email"))
	assert.Len(t, users, 1)
}

func TestCreateCustomerUserGroup(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-customer-user-group.json")

	group, err := client.CreateCustomerUserGroup(
		context.TODO(), CustomerUserGroupInput{Name: "Northwind Solutions", UserIDs: []string{"icu_0000AgZitpOnQtd3NQxjwX"}},
	)
	require.NoError(t, err)
	assert.Equal(t, "/identity/customer/user_groups", rec.Path)
	assert.JSONEq(t, `{"data":{"name":"Northwind Solutions","user_ids":["icu_0000AgZitpOnQtd3NQxjwX"]}}`, rec.Body)
	assert.Equal(t, []string{"icu_0000AgZitpOnQtd3NQxjwX"}, group.UserIDs)
}

func TestGetCustomerUserGroup(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-customer-user-group.json")

	group, err := client.GetCustomerUserGroup(context.TODO(), "usg_0000AgZitpOnQtd3NQxjwO")
	require.NoError(t, err)
	assert.Equal(t, "/identity/customer/user_groups/usg_0000AgZitpOnQtd3NQxjwO", rec.Path)
	assert.Equal(t, "Northwind Solutions", group.Name)
}

func TestUpdateCustomerUserGroup(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-customer-user-group.json")

	_, err := client.UpdateCustomerUserGroup(context.TODO(), "usg_0000AgZitpOnQtd3NQxjwO", CustomerUserGroupInput{Name: "Renamed"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, rec.Method)
	assert.JSONEq(t, `{"data":{"name":"Renamed"}}`, rec.Body)
}

func TestListCustomerUserGroups(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-list-customer-user-groups.json")

	groups, err := Collect(client.ListCustomerUserGroups(context.TODO()))
	require.NoError(t, err)
	assert.Equal(t, "/identity/customer/user_groups", rec.Path)
	assert.Len(t, groups, 1)
}

func TestCreateComponentClientKey(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-create-component-client-key.json")

	key, err := client.CreateComponentClientKey(
		context.TODO(), ComponentClientKeyInput{UserID: "icu_0000AgZitpOnQtd3NQxjwO", OrderID: "ord_0000ABdZnggSct7BoraU1o"},
	)
	require.NoError(t, err)
	assert.Equal(t, "/identity/component_client_keys", rec.Path)
	assert.JSONEq(t, `{"data":{"user_id":"icu_0000AgZitpOnQtd3NQxjwO","order_id":"ord_0000ABdZnggSct7BoraU1o"}}`, rec.Body)
	assert.NotEmpty(t, key.ComponentClientKey)
}

func TestCreateCard(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-card.json")

	card, err := client.CreateCard(context.TODO(), CreateCardInput{CardID: "tcd_00009hthhsUZ8W4LxQgkja", SecurityCode: "123"})
	require.NoError(t, err)
	a.Equal("/payments/cards", rec.Path)
	a.JSONEq(`{"data":{"card_id":"tcd_00009hthhsUZ8W4LxQgkja","cvc":"123"}}`, rec.Body)
	a.Equal("John Doe", card.Name)
	a.Equal(CardBrandMastercard, card.Brand)
}

func TestCreateThreeDSecureSession(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-three-d-secure-session.json")

	session, err := client.CreateThreeDSecureSession(
		context.TODO(), ThreeDSecureSessionInput{
			CardID:     "tcd_0000AWr2XsTRIF1Vp34gh5",
			ResourceID: "off_0000AKTurT99qsmGTujhis",
			Services:   []ServiceCreateInput{{ID: "ase_00009UhD4ongolulWd9123", Quantity: 1}},
		},
	)
	require.NoError(t, err)
	a.Equal("/payments/three_d_secure_sessions", rec.Path)
	a.JSONEq(
		`{"data":{"card_id":"tcd_0000AWr2XsTRIF1Vp34gh5","resource_id":"off_0000AKTurT99qsmGTujhis",
		"services":[{"id":"ase_00009UhD4ongolulWd9123","quantity":1}]}}`, rec.Body,
	)
	a.Equal(ThreeDSecureSessionStatusReadyForPayment, session.Status)
	a.Equal("3ds_0000AWr2XsTRIF1Vp34gh5", session.ID)
}
