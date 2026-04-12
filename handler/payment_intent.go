package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/paymentintent"
)

type createPaymentIntentResponse struct {
	ClientSecret string `json:"clientSecret"`
}

func CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := fmt.Sprintf("pi-%d", time.Now().UnixNano())

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(1000),
		Currency: stripe.String(string(stripe.CurrencyJPY)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	// 冪等性キーを設定する
	// 同じキーで同じリクエストを複数回送信しても、Stripe は最初のリクエストの結果を返し、二重処理を防ぐ仕組み
	params.SetIdempotencyKey(idempotencyKey)

	pi, err := paymentintent.New(params)
	if err != nil {
		stripeErr, ok := err.(*stripe.Error)
		if ok {
			log.Printf("Stripe error: type=%s code=%s msg=%s", stripeErr.Type, stripeErr.Code, stripeErr.Msg)
		} else {
			log.Printf("Unexpected error: %v", err)
		}
		http.Error(w, "failed to create payment intent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createPaymentIntentResponse{
		ClientSecret: pi.ClientSecret,
	})
}
