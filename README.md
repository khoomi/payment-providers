# payment-providers

Go client for payment gateways in Africa. [Khoomi](https://khoomi.com) uses it in production.

Paystack and Flutterwave are included. Other gateways implement the same `Provider` interface — see [CONTRIBUTING.md](CONTRIBUTING.md).

## Install

```bash
go get github.com/khoomi/payment-providers
```

## Example

```go
import (
    payproviders "github.com/khoomi/payment-providers"
    "github.com/khoomi/payment-providers/paystack"
    "github.com/khoomi/payment-providers/flutterwave"
)

mgr := payproviders.NewManager(payproviders.NamePaystack)
mgr.Register(paystack.New(paystack.Config{SecretKey: os.Getenv("PAYSTACK_SECRET_KEY")}))
mgr.Register(flutterwave.New(flutterwave.Config{
    SecretKey:   os.Getenv("FLUTTERWAVE_SECRET_KEY"),
    WebhookHash: os.Getenv("FLUTTERWAVE_WEBHOOK_HASH"),
}))

p, _ := mgr.Get(payproviders.NamePaystack)
result, err := p.Initialize(ctx, payproviders.InitRequest{
    Amount:    500000, // kobo
    Email:     "buyer@example.com",
    Reference: "order_ref_123",
})
```

## Included providers

| Provider | Init / Verify | Banks / Resolve | Webhook |
|----------|---------------|-----------------|---------|
| Paystack | Yes | Yes | HMAC-SHA512 (`x-paystack-signature`) |
| Flutterwave | Yes | Yes | `verif-hash` header |

## Webhook env

- **Paystack** — `PAYSTACK_SECRET_KEY` for API calls and webhook HMAC
- **Flutterwave** — `FLUTTERWAVE_SECRET_KEY` for API calls; `FLUTTERWAVE_WEBHOOK_HASH` for webhooks (from the Flutterwave dashboard, not the API secret)

## Structure

```
payment-providers/
├── client.go
├── types.go
├── manager.go
├── paystack/
└── flutterwave/
```

## Payment statuses

`PaymentStatus` covers `success`, `failed`, `pending`, `processing`, `abandoned`, `reversed`, `cancelled`, and `unknown`. Each provider maps its own API strings through `ParsePaymentStatus`.

## License

[MIT](LICENSE)