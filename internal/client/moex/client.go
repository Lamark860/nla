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

// GetSecurity fetches a single bond by SECID (including marketdata_yields)
func (c *Client) GetSecurity(ctx context.Context, secid string) (map[string]any, error) {
	path := fmt.Sprintf("engines/stock/markets/bonds/securities/%s.json", secid)
	return c.Get(ctx, path, map[string]string{
		"iss.only": "securities,marketdata,marketdata_yields",
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

// CCIRating represents a credit rating from MOEX CCI API
type CCIRating struct {
	AgencyName  string `json:"agency_name_short_ru"`
	RatingValue string `json:"rating_level_name_short_ru"`
	RatingDate  string `json:"rating_date"`
}

// GetCCIRatings fetches credit ratings for an emitter from MOEX CCI API.
// Uses /iss/cci/rating/companies/ecbd_{emitterID}.json with extended JSON format.
func (c *Client) GetCCIRatings(ctx context.Context, emitterID int64) ([]CCIRating, error) {
	u := fmt.Sprintf("%s/cci/rating/companies/ecbd_%d.json?iss.json=extended&iss.meta=off", baseURL, emitterID)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("moex cci request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read cci response: %w", err)
	}

	if len(body) == 0 || resp.StatusCode != http.StatusOK {
		return nil, nil // empty = no ratings available
	}

	// CCI extended format: [{charsetinfo}, {cci_rating_companies: [...], cursor: [...]}]
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode cci response: %w", err)
	}

	if len(raw) < 2 {
		return nil, nil
	}

	var data struct {
		Ratings []CCIRating `json:"cci_rating_companies"`
	}
	if err := json.Unmarshal(raw[1], &data); err != nil {
		return nil, fmt.Errorf("decode cci ratings: %w", err)
	}

	return data.Ratings, nil
}
