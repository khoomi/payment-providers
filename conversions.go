package paymentproviders

import (
	"math"
	"strconv"
)

// MinorUnitExponent returns the number of decimal places in a currency's major unit.
// Defaults to 2 when the currency is unknown.
func MinorUnitExponent(currency string) int {
	switch currency {
	case "BIF", "CLP", "DJF", "GNF", "ISK", "JPY", "KMF", "KRW", "PYG", "RWF", "UGX", "VND", "VUV", "XAF", "XOF", "XPF":
		return 0
	case "BHD", "IQD", "JOD", "KWD", "LYD", "OMR", "TND":
		return 3
	default:
		return 2
	}
}

// MinorToMajorUnit converts an amount in minor units to major units.
func MinorToMajorUnit(minor int64, currency string) float64 {
	exp := MinorUnitExponent(currency)
	return float64(minor) / math.Pow10(exp)
}

// MajorUnitToMinor converts an amount in major units to minor units.
func MajorUnitToMinor(major float64, currency string) int64 {
	exp := MinorUnitExponent(currency)
	return int64(math.Round(major * math.Pow10(exp)))
}

// FormatMajorUnitAmount formats minor units as a decimal string for gateways
// that expect major-unit amounts (e.g. Flutterwave).
func FormatMajorUnitAmount(minor int64, currency string) string {
	exp := MinorUnitExponent(currency)
	return strconv.FormatFloat(MinorToMajorUnit(minor, currency), 'f', exp, 64)
}

// AmountForGateway returns the amount value to send to a gateway API.
// Paystack expects minor units; Flutterwave expects a major-unit decimal string.
func AmountForGateway(minor int64, currency string, gateway Name) any {
	if gateway == NameFlutterwave {
		return FormatMajorUnitAmount(minor, currency)
	}
	return minor
}

// MinorFromGatewayAmount parses a gateway response amount into minor units.
func MinorFromGatewayAmount(major float64, currency string, gateway Name) int64 {
	if gateway == NameFlutterwave {
		return MajorUnitToMinor(major, currency)
	}
	return int64(major)
}