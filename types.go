package paymentproviders

import "time"

type Name string

const (
	NamePaystack    Name = "paystack"
	NameFlutterwave Name = "flutterwave"
)

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

// InitRequest amount fields are expressed in the currency's minor units
// (e.g. kobo for NGN). Use MinorUnitExponent to resolve the scale per currency.
type InitRequest struct {
	Amount      int64
	Email       string
	Metadata    map[string]any
	CallbackURL string
	Currency    Currency
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
	Currency        Currency
	PaidAt          time.Time
	GatewayResponse string
	Raw             map[string]any
}

type WebhookKind string

const (
	WebhookKindPayment WebhookKind = "payment"
	WebhookKindRefund  WebhookKind = "refund"
)

type WebhookEvent struct {
	Kind                 WebhookKind
	EventType            string
	Reference            string
	CustomReference      string
	TransactionReference string
	// RefundID is Paystack's numeric refund id (retry API path param).
	RefundID int64
	// RefundReference is Paystack's TRF_* refund reference when present.
	RefundReference   string
	Amount            int64
	Status            PaymentStatus
	RefundStatus      RefundStatus
	GatewayResponse   string
	PaidAt            time.Time
	RawData           map[string]any
}

func (e *WebhookEvent) IsRefund() bool {
	return e != nil && e.Kind == WebhookKindRefund
}

type Bank struct {
	ID       int    `json:"id"`
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

// RefundStatus is the normalized refund outcome across payment gateways.
type RefundStatus string

const (
	RefundStatusPending        RefundStatus = "pending"
	RefundStatusProcessing     RefundStatus = "processing"
	RefundStatusProcessed      RefundStatus = "processed"
	RefundStatusFailed         RefundStatus = "failed"
	RefundStatusNeedsAttention RefundStatus = "needs_attention"
	RefundStatusUnknown        RefundStatus = "unknown"
)

func ParseRefundStatus(raw string) RefundStatus {
	switch raw {
	case "pending", "queued":
		return RefundStatusPending
	case "processing", "ongoing":
		return RefundStatusProcessing
	case "processed", "success", "successful", "completed":
		return RefundStatusProcessed
	case "failed":
		return RefundStatusFailed
	case "needs-attention", "needs_attention":
		return RefundStatusNeedsAttention
	default:
		return RefundStatusUnknown
	}
}

// ParseRefundEventType maps Paystack refund.* webhook event names to a normalized status.
// Event type is authoritative when present; see https://paystack.com/docs/payments/refunds/
func ParseRefundEventType(eventType string) RefundStatus {
	switch eventType {
	case "refund.pending":
		return RefundStatusPending
	case "refund.processing":
		return RefundStatusProcessing
	case "refund.needs-attention":
		return RefundStatusNeedsAttention
	case "refund.failed":
		return RefundStatusFailed
	case "refund.processed":
		return RefundStatusProcessed
	default:
		return RefundStatusUnknown
	}
}

func (s RefundStatus) IsSuccessful() bool {
	return s == RefundStatusProcessed
}

// IsAccepted reports whether the gateway accepted the refund request.
// Paystack often returns "pending" while the refund is queued asynchronously.
func (s RefundStatus) IsAccepted() bool {
	switch s {
	case RefundStatusProcessed, RefundStatusPending, RefundStatusProcessing:
		return true
	default:
		return false
	}
}

type RefundRequest struct {
	TransactionReference string
	GatewayTransactionID int64
	Amount               int64
	Currency             Currency
	CustomerNote         string
}

type RefundResult struct {
	Reference string
	Status    RefundStatus
	Amount    int64
	Currency  Currency
	Raw       map[string]any
}

// RefundRetryRequest supplies buyer bank details for a Paystack refund in
// needs-attention status. See POST /refund/retry_with_customer_details/{id}.
type RefundRetryRequest struct {
	RefundID      int64
	Currency      Currency
	AccountNumber string
	BankID        string
}