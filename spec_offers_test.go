package duffel

import (
	"context"
	"net/http"
	"testing"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOfferDecodesSpecFields(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-offer-full.json")

	offer, err := client.GetOffer(context.TODO(), "off_00009htYpSCXrwaB9DnUm0", GetOfferParams{ReturnAvailableServices: true})
	require.NoError(t, err)
	a.Equal("/air/offers/off_00009htYpSCXrwaB9DnUm0", rec.Path)
	a.Equal("true", rec.Query.Get("return_available_services"))

	a.Equal([]string{"acd_00009htYpSCXrwaB9DnUm1", "acd_00009htYpSCXrwaB9DnUm2"}, offer.AvailableAirlineCreditIDs)
	require.Len(t, offer.IntendedPaymentMethods, 2)
	card, credit := offer.IntendedPaymentMethods[0], offer.IntendedPaymentMethods[1]
	a.Equal(IntendedPaymentMethodCard, card.Type)
	a.Equal("500.20 GBP", card.ChargeAmount().String())
	a.Equal("30.20 GBP", card.SurchargeAmount().String())
	a.Nil(card.PenaltyAmount())
	a.Equal(IntendedPaymentMethodAirlineCredit, credit.Type)
	a.Equal("10.20 GBP", credit.PenaltyAmount().String())

	require.Len(t, offer.TaxBreakdown, 1)
	a.Equal("GB", offer.TaxBreakdown[0].Code)
	a.Equal("26.00 GBP", offer.TaxBreakdown[0].Amount().String())
	a.False(*offer.TaxBreakdown[0].CarrierImposed)

	slice := offer.Slices[0]
	a.Equal("BmlZDw==", slice.ComparisonKey)
	a.Equal(1, *slice.NGSShelf)
	a.True(*slice.Conditions.PriorityBoarding)
	a.True(*slice.Conditions.PriorityCheckIn)
	a.True(*slice.Conditions.AdvanceSeatSelection)
	a.Equal("LON", slice.Origin.Airports[0].IATACityCode)

	cabin := slice.Segments[0].Passengers[0].Cabin
	require.NotNil(t, cabin)
	a.Equal(CabinClassEconomy, cabin.Name)
	a.Equal(WifiCostFree, cabin.Amenities.Wifi.Cost)
	a.Equal(SeatLegroomStandard, cabin.Amenities.Seat.Legroom)
	a.True(cabin.Amenities.Power.Available)

	a.Equal(23.0, *offer.AvailableServices[0].Metadata.MaximumWeightKg)
}

func TestAvailableServiceMetadataDecodesDecimalsAndNull(t *testing.T) {
	var m AvailableServiceMetadata
	require.NoError(t, json.Unmarshal([]byte(`{"type":"checked","maximum_weight_kg":23.5,"maximum_depth_cm":null}`), &m))
	assert.Equal(t, 23.5, *m.MaximumWeightKg)
	assert.Nil(t, m.MaximumDepthCM)
}

func TestPriceOffer(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-price-offer.json")

	offer, err := client.PriceOffer(
		context.TODO(), "off_00009htYpSCXrwaB9DnUm0", PriceOfferInput{
			IntendedPaymentMethods: []IntendedPaymentMethodInput{
				{Type: IntendedPaymentMethodCard, CardID: "tcd_0000A3t2sL46m5eDF45kWW"},
				{Type: IntendedPaymentMethodAirlineCredit, AirlineCreditID: "acd_00009htYpSCXrwaB9DnUm1"},
			},
			IntendedServices: []ServiceCreateInput{{ID: "ase_00009UhD4ongolulWd9123", Quantity: 1}},
		},
	)
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.Equal("/air/offers/off_00009htYpSCXrwaB9DnUm0/actions/price", rec.Path)
	a.JSONEq(
		`{"data":{"intended_payment_methods":[{"type":"card","card_id":"tcd_0000A3t2sL46m5eDF45kWW"},
		{"type":"airline_credit","airline_credit_id":"acd_00009htYpSCXrwaB9DnUm1"}],
		"intended_services":[{"id":"ase_00009UhD4ongolulWd9123","quantity":1}]}}`, rec.Body,
	)
	a.Equal("off_00009htYpSCXrwaB9DnUm0", offer.ID)
	a.Len(offer.IntendedPaymentMethods, 2)
}

func TestPriceOfferSendsEmptyServices(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-price-offer.json")

	_, err := client.PriceOffer(
		context.TODO(), "off_00009htYpSCXrwaB9DnUm0", PriceOfferInput{
			IntendedPaymentMethods: []IntendedPaymentMethodInput{{Type: IntendedPaymentMethodCard, CardID: "tcd_1"}},
		},
	)
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"intended_payment_methods":[{"type":"card","card_id":"tcd_1"}],"intended_services":[]}}`, rec.Body)
}

func TestGetUpsellOffers(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-upsell-offers.json")

	offers, err := client.GetUpsellOffers(context.TODO(), "off_00009htYpSCXrwaB9DnUm0")
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.Equal("/air/offers/off_00009htYpSCXrwaB9DnUm0/upsell_offers", rec.Path)
	require.Len(t, offers, 1)
	a.Equal("off_00009htYpSCXrwaB9DnUm0", offers[0].ID)

	_, err = client.GetUpsellOffers(context.TODO(), "orq_123")
	a.Error(err)
}

func TestCreateOfferRequestItineraries(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-offer-request-itineraries.json")

	res, err := client.CreateOfferRequestItineraries(
		context.TODO(), OfferRequestInput{
			Passengers:         []OfferRequestPassenger{{Type: PassengerTypeAdult, UserID: "icu_0000AgZitpOnQtd3NQxjwO"}},
			Slices:             []OfferRequestSlice{{Origin: "LHR", Destination: "JFK"}},
			AirlineCreditIDs:   []string{"acd_00009htYpSCXrwaB9DnUm1"},
			IncludeSplitTicket: true,
			SupplierTimeout:    10000,
		},
	)
	require.NoError(t, err)
	a.Equal("/air/offer_requests", rec.Path)
	a.Equal("itineraries", rec.Query.Get("view"))
	a.Equal("false", rec.Query.Get("return_offers"))
	a.Equal("10000", rec.Query.Get("supplier_timeout"))
	a.Contains(rec.Body, `"airline_credit_ids":["acd_00009htYpSCXrwaB9DnUm1"]`)
	a.Contains(rec.Body, `"include_split_ticket":true`)
	a.Contains(rec.Body, `"user_id":"icu_0000AgZitpOnQtd3NQxjwO"`)

	a.Equal("orq_00009hjdomFOCJyxHG7k7k", res.ID)
	a.Equal(2, *res.RemainingBatches)
	a.Equal("British Airways", res.References.Airlines["arl_00009VME7D6ivUu8dn35WK"].Name)
	a.Equal("LHR", res.References.Places["arp_lhr_gb"].IATACode)
	a.Equal("Airbus Industries A320", res.References.Aircraft["arc_00009VME7DBMt6U2Ah1Hpq"].Name)

	itinerary := res.Slices[0].Itineraries[0]
	a.Equal("arp_lhr_gb", itinerary.Segments[0].Origin)
	a.Equal("arl_00009VME7D6ivUu8dn35WK", itinerary.Segments[0].MarketingCarrier)
	brand := itinerary.Brands[0]
	a.Equal("Economy Basic", brand.FareBrandName)
	a.Equal(1, *brand.NGSShelf)
	a.Equal(ItineraryOfferTypeSingleTicket, brand.Offers[0].Type)
	a.Equal("off_00009htYpSCXrwaB9Dn456", brand.Offers[0].ID)
	a.Equal("470.00 GBP", brand.Offers[0].TotalAmount().String())
	a.Equal(CabinClassEconomy, brand.Passengers[0].Segments[0].Cabin.Name)
}

func TestGetOfferRequestItineraries(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/201-create-offer-request-itineraries.json")

	res, err := client.GetOfferRequestItineraries(context.TODO(), "orq_00009hjdomFOCJyxHG7k7k")
	require.NoError(t, err)
	a.Equal(http.MethodGet, rec.Method)
	a.Equal("/air/offer_requests/orq_00009hjdomFOCJyxHG7k7k", rec.Path)
	a.Equal("itineraries", rec.Query.Get("view"))
	a.Len(res.Slices, 1)
}

func TestCreateBatchOfferRequest(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusCreated, "fixtures/201-create-batch-offer-request.json")

	res, err := client.CreateBatchOfferRequest(
		context.TODO(), OfferRequestInput{
			Passengers:      []OfferRequestPassenger{{Type: PassengerTypeAdult}},
			Slices:          []OfferRequestSlice{{Origin: "LHR", Destination: "JFK"}},
			ReturnOffers:    true,
			SupplierTimeout: 5000,
		},
	)
	require.NoError(t, err)
	a.Equal(http.MethodPost, rec.Method)
	a.Equal("/air/batch_offer_requests", rec.Path)
	a.Equal("5000", rec.Query.Get("supplier_timeout"))
	a.False(rec.Query.Has("return_offers"))
	a.Equal("orq_00009hjdomFOCJyxHG7k7k", res.ID)
	a.Equal(2, res.TotalBatches)
	a.Equal(2, res.RemainingBatches)
	a.NotEmpty(res.ClientKey)
}

func TestGetBatchOfferRequest(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-batch-offer-request.json")

	res, err := client.GetBatchOfferRequest(context.TODO(), "orq_00009hjdomFOCJyxHG7k7k")
	require.NoError(t, err)
	a.Equal("/air/batch_offer_requests/orq_00009hjdomFOCJyxHG7k7k", rec.Path)
	a.False(rec.Query.Has("view"))
	a.Equal(2, res.RemainingBatches)
	require.Len(t, res.Offers, 1)
	a.Equal("off_00009htYpSCXrwaB9DnUm0", res.Offers[0].ID)
}

func TestGetBatchOfferRequestItineraries(t *testing.T) {
	a := assert.New(t)
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-batch-offer-request-itineraries.json")

	res, err := client.GetBatchOfferRequestItineraries(context.TODO(), "orq_00009hjdomFOCJyxHG7k7k")
	require.NoError(t, err)
	a.Equal("/air/batch_offer_requests/orq_00009hjdomFOCJyxHG7k7k", rec.Path)
	a.Equal("itineraries", rec.Query.Get("view"))
	a.Equal(2, *res.TotalBatches)
	a.Len(res.Slices[0].Itineraries[0].Brands, 1)
}

func TestGetLoyaltyProgrammeOwnerAirline(t *testing.T) {
	client, rec := newRecordingClient(t, http.StatusOK, "fixtures/200-get-loyalty-programme.json")

	programme, err := client.GetLoyaltyProgramme(context.TODO(), "loy_00001876aqC8c5umZmrRds")
	require.NoError(t, err)
	assert.Equal(t, "/air/loyalty_programmes/loy_00001876aqC8c5umZmrRds", rec.Path)
	require.NotNil(t, programme.OwnerAirline)
	assert.Equal(t, "BA", programme.OwnerAirline.IATACode)
}
