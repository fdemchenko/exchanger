package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	TestingRateServiceFetchInterval   = time.Second * 1
	TestingRateServiceWaitingDuration = time.Second * 2
)

type MockExchangeRateClient struct {
	mock.Mock
}

func (ec *MockExchangeRateClient) GetExchangeRate() (float32, error) {
	args := ec.Called()
	return args.Get(0).(float32), args.Error(1)
}

func TestRateService_RateIsCorrect(t *testing.T) {
	mockRateClient := new(MockExchangeRateClient)
	mockRateClient.On("GetExchangeRate").Return(float32(8.0), nil)
	rateService := NewRateService(TestingRateServiceFetchInterval, mockRateClient)
	rateService.StartBackgroundTask()

	rate, err := rateService.GetRate()
	assert.Equal(t, float32(8.0), rate)
	assert.NoError(t, err)
}

func TestRateService_RateIsCached(t *testing.T) {
	mockRateClient := new(MockExchangeRateClient)
	mockRateClient.On("GetExchangeRate").Return(float32(8.0), nil)
	rateService := NewRateService(TestingRateServiceFetchInterval, mockRateClient)
	rateService.StartBackgroundTask()

	_, _ = rateService.GetRate()
	_, _ = rateService.GetRate()

	// Make sure service cached result from previoues calls.
	mockRateClient.AssertNumberOfCalls(t, "GetExchangeRate", 1)
}

func TestRateService_RateIsFetchedAfterInterval(t *testing.T) {
	mockRateClient := new(MockExchangeRateClient)
	mockRateClient.On("GetExchangeRate").Return(float32(8.0), nil)
	rateService := NewRateService(TestingRateServiceFetchInterval, mockRateClient)
	rateService.StartBackgroundTask()

	// Make sure service re-fetch after update interval.
	time.Sleep(TestingRateServiceWaitingDuration)
	assert.GreaterOrEqual(t, len(mockRateClient.Calls), 2)
}
