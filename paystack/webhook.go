package paystack

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

func validateWebhookSignature(secretKey string, payload []byte, signature string) bool {
	hash := hmac.New(sha512.New, []byte(secretKey))
	hash.Write(payload)
	expected := hex.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}