package flutterwave

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	payproviders "github.com/khoomi/payment-providers"
)

const baseURL = "https://api.flutterwave.com/v3"

type Config struct {
	SecretKey   string
	WebhookHash string
	HTTPClient  *http.Client
	Logger      *slog.Logger
}

type Provider struct {
	client      *payproviders.HTTPClient
	webhookHash string
}

func New(cfg Config) *Provider {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	webhookHash := cfg.WebhookHash
	if webhookHash == "" {
		webhookHash = cfg.SecretKey
	}

	return &Provider{
		client: payproviders.NewHTTPClient(payproviders.ClientConfig{
			Name:       payproviders.NameFlutterwave,
			BaseURL:    baseURL,
			SecretKey:  cfg.SecretKey,
			HTTPClient: httpClient,
			Logger:     cfg.Logger,
		}),
		webhookHash: webhookHash,
	}
}

func (fws *Provider) Name() payproviders.Name {
	return payproviders.NameFlutterwave
}

func (fws *Provider) Initialize(ctx context.Context, req payproviders.InitRequest) (*payproviders.InitResult, error) {
	currency := req.Currency
	if currency == "" {
		currency = payproviders.DefaultCurrency
	}

	txRef := req.Reference
	if txRef == "" {
		if ref, ok := req.Metadata["reference"].(string); ok {
			txRef = ref
		}
	}
	if txRef == "" {
		return nil, errors.New("flutterwave tx_ref is required")
	}

	payload := map[string]any{
		"tx_ref":       txRef,
		"amount":       koboToMajorUnit(req.Amount),
		"currency":     currency,
		"redirect_url": req.CallbackURL,
		"customer": map[string]string{
			"email": req.Email,
		},
	}
	if len(req.Metadata) > 0 {
		payload["meta"] = req.Metadata
	}

	body, err := fws.client.Post(ctx, "/payments", payload)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Link string `json:"link"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if response.Status != "success" {
		return nil, fmt.Errorf("flutterwave initialization failed: %s", response.Message)
	}

	return &payproviders.InitResult{
		AuthorizationURL: response.Data.Link,
		Reference:        txRef,
	}, nil
}

func (fws *Provider) Verify(ctx context.Context, reference string) (*payproviders.VerifyResult, error) {
	endpoint := fmt.Sprintf("/transactions/verify_by_reference?tx_ref=%s", url.QueryEscape(reference))

	body, err := fws.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string  `json:"status"`
			TxRef     string  `json:"tx_ref"`
			Amount    float64 `json:"amount"`
			Currency  string  `json:"currency"`
			CreatedAt string  `json:"created_at"`
			Processor string  `json:"processor_response"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if response.Status != "success" {
		return nil, fmt.Errorf("flutterwave verification failed: %s", response.Message)
	}

	raw, _ := json.Marshal(response.Data)
	var rawMap map[string]any
	_ = json.Unmarshal(raw, &rawMap)

	var paidAt time.Time
	if response.Data.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, response.Data.CreatedAt); err == nil {
			paidAt = t
		}
	}

	return &payproviders.VerifyResult{
		Status:          payproviders.ParsePaymentStatus(response.Data.Status),
		Reference:       response.Data.TxRef,
		Amount:          majorUnitToKobo(response.Data.Amount),
		Currency:        response.Data.Currency,
		PaidAt:          paidAt,
		GatewayResponse: response.Data.Processor,
		Raw:             rawMap,
	}, nil
}

func (fws *Provider) ValidateWebhookSignature(_ context.Context, _ []byte, signature string) bool {
	if fws.webhookHash == "" || signature == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(fws.webhookHash), []byte(signature)) == 1
}

func (s *Provider) GetBanks(ctx context.Context) ([]payproviders.Bank, error) {
	body, err := s.client.Get(ctx, "/banks/NG")
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if response.Status != "success" {
		return nil, fmt.Errorf("failed to get banks: %s", response.Message)
	}

	banks := make([]payproviders.Bank, len(response.Data))
	for i, b := range response.Data {
		banks[i] = payproviders.Bank{Name: b.Name, Code: b.Code}
	}
	return banks, nil
}

func (s *Provider) ValidateAccount(ctx context.Context, accountNumber, bankCode string) (*payproviders.AccountValidation, error) {
	body, err := s.client.Post(ctx, "/accounts/resolve", map[string]string{
		"account_number": accountNumber,
		"account_bank":   bankCode,
	})
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccountNumber string `json:"account_number"`
			AccountName   string `json:"account_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if response.Status != "success" {
		return nil, errors.New("account validation failed")
	}

	return &payproviders.AccountValidation{
		AccountNumber: response.Data.AccountNumber,
		AccountName:   response.Data.AccountName,
	}, nil
}

func (fws *Provider) ParseWebhook(payload []byte) (*payproviders.WebhookEvent, error) {
	var webhook struct {
		Event string `json:"event"`
		Data  struct {
			TxRef             string         `json:"tx_ref"`
			Amount            float64        `json:"amount"`
			Status            string         `json:"status"`
			ProcessorResponse string         `json:"processor_response"`
			CreatedAt         string         `json:"created_at"`
			Meta              map[string]any `json:"meta"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse flutterwave webhook: %w", err)
	}

	raw, _ := json.Marshal(webhook)
	var rawMap map[string]any
	_ = json.Unmarshal(raw, &rawMap)

	status := payproviders.ParsePaymentStatus(webhook.Data.Status)
	switch webhook.Event {
	case "charge.completed":
		status = payproviders.PaymentStatusSuccess
	case "charge.failed":
		status = payproviders.PaymentStatusFailed
	}

	event := &payproviders.WebhookEvent{
		EventType:       webhook.Event,
		Reference:       webhook.Data.TxRef,
		Amount:          majorUnitToKobo(webhook.Data.Amount),
		Status:          status,
		GatewayResponse: webhook.Data.ProcessorResponse,
		RawData:         rawMap,
	}

	if customRef, ok := webhook.Data.Meta["reference"].(string); ok {
		event.CustomReference = customRef
	}
	if webhook.Data.CreatedAt != "" {
		if paidAt, err := time.Parse(time.RFC3339, webhook.Data.CreatedAt); err == nil {
			event.PaidAt = paidAt
		}
	}

	return event, nil
}

func koboToMajorUnit(amount int64) string {
	return strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
}

func majorUnitToKobo(amount float64) int64 {
	return int64(amount * 100)
}

var _ payproviders.Provider = (*Provider)(nil)