package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kashmii/stripe-go-practice/internal/fraud"
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	// リクエストボディを読み込む
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// 署名検証：Stripe から届いたリクエストが本物かを確認する
	// Stripe-Signature ヘッダーと STRIPE_WEBHOOK_SECRET を使って検証
	// IgnoreAPIVersionMismatch: Stripe CLI のバージョンと stripe-go のバージョンがずれる場合に無視する（ローカル開発用）
	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"), webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// イベントタイプで分岐
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("Failed to unmarshal PaymentIntent: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		log.Printf("[succeeded] PaymentIntent ID: %s, Amount: %d", pi.ID, pi.Amount)

		// fraudCheck：ビジネスロジック実行前に不正検知を挟む
		if err := fraud.Check(&pi); err != nil {
			log.Printf("[fraud detected] %v", err)
			return
		}
		log.Printf("[fraud ok] proceeding to business logic")

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("Failed to unmarshal PaymentIntent: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		log.Printf("[payment_failed] PaymentIntent ID: %s, Amount: %d", pi.ID, pi.Amount)

	// Subscription の Invoice 決済が成功したとき（毎月の自動更新を含む）
	case "invoice.payment_succeeded":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			log.Printf("Failed to unmarshal Invoice: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		log.Printf("[invoice.payment_succeeded] Invoice ID: %s, Amount: %d, Customer: %s", inv.ID, inv.AmountPaid, inv.Customer.ID)

	// Subscription のステータスが変化したとき（プラン変更・解約・更新など）
	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("Failed to unmarshal Subscription: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		log.Printf("[subscription.updated] Subscription ID: %s, Status: %s", sub.ID, sub.Status)

	default:
		log.Printf("[ignored] event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}
