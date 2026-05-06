package subscription

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/customer"
)

type createCustomerRequest struct {
	Email string `json:"email"`
}

type createCustomerResponse struct {
	CustomerID string `json:"customerId"`
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