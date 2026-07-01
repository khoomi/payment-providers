package paymentproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payment-providers"
)

func TestMinorUnitExponent(t *testing.T) {
	if got := payproviders.MinorUnitExponent("NGN"); got != 2 {
		t.Fatalf("NGN exponent %d, want 2", got)
	}
	if got := payproviders.MinorUnitExponent("JPY"); got != 0 {
		t.Fatalf("JPY exponent %d, want 0", got)
	}
	if got := payproviders.MinorUnitExponent("KWD"); got != 3 {
		t.Fatalf("KWD exponent %d, want 3", got)
	}
}

func TestMinorToMajorUnit(t *testing.T) {
	if got := payproviders.MinorToMajorUnit(250000, "NGN"); got != 2500 {
		t.Fatalf("got %v, want 2500", got)
	}
}

func TestMajorUnitToMinor(t *testing.T) {
	if got := payproviders.MajorUnitToMinor(2500.50, "NGN"); got != 250050 {
		t.Fatalf("got %d, want 250050", got)
	}
}

func TestFormatMajorUnitAmount(t *testing.T) {
	if got := payproviders.FormatMajorUnitAmount(250000, "NGN"); got != "2500.00" {
		t.Fatalf("got %q, want 2500.00", got)
	}
}

func TestAmountForGateway(t *testing.T) {
	if got := payproviders.AmountForGateway(250000, "NGN", payproviders.NamePaystack); got != int64(250000) {
		t.Fatalf("paystack amount %v, want 250000", got)
	}
	if got := payproviders.AmountForGateway(250000, "NGN", payproviders.NameFlutterwave); got != "2500.00" {
		t.Fatalf("flutterwave amount %v, want 2500.00", got)
	}
}

func TestMinorFromGatewayAmount(t *testing.T) {
	if got := payproviders.MinorFromGatewayAmount(2500.50, "NGN", payproviders.NameFlutterwave); got != 250050 {
		t.Fatalf("got %d, want 250050", got)
	}
	if got := payproviders.MinorFromGatewayAmount(250000, "NGN", payproviders.NamePaystack); got != 250000 {
		t.Fatalf("got %d, want 250000", got)
	}
}