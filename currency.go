package paymentproviders

// Currency is an ISO 4217 African currency code supported by Khoomi gateways.
type Currency string

const (
	CurrencyNGN Currency = "NGN" // Nigerian naira
	CurrencyGHS Currency = "GHS" // Ghanaian cedi
	CurrencyKES Currency = "KES" // Kenyan shilling
	CurrencyZAR Currency = "ZAR" // South African rand
	CurrencyEGP Currency = "EGP" // Egyptian pound
	CurrencyMAD Currency = "MAD" // Moroccan dirham
	CurrencyTZS Currency = "TZS" // Tanzanian shilling
	CurrencyZMW Currency = "ZMW" // Zambian kwacha
	CurrencyAOA Currency = "AOA" // Angolan kwanza
	CurrencyBWP Currency = "BWP" // Botswana pula
	CurrencyCDF Currency = "CDF" // Congolese franc
	CurrencyETB Currency = "ETB" // Ethiopian birr
	CurrencyGMD Currency = "GMD" // Gambian dalasi
	CurrencyLSL Currency = "LSL" // Lesotho loti
	CurrencyMWK Currency = "MWK" // Malawian kwacha
	CurrencyMZN Currency = "MZN" // Mozambican metical
	CurrencyNAD Currency = "NAD" // Namibian dollar
	CurrencySCR Currency = "SCR" // Seychellois rupee
	CurrencySDG Currency = "SDG" // Sudanese pound
	CurrencySOS Currency = "SOS" // Somali shilling
	CurrencySZL Currency = "SZL" // Eswatini lilangeni

	// Currencies with no fractional minor unit
	CurrencyBIF Currency = "BIF" // Burundian franc
	CurrencyDJF Currency = "DJF" // Djiboutian franc
	CurrencyGNF Currency = "GNF" // Guinean franc
	CurrencyKMF Currency = "KMF" // Comorian franc
	CurrencyRWF Currency = "RWF" // Rwandan franc
	CurrencyUGX Currency = "UGX" // Ugandan shilling
	CurrencyXAF Currency = "XAF" // Central African CFA franc
	CurrencyXOF Currency = "XOF" // West African CFA franc

	// Currencies with three decimal places
	CurrencyLYD Currency = "LYD" // Libyan dinar
	CurrencyTND Currency = "TND" // Tunisian dinar
)

const DefaultCurrency = CurrencyNGN

func (c Currency) String() string {
	return string(c)
}

func (c Currency) OrDefault() Currency {
	if c == "" {
		return DefaultCurrency
	}
	return c
}

func ParseCurrency(raw string) Currency {
	return Currency(raw)
}