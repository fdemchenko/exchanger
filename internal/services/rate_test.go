package services

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	TestingRateServiceFetchInterval   = time.Second * 1
	TestingRateServiceWaitingDuration = time.Second * 2
)

type MockExchangeRateFetcher struct {
	mock.Mock
}

func (ec *MockExchangeRateFetcher) FirstFetcher(code string, client *http.Client) (float32, error) {
	args := ec.Called()
	return args.Get(0).(float32), args.Error(1)
}

func (ec *MockExchangeRateFetcher) SecondFetcher(code string, client *http.Client) (float32, error) {
	args := ec.Called()
	return args.Get(0).(float32), args.Error(1)
}

func TestRateService_RateIsCorrect(t *testing.T) {
	mockRateFetcher := new(MockExchangeRateFetcher)
	mockRateFetcher.On("FirstFetcher").Return(float32(8.0), nil)

	rateService := NewRateService(WithFetchers(mockRateFetcher.FirstFetcher))
	rate, err := rateService.GetRate("usd")

	assert.Equal(t, float32(8.0), rate)
	assert.NoError(t, err)
}

func TestRateService_RateIsCached(t *testing.T) {
	mockRateFetcher := new(MockExchangeRateFetcher)
	mockRateFetcher.On("FirstFetcher").Return(float32(8.0), nil)
	rateService := NewRateService(WithFetchers(mockRateFetcher.FirstFetcher))
	_, _ = rateService.GetRate("usd")
	_, _ = rateService.GetRate("usd")

	// Make sure service cached result from previoues calls.
	mockRateFetcher.AssertNumberOfCalls(t, "FirstFetcher", 1)
}

func TestRateService_RateIsFetchedAfterInterval(t *testing.T) {
	mockRateFetcher := new(MockExchangeRateFetcher)
	mockRateFetcher.On("FirstFetcher").Return(float32(8.0), nil)
	rateService := NewRateService(
		WithFetchers(mockRateFetcher.FirstFetcher),
		WithUpdateInterval(TestingRateServiceFetchInterval))

	// Make sure service re-fetch after update interval.
	_, _ = rateService.GetRate("usd")
	time.Sleep(TestingRateServiceWaitingDuration)
	_, _ = rateService.GetRate("usd")

	mockRateFetcher.AssertNumberOfCalls(t, "FirstFetcher", 2)
}

func TestRateService_FallbackToAnotherFetcher(t *testing.T) {
	mockRateFetcher := new(MockExchangeRateFetcher)
	mockRateFetcher.On("FirstFetcher").Return(float32(0), ErrInvalidCurrencyCode)
	mockRateFetcher.On("SecondFetcher").Return(float32(8.0), nil)
	rateService := NewRateService(WithFetchers(mockRateFetcher.FirstFetcher, mockRateFetcher.SecondFetcher))

	rate, err := rateService.GetRate("usd")
	assert.NoError(t, err)
	assert.Equal(t, float32(8.0), rate)

	mockRateFetcher.AssertCalled(t, "FirstFetcher")
	mockRateFetcher.AssertCalled(t, "SecondFetcher")
}
