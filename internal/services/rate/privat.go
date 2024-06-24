package rate

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	PrivatExchangeRateURL = "https://api.privatbank.ua/p24api/pubinfo?json&exchange&coursid=5"
)

type PrivatResponse struct {
	Base     string  `json:"base"`
	Currency string  `json:"ccy"`
	Buy      float32 `json:"buy"`
	Sale     float32 `json:"sale"`
}

type PrivatRateFetcher struct {
	name string
}

func (nrf PrivatRateFetcher) Name() string {
	return nrf.name
}

func NewPrivatRateFetcher(name string) PrivatRateFetcher {
	return PrivatRateFetcher{name: name}
}

func (nfr PrivatRateFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(ctx, RateFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, PrivatExchangeRateURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var privatResponse []PrivatResponse
	err = json.NewDecoder(resp.Body).Decode(&privatResponse)
	if err != nil {
		return 0, err
	}
	for _, rate := range privatResponse {
		if strings.EqualFold(code, rate.Currency) {
			return rate.Buy, nil
		}
	}
	return 0, ErrInvalidCurrencyCode
}
