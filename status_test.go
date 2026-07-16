package payproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payproviders"
)

func TestParsePaymentStatus(t *testing.T) {
	tests := []struct {
		raw  string
		want payproviders.PaymentStatus
	}{
		{"success", payproviders.PaymentStatusSuccess},
		{"successful", payproviders.PaymentStatusSuccess},
		{"failed", payproviders.PaymentStatusFailed},
		{"pending", payproviders.PaymentStatusPending},
		{"processing", payproviders.PaymentStatusProcessing},
		{"abandoned", payproviders.PaymentStatusAbandoned},
		{"reversed", payproviders.PaymentStatusReversed},
		{"cancelled", payproviders.PaymentStatusCancelled},
		{"weird", payproviders.PaymentStatusUnknown},
	}

	for _, tt := range tests {
		if got := payproviders.ParsePaymentStatus(tt.raw); got != tt.want {
			t.Fatalf("ParsePaymentStatus(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestParseRefundStatus(t *testing.T) {
	tests := []struct {
		raw  string
		want payproviders.RefundStatus
	}{
		{"pending", payproviders.RefundStatusPending},
		{"pending-momo", payproviders.RefundStatusPending},
		{"processed", payproviders.RefundStatusProcessed},
		{"completed", payproviders.RefundStatusProcessing}, // FW: initiated, not disbursed
		{"completed-bank-transfer", payproviders.RefundStatusProcessed},
		{"completed-momo", payproviders.RefundStatusProcessed},
		{"failed", payproviders.RefundStatusFailed},
		{"weird", payproviders.RefundStatusUnknown},
	}

	for _, tt := range tests {
		if got := payproviders.ParseRefundStatus(tt.raw); got != tt.want {
			t.Fatalf("ParseRefundStatus(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}