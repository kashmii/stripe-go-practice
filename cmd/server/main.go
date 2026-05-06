package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/kashmii/stripe-go-practice/internal/payment"
	"github.com/kashmii/stripe-go-practice/internal/subscription"
	"github.com/kashmii/stripe-go-practice/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v85"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to environment variables")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Fatal("STRIPE_SECRET_KEY is not set")
	}

	r := chi.NewRouter()
	r.Post("/create-payment-intent", payment.CreatePaymentIntent)
	r.Post("/subscriptions/customer", subscription.CreateCustomer)
	r.Post("/webhook", webhook.Handle)
	r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"publishableKey": os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		})
	})
	r.Handle("/*", http.FileServer(http.Dir("static")))

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
