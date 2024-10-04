package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/thetreep/duffel"
)

func main() {
	token := os.Getenv("DUFFEL_TOKEN")
	if token == "" {
		log.Fatal("DUFFEL_TOKEN environment variable not set")
	}

	client := duffel.New(token)
	ctx := context.Background()

	cardsAPIClient := duffel.New(token, duffel.WithHost("https://api.duffel.cards"), duffel.WithDebug())

	// Offer Request Tests
	testNoFlights(ctx, client)
	testHoldOrder(ctx, client)
	testConnectingFlights(ctx, client)
	testNoBaggages(ctx, client)
	// // testOfferTimeout(ctx, client)
	//
	// // Offer Tests
	testNoServices(ctx, client)
	testOfferUnavailable(ctx, client)
	testOfferPriceChange(ctx, client)
	//
	// // Order Tests
	testOrderCreationError(ctx, client)
	testInsufficientBalance(ctx, client)
	testOfferUnavailable(ctx, client)
	// testCardPaymentSuccess(ctx, client, cardsAPIClient) // Needs specific API access
	// testCardPaymentAccepted(ctx, client)
	//
	// // Payment Tests
	// testPaymentSuccess(ctx, client)
	// testPaymentAccepted(ctx, client)
	//
	// // Airline-Initiated Change Test
	// testAirlineChange(ctx, client)
	//
	// // Airline Credit Test
	// testAirlineCredit(ctx, client)
}

func testNoFlights(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing no flights returned...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "PVD",
					Destination:   "RAI",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) == 0 {
		t.AppendRow(table.Row{"No Flights", "PASSED", "No offers returned as expected"})
	} else {
		t.AppendRow(table.Row{"No Flights", "FAILED", fmt.Sprintf("Expected 0 offers, got %d", len(allOffers))})
	}

	t.Render()
}

func testHoldOrder(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing hold order...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "JFK",
					Destination:   "EWR",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	holdOrderFound := false
	for _, offer := range allOffers {
		if !offer.PaymentRequirements.RequiresInstantPayment {
			holdOrderFound = true
			t.AppendRow(table.Row{"Hold Order", "PASSED", fmt.Sprintf("Offer ID: %s", offer.ID)})
			break
		}
	}

	if !holdOrderFound {
		t.AppendRow(table.Row{"Hold Order", "FAILED", "All offers require instant payment"})
	}

	t.Render()
}

func testConnectingFlights(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing connecting flights...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LHR",
					Destination:   "DXB",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	connectingFlightFound := false
	for _, offer := range allOffers {
		if len(offer.Slices[0].Segments) > 1 {
			connectingFlightFound = true
			t.AppendRow(
				table.Row{
					"Connecting Flights", "PASSED",
					fmt.Sprintf("Offer ID: %s, Segments: %d", offer.ID, len(offer.Slices[0].Segments)),
				},
			)
			break
		}
	}

	if !connectingFlightFound {
		t.AppendRow(table.Row{"Connecting Flights", "FAILED", "Only direct flights found"})
	}

	t.Render()
}

func testNoBaggages(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing no baggages...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "BTS",
					Destination:   "MRU",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	baggageOfferFound := false
	for _, offer := range allOffers {
		if len(offer.Slices[0].Segments[0].Passengers[0].Baggages) > 0 {
			baggageOfferFound = true
			t.AppendRow(table.Row{"No Baggages", "FAILED", fmt.Sprintf("Offer ID: %s", offer.ID)})
			break
		}
	}

	if !baggageOfferFound {
		t.AppendRow(table.Row{"No Baggages", "PASSED", "No offers include baggage as expected"})
	}

	t.Render()
}

func testNoServices(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing no additional services...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "BTS",
					Destination:   "ABV",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		offer, err := client.GetOffer(ctx, allOffers[0].ID, duffel.GetOfferParams{ReturnAvailableServices: true})
		handleErr(err)

		if len(offer.AvailableServices) == 0 {
			t.AppendRow(table.Row{"No Additional Services", "PASSED", "No services available as expected"})
		} else {
			t.AppendRow(
				table.Row{
					"No Additional Services", "FAILED", fmt.Sprintf("Found %d services", len(offer.AvailableServices)),
				},
			)
		}
	} else {
		t.AppendRow(table.Row{"No Additional Services", "FAILED", "No offers found"})
	}

	t.Render()
}

func testOfferUnavailable(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing offer unavailable...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LGW",
					Destination:   "LHR",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		_, err := client.GetOffer(ctx, allOffers[0].ID)
		if err != nil && duffel.IsErrorCode(err, duffel.OfferNoLongerAvailable) {
			t.AppendRow(table.Row{"Offer Unavailable", "PASSED", "Offer no longer available as expected"})
		} else {
			t.AppendRow(table.Row{"Offer Unavailable", "FAILED", "Offer still available or unexpected error"})
		}
	} else {
		t.AppendRow(table.Row{"Offer Unavailable", "FAILED", "No offers found"})
	}

	t.Render()
}

func testOfferPriceChange(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing offer price change...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LHR",
					Destination:   "STN",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		originalOffer := allOffers[0]
		updatedOffer, err := client.GetOffer(ctx, originalOffer.ID)
		handleErr(err)

		if originalOffer.TotalAmount().String() != updatedOffer.TotalAmount().String() {
			t.AppendRow(
				table.Row{
					"Offer Price Change", "PASSED", fmt.Sprintf(
						"Price changed from %s to %s", originalOffer.TotalAmount().String(),
						updatedOffer.TotalAmount().String(),
					),
				},
			)
		} else {
			t.AppendRow(table.Row{"Offer Price Change", "FAILED", "Price remained the same"})
		}
	} else {
		t.AppendRow(table.Row{"Offer Price Change", "FAILED", "No offers found"})
	}

	t.Render()
}

func testOrderCreationError(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing order creation error...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LHR",
					Destination:   "LGW",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		_, err := client.CreateOrder(
			ctx, duffel.CreateOrderInput{
				Type:           duffel.OrderTypeInstant,
				SelectedOffers: []string{allOffers[0].ID},
				Passengers: []duffel.OrderPassenger{
					{
						ID:          allOffers[0].Passengers[0].ID,
						Title:       duffel.PassengerTitleMrs,
						GivenName:   "Amelia",
						FamilyName:  "Earhart",
						Gender:      duffel.GenderFemale,
						BornOn:      duffel.Date(time.Now().AddDate(-30, 0, 0)),
						Email:       "amelia@duffel.com",
						PhoneNumber: "+442080160509",
					},
				},
				Payments: []duffel.PaymentCreateInput{
					{
						Type:     duffel.PaymentMethodBalance,
						Amount:   allOffers[0].RawTotalAmount,
						Currency: allOffers[0].RawTotalCurrency,
					},
				},
			},
		)

		if err != nil && duffel.IsErrorType(err, duffel.AirlineError) {
			t.AppendRow(table.Row{"Order Creation Error", "PASSED", "Order creation failed as expected"})
		} else {
			t.AppendRow(table.Row{"Order Creation Error", "FAILED", "Order creation succeeded or unexpected error"})
		}
	} else {
		t.AppendRow(table.Row{"Order Creation Error", "FAILED", "No offers found"})
	}

	t.Render()
}

func testInsufficientBalance(ctx context.Context, client duffel.Duffel) {
	fmt.Println("Testing insufficient balance...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LGW",
					Destination:   "STN",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		_, err := client.CreateOrder(
			ctx, duffel.CreateOrderInput{
				Type:           duffel.OrderTypeInstant,
				SelectedOffers: []string{allOffers[0].ID},
				Passengers: []duffel.OrderPassenger{
					{
						ID:          allOffers[0].Passengers[0].ID,
						Title:       duffel.PassengerTitleMrs,
						GivenName:   "Amelia",
						FamilyName:  "Earhart",
						Gender:      duffel.GenderFemale,
						BornOn:      duffel.Date(time.Now().AddDate(-30, 0, 0)),
						Email:       "amelia@duffel.com",
						PhoneNumber: "+442080160509",
					},
				},
				Payments: []duffel.PaymentCreateInput{
					{
						Type:     duffel.PaymentMethodBalance,
						Amount:   allOffers[0].RawTotalAmount,
						Currency: allOffers[0].RawTotalCurrency,
					},
				},
			},
		)

		if err != nil && duffel.IsErrorCode(err, duffel.InsufficientBalance) {
			t.AppendRow(table.Row{"Insufficient Balance", "PASSED", "Insufficient balance error as expected"})
		} else {
			t.AppendRow(table.Row{"Insufficient Balance", "FAILED", "Order creation succeeded or unexpected error"})
		}
	} else {
		t.AppendRow(table.Row{"Insufficient Balance", "FAILED", "No offers found"})
	}

	t.Render()
}

func testCardPaymentSuccess(ctx context.Context, client duffel.Duffel, cardsAPIClient duffel.Duffel) {
	fmt.Println("Testing card payment success...")
	offerReq, err := client.CreateOfferRequest(
		ctx, duffel.OfferRequestInput{
			CabinClass: duffel.CabinClassEconomy,
			Passengers: []duffel.OfferRequestPassenger{{Type: duffel.PassengerTypeAdult}},
			Slices: []duffel.OfferRequestSlice{
				{
					Origin:        "LTN",
					Destination:   "STN",
					DepartureDate: duffel.Date(time.Now().AddDate(0, 0, 7)),
				},
			},
		},
	)
	handleErr(err)

	// Create a temporary payment card
	card, err := cardsAPIClient.CreatePaymentCardRecord(
		ctx, &duffel.CreatePaymentCardRecordRequest{
			AddressCity:        "London",
			AddressCountryCode: "GB",
			AddressLine1:       "1 Downing St",
			AddressLine2:       "First floot",
			AddressPostalCode:  "EC2A 4RQ",
			AddressRegion:      "London",
			ExpireMonth:        "03",
			ExpireYear:         "30",
			Name:               "Neil Armstrong",
			Number:             "4242424242424242",
			SecurityCode:       "123",
			MultiUse:           false,
		},
	)
	handleErr(err)

	offers := client.ListOffers(ctx, offerReq.ID)
	allOffers, err := duffel.Collect(offers)
	handleErr(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Result", "Details"})

	if len(allOffers) > 0 {
		order, err := client.CreateOrder(
			ctx, duffel.CreateOrderInput{
				Type:           duffel.OrderTypeInstant,
				SelectedOffers: []string{allOffers[0].ID},
				Passengers: []duffel.OrderPassenger{
					{
						ID:          allOffers[0].Passengers[0].ID,
						Title:       duffel.PassengerTitleMrs,
						GivenName:   "Amelia",
						FamilyName:  "Earhart",
						Gender:      duffel.GenderFemale,
						BornOn:      duffel.Date(time.Now().AddDate(-30, 0, 0)),
						Email:       "amelia@duffel.com",
						PhoneNumber: "+442080160509",
					},
				},
				Payments: []duffel.PaymentCreateInput{
					{
						Type:     duffel.PaymentMethodCard,
						Amount:   allOffers[0].RawTotalAmount,
						Currency: allOffers[0].RawTotalCurrency,
						CardID:   card.ID,
					},
				},
			},
		)

		if err == nil {
			t.AppendRow(
				table.Row{
					"Card Payment Success", "PASSED", fmt.Sprintf("Order created successfully: %s", order.ID),
				},
			)
		} else {
			t.AppendRow(table.Row{"Card Payment Success", "FAILED", fmt.Sprintf("Error creating order: %v", err)})
		}
	} else {
		t.AppendRow(table.Row{"Card Payment Success", "FAILED", "No offers found"})
	}

	t.Render()
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
