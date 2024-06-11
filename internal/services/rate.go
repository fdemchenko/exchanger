package services

import (
	"encoding/json"
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

type RateService struct {
	currentRate *Rate
	mutex       sync.RWMutex
	fetchError  error
}

// Fetches currency data periodically, period is defined by updateInterval.
// Makes 3 tries before giving up by default
func NewRateService(updateInterval time.Duration) *RateService {
	rateService := &RateService{mutex: sync.RWMutex{}}

	// initial data fetch
	rateService.currentRate, rateService.fetchError = rateService.fetchExchangeRate()
	go func() {
		for range time.Tick(updateInterval) {
			for i := 0; i < RetryCount; i++ {
				rate, err := rateService.fetchExchangeRate()
				if err == nil {
					rateService.mutex.Lock()
					rateService.currentRate = rate
					rateService.fetchError = nil
					rateService.mutex.Unlock()
					break
				}
				if i == RetryCount-1 {
					// Give up
					rateService.mutex.Lock()
					rateService.fetchError = err
					rateService.mutex.Unlock()
				}
				time.Sleep(RetryInterval)
			}
		}
	}()
	return rateService
}

func (rs *RateService) GetRate() (*Rate, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if rs.fetchError != nil {
		return nil, rs.fetchError
	}
	return rs.currentRate, nil
}

func (rs *RateService) fetchExchangeRate() (*Rate, error) {
	resp, err := http.Get(RatesAPIBaseURL + "/usd.json")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var rate Rate

	err = json.NewDecoder(resp.Body).Decode(&rate)
	if err != nil {
		return nil, err
	}

	return &rate, nil
}
