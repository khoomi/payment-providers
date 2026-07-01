# payment-providers

Go client for payment gateways in Africa. [Khoomi](https://khoomi.com) uses it in production.

Paystack and Flutterwave are included. Gateways share a common `Provider` interface because init, verify, banks, and webhooks have the same shape across providers. See [CONTRIBUTING.md](CONTRIBUTING.md) to add another gateway.

## Install

```bash
go get github.com/khoomi/payment-providers
```

## Quick start

```go
import (
    payproviders "github.com/khoomi/payment-providers"
    "github.com/khoomi/payment-providers/paystack"
)

mgr := payproviders.NewManager(payproviders.NamePaystack)
mgr.Register(paystack.New(paystack.Config{SecretKey: os.Getenv("PAYSTACK_SECRET_KEY")}))

p, _ := mgr.Get(payproviders.NamePaystack)
result, err := p.Initialize(ctx, payproviders.InitRequest{
    Amount:    500000, // kobo
    Email:     "buyer@example.com",
    Reference: "order_ref_123",
})
```

## Using in Khoomi (core-api)

Khoomi pins this module in `go.mod` and passes a `payproviders.Manager` directly into `payment_service`. There is no second facade in `lib/` — the service layer owns provider selection.

### 1. Startup — register providers

`lib/container/services.go` builds the manager at container init. Paystack is the default; Flutterwave is registered when configured:

```go
// lib/container/services.go
paymentManager := payproviders.NewManager(payproviders.NamePaystack)
paymentManager.Register(paystack.New(paystack.Config{
    SecretKey: cfg.Paystack.SecretKey,
}))
if cfg.Flutterwave.SecretKey != "" {
    paymentManager.Register(flutterwave.New(flutterwave.Config{
        SecretKey:   cfg.Flutterwave.SecretKey,
        WebhookHash: cfg.Flutterwave.WebhookHash,
    }))
}

paymentService := payment.NewPaymentService(payment.PaymentServiceConfig{
    DB: db, Cache: runtime.Cache, EventBus: eventBus,
    WalletService: walletSvc, PlatformService: platformService,
    Outbox: outboxSvc, Manager: paymentManager,
})
```

### 2. Checkout — initialize a charge

When a buyer pays for an order, `payment_service` resolves the provider by name, then calls `Initialize` with the transaction reference and amount in kobo:

```go
// services/payment/payment_service.go
provider, err := ps.GetProvider(selectedProvider) // models.PaymentProvider → payproviders.Name

initResult, err := provider.Initialize(ctx, payproviders.InitRequest{
    Amount:      int64(tx.Amount),
    Email:       string(tx.CustomerEmail),
    Metadata:    tx.Provider.Metadata,
    CallbackURL: callbackURI.String(),
    Currency:    string(tx.Currency),
    Reference:   tx.Reference,
})
// → store AuthorizationURL, AccessCode, Reference on the transaction
```

The buyer is redirected to `initResult.AuthorizationURL`. Khoomi generates references like `KHM_<timestamp>_<id>` via `payment.GenerateReference()`.

### 3. Verify — confirm payment status

After redirect or on polling, Khoomi verifies with the provider and maps normalized statuses into domain transaction states:

```go
provider, _ := ps.GetProvider(paymentProvider)
verifyResult, err := provider.Verify(ctx, reference)

switch verifyResult.Status {
case payproviders.PaymentStatusSuccess:
    status = models.TransactionStatusSuccessful
case payproviders.PaymentStatusFailed:
    status = models.TransactionStatusFailed
case payproviders.PaymentStatusAbandoned:
    status = models.TransactionStatusAbandoned
// ...
}
```

### 4. Webhooks — verify, parse, process

A single webhook endpoint handles both gateways. The handler detects the provider from the signature header, then delegates to the module:

```go
// handlers/payment/payment_handler.go
// Detect provider from header
if c.GetHeader("x-paystack-signature") != "" {
    provider = models.PaymentProviderPaystack
} else if c.GetHeader("verif-hash") != "" {
    provider = models.PaymentProviderFlutterWave
}

// Verify + parse via module
if !self.paymentService.ValidateWebhookSignature(ctx, provider, body, signature) {
    // 401
}
paymentProvider, _ := self.paymentService.GetProvider(provider)
parsedEvent, err := paymentProvider.ParseWebhook(body)

// Process normalized event
self.paymentService.ProcessWebhookEvent(ctx, provider, parsedEvent)
```

`ProcessWebhookEvent` maps `payproviders.PaymentStatus` to order/wallet side effects (credit seller wallet on success, mark failed on terminal failure, etc.).

### 5. Seller payouts — banks and account resolution

Khoomi exposes bank lists and account name resolution for seller payout setup. These use the default provider (Paystack):

```go
// services/payment/payment_service.go
provider := ps.manager.Default()
banks, err := provider.GetBanks(ctx)           // cached in Redis
result, err := provider.ValidateAccount(ctx, accountNumber, bankCode)
```

`shop/payout_info_service` calls `ValidateAccount` when a seller adds a bank account. Results are exposed at `GET /api/banks` and `GET /api/banks/:code/accounts/:number`.

### Environment variables

| Variable | Provider | Used for |
|----------|----------|----------|
| `PAYSTACK_SECRET_KEY` | Paystack | API calls and webhook HMAC (`x-paystack-signature`) |
| `FLUTTERWAVE_SECRET_KEY` | Flutterwave | API calls |
| `FLUTTERWAVE_WEBHOOK_HASH` | Flutterwave | Webhook verification (`verif-hash` header) |

## Included providers

| Provider | Init / Verify | Banks / Resolve | Webhook |
|----------|---------------|-----------------|---------|
| Paystack | Yes | Yes | HMAC-SHA512 (`x-paystack-signature`) |
| Flutterwave | Yes | Yes | `verif-hash` header |

## Payment statuses

`PaymentStatus` covers `success`, `failed`, `pending`, `processing`, `abandoned`, `reversed`, `cancelled`, and `unknown`. Each provider maps its own API strings through `ParsePaymentStatus`.

## Structure

```
payment-providers/
├── client.go
├── types.go
├── manager.go
├── paystack/
└── flutterwave/
```

## License

[MIT](LICENSE)