package paystack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	payproviders "github.com/khoomi/payment-providers"
	"github.com/khoomi/payment-providers/paystack"
)

func TestRetryRefundWithCustomerDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refund/retry_with_customer_details/4128514" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": true,
			"message": "Refund retry initiated",
			"data": {
				"id": 4128514,
				"amount": 250000,
				"currency": "NGN",
				"status": "processing"
			}
		}`))
	}))
	defer server.Close()

	provider := paystack.New(paystack.Config{
		SecretKey:  "test_secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	result, err := provider.RetryRefundWithCustomerDetails(context.Background(), payproviders.RefundRetryRequest{
		RefundID:      4128514,
		Currency:      payproviders.CurrencyNGN,
		AccountNumber: "0123456789",
		BankID:        "9",
	})
	if err != nil {
		t.Fatalf("RetryRefundWithCustomerDetails: %v", err)
	}
	if result.Status != payproviders.RefundStatusProcessing {
		t.Fatalf("status %q, want processing", result.Status)
	}
}