# Contributing

## New provider

Implement `paymentproviders.Provider` in its own package. Copy `paystack/` or `flutterwave/` and work from there.

### Name

Add a constant in `types.go`:

```go
const (
    NamePaystack    Name = "paystack"
    NameFlutterwave Name = "flutterwave"
    NameYourGateway Name = "your_gateway"
)
```

### Package layout

```
yourgateway/
├── yourgateway.go
├── webhook.go
└── webhook_test.go
```

### Interface

```go
type Provider interface {
    Name() Name
    Initialize(ctx context.Context, req InitRequest) (*InitResult, error)
    Verify(ctx context.Context, reference string) (*VerifyResult, error)
    GetBanks(ctx context.Context) ([]Bank, error)
    ValidateAccount(ctx context.Context, accountNumber, bankCode string) (*AccountValidation, error)
    ValidateWebhookSignature(ctx context.Context, payload []byte, signature string) bool
    ParseWebhook(payload []byte) (*WebhookEvent, error)
}
```

- Amounts in kobo at the package boundary. Convert inside your package if the API uses major units.
- Webhook validation belongs in your package, not `client.go`.
- Return `InitResult`, `VerifyResult`, `WebhookEvent` — not gateway-specific structs on the interface.
- Extend `ParsePaymentStatus` if the gateway uses status strings we don't map yet.

### Config

```go
type Config struct {
    SecretKey  string
    HTTPClient *http.Client // optional
    Logger     *slog.Logger // optional
}

func New(cfg Config) *Provider { ... }
```

### Registration

```go
mgr := payproviders.NewManager(payproviders.NamePaystack)
mgr.Register(yourgateway.New(yourgateway.Config{SecretKey: "..."}))
```

`Manager.Get` falls back to the default provider if the requested one isn't registered.

### Tests

No `integration` tag needed. At minimum:

- `ParseWebhook` — valid payload, bad JSON
- `ValidateWebhookSignature` — pass and fail cases
- Amount conversion if the API doesn't use kobo

```bash
go test ./...
```

### PR checklist

- [ ] `Name` constant in `types.go`
- [ ] `var _ payproviders.Provider = (*Provider)(nil)`
- [ ] Webhook tests
- [ ] README row + env vars
- [ ] No imports from `khoomi/core`

## Other changes

Match the existing style. Keep PRs small. Include a test for bug fixes where it makes sense.

Unclear on a gateway's webhook format? Open an issue before writing a lot of code.