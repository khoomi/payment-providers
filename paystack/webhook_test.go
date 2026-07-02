package paystack_test

import (
	"testing"

	payproviders "github.com/khoomi/payproviders"
	"github.com/khoomi/payproviders/paystack"
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
	if event.Kind != payproviders.WebhookKindPayment {
		t.Fatalf("kind %q, want payment", event.Kind)
	}
}

func TestParseWebhook_RefundProcessed(t *testing.T) {
	payload := []byte(`{
		"event": "refund.processed",
		"data": {
			"id": 4128514,
			"amount": 250000,
			"currency": "NGN",
			"status": "processed",
			"transaction": {
				"reference": "PSK_ref_123"
			}
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.Kind != payproviders.WebhookKindRefund {
		t.Fatalf("kind %q, want refund", event.Kind)
	}
	if event.EventType != "refund.processed" {
		t.Fatalf("event type %q, want refund.processed", event.EventType)
	}
	if event.RefundID != 4128514 {
		t.Fatalf("refund id %d, want 4128514", event.RefundID)
	}
	if event.TransactionReference != "PSK_ref_123" {
		t.Fatalf("transaction reference %q, want PSK_ref_123", event.TransactionReference)
	}
	if event.RefundStatus != payproviders.RefundStatusProcessed {
		t.Fatalf("refund status %q, want processed", event.RefundStatus)
	}
}

func TestParseWebhook_RefundPending(t *testing.T) {
	payload := []byte(`{
		"event": "refund.pending",
		"data": {
			"id": 4128514,
			"amount": 250000,
			"status": "pending",
			"transaction_reference": "KHM_custom_ref"
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.Kind != payproviders.WebhookKindRefund {
		t.Fatalf("kind %q, want refund", event.Kind)
	}
	if event.RefundStatus != payproviders.RefundStatusPending {
		t.Fatalf("refund status %q, want pending", event.RefundStatus)
	}
	if event.TransactionReference != "KHM_custom_ref" {
		t.Fatalf("transaction reference %q, want KHM_custom_ref", event.TransactionReference)
	}
	if event.EventType != "refund.pending" {
		t.Fatalf("event type %q, want refund.pending", event.EventType)
	}
}

func TestParseWebhook_RefundProcessing(t *testing.T) {
	payload := []byte(`{
		"event": "refund.processing",
		"data": {
			"id": 4128514,
			"amount": 250000,
			"status": "processing",
			"transaction": {"reference": "PSK_ref_123"}
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.EventType != "refund.processing" {
		t.Fatalf("event type %q, want refund.processing", event.EventType)
	}
	if event.RefundStatus != payproviders.RefundStatusProcessing {
		t.Fatalf("refund status %q, want processing", event.RefundStatus)
	}
}

func TestParseWebhook_RefundNeedsAttention(t *testing.T) {
	payload := []byte(`{
		"event": "refund.needs-attention",
		"data": {
			"status": "needs-attention",
			"transaction_reference": "88bfa94509eb96aa9785641c26cc57cc",
			"refund_reference": "TRF_7jn17u9vkqm91efk",
			"amount": 5306,
			"currency": "NGN",
			"id": "123456",
			"customer_note": "Refund for transaction 88bfa94509eb96aa9785641c26cc57cc",
			"merchant_note": "Refund for transaction 88bfa94509eb96aa9785641c26cc57cc by paystack@email.com"
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.RefundStatus != payproviders.RefundStatusNeedsAttention {
		t.Fatalf("refund status %q, want needs_attention", event.RefundStatus)
	}
	if event.Reference != "TRF_7jn17u9vkqm91efk" {
		t.Fatalf("reference %q, want TRF_7jn17u9vkqm91efk", event.Reference)
	}
	if event.RefundID != 123456 {
		t.Fatalf("refund id %d, want 123456", event.RefundID)
	}
	if event.RefundReference != "TRF_7jn17u9vkqm91efk" {
		t.Fatalf("refund reference %q, want TRF_7jn17u9vkqm91efk", event.RefundReference)
	}
	if event.TransactionReference != "88bfa94509eb96aa9785641c26cc57cc" {
		t.Fatalf("transaction reference %q", event.TransactionReference)
	}
	if event.Amount != 5306 {
		t.Fatalf("amount %d, want 5306", event.Amount)
	}
}

func TestParseWebhook_RefundFailed(t *testing.T) {
	payload := []byte(`{
		"event": "refund.failed",
		"data": {
			"id": 4128514,
			"amount": 250000,
			"status": "failed",
			"merchant_note": "Processor rejected refund",
			"transaction": {"reference": "PSK_ref_123"}
		}
	}`)

	p := paystack.New(paystack.Config{SecretKey: "test"})
	event, err := p.ParseWebhook(payload)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.RefundStatus != payproviders.RefundStatusFailed {
		t.Fatalf("refund status %q, want failed", event.RefundStatus)
	}
	if event.GatewayResponse != "Processor rejected refund" {
		t.Fatalf("gateway response %q, want Processor rejected refund", event.GatewayResponse)
	}
}