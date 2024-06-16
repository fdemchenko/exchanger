package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRateEndpointIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rateService := services.NewRateService(time.Minute, services.NewHTTPExchangeRateClient(http.DefaultClient))
	rateService.StartBackgroundTask()
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

	_, err = strconv.ParseFloat(string(body), 32)
	assert.NoError(t, err)
}
