// Copyright 2021-present Airheart, Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package duffel

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestCancelOrder(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	a := assert.New(t)
	gock.New("https://api.duffel.com").
		Post("/air/order_cancellations").
		Reply(201).
		JSON(`{"data":{"order_id":"ord_00009hthhsUZ8W4LxQgkjo"}}`).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/201-create-order-cancellation.json")

	ctx := context.TODO()

	client := New("duffel_test_123")
	data, err := client.CreateOrderCancellation(ctx, "ord_00009hthhsUZ8W4LxQgkjo")
	a.NoError(err)
	a.NotNil(data)
	a.Equal("90.80 GBP", data.RefundAmount().String())
}

func TestConfirmCancelOrder(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)
	gock.New("https://api.duffel.com").
		Post("/air/order_cancellations/ore_00009qzZWzjDipIkqpaUAj/actions/confirm").
		BodyString("").
		Reply(201).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/201-create-order-cancellation.json") // Uses the same payload as the create endpoint.

	ctx := context.TODO()

	client := New("duffel_test_123")
	data, err := client.ConfirmOrderCancellation(ctx, "ore_00009qzZWzjDipIkqpaUAj")
	a.NoError(err)
	a.NotNil(data)
	a.Equal("ore_00009qzZWzjDipIkqpaUAj", data.ID)
	a.Equal("90.80 GBP", data.RefundAmount().String())
}

func TestGetOrderCancellation(t *testing.T) {
	defer gock.Off()
	a := assert.New(t)
	gock.New("https://api.duffel.com").
		Get("/air/order_cancellations/ore_00009qzZWzjDipIkqpaUAj").
		Reply(200).
		SetHeader("Ratelimit-Limit", "5").
		SetHeader("Ratelimit-Remaining", "5").
		SetHeader("Ratelimit-Reset", time.Now().Format(time.RFC1123)).
		SetHeader("Date", time.Now().Format(time.RFC1123)).
		File("fixtures/200-get-order-cancellation.json")

	ctx := context.TODO()

	client := New("duffel_test_123")
	data, err := client.GetOrderCancellation(ctx, "ore_00009qzZWzjDipIkqpaUAj")
	a.NoError(err)
	a.NotNil(data)
	a.Equal("90.80 GBP", data.RefundAmount().String())
}

func TestListOrderCancellationsPath(t *testing.T) {
	a := assert.New(t)
	var got string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"ore_1","order_id":"ord_1"}],"meta":{}}`))
	})

	items, err := Collect(client.ListOrderCancellations(context.TODO(), ListOrderCancellationParams{OrderID: "ord_1"}))
	a.NoError(err)
	a.Len(items, 1)
	a.Equal("ore_1", items[0].ID)
	a.Equal("/air/order_cancellations?order_id=ord_1", got)
}

func TestOrderCancellationShortIDsDoNotPanic(t *testing.T) {
	client := New("duffel_test_123")
	for _, id := range []string{"", "ab", "ord_123"} {
		_, err := client.ConfirmOrderCancellation(context.TODO(), id)
		assert.Error(t, err)
		_, err = client.GetOrderCancellation(context.TODO(), id)
		assert.Error(t, err)
	}
}

func TestOrderCancellationTimeAccessors(t *testing.T) {
	a := assert.New(t)
	var oc OrderCancellation
	a.NoError(json.Unmarshal([]byte(`{"expires_at":"2020-01-17T10:42:14Z","created_at":"2020-04-11T15:48:11.642Z","confirmed_at":null}`), &oc))

	expires, ok := oc.ExpiresAtTime()
	a.True(ok)
	a.Equal(time.Date(2020, 1, 17, 10, 42, 14, 0, time.UTC), expires)

	created, ok := oc.CreatedAtTime()
	a.True(ok)
	a.Equal(642*time.Millisecond, time.Duration(created.Nanosecond()))

	_, ok = oc.ConfirmedAtTime()
	a.False(ok)
}
