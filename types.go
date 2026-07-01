package paymentproviders

import "time"

type Name string

const (
	NamePaystack    Name = "paystack"
	NameFlutterwave Name = "flutterwave"
)

const DefaultCurrency = "NGN"

// PaymentStatus is the normalized charge outcome across payment gateways.
type PaymentStatus string

const (
	PaymentStatusSuccess    PaymentStatus = "success"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusAbandoned  PaymentStatus = "abandoned"
	PaymentStatusReversed   PaymentStatus = "reversed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusUnknown    PaymentStatus = "unknown"
)

func ParsePaymentStatus(raw string) PaymentStatus {
	switch raw {
	case "success", "successful":
		return PaymentStatusSuccess
	case "failed":
		return PaymentStatusFailed
	case "pending", "queued":
		return PaymentStatusPending
	case "processing", "ongoing":
		return PaymentStatusProcessing
	case "abandoned":
		return PaymentStatusAbandoned
	case "reversed":
		return PaymentStatusReversed
	case "cancelled", "canceled":
		return PaymentStatusCancelled
	default:
		return PaymentStatusUnknown
	}
}

func (s PaymentStatus) IsTerminal() bool {
	switch s {
	case PaymentStatusSuccess, PaymentStatusFailed, PaymentStatusAbandoned, PaymentStatusReversed, PaymentStatusCancelled:
		return true
	default:
		return false
	}
}

type InitRequest struct {
	Amount      int64
	Email       string
	Metadata    map[string]any
	CallbackURL string
	Currency    string
	Reference   string
}

type InitResult struct {
	AuthorizationURL string
	AccessCode       string
	Reference        string
}

type VerifyResult struct {
	Status          PaymentStatus
	Reference       string
	Amount          int64
	Currency        string
	PaidAt          time.Time
	GatewayResponse string
	Raw             map[string]any
}

type WebhookEvent struct {
	EventType       string
	Reference       string
	CustomReference string
	Amount          int64
	Status          PaymentStatus
	GatewayResponse string
	PaidAt          time.Time
	RawData         map[string]any
}

type Bank struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Code     string `json:"code"`
	LongCode string `json:"longcode"`
	Type     string `json:"type"`
	Active   bool   `json:"active"`
}

type AccountValidation struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankID        int    `json:"bank_id"`
}