package flutterwave_test

import (
	"testing"

	payproviders "github.com/khoomi/payproviders"
	"github.com/khoomi/payproviders/flutterwave"
)

func TestParseWebhook_ChargeCompleted(t *testing.T) {
	payload := []byte(`{
		"event": "charge.completed",
		"data": {
			"tx_ref": "KHM_ref_456",
			"amount": 2500.50,
			"status": "successful",
			"processor_response": "Approved",
			"created_at": "2026-01-15T10:00:00Z",
			"meta": {"reference": "KHM_custom_ref"}
		}
	}`)

	p := flutterwave.New(flutterwave.Config{SecretKey: "test", WebhookHash: "hash"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.EventType != "charge.completed" {
		t.Fatalf("event type %q, want charge.completed", event.EventType)
	}
	if event.Reference != "KHM_ref_456" {
		t.Fatalf("reference %q, want KHM_ref_456", event.Reference)
	}
	if event.CustomReference != "KHM_custom_ref" {
		t.Fatalf("custom reference %q, want KHM_custom_ref", event.CustomReference)
	}
	if event.Amount != 250050 {
		t.Fatalf("amount %d, want 250050 minor units", event.Amount)
	}
	if event.Status != payproviders.PaymentStatusSuccess {
		t.Fatalf("status %q, want success", event.Status)
	}
}

func TestValidateWebhookSignature(t *testing.T) {
	p := flutterwave.New(flutterwave.Config{SecretKey: "api-secret", WebhookHash: "webhook-hash"})
	if !p.ValidateWebhookSignature(nil, nil, "webhook-hash") {
		t.Fatal("expected valid webhook hash")
	}
	if p.ValidateWebhookSignature(nil, nil, "wrong-hash") {
		t.Fatal("expected invalid webhook hash")
	}
}