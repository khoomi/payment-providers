package paymentproviders

import "context"

type Provider interface {
	Name() Name
	Initialize(ctx context.Context, req InitRequest) (*InitResult, error)
	Verify(ctx context.Context, reference string) (*VerifyResult, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)
	GetBanks(ctx context.Context) ([]Bank, error)
	ValidateAccount(ctx context.Context, accountNumber, bankCode string) (*AccountValidation, error)
	ValidateWebhookSignature(ctx context.Context, payload []byte, signature string) bool
	ParseWebhook(payload []byte) (*WebhookEvent, error)
}