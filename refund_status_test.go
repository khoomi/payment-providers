package payproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payproviders"
)

func TestParseRefundEventType(t *testing.T) {
	cases := []struct {
		event string
		want  payproviders.RefundStatus
	}{
		{"refund.pending", payproviders.RefundStatusPending},
		{"refund.processing", payproviders.RefundStatusProcessing},
		{"refund.needs-attention", payproviders.RefundStatusNeedsAttention},
		{"refund.failed", payproviders.RefundStatusFailed},
		{"refund.processed", payproviders.RefundStatusProcessed},
		{"charge.success", payproviders.RefundStatusUnknown},
	}

	for _, tc := range cases {
		if got := payproviders.ParseRefundEventType(tc.event); got != tc.want {
			t.Fatalf("ParseRefundEventType(%q) = %q, want %q", tc.event, got, tc.want)
		}
	}
}