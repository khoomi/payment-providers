package paystack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	payproviders "github.com/khoomi/payment-providers"
)

const baseURL = "https://api.paystack.co"

type Config struct {
	SecretKey  string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

type Provider struct {
	client    *payproviders.HTTPClient
	secretKey string
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

	apiURL := cfg.BaseURL
	if apiURL == "" {
		apiURL = baseURL
	}

	return &Provider{
		secretKey: cfg.SecretKey,
		client: payproviders.NewHTTPClient(payproviders.ClientConfig{
			Name:       payproviders.NamePaystack,
			BaseURL:    apiURL,
			SecretKey:  cfg.SecretKey,
			HTTPClient: httpClient,
			Logger:     cfg.Logger,
		}),
	}
}

func (ps *Provider) Name() payproviders.Name {
	return payproviders.NamePaystack
}

func (ps *Provider) Initialize(ctx context.Context, req payproviders.InitRequest) (*payproviders.InitResult, error) {
	currency := req.Currency.OrDefault()

	payload := map[string]any{
		"amount":   payproviders.AmountForGateway(req.Amount, currency, payproviders.NamePaystack),
		"email":    req.Email,
		"metadata": req.Metadata,
		"currency": currency.String(),
	}
	if req.CallbackURL != "" {
		payload["callback_url"] = req.CallbackURL
	}

	body, err := ps.client.Post(ctx, "/transaction/initialize", payload)
	if err != nil {
		return nil, err
	}

	var initResponse initResponse
	if err := json.Unmarshal(body, &initResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !initResponse.Status {
		return nil, fmt.Errorf("paystack initialization failed: %s", initResponse.Message)
	}

	return &payproviders.InitResult{
		AuthorizationURL: initResponse.Data.AuthorizationURL,
		AccessCode:       initResponse.Data.AccessCode,
		Reference:        initResponse.Data.Reference,
	}, nil
}

func (ps *Provider) Verify(ctx context.Context, reference string) (*payproviders.VerifyResult, error) {
	endpoint := fmt.Sprintf("/transaction/verify/%s", url.PathEscape(reference))

	body, err := ps.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var verifyResponse verifyResponse
	if err := json.Unmarshal(body, &verifyResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !verifyResponse.Status {
		return nil, fmt.Errorf("paystack verification failed: %s", verifyResponse.Message)
	}

	raw, _ := json.Marshal(verifyResponse.Data)
	var rawMap map[string]any
	_ = json.Unmarshal(raw, &rawMap)

	return &payproviders.VerifyResult{
		Status:          payproviders.ParsePaymentStatus(verifyResponse.Data.Status),
		Reference:       verifyResponse.Data.Reference,
		Amount:          verifyResponse.Data.Amount,
		Currency:        payproviders.ParseCurrency(verifyResponse.Data.Currency),
		PaidAt:          verifyResponse.Data.PaidAt,
		GatewayResponse: verifyResponse.Data.GatewayResponse,
		Raw:             rawMap,
	}, nil
}

func (ps *Provider) Refund(ctx context.Context, req payproviders.RefundRequest) (*payproviders.RefundResult, error) {
	if req.TransactionReference == "" {
		return nil, errors.New("paystack transaction reference is required")
	}

	payload := map[string]any{
		"transaction": req.TransactionReference,
	}
	currency := req.Currency.OrDefault()

	if req.Amount > 0 {
		payload["amount"] = payproviders.AmountForGateway(req.Amount, currency, payproviders.NamePaystack)
	}
	if req.CustomerNote != "" {
		payload["customer_note"] = req.CustomerNote
	}

	body, err := ps.client.Post(ctx, "/refund", payload)
	if err != nil {
		return nil, err
	}

	var response refundResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !response.Status {
		return nil, fmt.Errorf("paystack refund failed: %s", response.Message)
	}

	raw, _ := json.Marshal(response.Data)
	var rawMap map[string]any
	_ = json.Unmarshal(raw, &rawMap)

	return &payproviders.RefundResult{
		Reference: fmt.Sprintf("%d", response.Data.ID),
		Status:    payproviders.ParseRefundStatus(response.Data.Status),
		Amount:    response.Data.Amount,
		Currency:  payproviders.ParseCurrency(response.Data.Currency),
		Raw:       rawMap,
	}, nil
}

func (ps *Provider) ValidateWebhookSignature(_ context.Context, payload []byte, signature string) bool {
	return validateWebhookSignature(ps.secretKey, payload, signature)
}

func (ps *Provider) GetBanks(ctx context.Context) ([]payproviders.Bank, error) {
	body, err := ps.client.Get(ctx, "/bank")
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  bool                 `json:"status"`
		Message string               `json:"message"`
		Data    []payproviders.Bank  `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !response.Status {
		return nil, fmt.Errorf("failed to get banks: %s", response.Message)
	}
	return response.Data, nil
}

func (ps *Provider) ValidateAccount(ctx context.Context, accountNumber, bankCode string) (*payproviders.AccountValidation, error) {
	endpoint := fmt.Sprintf("/bank/resolve?account_number=%s&bank_code=%s",
		url.QueryEscape(accountNumber), url.QueryEscape(bankCode))

	body, err := ps.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  bool                           `json:"status"`
		Message string                         `json:"message"`
		Data    payproviders.AccountValidation `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !response.Status {
		return nil, errors.New("account validation failed")
	}
	return &response.Data, nil
}

func (ps *Provider) ParseWebhook(payload []byte) (*payproviders.WebhookEvent, error) {
	var webhook webhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse paystack webhook: %w", err)
	}

	raw, _ := json.Marshal(webhook)
	var rawMap map[string]any
	_ = json.Unmarshal(raw, &rawMap)

	event := &payproviders.WebhookEvent{
		EventType:       webhook.Event,
		Reference:       webhook.Data.Reference,
		Amount:          webhook.Data.Amount,
		Status:          payproviders.ParsePaymentStatus(webhook.Data.Status),
		GatewayResponse: webhook.Data.GatewayResponse,
		RawData:         rawMap,
	}

	if customRef, ok := webhook.Data.Metadata["reference"].(string); ok {
		event.CustomReference = customRef
	}
	if webhook.Data.PaidAt != "" {
		if paidAt, err := time.Parse(time.RFC3339, webhook.Data.PaidAt); err == nil {
			event.PaidAt = paidAt
		}
	}

	return event, nil
}

type initResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type verifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status          string    `json:"status"`
		Reference       string    `json:"reference"`
		Amount          int64     `json:"amount"`
		Currency        string    `json:"currency"`
		PaidAt          time.Time `json:"paid_at"`
		GatewayResponse string    `json:"gateway_response"`
	} `json:"data"`
}

type refundResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID       int64  `json:"id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
		Status   string `json:"status"`
	} `json:"data"`
}

type webhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Reference       string         `json:"reference"`
		Amount          int64          `json:"amount"`
		Status          string         `json:"status"`
		GatewayResponse string         `json:"gateway_response"`
		PaidAt          string         `json:"paid_at"`
		Metadata        map[string]any `json:"metadata"`
	} `json:"data"`
}

var _ payproviders.Provider = (*Provider)(nil)