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

type CachingRateService struct {
	currentRate    *Rate
	mutex          sync.RWMutex
	fetchError     error
	updateInterval time.Duration
}

func NewRateService(updateInterval time.Duration) *CachingRateService {
	return &CachingRateService{
		mutex:          sync.RWMutex{},
		updateInterval: updateInterval,
		fetchError:     ErrNoFetchOccurred,
	}
}

// Fetches currency data periodically, period is defined by updateInterval.
// Makes 3 tries before giving up by default
func (rs *CachingRateService) StartBackgroundTask() {
	// initial fetch
	rs.mutex.Lock()
	rs.currentRate, rs.fetchError = rs.fetchExchangeRate()
	rs.mutex.Unlock()
	go func() {
		for range time.Tick(rs.updateInterval) {
			for i := 0; i < RetryCount; i++ {
				rate, err := rs.fetchExchangeRate()
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

func (rs *CachingRateService) GetRate() (*Rate, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if rs.fetchError != nil {
		return nil, rs.fetchError
	}
	return rs.currentRate, nil
}

func (rs *CachingRateService) fetchExchangeRate() (*Rate, error) {
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
