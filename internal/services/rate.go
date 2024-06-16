package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

const RatesAPIBaseURL = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies"

const (
	RetryInterval = 500 * time.Millisecond
	RetryCount    = 3
)

type Rate struct {
	Rates struct {
		UAH float32 `json:"uah"`
	} `json:"usd"`
}

var ErrNoFetchOccurred = errors.New("no fecth occurred yet")

type ExchangeRateClient interface {
	GetExchangeRate() (float32, error)
}

type CachingRateService struct {
	currentRate    float32
	mutex          sync.RWMutex
	fetchError     error
	updateInterval time.Duration
	client         ExchangeRateClient
}

// Creates new rate service instance.
// Pass nil to client to use default http client.
func NewRateService(updateInterval time.Duration, client ExchangeRateClient) *CachingRateService {
	return &CachingRateService{
		mutex:          sync.RWMutex{},
		updateInterval: updateInterval,
		fetchError:     ErrNoFetchOccurred,
		client:         client,
	}
}

// Fetches currency data periodically, period is defined by updateInterval.
// Makes 3 tries before giving up by default
func (rs *CachingRateService) StartBackgroundTask() {
	// initial fetch
	rs.mutex.Lock()
	rs.currentRate, rs.fetchError = rs.client.GetExchangeRate()
	rs.mutex.Unlock()
	go func() {
		for range time.Tick(rs.updateInterval) {
			for i := 0; i < RetryCount; i++ {
				rate, err := rs.client.GetExchangeRate()
				if err == nil {
					rs.mutex.Lock()
					rs.currentRate = rate
					rs.fetchError = nil
					rs.mutex.Unlock()
					break
				}
				if i == RetryCount-1 {
					// Give up
					rs.mutex.Lock()
					rs.fetchError = err
					rs.mutex.Unlock()
				}
				time.Sleep(RetryInterval)
			}
		}
	}()
}

func (rs *CachingRateService) GetRate() (float32, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if rs.fetchError != nil {
		return 0, rs.fetchError
	}
	return rs.currentRate, nil
}

type HttpExchangeRateClient struct {
	client *http.Client
}

func NewHttpExchangeRateClient(client *http.Client) *HttpExchangeRateClient {
	return &HttpExchangeRateClient{client: client}
}

func (ec *HttpExchangeRateClient) GetExchangeRate() (float32, error) {
	resp, err := ec.client.Get(RatesAPIBaseURL + "/usd.json")
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	var rate Rate

	err = json.NewDecoder(resp.Body).Decode(&rate)
	if err != nil {
		return 0, err
	}

	return rate.Rates.UAH, nil
}
