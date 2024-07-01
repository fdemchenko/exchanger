package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/fdemchenko/exchanger/internal/integration"
	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/fdemchenko/exchanger/internal/services/rate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

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

	rate, err := strconv.ParseFloat(string(body), 32)
	assert.NoError(t, err)
	assert.Greater(t, rate, float64(0))
}

type SubscribeEndpointTestSuite struct {
	suite.Suite
	container  *postgres.PostgresContainer
	testServer *httptest.Server
}

func (sets *SubscribeEndpointTestSuite) SetupSuite() {
	t := sets.T()
	container, err := integration.CreateTestDBContainer()
	if err != nil {
		t.Fatal(err)
	}
	sets.container = container

	dsn, err := container.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	postgresRepo := &repositories.PostgresEmailRepository{DB: db}
	emailService := services.NewEmailService(postgresRepo)

	app := application{
		emailService: emailService,
	}
	ts := httptest.NewServer(app.routes())
	sets.testServer = ts
}

func (sets *SubscribeEndpointTestSuite) TestSubscribe_Success() {
	t := sets.T()
	client := sets.testServer.Client()
	data := url.Values{}
	data.Set("email", "someemail@gmail.com")
	encodedData := data.Encode()
	resp, err := client.Post(sets.testServer.URL+"/subscribe", "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (sets *SubscribeEndpointTestSuite) TestSubscribe_InvalidEmail() {
	t := sets.T()
	client := sets.testServer.Client()
	data := url.Values{}
	data.Set("email", "some^!invalid@@gmail_com")
	encodedData := data.Encode()
	resp, err := client.Post(sets.testServer.URL+"/subscribe", "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (sets *SubscribeEndpointTestSuite) TestSubscribe_DuplicateEmail() {
	t := sets.T()
	client := sets.testServer.Client()
	data := url.Values{}
	data.Set("email", "mail@mail.com")
	encodedData := data.Encode()
	resp, err := client.Post(sets.testServer.URL+"/subscribe", "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = client.Post(sets.testServer.URL+"/subscribe", "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func (sets *SubscribeEndpointTestSuite) TearDownTest() {
	err := sets.container.Restore(context.Background())
	if err != nil {
		sets.T().Fatal(err)
	}
}

func (sets *SubscribeEndpointTestSuite) TearDownSuite() {
	sets.testServer.Close()
	if err := sets.container.Terminate(context.Background()); err != nil {
		sets.T().Fatal(err)
	}
}

func TestSubscribeEndpointSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping email service integrtion test...")
	}
	suite.Run(t, new(SubscribeEndpointTestSuite))
}
