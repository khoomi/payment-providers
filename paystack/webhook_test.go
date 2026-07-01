package paystack_test

import (
	"testing"

	payproviders "github.com/khoomi/payment-providers"
	"github.com/khoomi/payment-providers/paystack"
)

func TestParseWebhook_ChargeSuccess(t *testing.T) {
	payload := []byte(`{
		"event": "charge.success",
		"data": {
			"reference": "PSK_ref_123",
			"amount": 500000,
			"status": "success",
			"gateway_response": "Approved",
			"paid_at": "2026-01-15T10:00:00Z",
			"metadata": {"reference": "KHM_custom_ref"}
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.EventType != "charge.success" {
		t.Fatalf("event type %q, want charge.success", event.EventType)
	}
	if event.Reference != "PSK_ref_123" {
		t.Fatalf("reference %q, want PSK_ref_123", event.Reference)
	}
	if event.CustomReference != "KHM_custom_ref" {
		t.Fatalf("custom reference %q, want KHM_custom_ref", event.CustomReference)
	}
	if event.Amount != 500000 {
		t.Fatalf("amount %d, want 500000", event.Amount)
	}
	if event.Status != payproviders.PaymentStatusSuccess {
		t.Fatalf("status %q, want success", event.Status)
	}
}