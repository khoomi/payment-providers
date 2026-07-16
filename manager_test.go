package payproviders_test

import (
	"context"
	"testing"

	payproviders "github.com/khoomi/payproviders"
)

type stubProvider struct {
	name payproviders.Name
}

func (s stubProvider) Name() payproviders.Name { return s.name }
func (s stubProvider) Initialize(context.Context, payproviders.InitRequest) (*payproviders.InitResult, error) {
	return nil, nil
}
func (s stubProvider) Verify(context.Context, string) (*payproviders.VerifyResult, error) {
	return nil, nil
}
func (s stubProvider) Refund(context.Context, payproviders.RefundRequest) (*payproviders.RefundResult, error) {
	return nil, nil
}
func (s stubProvider) GetRefund(context.Context, string) (*payproviders.RefundResult, error) {
	return nil, nil
}
func (s stubProvider) RetryRefundWithCustomerDetails(context.Context, payproviders.RefundRetryRequest) (*payproviders.RefundResult, error) {
	return nil, nil
}
func (s stubProvider) GetBanks(context.Context) ([]payproviders.Bank, error) { return nil, nil }
func (s stubProvider) ValidateAccount(context.Context, string, string) (*payproviders.AccountValidation, error) {
	return nil, nil
}
func (s stubProvider) ValidateWebhookSignature(context.Context, []byte, string) bool { return true }
func (s stubProvider) ParseWebhook([]byte) (*payproviders.WebhookEvent, error) { return nil, nil }

func TestManager_GetProvider(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)
	mgr.Register(stubProvider{name: payproviders.NamePaystack})
	mgr.Register(stubProvider{name: payproviders.NameFlutterwave})

	got, err := mgr.Get(payproviders.NamePaystack)
	if err != nil || got.Name() != payproviders.NamePaystack {
		t.Fatalf("Get(paystack): %v", err)
	}

	got, err = mgr.Get(payproviders.NameFlutterwave)
	if err != nil || got.Name() != payproviders.NameFlutterwave {
		t.Fatalf("Get(flutterwave): %v", err)
	}

	// Empty name uses default.
	got, err = mgr.Get("")
	if err != nil || got.Name() != payproviders.NamePaystack {
		t.Fatalf("Get(\"\") should use default paystack, got %v err %v", got, err)
	}
}

func TestManager_GetMissingProvider(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)
	mgr.Register(stubProvider{name: payproviders.NamePaystack})
	// Must not silently fall back to paystack when flutterwave was requested.
	if _, err := mgr.Get(payproviders.NameFlutterwave); err == nil {
		t.Fatal("expected error when flutterwave is not registered")
	}
	if _, err := mgr.Get(payproviders.NameFlutterwave); err == nil {
		t.Fatal("expected error when no providers registered for name")
	}
}