package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fdemchenko/exchanger/internal/cache"
)

const (
	NBUExchangeRateURL        = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
	FawazAhmedExchangeRateURL = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies"

	DefaultCachingDuration = 15 * time.Minute
	RateFetchTimeout       = 10 * time.Second
)

var ErrInvalidCurrencyCode = errors.New("invalid currency code")

type ExchangeRateFetcher func(code string, client *http.Client) (float32, error)

type NBUResponse struct {
	Rates []struct {
		Rate float32 `json:"rate"`
		Code string  `json:"cc"`
	}
}

func NBURateFetcher(code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RateFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, NBUExchangeRateURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var nbuResponse NBUResponse
	err = json.NewDecoder(resp.Body).Decode(&nbuResponse)
	if err != nil {
		return 0, err
	}
	for _, rate := range nbuResponse.Rates {
		if strings.EqualFold(code, rate.Code) {
			return rate.Rate, nil
		}
	}
	return 0, ErrInvalidCurrencyCode
}

type FawazAhmedResponse struct {
	Rates struct {
		UAH float32 `json:"uah"`
	} `json:"usd"`
}

func FawazAhmedRateFetcher(code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RateFetchTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/%s.json", FawazAhmedExchangeRateURL, strings.ToLower(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, ErrInvalidCurrencyCode
	}

	defer resp.Body.Close()
	var fawazAhmedResponse FawazAhmedResponse
	err = json.NewDecoder(resp.Body).Decode(&fawazAhmedResponse)
	if err != nil {
		return 0, err
	}
	return fawazAhmedResponse.Rates.UAH, nil
}

type cachingRateService struct {
	fetchers       []ExchangeRateFetcher
	client         *http.Client
	updateInterval time.Duration
	cache          *cache.Cache[string, float32]
}

type Option func(*cachingRateService)

func WithClient(client *http.Client) Option {
	return func(crs *cachingRateService) {
		crs.client = client
	}
}

func WithUpdateInterval(updateInterval time.Duration) Option {
	return func(crs *cachingRateService) {
		crs.updateInterval = updateInterval
	}
}

func WithFetchers(fetchers ...ExchangeRateFetcher) Option {
	return func(crs *cachingRateService) {
		crs.fetchers = fetchers
	}
}

func NewRateService(options ...Option) *cachingRateService {
	// caching rate service with default values.
	service := &cachingRateService{
		client:         http.DefaultClient,
		updateInterval: DefaultCachingDuration,
		fetchers:       []ExchangeRateFetcher{NBURateFetcher},
		cache:          cache.New[string, float32](),
	}

	for _, option := range options {
		option(service)
	}
	return service
}

func (crs *cachingRateService) GetRate(currencyCode string) (float32, error) {
	if rate, exists := crs.cache.Get(strings.ToLower(currencyCode)); exists {
		return rate, nil
	}

	var err error
	var rate float32
	for _, fetcher := range crs.fetchers {
		rate, err = fetcher(currencyCode, crs.client)
		if err == nil {
			crs.cache.Set(strings.ToLower(currencyCode), rate, crs.updateInterval)
			return rate, nil
		}
	}

	return 0, err
}
