package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/thetreep/duffel"
)

var rowConfigAutoMerge table.RowConfig

func main() {
	token := os.Getenv("DUFFEL_TOKEN")
	if token == "" {
		log.Fatal("DUFFEL_TOKEN environment variable not set")
	}

	client := duffel.New(token)
	cardsAPIClient := duffel.New(token, duffel.WithDebug())
	ctx := context.Background()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleBold)
	rowConfigAutoMerge = table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"Test", "Step", "Result", "Details"})
	t.SetColumnConfigs(
		[]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 2, AutoMerge: false},
			{
				Number: 3, AutoMerge: false, Transformer: func(val interface{}) string {
					if val == "PASSED" {
						return text.FgGreen.Sprint(val)
					}
					return text.FgRed.Sprint(val)
				},
			},
			{Number: 4, AutoMerge: false},
		},
	)
	t.Style().Options.SeparateRows = true

	tests := []struct {
		name string
		fn   func(context.Context, duffel.Duffel, duffel.Duffel, table.Writer)
	}{
		{"No Flights", testNoFlights},
		{"Hold Order", testHoldOrder},
		{"Connecting Flights", testConnectingFlights},
		{"No Baggages", testNoBaggages},
		{"No Services", testNoServices},
		{"Offer Unavailable", testOfferUnavailable},
		{"Offer Price Change", testOfferPriceChange},
		{"Order Creation Error", testOrderCreationError},
		{"Insufficient Balance", testInsufficientBalance},
		{"Card Payment Success", testCardPaymentSuccess},
		{"Airline Initiated Change", testAirlineInitiatedChange},
	}

	for _, test := range tests {
		fmt.Printf("Testing %s...\n", test.name)
		test.fn(ctx, client, cardsAPIClient, t)
	}

	t.Render()
}

func createOfferRequest(
	ctx context.Context, client duffel.Duffel, t table.Writer, testName, origin, destination string,
) (*duffel.OfferRequest, []*duffel.Offer) {
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        origin,
					Destination:   destination,
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)

	if err != nil {
		t.AppendRow(
			table.Row{testName, "Create Offer Request", "FAILED", fmt.Sprintf("Error: %v", err)}, rowConfigAutoMerge,
		)
		return nil, nil
	}
	t.AppendRow(
		table.Row{testName, "Create Offer Request", "PASSED", fmt.Sprintf("Offer Request ID: %s", offerReq.ID)},
		rowConfigAutoMerge,
	)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	if err != nil {
		t.AppendRow(table.Row{testName, "List Offers", "FAILED", fmt.Sprintf("Error: %v", err)}, rowConfigAutoMerge)
		return offerReq, nil
	}
	t.AppendRow(
		table.Row{testName, "List Offers", "PASSED", fmt.Sprintf("Found %d offers", len(allOffers))},
		rowConfigAutoMerge,
	)

	return offerReq, allOffers
}

func testNoFlights(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "No Flights", "PVD", "RAI")
	if allOffers != nil && len(allOffers) == 0 {
		t.AppendRow(
			table.Row{"No Flights", "Check Offers", "PASSED", "No offers returned as expected"}, rowConfigAutoMerge,
		)
	} else if allOffers != nil {
		t.AppendRow(
			table.Row{
				"No Flights", "Check Offers", "FAILED", fmt.Sprintf("Expected 0 offers, got %d", len(allOffers)),
			}, rowConfigAutoMerge,
		)
	}
}

func testHoldOrder(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Hold Order", "JFK", "EWR")
	if allOffers == nil {
		return
	}

	holdOrderFound := false
	for _, offer := range allOffers {
		if !offer.PaymentRequirements.RequiresInstantPayment {
			holdOrderFound = true
			t.AppendRow(
				table.Row{"Hold Order", "Check Offer", "PASSED", fmt.Sprintf("Offer ID: %s", offer.ID)},
				rowConfigAutoMerge,
			)
			break
		}
	}

	if !holdOrderFound {
		t.AppendRow(
			table.Row{"Hold Order", "Check Offer", "FAILED", "All offers require instant payment"}, rowConfigAutoMerge,
		)
	}
}

func testConnectingFlights(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Connecting Flights", "LHR", "DXB")
	if allOffers == nil {
		return
	}

	connectingFlightFound := false
	for _, offer := range allOffers {
		if len(offer.Slices[0].Segments) > 1 {
			connectingFlightFound = true
			t.AppendRow(
				table.Row{
					"Connecting Flights", "Check Offer", "PASSED",
					fmt.Sprintf("Offer ID: %s, Segments: %d", offer.ID, len(offer.Slices[0].Segments)),
				}, rowConfigAutoMerge,
			)
			break
		}
	}

	if !connectingFlightFound {
		t.AppendRow(
			table.Row{"Connecting Flights", "Check Offer", "FAILED", "Only direct flights found"}, rowConfigAutoMerge,
		)
	}
}

func testNoBaggages(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "No Baggages", "BTS", "MRU")
	if allOffers == nil {
		return
	}

	baggageOfferFound := false
	for _, offer := range allOffers {
		if len(offer.Slices[0].Segments[0].Passengers[0].Baggages) > 0 {
			baggageOfferFound = true
			t.AppendRow(
				table.Row{"No Baggages", "Check Offer", "FAILED", fmt.Sprintf("Offer ID: %s", offer.ID)},
				rowConfigAutoMerge,
			)
			break
		}
	}

	if !baggageOfferFound {
		t.AppendRow(
			table.Row{"No Baggages", "Check Offer", "PASSED", "No offers include baggage as expected"},
			rowConfigAutoMerge,
		)
	}
}

func testNoServices(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "No Additional Services", "BTS", "ABV")
	if len(allOffers) == 0 {
		return
	}

	offer, err := client.GetOffer(ctx, allOffers[0].ID, duffel.GetOfferParams{ReturnAvailableServices: true})
	if err != nil {
		t.AppendRow(
			table.Row{"No Additional Services", "Get Offer", "FAILED", fmt.Sprintf("Error: %v", err)},
			rowConfigAutoMerge,
		)
		return
	}
	t.AppendRow(
		table.Row{"No Additional Services", "Get Offer", "PASSED", fmt.Sprintf("Offer ID: %s", offer.ID)},
		rowConfigAutoMerge,
	)

	if len(offer.AvailableServices) == 0 {
		t.AppendRow(
			table.Row{
				"No Additional Services", "Check Services", "PASSED", "No services available as expected",
			}, rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{
				"No Additional Services", "Check Services", "FAILED",
				fmt.Sprintf("Found %d services", len(offer.AvailableServices)),
			}, rowConfigAutoMerge,
		)
	}
}

func testOfferUnavailable(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Offer Unavailable", "LGW", "LHR")
	if len(allOffers) == 0 {
		return
	}

	_, err := client.GetOffer(ctx, allOffers[0].ID)
	if err != nil && duffel.IsErrorCode(err, duffel.OfferNoLongerAvailable) {
		t.AppendRow(
			table.Row{"Offer Unavailable", "Get Offer", "PASSED", "Offer no longer available as expected"},
			rowConfigAutoMerge,
		)
	} else if err != nil {
		t.AppendRow(
			table.Row{"Offer Unavailable", "Get Offer", "FAILED", fmt.Sprintf("Unexpected error: %v", err)},
			rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(table.Row{"Offer Unavailable", "Get Offer", "FAILED", "Offer still available"}, rowConfigAutoMerge)
	}
}

func testOfferPriceChange(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Offer Price Change", "LHR", "STN")
	if len(allOffers) == 0 {
		return
	}

	originalOffer := allOffers[0]
	updatedOffer, err := client.GetOffer(ctx, originalOffer.ID)
	if err != nil {
		t.AppendRow(
			table.Row{"Offer Price Change", "Get Updated Offer", "FAILED", fmt.Sprintf("Error: %v", err)},
			rowConfigAutoMerge,
		)
		return
	}
	t.AppendRow(
		table.Row{
			"Offer Price Change", "Get Updated Offer", "PASSED", fmt.Sprintf("Updated Offer ID: %s", updatedOffer.ID),
		}, rowConfigAutoMerge,
	)

	if originalOffer.TotalAmount().String() != updatedOffer.TotalAmount().String() {
		t.AppendRow(
			table.Row{
				"Offer Price Change", "Compare Prices", "PASSED", fmt.Sprintf(
					"Price changed from %s to %s", originalOffer.TotalAmount().String(),
					updatedOffer.TotalAmount().String(),
				),
			}, rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{"Offer Price Change", "Compare Prices", "FAILED", "Price remained the same"}, rowConfigAutoMerge,
		)
	}
}

func testOrderCreationError(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Order Creation Error", "LHR", "LGW")
	if len(allOffers) == 0 {
		return
	}

	_, err := createOrder(ctx, client, allOffers[0], duffel.PaymentMethodBalance)
	if err != nil && duffel.IsErrorType(err, duffel.AirlineError) {
		t.AppendRow(
			table.Row{"Order Creation Error", "Create Order", "PASSED", "Order creation failed as expected"},
			rowConfigAutoMerge,
		)
	} else if err != nil {
		t.AppendRow(
			table.Row{
				"Order Creation Error", "Create Order", "FAILED", fmt.Sprintf("Unexpected error: %v", err),
			}, rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{"Order Creation Error", "Create Order", "FAILED", "Order creation succeeded"}, rowConfigAutoMerge,
		)
	}
}

func testInsufficientBalance(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Insufficient Balance", "LGW", "STN")
	if len(allOffers) == 0 {
		return
	}

	_, err := createOrder(ctx, client, allOffers[0], duffel.PaymentMethodBalance)
	if err != nil && duffel.IsErrorCode(err, duffel.InsufficientBalance) {
		t.AppendRow(
			table.Row{
				"Insufficient Balance", "Create Order", "PASSED", "Insufficient balance error as expected",
			}, rowConfigAutoMerge,
		)
	} else if err != nil {
		t.AppendRow(
			table.Row{
				"Insufficient Balance", "Create Order", "FAILED", fmt.Sprintf("Unexpected error: %v", err),
			}, rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{"Insufficient Balance", "Create Order", "FAILED", "Order creation succeeded"}, rowConfigAutoMerge,
		)
	}
}

func testCardPaymentSuccess(ctx context.Context, client duffel.Duffel, cardsAPIClient duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Card Payment Success", "LTN", "STN")
	if len(allOffers) == 0 {
		return
	}

	card, err := createTemporaryPaymentCard(ctx, cardsAPIClient)
	if err != nil {
		t.AppendRow(
			table.Row{"Card Payment Success", "Create Payment Card", "FAILED", fmt.Sprintf("Error: %v", err)},
			rowConfigAutoMerge,
		)
		return
	}
	t.AppendRow(
		table.Row{"Card Payment Success", "Create Payment Card", "PASSED", fmt.Sprintf("Card ID: %s", card.ID)},
		rowConfigAutoMerge,
	)

	order, err := createOrder(ctx, client, allOffers[0], duffel.PaymentMethodCard, card.ID)
	if err != nil {
		t.AppendRow(
			table.Row{"Card Payment Success", "Create Order", "FAILED", fmt.Sprintf("Error: %v", err)},
			rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{"Card Payment Success", "Create Order", "PASSED", fmt.Sprintf("Order ID: %s", order.ID)},
			rowConfigAutoMerge,
		)
	}
}

func testAirlineInitiatedChange(ctx context.Context, client duffel.Duffel, _ duffel.Duffel, t table.Writer) {
	_, allOffers := createOfferRequest(ctx, client, t, "Airline-Initiated Change", "LHR", "LTN")
	if len(allOffers) == 0 {
		return
	}

	order, err := createOrder(ctx, client, allOffers[0], duffel.PaymentMethodBalance)
	if err != nil {
		t.AppendRow(
			table.Row{"Airline-Initiated Change", "Create Order", "FAILED", fmt.Sprintf("Error: %v", err)},
			rowConfigAutoMerge,
		)
		return
	}

	changes, err := client.ListAirlineInitiatedChanges(
		ctx, duffel.ListAirlineInitiatedChangesParams{
			OrderID: order.ID,
		},
	)
	if err != nil {
		t.AppendRow(
			table.Row{
				"Airline-Initiated Change", "List Changes", "FAILED", fmt.Sprintf("Error: %v", err),
			}, rowConfigAutoMerge,
		)
		return
	}

	if len(changes) > 0 {
		t.AppendRow(
			table.Row{
				"Airline-Initiated Change", "Check Changes", "PASSED",
				fmt.Sprintf("Found %d changes", len(changes)),
			},
			rowConfigAutoMerge,
		)
	} else {
		t.AppendRow(
			table.Row{"Airline-Initiated Change", "Check Changes", "FAILED", "No changes found"},
			rowConfigAutoMerge,
		)
	}
}

func createOrder(
	ctx context.Context, client duffel.Duffel, offer *duffel.Offer, paymentMethod duffel.PaymentMethod,
	cardID ...string,
) (*duffel.Order, error) {
	payment := duffel.PaymentCreateInput{
		Type:     paymentMethod,
		Amount:   offer.RawTotalAmount,
		Currency: offer.RawTotalCurrency,
	}
	if paymentMethod == duffel.PaymentMethodCard && len(cardID) > 0 {
		payment.CardID = cardID[0]
	}

	return client.CreateOrder(
		ctx, duffel.CreateOrderInput{
			Type:           duffel.OrderTypeInstant,
			SelectedOffers: []string{offer.ID},
			Passengers: []duffel.OrderPassenger{
				{
					ID:          offer.Passengers[0].ID,
					Title:       duffel.PassengerTitleMrs,
					GivenName:   "Amelia",
					FamilyName:  "Earhart",
					Gender:      duffel.GenderFemale,
					BornOn:      duffel.Date(time.Now().AddDate(-30, 0, 0)),
					Email:       "amelia@duffel.com",
					PhoneNumber: "+442080160509",
				},
			},
			Payments: []duffel.PaymentCreateInput{payment},
		},
	)
}

func createTemporaryPaymentCard(ctx context.Context, cardsAPIClient duffel.Duffel) (*duffel.PaymentCard, error) {
	return cardsAPIClient.CreatePaymentCardRecord(
		ctx, &duffel.CreatePaymentCardRecordRequest{
			AddressCity:        "London",
			AddressCountryCode: "GB",
			AddressLine1:       "1 Downing St",
			AddressLine2:       "First floor",
			AddressPostalCode:  "EC2A 4RQ",
			AddressRegion:      "London",
			ExpiryMonth:        "07",
			ExpiryYear:         "30",
			Name:               "Neil Armstrong",
			Number:             "347828429964915",
			SecurityCode:       "2271",
			MultiUse:           false,
		},
	)
}
