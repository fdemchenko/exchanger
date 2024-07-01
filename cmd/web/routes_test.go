package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fdemchenko/exchanger/internal/services/rate"
	"github.com/stretchr/testify/assert"
)

type RateResponse struct {
	Rate float32 `json:"rate"`
}

func TestRateEndpointIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rateService := rate.NewRateService(
		rate.WithFetchers(
			rate.NewNBURateFetcher("nbu fetcher"),
			rate.NewFawazRateFetcher("fawaz fetcher"),
			rate.NewPrivatRateFetcher("privat fetcher"),
		),
		rate.WithUpdateInterval(RateCachingDuration),
	)
	app := application{
		rateService: rateService,
	}

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/rate")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, rs.StatusCode)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	var rateResponse RateResponse
	err = json.Unmarshal(body, &rateResponse)
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, rateResponse.Rate, float32(0))
}
