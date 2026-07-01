package paymentproviders_test

import (
	"context"
	"testing"

	payproviders "github.com/khoomi/payment-providers"
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
func (s stubProvider) GetBanks(context.Context) ([]payproviders.Bank, error) { return nil, nil }
func (s stubProvider) ValidateAccount(context.Context, string, string) (*payproviders.AccountValidation, error) {
	return nil, nil
}
func (s stubProvider) ValidateWebhookSignature(context.Context, []byte, string) bool { return true }
func (s stubProvider) ParseWebhook([]byte) (*payproviders.WebhookEvent, error) { return nil, nil }

func TestManager_RegisterAndGet(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)
	paystack := stubProvider{name: payproviders.NamePaystack}
	mgr.Register(paystack)

	got, err := mgr.Get(payproviders.NamePaystack)
	if err != nil {
		t.Fatalf("Get(paystack): %v", err)
	}
	if got.Name() != payproviders.NamePaystack {
		t.Fatalf("got provider %q, want paystack", got.Name())
	}
}

func TestManager_GetFallsBackToDefault(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)
	mgr.Register(stubProvider{name: payproviders.NamePaystack})

	got, err := mgr.Get(payproviders.NameFlutterwave)
	if err != nil {
		t.Fatalf("Get(flutterwave) fallback: %v", err)
	}
	if got.Name() != payproviders.NamePaystack {
		t.Fatalf("fallback provider %q, want paystack", got.Name())
	}
}

func TestManager_GetMissingProvider(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)

	_, err := mgr.Get(payproviders.NameFlutterwave)
	if err == nil {
		t.Fatal("expected error when no providers registered")
	}
}

func TestManager_DefaultAndHasProvider(t *testing.T) {
	mgr := payproviders.NewManager(payproviders.NamePaystack)
	mgr.Register(stubProvider{name: payproviders.NamePaystack})

	if !mgr.HasProvider(payproviders.NamePaystack) {
		t.Fatal("expected HasProvider(paystack) true")
	}
	if mgr.HasProvider(payproviders.NameFlutterwave) {
		t.Fatal("expected HasProvider(flutterwave) false")
	}
	if mgr.Default() == nil {
		t.Fatal("expected Default() to return paystack provider")
	}
}