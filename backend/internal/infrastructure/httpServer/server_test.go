package httpServer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/billing"
	"learning/internal/meter"

	"github.com/gin-gonic/gin"
)

type testMeterService struct {
	items   []meter.Summary
	get     meter.Detail
	history []meter.SitePoint
}

func (service testMeterService) List(context.Context) ([]meter.Summary, error) {
	return service.items, nil
}

func (service testMeterService) Get(context.Context, string) (meter.Detail, error) {
	return service.get, nil
}

func (service testMeterService) SiteHistory(context.Context) ([]meter.SitePoint, error) {
	return service.history, nil
}

type testAnomalyService struct{}

func (testAnomalyService) List(context.Context) ([]anomaly.Anomaly, error) {
	return nil, nil
}

func (testAnomalyService) Get(context.Context, string) (anomaly.Anomaly, error) {
	return anomaly.Anomaly{}, nil
}

type testAnalysisService struct {
	result ai.Analysis
}

type testHealthChecker struct {
	err error
}

func (checker testHealthChecker) Ping(context.Context) error {
	return checker.err
}

func (service testAnalysisService) Analyze(context.Context) (ai.Analysis, error) {
	return service.result, nil
}

func (service testAnalysisService) GetAnalysis(context.Context, string) (ai.Analysis, error) {
	return service.result, nil
}

func (service testAnalysisService) GetLatestAnalysis(context.Context) (ai.Analysis, error) {
	return service.result, nil
}

func TestListMetersFiltersAndMapsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := newTestRouter(t, testMeterService{items: []meter.Summary{
		{Meter: meter.Meter{MeterID: "M-109"}, Status: "CRITICAL", CurrentConsumptionKWh: 91, ChangePercent: 110, Severity: anomaly.High},
		{Meter: meter.Meter{MeterID: "M-104"}, Status: "ALERT", CurrentConsumptionKWh: 72, ChangePercent: 47, Severity: anomaly.Medium},
	}})
	request := httptest.NewRequest(http.MethodGet, "/meters?filter=critical&meter_id=M-109&sort=consumption", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("local frontend origin was not allowed")
	}
	var body struct {
		Meters []struct {
			MeterID string `json:"meter_id"`
			Status  string `json:"status"`
		} `json:"meters"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Meters) != 1 || body.Meters[0].MeterID != "M-109" || body.Meters[0].Status != "CRITICAL" {
		t.Fatalf("unexpected meter response: %s", response.Body.String())
	}
}

func TestVersionedRoutesAreAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := newTestRouter(t, testMeterService{items: []meter.Summary{}})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/meters", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
}

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	meters := testMeterService{}
	analyses := testAnalysisService{result: ai.Analysis{Anomalies: []anomaly.Anomaly{}}}
	router, err := NewRouterWithDashboardAndHealth(meters, testAnomalyService{}, analyses, nil, testHealthChecker{}, nil, logger)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path    string
		message string
	}{
		{path: "/api/v1/health", message: "healthy"},
		{path: "/api/v1/isalive", message: "alive"},
		{path: "/api/v1/ping", message: "pong"},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", test.path, response.Code, http.StatusOK)
		}
		if !strings.Contains(response.Body.String(), test.message) {
			t.Fatalf("%s response = %s, want %q", test.path, response.Body.String(), test.message)
		}
	}
}

func TestHealthReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, err := NewRouterWithDashboardAndHealth(testMeterService{}, testAnomalyService{}, testAnalysisService{}, nil, testHealthChecker{err: errors.New("database unavailable")}, nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
func TestMeterRejectsMalformedIDBeforeLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := newTestRouter(t, testMeterService{})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/meters/M-abc", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAnalyzeReturnsStructuredResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	startedAt := time.Date(2026, time.September, 14, 12, 0, 0, 0, time.UTC)
	server := newTestRouter(t, testMeterService{}, testAnalysisService{result: ai.Analysis{
		ID: "f310f1fb-f500-44c3-a6a6-2d4ef4a60a43", Status: "COMPLETED", StartedAt: startedAt, Anomalies: []anomaly.Anomaly{},
	}})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/ai/analyze", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID == "" || body.Status != "COMPLETED" {
		t.Fatalf("unexpected analysis response: %s", response.Body.String())
	}
}

func TestPatchMeterConfigPersistsProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	profiles := &testMeterProfileStore{}
	profileRouter, err := NewRouterWithDashboardAndHealth(testMeterService{}, testAnomalyService{}, testAnalysisService{}, nil, testHealthChecker{}, profiles, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/meters/M-101", strings.NewReader(`{"provider":"EPM","region":"Medellín","rate_type":"industrial"}`))
	request.Header.Set("Content-Type", "application/json")
	profileRouter.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var decoded meterConfigResponseDTO
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.MeterID != "M-101" || decoded.Provider != "EPM" || decoded.Region != "Medellín" || decoded.RateType != "industrial" {
		t.Fatalf("unexpected profile response: %s", response.Body.String())
	}

	rejected := httptest.NewRecorder()
	badRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/meters/M-abc", strings.NewReader(`{"provider":"EPM"}`))
	badRequest.Header.Set("Content-Type", "application/json")
	profileRouter.ServeHTTP(rejected, badRequest)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rejected.Code, http.StatusBadRequest)
	}
}

type testMeterProfileStore struct {
	provider string
	region   string
	rateType string
}

func (fake *testMeterProfileStore) GetMeterProfile(_ context.Context, meterID string) (billing.MeterProfile, error) {
	return billing.MeterProfile{MeterID: meterID, Provider: fake.provider, Region: fake.region, RateType: fake.rateType}, nil
}

func (fake *testMeterProfileStore) UpdateMeterProfile(_ context.Context, _ string, provider, region, rateType string) error {
	fake.provider = provider
	fake.region = region
	fake.rateType = rateType
	return nil
}

func TestSiteHistoryReturnsOrderedPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	first := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	server := newTestRouter(t, testMeterService{history: []meter.SitePoint{
		{Timestamp: first, ConsumptionKWh: 250.5},
		{Timestamp: first.Add(time.Hour), ConsumptionKWh: 300.25},
	}})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/history", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Points []struct {
			Timestamp      time.Time `json:"timestamp"`
			ConsumptionKWh float64   `json:"consumption_kwh"`
		} `json:"points"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Points) != 2 || body.Points[0].ConsumptionKWh != 250.5 || body.Points[1].ConsumptionKWh != 300.25 {
		t.Fatalf("unexpected history response: %s", response.Body.String())
	}
	if body.Points[0].Timestamp.After(body.Points[1].Timestamp) {
		t.Fatalf("points not ordered: %s", response.Body.String())
	}
}

func newTestRouter(t *testing.T, meters testMeterService, analyses ...testAnalysisService) http.Handler {
	t.Helper()
	analysisService := testAnalysisService{result: ai.Analysis{Anomalies: []anomaly.Anomaly{}}}
	if len(analyses) > 0 {
		analysisService = analyses[0]
	}
	router, err := NewRouter(meters, testAnomalyService{}, analysisService, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return router
}
