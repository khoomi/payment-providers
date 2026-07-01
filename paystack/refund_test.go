package paystack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	payproviders "github.com/khoomi/payment-providers"
	"github.com/khoomi/payment-providers/paystack"
)

func TestRefund_PartialRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refund" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": true,
			"message": "Refund has been queued for processing",
			"data": {
				"id": 4128514,
				"amount": 250000,
				"currency": "NGN",
				"status": "pending"
			}
		}`))
	}))
	defer server.Close()

	provider := paystack.New(paystack.Config{
		SecretKey:  "test_secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	result, err := provider.Refund(context.Background(), payproviders.RefundRequest{
		TransactionReference: "PSK_ref_123",
		Amount:               250000,
		CustomerNote:         "Partial refund",
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if result.Reference != "4128514" {
		t.Fatalf("reference %q, want 4128514", result.Reference)
	}
	if result.Status != payproviders.RefundStatusPending {
		t.Fatalf("status %q, want pending", result.Status)
	}
	if result.Amount != 250000 {
		t.Fatalf("amount %d, want 250000", result.Amount)
	}
}

func TestRefund_MissingReference(t *testing.T) {
	p := paystack.New(paystack.Config{SecretKey: "test"})
	_, err := p.Refund(context.Background(), payproviders.RefundRequest{})
	if err == nil {
		t.Fatal("expected error for missing transaction reference")
	}
}