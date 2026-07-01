package paymentproviders_test

import (
	"testing"

	payproviders "github.com/khoomi/payment-providers"
)

func TestCurrency_OrDefault(t *testing.T) {
	if got := payproviders.Currency("").OrDefault(); got != payproviders.DefaultCurrency {
		t.Fatalf("got %q, want %q", got, payproviders.DefaultCurrency)
	}
	if got := payproviders.CurrencyGHS.OrDefault(); got != payproviders.CurrencyGHS {
		t.Fatalf("got %q, want GHS", got)
	}
}

func TestParseCurrency(t *testing.T) {
	if got := payproviders.ParseCurrency("NGN"); got != payproviders.CurrencyNGN {
		t.Fatalf("got %q, want NGN", got)
	}
}