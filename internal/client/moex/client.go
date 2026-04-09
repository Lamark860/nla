package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://iss.moex.com/iss"

// Client wraps HTTP calls to MOEX ISS API
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// Get makes a request to MOEX ISS API and returns decoded JSON.
// path is relative to /iss/ (e.g. "engines/stock/markets/bonds/securities.json")
func (c *Client) Get(ctx context.Context, path string, params map[string]string) (map[string]any, error) {
	u, err := url.Parse(fmt.Sprintf("%s/%s", baseURL, path))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("iss.meta", "off")
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("moex request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("moex returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode moex response: %w", err)
	}

	return result, nil
}

// GetSecurities fetches all bonds from the market
func (c *Client) GetSecurities(ctx context.Context) (map[string]any, error) {
	return c.Get(ctx, "engines/stock/markets/bonds/securities.json", map[string]string{
		"iss.only": "securities,marketdata",
	})
}

// GetSecurity fetches a single bond by SECID
func (c *Client) GetSecurity(ctx context.Context, secid string) (map[string]any, error) {
	path := fmt.Sprintf("engines/stock/markets/bonds/securities/%s.json", secid)
	return c.Get(ctx, path, map[string]string{
		"iss.only": "securities,marketdata",
	})
}

// GetBondization fetches coupon schedule for a bond
func (c *Client) GetBondization(ctx context.Context, secid string) (map[string]any, error) {
	path := fmt.Sprintf("securities/%s/bondization.json", secid)
	return c.Get(ctx, path, nil)
}

// GetCandles fetches OHLC price history
func (c *Client) GetCandles(ctx context.Context, secid string, from, till string) (map[string]any, error) {
	path := fmt.Sprintf("engines/stock/markets/bonds/securities/%s/candles.json", secid)
	params := map[string]string{"interval": "24"}
	if from != "" {
		params["from"] = from
	}
	if till != "" {
		params["till"] = till
	}
	return c.Get(ctx, path, params)
}

// GetDisclosure fetches emitter metadata for a security
func (c *Client) GetDisclosure(ctx context.Context, secid string) (map[string]any, error) {
	path := fmt.Sprintf("securities/%s.json", secid)
	return c.Get(ctx, path, map[string]string{
		"iss.only": "description",
	})
}
