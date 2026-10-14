package payment

import (
	"encoding/json"
	"net/http"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/refund"
)

type refundRequest struct {
	PaymentIntentID string `json:"paymentIntentId"`
	Amount          int64  `json:"amount"` // 0 の場合は全額返金
}

type refundResponse struct {
	RefundID string `json:"refundId"`
	Amount   int64  `json:"amount"`
	Status   string `json:"status"`
}

func CreateRefund(w http.ResponseWriter, r *http.Request) {
	var req refundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.PaymentIntentID == "" {
		http.Error(w, "paymentIntentId is required", http.StatusBadRequest)
		return
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
	}
	// Amount が指定されている場合は一部返金、0 の場合は全額返金
	if req.Amount > 0 {
		params.Amount = stripe.Int64(req.Amount)
	}

	re, err := refund.New(params)
	if err != nil {
		handleStripeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refundResponse{
		RefundID: re.ID,
		Amount:   re.Amount,
		Status:   string(re.Status),
	})
}
