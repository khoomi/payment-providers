package flutterwave_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khoomi/payproviders/flutterwave"
)

func TestVerify_PreservesPaymentMethodDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/verify_by_reference" ||
			r.URL.Query().Get("tx_ref") != "order-123" ||
			r.Method != http.MethodGet {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": "success",
			"message": "Transaction fetched successfully",
			"data": {
				"id": 5708,
				"status": "successful",
				"tx_ref": "order-123",
				"amount": 22628.50,
				"currency": "NGN",
				"created_at": "2026-07-24T12:00:00Z",
				"processor_response": "Approved",
				"payment_type": "card",
				"card": {
					"type": "MASTERCARD",
					"last_4digits": "1234",
					"issuer": "Example Bank"
				}
			}
		}`))
	}))
	defer server.Close()

	provider := flutterwave.New(flutterwave.Config{
		SecretKey:  "test_secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	result, err := provider.Verify(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if result.Raw["payment_type"] != "card" {
		t.Fatalf("payment_type %#v, want card", result.Raw["payment_type"])
	}

	card, ok := result.Raw["card"].(map[string]any)
	if !ok {
		t.Fatalf("card %#v, want object", result.Raw["card"])
	}
	if card["type"] != "MASTERCARD" {
		t.Fatalf("card type %#v, want MASTERCARD", card["type"])
	}
	if card["last_4digits"] != "1234" {
		t.Fatalf("last_4digits %#v, want 1234", card["last_4digits"])
	}
}
