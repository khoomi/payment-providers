package payproviders

import (
	"math"
	"strconv"
)

// MinorUnitExponent returns the number of decimal places in a currency's major unit.
// Covers African ISO 4217 codes Khoomi is likely to support; defaults to 2.
func MinorUnitExponent(currency Currency) int {
	switch currency {
	// No fractional minor unit
	case "BIF", // Burundian franc
		"DJF", // Djiboutian franc
		"GNF", // Guinean franc
		"KMF", // Comorian franc
		"RWF", // Rwandan franc
		"UGX", // Ugandan shilling
		"XAF", // Central African CFA franc
		"XOF": // West African CFA franc
		return 0
	// Three decimal places
	case "LYD", // Libyan dinar
		"TND": // Tunisian dinar
		return 3
	default:
		// NGN, GHS, KES, ZAR, EGP, MAD, TZS, ZMW, and most others
		return 2
	}
}

// MinorToMajorUnit converts an amount in minor units to major units.
func MinorToMajorUnit(minor int64, currency Currency) float64 {
	exp := MinorUnitExponent(currency)
	return float64(minor) / math.Pow10(exp)
}

// MajorUnitToMinor converts an amount in major units to minor units.
func MajorUnitToMinor(major float64, currency Currency) int64 {
	exp := MinorUnitExponent(currency)
	return int64(math.Round(major * math.Pow10(exp)))
}

// FormatMajorUnitAmount formats minor units as a decimal string for gateways
// that expect major-unit amounts (e.g. Flutterwave).
func FormatMajorUnitAmount(minor int64, currency Currency) string {
	exp := MinorUnitExponent(currency)
	return strconv.FormatFloat(MinorToMajorUnit(minor, currency), 'f', exp, 64)
}

// AmountForGateway returns the amount value to send to a gateway API.
// Paystack expects minor units; Flutterwave expects a major-unit decimal string.
func AmountForGateway(minor int64, currency Currency, gateway Name) any {
	if gateway == NameFlutterwave {
		return FormatMajorUnitAmount(minor, currency)
	}
	return minor
}

// MinorFromGatewayAmount parses a gateway response amount into minor units.
func MinorFromGatewayAmount(major float64, currency Currency, gateway Name) int64 {
	if gateway == NameFlutterwave {
		return MajorUnitToMinor(major, currency)
	}
	return int64(major)
}