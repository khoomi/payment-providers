package payproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payproviders"
)

func TestAmountForGateway(t *testing.T) {
	// Paystack takes minor units; Flutterwave takes major-unit strings.
	if got := payproviders.AmountForGateway(250000, payproviders.CurrencyNGN, payproviders.NamePaystack); got != int64(250000) {
		t.Fatalf("paystack amount %v, want 250000", got)
	}
	if got := payproviders.AmountForGateway(250000, payproviders.CurrencyNGN, payproviders.NameFlutterwave); got != "2500.00" {
		t.Fatalf("flutterwave amount %v, want 2500.00", got)
	}

	// XOF has no fractional minor unit.
	if got := payproviders.FormatMajorUnitAmount(1500, payproviders.CurrencyXOF); got != "1500" {
		t.Fatalf("XOF format %q, want 1500", got)
	}

	// Round-trip Flutterwave response parsing.
	if got := payproviders.MinorFromGatewayAmount(2500.50, payproviders.CurrencyNGN, payproviders.NameFlutterwave); got != 250050 {
		t.Fatalf("got %d, want 250050", got)
	}
}