package paymentproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payment-providers"
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
		{"queued", payproviders.PaymentStatusPending},
		{"processing", payproviders.PaymentStatusProcessing},
		{"ongoing", payproviders.PaymentStatusProcessing},
		{"abandoned", payproviders.PaymentStatusAbandoned},
		{"reversed", payproviders.PaymentStatusReversed},
		{"cancelled", payproviders.PaymentStatusCancelled},
		{"canceled", payproviders.PaymentStatusCancelled},
		{"", payproviders.PaymentStatusUnknown},
		{"weird", payproviders.PaymentStatusUnknown},
	}

	for _, tt := range tests {
		if got := payproviders.ParsePaymentStatus(tt.raw); got != tt.want {
			t.Fatalf("ParsePaymentStatus(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestPaymentStatus_IsTerminal(t *testing.T) {
	terminal := []payproviders.PaymentStatus{
		payproviders.PaymentStatusSuccess,
		payproviders.PaymentStatusFailed,
		payproviders.PaymentStatusAbandoned,
		payproviders.PaymentStatusReversed,
		payproviders.PaymentStatusCancelled,
	}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Fatalf("%q should be terminal", s)
		}
	}

	nonTerminal := []payproviders.PaymentStatus{
		payproviders.PaymentStatusPending,
		payproviders.PaymentStatusProcessing,
		payproviders.PaymentStatusUnknown,
	}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Fatalf("%q should not be terminal", s)
		}
	}
}