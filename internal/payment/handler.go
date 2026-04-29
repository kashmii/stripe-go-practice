package payment

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
		handleStripeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createPaymentIntentResponse{
		ClientSecret: pi.ClientSecret,
	})
}

// handleStripeError は Stripe APIから返ったエラーを種類ごとに分類してログ出力・レスポンスを返す
func handleStripeError(w http.ResponseWriter, err error) {
	stripeErr, ok := err.(*stripe.Error)
	if !ok {
		// Stripe に届く前に発生したネットワークエラーなど
		log.Printf("[error] unexpected error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	switch stripeErr.Type {
	case stripe.ErrorTypeCard:
		// カードエラー：拒否理由（DeclineCode）でさらに分類する
		switch stripeErr.DeclineCode {
		case stripe.DeclineCodeFraudulent:
			// Stripe Radar が不正と判定した
			log.Printf("[error] fraudulent card: %s", stripeErr.Msg)
			http.Error(w, "payment not allowed", http.StatusForbidden)
		default:
			// 残高不足・有効期限切れなど一般的なカード拒否
			log.Printf("[error] card declined: code=%s msg=%s", stripeErr.DeclineCode, stripeErr.Msg)
			http.Error(w, "card declined", http.StatusPaymentRequired)
		}

	case stripe.ErrorTypeInvalidRequest:
		// 必須パラメータ不足など（実装ミス）
		log.Printf("[error] invalid request: %s", stripeErr.Msg)
		http.Error(w, "bad request", http.StatusBadRequest)

	default:
		// APIエラー・認証エラーなどその他
		log.Printf("[error] stripe error: type=%s code=%s msg=%s", stripeErr.Type, stripeErr.Code, stripeErr.Msg)
		http.Error(w, "payment error", http.StatusInternalServerError)
	}
}
