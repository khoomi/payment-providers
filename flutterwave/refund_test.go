package flutterwave_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	payproviders "github.com/khoomi/payment-providers"
	"github.com/khoomi/payment-providers/flutterwave"
)

func TestRefund_PartialRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/5708/refund" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "success",
			"message": "Transaction refund initiated",
			"data": {
				"id": 223,
				"amount_refunded": 2500,
				"status": "completed",
				"flw_ref": "FLW-REF-123"
			}
		}`))
	}))
	defer server.Close()

	provider := flutterwave.New(flutterwave.Config{
		SecretKey:  "test_secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	result, err := provider.Refund(context.Background(), payproviders.RefundRequest{
		GatewayTransactionID: 5708,
		Amount:               250000,
		CustomerNote:         "Partial refund",
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if result.Reference != "FLW-REF-123" {
		t.Fatalf("reference %q, want FLW-REF-123", result.Reference)
	}
	if result.Status != payproviders.RefundStatusProcessed {
		t.Fatalf("status %q, want processed", result.Status)
	}
	if result.Amount != 250000 {
		t.Fatalf("amount %d, want 250000", result.Amount)
	}
}