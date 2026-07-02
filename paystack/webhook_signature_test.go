package paystack_test

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/khoomi/payproviders/paystack"
)

func signPaystackPayload(secret string, payload []byte) string {
	hash := hmac.New(sha512.New, []byte(secret))
	hash.Write(payload)
	return hex.EncodeToString(hash.Sum(nil))
}

func TestValidateWebhookSignature(t *testing.T) {
	secret := "sk_test_secret"
	payload := []byte(`{"event":"charge.success"}`)
	validSig := signPaystackPayload(secret, payload)

	p := paystack.New(paystack.Config{SecretKey: secret})

	if !p.ValidateWebhookSignature(nil, payload, validSig) {
		t.Fatal("expected valid paystack webhook signature")
	}
	if p.ValidateWebhookSignature(nil, payload, "invalid") {
		t.Fatal("expected invalid paystack webhook signature")
	}
}