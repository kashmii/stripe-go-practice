package subscription

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/customer"
	"github.com/stripe/stripe-go/v85/subscription"
)

type createCustomerRequest struct {
	Email string `json:"email"`
}

type createCustomerResponse struct {
	CustomerID string `json:"customerId"`
}

type createSubscriptionRequest struct {
	CustomerID string `json:"customerId"`
	PriceID    string `json:"priceId"`
}

type createSubscriptionResponse struct {
	ClientSecret string `json:"clientSecret"`
}

// CreateCustomer はメールアドレスを受け取り Stripe に Customer を作成する
// (Subscription を作る際に必要となる)
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	// リクエストボディを構造体にデコード
	var req createCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	// Stripe に Customer を作成する
	params := &stripe.CustomerParams{
		Email: stripe.String(req.Email),
	}

	c, err := customer.New(params)
	if err != nil {
		stripeErr, ok := err.(*stripe.Error)
		if ok {
			log.Printf("[error] stripe error: type=%s code=%s msg=%s", stripeErr.Type, stripeErr.Code, stripeErr.Msg)
		} else {
			log.Printf("[error] unexpected error: %v", err)
		}
		http.Error(w, "failed to create customer", http.StatusInternalServerError)
		return
	}

	// 作成した Customer の ID をフロントに返す（以降の Subscription 作成に使う）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createCustomerResponse{
		CustomerID: c.ID,
	})
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" || req.PriceID == "" {
		http.Error(w, "customerId and priceId are required", http.StatusBadRequest)
		return
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(req.PriceID)},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
	}

	params.AddExpand("latest_invoice.confirmation_secret")

	idempotencyKey := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	params.SetIdempotencyKey(idempotencyKey)

	sub, err := subscription.New(params)
	if err != nil {
		stripeErr, ok := err.(*stripe.Error)
		if ok {
			log.Printf("[error] stripe error: type=%s code=%s msg=%s", stripeErr.Type, stripeErr.Code, stripeErr.Msg)
		} else {
			log.Printf("[error] unexpected error: %v", err)
		}
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	if sub.LatestInvoice == nil || sub.LatestInvoice.ConfirmationSecret == nil {
		log.Printf("[error] confirmation secret not found")
		http.Error(w, "confirmation secret not found", http.StatusInternalServerError)
		return
	}

	clientSecret := sub.LatestInvoice.ConfirmationSecret.ClientSecret

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createSubscriptionResponse{
		ClientSecret: clientSecret,
	})
}
