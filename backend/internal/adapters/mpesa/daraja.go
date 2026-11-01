package mpesa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DarajaConfig contains seller-specific credentials. Callback URLs are supplied
// by the POS deployment and must be publicly reachable by Safaricom.
type DarajaConfig struct {
	Environment       string `json:"environment"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	BusinessShortCode string `json:"business_short_code"`
	Passkey           string `json:"passkey"`
	TransactionType   string `json:"transaction_type"`
	CallbackURL       string `json:"callback_url"`
	ConfirmationURL   string `json:"confirmation_url"`
	ValidationURL     string `json:"validation_url"`
	C2BResponseType   string `json:"c2b_response_type"`
}

type Daraja struct {
	Config  DarajaConfig
	Client  *http.Client
	now     func() time.Time
	mu      sync.Mutex
	token   string
	expires time.Time
}

func NewDaraja(config DarajaConfig, client *http.Client) (*Daraja, error) {
	if config.Environment != "sandbox" && config.Environment != "production" {
		return nil, errors.New("M-Pesa environment must be sandbox or production")
	}
	if config.ConsumerKey == "" || config.ConsumerSecret == "" || config.BusinessShortCode == "" || config.Passkey == "" || config.CallbackURL == "" {
		return nil, errors.New("M-Pesa credentials, shortcode, passkey, and callback URL are required")
	}
	if config.TransactionType == "" {
		config.TransactionType = "CustomerPayBillOnline"
	}
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &Daraja{Config: config, Client: client, now: time.Now}, nil
}

func (d *Daraja) baseURL() string {
	if d.Config.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (d *Daraja) accessToken(ctx context.Context) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.token != "" && d.now().Before(d.expires.Add(-30*time.Second)) {
		return d.token, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL()+"/oauth/v1/generate?grant_type=client_credentials", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(d.Config.ConsumerKey, d.Config.ConsumerSecret)
	resp, err := d.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("M-Pesa OAuth request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("M-Pesa OAuth returned HTTP %d", resp.StatusCode)
	}
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&body); err != nil {
		return "", err
	}
	seconds, _ := strconv.Atoi(body.ExpiresIn)
	if body.AccessToken == "" || seconds < 1 {
		return "", errors.New("M-Pesa OAuth response is incomplete")
	}
	d.token, d.expires = body.AccessToken, d.now().Add(time.Duration(seconds)*time.Second)
	return d.token, nil
}

func (d *Daraja) InitiateSTK(ctx context.Context, in STKRequest) (STKResponse, error) {
	if in.AmountMinor <= 0 || in.AmountMinor%100 != 0 {
		return STKResponse{}, errors.New("M-Pesa STK amount must be a positive whole shilling in minor units")
	}
	phone := normalizePhone(in.Phone)
	if phone == "" {
		return STKResponse{}, errors.New("M-Pesa phone must be a valid Kenyan mobile number")
	}
	token, err := d.accessToken(ctx)
	if err != nil {
		return STKResponse{}, err
	}
	timestamp := darajaTimestamp(d.now())
	password := base64.StdEncoding.EncodeToString([]byte(d.Config.BusinessShortCode + d.Config.Passkey + timestamp))
	shortcode := d.Config.BusinessShortCode
	if in.BusinessShortCode != "" {
		shortcode = in.BusinessShortCode
	}
	callback := map[string]string{"BusinessShortCode": shortcode, "Password": password, "Timestamp": timestamp, "TransactionType": d.Config.TransactionType, "Amount": strconv.FormatInt(in.AmountMinor/100, 10), "PartyA": phone, "PartyB": shortcode, "PhoneNumber": phone, "CallBackURL": d.Config.CallbackURL, "AccountReference": limit(in.AccountReference, 12), "TransactionDesc": limit(in.Description, 13)}
	var out STKResponse
	if err := d.post(ctx, token, "/mpesa/stkpush/v1/processrequest", callback, &out); err != nil {
		return STKResponse{}, err
	}
	if out.ResponseCode != "0" || out.CheckoutRequestID == "" {
		return STKResponse{}, fmt.Errorf("M-Pesa STK rejected request: %s", limit(out.ResponseDescription, 120))
	}
	return out, nil
}

func (d *Daraja) QueryStatus(ctx context.Context, checkoutID string) (string, error) {
	token, err := d.accessToken(ctx)
	if err != nil {
		return "", err
	}
	timestamp := darajaTimestamp(d.now())
	password := base64.StdEncoding.EncodeToString([]byte(d.Config.BusinessShortCode + d.Config.Passkey + timestamp))
	var out struct {
		ResultCode json.RawMessage `json:"ResultCode"`
		ResultDesc string          `json:"ResultDesc"`
	}
	err = d.post(ctx, token, "/mpesa/stkpushquery/v1/query", map[string]string{"BusinessShortCode": d.Config.BusinessShortCode, "Password": password, "Timestamp": timestamp, "CheckoutRequestID": checkoutID}, &out)
	if err != nil {
		return "", err
	}
	resultCode := strings.Trim(string(out.ResultCode), `"`)
	if resultCode == "0" {
		return "Completed", nil
	}
	if resultCode == "1032" {
		return "Failed", nil
	}
	if resultCode == "" || resultCode == "1037" {
		return "Pending", nil
	}
	return "Failed", nil
}

func (d *Daraja) Reverse(context.Context, string, int64) error {
	return errors.New("M-Pesa reversal requires a separately configured initiator and security credential")
}

func (d *Daraja) RegisterC2B(ctx context.Context) error {
	if d.Config.ValidationURL == "" || d.Config.ConfirmationURL == "" {
		return errors.New("M-Pesa validation and confirmation URLs are required for C2B registration")
	}
	token, err := d.accessToken(ctx)
	if err != nil {
		return err
	}
	responseType := d.Config.C2BResponseType
	if responseType == "" {
		responseType = "Completed"
	}
	var out struct {
		ResponseCode        json.RawMessage `json:"ResponseCode"`
		ResponseDescription string          `json:"ResponseDescription"`
	}
	if err := d.post(ctx, token, "/mpesa/c2b/v2/registerurl", map[string]string{"ShortCode": d.Config.BusinessShortCode, "ResponseType": responseType, "ConfirmationURL": d.Config.ConfirmationURL, "ValidationURL": d.Config.ValidationURL}, &out); err != nil {
		return err
	}
	code := strings.Trim(string(out.ResponseCode), `"`)
	if code != "" && code != "0" {
		return fmt.Errorf("M-Pesa C2B URL registration rejected: %s", limit(out.ResponseDescription, 120))
	}
	return nil
}

func (d *Daraja) post(ctx context.Context, token, path string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL()+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.Client.Do(req)
	if err != nil {
		return fmt.Errorf("M-Pesa request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("M-Pesa returned HTTP %d: %s", resp.StatusCode, limit(string(raw), 240))
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(result)
}

// ParseCallback accepts Safaricom's nested STK callback schema.
func (d *Daraja) ParseCallback(body []byte) (Callback, error) {
	var envelope struct {
		Body struct {
			STKCallback struct {
				CheckoutRequestID string `json:"CheckoutRequestID"`
				ResultCode        int    `json:"ResultCode"`
				ResultDesc        string `json:"ResultDesc"`
				CallbackMetadata  struct {
					Items []struct {
						Name  string `json:"Name"`
						Value any    `json:"Value"`
					} `json:"Item"`
				} `json:"CallbackMetadata"`
			} `json:"stkCallback"`
		} `json:"Body"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return Callback{}, err
	}
	c := envelope.Body.STKCallback
	if c.CheckoutRequestID == "" {
		return Callback{}, errors.New("missing CheckoutRequestID")
	}
	out := Callback{CheckoutRequestID: c.CheckoutRequestID, ResultCode: c.ResultCode, ResultDescription: c.ResultDesc}
	for _, item := range c.CallbackMetadata.Items {
		switch strings.ToLower(item.Name) {
		case "amount":
			out.AmountMinor = int64(number(item.Value)) * 100
		case "mpesareceiptnumber":
			out.ReceiptNumber = scalarText(item.Value)
		case "phonenumber":
			out.Phone = scalarText(item.Value)
		case "accountreferencenumber":
			out.AccountReference = scalarText(item.Value)
		}
	}
	return out, nil
}

func normalizePhone(s string) string {
	s = strings.TrimSpace(strings.NewReplacer("+", "", " ", "", "-", "").Replace(s))
	if strings.HasPrefix(s, "0") && len(s) == 10 {
		s = "254" + s[1:]
	}
	if strings.HasPrefix(s, "7") && len(s) == 9 {
		s = "254" + s
	}
	if strings.HasPrefix(s, "1") && len(s) == 9 {
		s = "254" + s
	}
	if len(s) != 12 || !strings.HasPrefix(s, "254") {
		return ""
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return s
}

func darajaTimestamp(t time.Time) string {
	return t.In(time.FixedZone("EAT", 3*60*60)).Format("20060102150405")
}

func number(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		i, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		return i
	}
}

func scalarText(v any) string {
	if n, ok := v.(float64); ok {
		return strconv.FormatFloat(n, 'f', 0, 64)
	}
	return fmt.Sprint(v)
}
func limit(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
