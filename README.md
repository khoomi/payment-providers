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

## Using in Khoomi

Khoomi holds a `Manager` with Paystack as the default provider. Flutterwave is registered when configured. The payment service resolves providers by name — no second wrapper layer.

### 1. Startup — register providers

```go
mgr := payproviders.NewManager(payproviders.NamePaystack)

mgr.Register(paystack.New(paystack.Config{
    SecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
}))
if os.Getenv("FLUTTERWAVE_SECRET_KEY") != "" {
    mgr.Register(flutterwave.New(flutterwave.Config{
        SecretKey:   os.Getenv("FLUTTERWAVE_SECRET_KEY"),
        WebhookHash: os.Getenv("FLUTTERWAVE_WEBHOOK_HASH"),
    }))
}
```

### 2. Checkout — initialize a charge

When a buyer pays, Khoomi resolves the gateway and calls `Initialize` with the amount in kobo and a unique reference:

```go
provider, _ := mgr.Get(payproviders.NamePaystack)

initResult, err := provider.Initialize(ctx, payproviders.InitRequest{
    Amount:      amountKobo,
    Email:       buyerEmail,
    Metadata:    map[string]any{"order_id": orderID},
    CallbackURL: callbackURL,
    Currency:    payproviders.CurrencyNGN,
    Reference:   reference,
})
// → redirect buyer to initResult.AuthorizationURL
```

### 3. Verify — confirm payment status

After redirect or polling, Khoomi calls `Verify` and maps the normalized status to its own transaction states:

```go
provider, _ := mgr.Get(providerName)
verifyResult, err := provider.Verify(ctx, reference)

switch verifyResult.Status {
case payproviders.PaymentStatusSuccess:
    // mark paid, credit seller wallet
case payproviders.PaymentStatusFailed:
    // mark failed
case payproviders.PaymentStatusAbandoned:
    // mark abandoned
}
```

### 4. Webhooks — verify, parse, process

A single webhook endpoint handles both gateways. Khoomi detects the provider from the signature header, then delegates to the module:

```go
var providerName payproviders.Name
var signature string

if signature = r.Header.Get("x-paystack-signature"); signature != "" {
    providerName = payproviders.NamePaystack
} else if signature = r.Header.Get("verif-hash"); signature != "" {
    providerName = payproviders.NameFlutterwave
}

provider, _ := mgr.Get(providerName)

if !provider.ValidateWebhookSignature(ctx, body, signature) {
    return // 401
}
event, err := provider.ParseWebhook(body)

if event.IsRefund() {
    switch event.RefundStatus {
    case payproviders.RefundStatusPending, payproviders.RefundStatusProcessing:
        // acknowledged — wait for a terminal webhook
    case payproviders.RefundStatusProcessed:
        // complete refund in your ledger
    case payproviders.RefundStatusNeedsAttention:
        // collect buyer bank details and call RetryRefundWithCustomerDetails
    case payproviders.RefundStatusFailed:
        // mark refund failed; Paystack credits the merchant
    }
    return
}

switch event.Status {
case payproviders.PaymentStatusSuccess:
    // fulfil order, credit wallet
case payproviders.PaymentStatusFailed, payproviders.PaymentStatusAbandoned:
    // release inventory, notify buyer
}
```

### 5. Seller payouts — banks and account resolution

Bank lists and account name resolution for seller payout setup use the default provider (Paystack). Results are cached:

```go
provider := mgr.Default()

banks, err := provider.GetBanks(ctx)
result, err := provider.ValidateAccount(ctx, accountNumber, bankCode)
// result.AccountName → confirm before saving payout details
```

### 6. Refunds — initiate, webhook, and retry

Amounts are in the currency's **minor units** (e.g. kobo for NGN). Use `Currency` and the conversion helpers (`AmountForGateway`, `MinorFromGatewayAmount`) so each gateway gets the right scale.

**Initiate a refund**

```go
provider, _ := mgr.Get(payproviders.NamePaystack)

result, err := provider.Refund(ctx, payproviders.RefundRequest{
    TransactionReference: txRef,       // Paystack charge reference
    Amount:               amountKobo, // 0 = full refund
    Currency:             payproviders.CurrencyNGN,
    CustomerNote:         "Order cancelled",
})
```

Paystack needs `TransactionReference`. Flutterwave needs `GatewayTransactionID` instead.

**Do not treat sync `pending` as completion.** Paystack often returns `pending` while the refund is queued. Use `RefundStatus.IsAccepted()` to know the gateway accepted the request, and `RefundStatus.IsSuccessful()` only when the refund is actually done (`processed`).

```go
if result.Status.IsAccepted() {
    // request accepted — keep internal status as processing
}
if result.Status.IsSuccessful() {
    // rare on sync response; usually arrives via webhook
}
```

**Refund webhooks (Paystack)**

`ParseWebhook` sets `Kind: WebhookKindRefund` for `refund.*` events. Prefer `ParseRefundEventType(event.EventType)` — the event name is authoritative:

| Paystack event | Normalized status |
|----------------|-------------------|
| `refund.pending` | `pending` |
| `refund.processing` | `processing` |
| `refund.needs-attention` | `needs_attention` |
| `refund.failed` | `failed` |
| `refund.processed` | `processed` |

Parsed refund webhooks expose:

- `RefundID` — numeric id (retry API path param)
- `RefundReference` — `TRF_*` reference when present
- `TransactionReference` — original charge reference

**Retry after `needs-attention` (Paystack only)**

When Paystack cannot return funds to the original payment method, collect the buyer's bank details and retry:

```go
banks, _ := provider.GetBanks(ctx)
// resolve bank_id from bank code via Bank.ID

result, err := provider.RetryRefundWithCustomerDetails(ctx, payproviders.RefundRetryRequest{
    RefundID:      event.RefundID, // numeric, not TRF_*
    Currency:      payproviders.CurrencyNGN,
    AccountNumber: accountNumber,
    BankID:        bankID,         // Paystack numeric bank id as string
})
```

Flutterwave returns an error from `RetryRefundWithCustomerDetails` — that flow is Paystack-specific.

### Environment variables

| Variable | Provider | Used for |
|----------|----------|----------|
| `PAYSTACK_SECRET_KEY` | Paystack | API calls and webhook HMAC (`x-paystack-signature`) |
| `FLUTTERWAVE_SECRET_KEY` | Flutterwave | API calls |
| `FLUTTERWAVE_WEBHOOK_HASH` | Flutterwave | Webhook verification (`verif-hash` header) |

## Included providers

| Provider | Init / Verify | Refund | Retry (needs-attention) | Banks / Resolve | Webhook |
|----------|---------------|--------|-------------------------|-----------------|---------|
| Paystack | Yes | Yes | Yes | Yes | HMAC-SHA512 (`x-paystack-signature`) |
| Flutterwave | Yes | Yes | No | Yes | `verif-hash` header |

## Payment statuses

`PaymentStatus` covers `success`, `failed`, `pending`, `processing`, `abandoned`, `reversed`, `cancelled`, and `unknown`. Each provider maps its own API strings through `ParsePaymentStatus`.

## Refund statuses

`RefundStatus` covers `pending`, `processing`, `processed`, `failed`, `needs_attention`, and `unknown`.

- `IsAccepted()` — gateway queued the refund (`pending`, `processing`, or `processed`)
- `IsSuccessful()` — refund completed (`processed` only)

Paystack refund webhooks should be routed with `ParseRefundEventType`. API response strings use `ParseRefundStatus`.

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