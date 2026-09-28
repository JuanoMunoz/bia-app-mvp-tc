package httpServer

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	anomalyapp "learning/internal/anomaly/application"
	dashboardapp "learning/internal/dashboard/application"
	"learning/internal/billing"
	"learning/internal/infrastructure/middleware"
	"learning/internal/meter"
	"learning/internal/reading"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalysisService interface {
	Analyze(context.Context) (ai.Analysis, error)
	GetAnalysis(context.Context, string) (ai.Analysis, error)
	GetLatestAnalysis(context.Context) (ai.Analysis, error)
}

type MeterService interface {
	List(context.Context) ([]meter.Summary, error)
	Get(context.Context, string) (meter.Detail, error)
	SiteHistory(context.Context) ([]meter.SitePoint, error)
}

type AnomalyService interface {
	List(context.Context) ([]anomaly.Anomaly, error)
	Get(context.Context, string) (anomaly.Anomaly, error)
}

type DashboardService interface {
	GetSummary(context.Context) (dashboardapp.Summary, error)
}

type HealthChecker interface {
	Ping(context.Context) error
}

type MeterProfileStore interface {
	GetMeterProfile(context.Context, string) (billing.MeterProfile, error)
	UpdateMeterProfile(context.Context, string, string, string, string) error
}

type BillingService interface {
	CreateTariff(ctx context.Context, input tariffRequestDTO) (tariffResponseDTO, error)
	ListTariffs(ctx context.Context, provider, region string, active bool) ([]tariffResponseDTO, error)
	EstimateMeterConsumptionCost(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (estimateConsumptionResponseDTO, error)
	GetBillingHistoryByMeter(ctx context.Context, meterID string) ([]billingResponseDTO, error)
	GetBillingSummary(ctx context.Context, from, to time.Time) (billingSummaryResponseDTO, error)
	GetCostSummary(ctx context.Context, from, to time.Time) (billingSummaryResponseDTO, error)
	GetCostTimeSeries(ctx context.Context, from, to time.Time, granularity string) ([]costTimePointResponseDTO, error)
	GetCostByMeter(ctx context.Context, from, to time.Time) ([]costByMeterRowResponseDTO, error)
	GenerateBillingReport(ctx context.Context, format string, from, to time.Time, meterID *string) (billingReportFileResponseDTO, error)
	GetMeterOptimizationInsight(ctx context.Context, meterID string, from, to time.Time) (optimizationInsightResponseDTO, error)
}

type Server struct {
	meters    MeterService
	anomalies AnomalyService
	analysis  AnalysisService
	dashboard DashboardService
	health    HealthChecker
	profiles  MeterProfileStore
	billing   BillingService
	logger    *slog.Logger
}

func NewRouter(meters MeterService, anomalies AnomalyService, analysis AnalysisService, logger *slog.Logger) (*gin.Engine, error) {
	return NewRouterWithDashboardAndHealth(meters, anomalies, analysis, dashboardapp.NewService(meters, anomalies, analysis), nil, nil, logger, nil)
}

func NewRouterWithDashboard(meters MeterService, anomalies AnomalyService, analysis AnalysisService, dashboard DashboardService, logger *slog.Logger) (*gin.Engine, error) {
	return NewRouterWithDashboardAndHealth(meters, anomalies, analysis, dashboard, nil, nil, logger, nil)
}

func NewRouterWithDashboardAndHealth(meters MeterService, anomalies AnomalyService, analysis AnalysisService, dashboard DashboardService, health HealthChecker, profiles MeterProfileStore, logger *slog.Logger, allowedOrigins []string, billingServices ...BillingService) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	router.Use(middleware.JSONMiddleware(), middleware.CORSMiddleware(allowedOrigins), middleware.RateLimitMiddleware(60, time.Minute), middleware.LoggingMiddleware(logger), gin.Recovery())
	var billing BillingService
	if len(billingServices) > 0 {
		billing = billingServices[0]
	}
	server := &Server{meters: meters, anomalies: anomalies, analysis: analysis, dashboard: dashboard, health: health, profiles: profiles, billing: billing, logger: logger}
	registerRoutes(router.Group("/api/v1"), server)
	registerRoutes(router.Group(""), server)
	return router, nil
}

func registerRoutes(group *gin.RouterGroup, server *Server) {
	group.GET("/health", server.healthCheck)
	group.GET("/isalive", server.isAlive)
	group.GET("/ping", server.ping)
	group.GET("/meters", server.listMeters)
	group.GET("/meters/:meterId", server.getMeter)
	group.PATCH("/meters/:meterId", server.patchMeterConfig)
	group.GET("/meters/:meterId/readings", server.getMeterReadings)
	group.GET("/meters/:meterId/estimated-consumption", server.getEstimatedConsumption)
	group.GET("/meters/:meterId/optimization-insight", server.getMeterOptimizationInsight)
	group.GET("/meters/:meterId/billing-history", server.getMeterBillingHistory)
	group.GET("/anomalies", server.listAnomalies)
	group.GET("/anomalies/:id", server.getAnomaly)
	group.POST("/ai/analyze", server.analyze)
	group.GET("/ai/analysis/:id", server.getAnalysis)
	group.GET("/dashboard/summary", server.dashboardSummary)
	group.GET("/dashboard/history", server.siteHistory)
	group.POST("/tariffs", server.createTariff)
	group.GET("/tariffs", server.listTariffs)
	group.GET("/billing/summary", server.billingSummary)
	group.GET("/billing/history", server.billingHistory)
	group.GET("/billing/meters", server.billingByMeter)
	group.GET("/reports/billing", server.billingReport)
}

func (server *Server) healthCheck(context *gin.Context) {
	if server.health != nil {
		if err := server.health.Ping(context.Request.Context()); err != nil {
			server.logger.Error("health check failed", "error", err)
			context.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "service": "api"})
			return
		}
	}
	context.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "api", "timestamp": time.Now().UTC()})
}

func (server *Server) isAlive(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (server *Server) ping(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (server *Server) createTariff(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "create tariff", errors.New("billing service unavailable"))
		return
	}
	var input tariffRequestDTO
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := server.billing.CreateTariff(context.Request.Context(), input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, response)
}

func (server *Server) listTariffs(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "list tariffs", errors.New("billing service unavailable"))
		return
	}
	provider := strings.TrimSpace(context.Query("provider"))
	region := strings.TrimSpace(context.Query("region"))
	active := strings.EqualFold(strings.TrimSpace(context.Query("active")), "true")
	response, err := server.billing.ListTariffs(context.Request.Context(), provider, region, active)
	if err != nil {
		server.internalError(context, "list tariffs", err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"tariffs": response})
}

func (server *Server) getEstimatedConsumption(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "estimate consumption", errors.New("billing service unavailable"))
		return
	}
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	periodStart, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("period_start")))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start"})
		return
	}
	periodEnd, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("period_end")))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end"})
		return
	}
	response, err := server.billing.EstimateMeterConsumptionCost(context.Request.Context(), meterID, periodStart, periodEnd)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, response)
}

func (server *Server) getMeterBillingHistory(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "billing history", errors.New("billing service unavailable"))
		return
	}
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	response, err := server.billing.GetBillingHistoryByMeter(context.Request.Context(), meterID)
	if err != nil {
		server.internalError(context, "billing history", err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"billing_history": response})
}

func (server *Server) getMeterOptimizationInsight(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "optimization insight", errors.New("billing service unavailable"))
		return
	}
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("from")))
	if err != nil && strings.TrimSpace(context.Query("from")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
		return
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("to")))
	if err != nil && strings.TrimSpace(context.Query("to")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
		return
	}
	response, err := server.billing.GetMeterOptimizationInsight(context.Request.Context(), meterID, from, to)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, response)
}

func (server *Server) billingSummary(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "billing summary", errors.New("billing service unavailable"))
		return
	}
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("from")))
	if err != nil && strings.TrimSpace(context.Query("from")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
		return
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("to")))
	if err != nil && strings.TrimSpace(context.Query("to")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
		return
	}
	response, err := server.billing.GetCostSummary(context.Request.Context(), from, to)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, response)
}

func (server *Server) billingHistory(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "billing history", errors.New("billing service unavailable"))
		return
	}
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("from")))
	if err != nil && strings.TrimSpace(context.Query("from")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
		return
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("to")))
	if err != nil && strings.TrimSpace(context.Query("to")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
		return
	}
	granularity := strings.TrimSpace(context.DefaultQuery("granularity", "daily"))
	response, err := server.billing.GetCostTimeSeries(context.Request.Context(), from, to, granularity)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"series": response})
}

func (server *Server) billingByMeter(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "billing by meter", errors.New("billing service unavailable"))
		return
	}
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("from")))
	if err != nil && strings.TrimSpace(context.Query("from")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
		return
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("to")))
	if err != nil && strings.TrimSpace(context.Query("to")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
		return
	}
	response, err := server.billing.GetCostByMeter(context.Request.Context(), from, to)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"rows": response})
}

func (server *Server) billingReport(context *gin.Context) {
	if server.billing == nil {
		server.internalError(context, "billing report", errors.New("billing service unavailable"))
		return
	}
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("from")))
	if err != nil && strings.TrimSpace(context.Query("from")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
		return
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(context.Query("to")))
	if err != nil && strings.TrimSpace(context.Query("to")) != "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
		return
	}
	format := strings.ToLower(strings.TrimSpace(context.Query("format")))
	meterID := strings.TrimSpace(context.Query("meter_id"))
	var meterRef *string
	if meterID != "" {
		meterRef = &meterID
	}
	response, err := server.billing.GenerateBillingReport(context.Request.Context(), format, from, to, meterRef)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.Header("Content-Disposition", "attachment; filename=\""+response.FileName+"\"")
	context.Data(http.StatusOK, response.ContentType, response.Content)
}

func (server *Server) listMeters(context *gin.Context) {
	results, err := server.meters.List(context.Request.Context())
	if err != nil {
		server.internalError(context, "list meters", err)
		return
	}
	filter := strings.ToLower(context.DefaultQuery("filter", "all"))
	search := strings.ToLower(strings.TrimSpace(context.DefaultQuery("meter_id", context.Query("q"))))
	filtered := make([]meter.Summary, 0, len(results))
	for _, result := range results {
		if search != "" && !strings.Contains(strings.ToLower(result.MeterID), search) {
			continue
		}
		if !matchesFilter(result.Status, filter) {
			continue
		}
		filtered = append(filtered, result)
	}
	sortMeters(filtered, context.DefaultQuery("sort", "severity"), context.DefaultQuery("order", "desc"))
	offset, limit := pagination(context)
	paged, hasMore := page(filtered, offset, limit)
	responses := make([]meterResponse, 0, len(paged))
	for _, result := range paged {
		responses = append(responses, mapMeter(result))
	}
	context.JSON(http.StatusOK, gin.H{"meters": responses, "pagination": paginationResponse(offset, limit, len(filtered), hasMore)})
}

func (server *Server) getMeter(context *gin.Context) {
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	result, err := server.meters.Get(context.Request.Context(), meterID)
	if err != nil {
		if errors.Is(err, meter.ErrNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"error": "meter not found"})
			return
		}
		server.internalError(context, "get meter", err)
		return
	}
	response := meterDetailResponse{
		meterResponse: mapMeter(result.Summary),
		VoltageV:      result.VoltageV, CurrentA: result.CurrentA, PowerFactor: result.PowerFactor,
		History: mapReadings(result.Readings),
	}
	context.JSON(http.StatusOK, response)
}

func (server *Server) patchMeterConfig(context *gin.Context) {
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	var input meterConfigRequestDTO
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if server.profiles == nil {
		context.JSON(http.StatusServiceUnavailable, gin.H{"error": "meter profile storage unavailable"})
		return
	}
	provider := strings.TrimSpace(input.Provider)
	region := strings.TrimSpace(input.Region)
	rateType := strings.TrimSpace(input.RateType)
	if rateType == "" {
		rateType = "industrial"
	}
	if provider == "" && region == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "at least provider or region must be provided"})
		return
	}
	if err := server.profiles.UpdateMeterProfile(context.Request.Context(), meterID, provider, region, rateType); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := server.profiles.GetMeterProfile(context.Request.Context(), meterID)
	if err != nil {
		server.internalError(context, "load updated meter config", err)
		return
	}
	context.JSON(http.StatusOK, meterConfigResponseDTO{MeterID: updated.MeterID, Provider: updated.Provider, Region: updated.Region, RateType: updated.RateType, UpdatedAt: time.Now().UTC()})
}

func (server *Server) getMeterReadings(context *gin.Context) {
	meterID := strings.TrimSpace(context.Param("meterId"))
	if !validMeterID(meterID) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid meter id"})
		return
	}
	result, err := server.meters.Get(context.Request.Context(), meterID)
	if err != nil {
		if errors.Is(err, meter.ErrNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"error": "meter not found"})
			return
		}
		server.internalError(context, "get meter readings", err)
		return
	}
	offset, limit := pagination(context)
	paged, hasMore := page(result.Readings, offset, limit)
	context.JSON(http.StatusOK, gin.H{"meter_id": meterID, "readings": mapReadings(paged), "pagination": paginationResponse(offset, limit, len(result.Readings), hasMore)})
}

func (server *Server) listAnomalies(context *gin.Context) {
	results, err := server.anomalies.List(context.Request.Context())
	if err != nil {
		server.internalError(context, "list anomalies", err)
		return
	}
	filterType := strings.ToUpper(strings.TrimSpace(context.Query("type")))
	severityFilter := strings.ToUpper(strings.TrimSpace(context.Query("severity")))
	meterSearch := strings.ToLower(strings.TrimSpace(context.DefaultQuery("meter_id", context.Query("q"))))
	filtered := make([]anomaly.Anomaly, 0, len(results))
	for _, result := range results {
		if filterType != "" && filterType != "ALL" && string(result.Type) != filterType {
			continue
		}
		if severityFilter != "" && severityFilter != "ALL" && string(result.Severity) != severityFilter {
			continue
		}
		if meterSearch != "" && !strings.Contains(strings.ToLower(result.MeterID), meterSearch) {
			continue
		}
		filtered = append(filtered, result)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if context.DefaultQuery("sort", "severity") == "meter_id" {
			return filtered[i].MeterID < filtered[j].MeterID
		}
		return severityRank(filtered[i].Severity) < severityRank(filtered[j].Severity)
	})
	offset, limit := pagination(context)
	paged, hasMore := page(filtered, offset, limit)
	responses := make([]anomalyResponse, 0, len(paged))
	for _, result := range paged {
		responses = append(responses, mapAnomaly(result))
	}
	context.JSON(http.StatusOK, gin.H{"anomalies": responses, "pagination": paginationResponse(offset, limit, len(filtered), hasMore)})
}

type paginationMetadata struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

func pagination(context *gin.Context) (int, int) {
	offset, _ := strconv.Atoi(context.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(context.DefaultQuery("limit", "50"))
	return max(0, offset), min(100, max(1, limit))
}

func paginationResponse(offset, limit, total int, hasMore bool) paginationMetadata {
	return paginationMetadata{Offset: offset, Limit: limit, Total: total, HasMore: hasMore}
}

func page[T any](items []T, offset, limit int) ([]T, bool) {
	if offset >= len(items) {
		return []T{}, false
	}
	end := min(len(items), offset+limit)
	return items[offset:end], end < len(items)
}

func (server *Server) getAnomaly(context *gin.Context) {
	id := context.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid anomaly id"})
		return
	}
	result, err := server.anomalies.Get(context.Request.Context(), id)
	if err != nil {
		if errors.Is(err, anomalyapp.ErrNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"error": "anomaly not found"})
			return
		}
		server.internalError(context, "get anomaly", err)
		return
	}
	context.JSON(http.StatusOK, mapAnomaly(result))
}

func (server *Server) analyze(context *gin.Context) {
	result, err := server.analysis.Analyze(context.Request.Context())
	if err != nil {
		server.internalError(context, "run analysis", err)
		return
	}
	context.JSON(http.StatusOK, mapAnalysis(result))
}

func (server *Server) getAnalysis(context *gin.Context) {
	id := context.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis id"})
		return
	}
	result, err := server.analysis.GetAnalysis(context.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ai.ErrNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
			return
		}
		server.internalError(context, "get analysis", err)
		return
	}
	context.JSON(http.StatusOK, mapAnalysis(result))
}

func (server *Server) dashboardSummary(context *gin.Context) {
	result, err := server.dashboard.GetSummary(context.Request.Context())
	if err != nil {
		server.internalError(context, "load dashboard summary", err)
		return
	}
	var latestResponse *analysisResponse
	if result.LatestAnalysis != nil {
		response := mapAnalysis(*result.LatestAnalysis)
		latestResponse = &response
	}
	context.JSON(http.StatusOK, gin.H{
		"meter_count": result.MeterCount, "total_consumption_kwh": round(result.TotalConsumptionKWh),
		"anomaly_count": result.AnomalyCount, "high_priority_count": result.HighPriorityCount,
		"aggregate_confidence": result.AggregateConfidence, "latest_analysis": latestResponse,
		"period_start": result.PeriodStart, "period_end": result.PeriodEnd,
	})
}

func (server *Server) siteHistory(context *gin.Context) {
	points, err := server.meters.SiteHistory(context.Request.Context())
	if err != nil {
		server.internalError(context, "load site history", err)
		return
	}
	response := make([]siteHistoryPointResponse, 0, len(points))
	for _, point := range points {
		response = append(response, siteHistoryPointResponse{Timestamp: point.Timestamp, ConsumptionKWh: point.ConsumptionKWh})
	}
	context.JSON(http.StatusOK, gin.H{"points": response})
}

func mapMeter(result meter.Summary) meterResponse {
	response := meterResponse{
		MeterID: result.MeterID, Provider: result.Provider, Region: result.Region, RateType: result.RateType,
		CurrentConsumptionKWh: result.CurrentConsumptionKWh,
		PeriodConsumptionKWh:  result.PeriodConsumptionKWh, BaselineKWh: result.BaselineKWh,
		ChangePercent: result.ChangePercent, Status: result.Status, Severity: result.Severity,
	}
	if result.Anomaly != nil {
		mapped := mapAnomaly(*result.Anomaly)
		response.Anomaly = &mapped
	}
	return response
}

func mapReadings(results []reading.Reading) []readingResponse {
	responses := make([]readingResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, readingResponse{Timestamp: result.Timestamp, ConsumptionKWh: result.ConsumptionKWh, VoltageV: result.VoltageV, CurrentA: result.CurrentA, PowerFactor: result.PowerFactor, Status: result.Status})
	}
	return responses
}

func mapAnomaly(result anomaly.Anomaly) anomalyResponse {
	return anomalyResponse{
		ID: result.ID, MeterID: result.MeterID, Anomaly: result.Anomaly, Type: result.Type, Severity: result.Severity,
		Confidence: result.Confidence, Reason: result.Reason, RecommendedAction: result.RecommendedAction,
		DetectedAt: result.DetectedAt,
		Evidence: evidenceResponse{BaselineKWh: result.Evidence.BaselineKWh, CurrentKWh: result.Evidence.CurrentKWh,
			ChangePercent: result.Evidence.ChangePercent, VoltageV: result.Evidence.VoltageV, CurrentA: result.Evidence.CurrentA,
			PowerFactor: result.Evidence.PowerFactor, CurrentVariationPercent: result.Evidence.CurrentVariationPercent,
			RelatedEvents: result.Evidence.RelatedEvents, AbnormalHours: result.Evidence.AbnormalHours, OutlierCount: result.Evidence.OutlierCount},
	}
}

func matchesFilter(status, filter string) bool {
	switch filter {
	case "all", "":
		return true
	case "normal", "normals":
		return status == "NORMAL"
	case "alert", "alerts":
		return status == "ALERT"
	case "critical", "criticals":
		return status == "CRITICAL"
	default:
		return true
	}
}

func sortMeters(results []meter.Summary, field, direction string) {
	descending := strings.ToLower(direction) != "asc"
	sort.SliceStable(results, func(i, j int) bool {
		switch field {
		case "consumption":
			if descending {
				return results[i].CurrentConsumptionKWh > results[j].CurrentConsumptionKWh
			}
			return results[i].CurrentConsumptionKWh < results[j].CurrentConsumptionKWh
		case "variation":
			if descending {
				return results[i].ChangePercent > results[j].ChangePercent
			}
			return results[i].ChangePercent < results[j].ChangePercent
		default:
			if descending {
				return severityRank(results[i].Severity) < severityRank(results[j].Severity)
			}
			return severityRank(results[i].Severity) > severityRank(results[j].Severity)
		}
	})
}

func severityRank(severity anomaly.Severity) int {
	switch severity {
	case anomaly.Critical:
		return 0
	case anomaly.High:
		return 1
	case anomaly.Medium:
		return 2
	case anomaly.Low:
		return 3
	default:
		return 4
	}
}

func validMeterID(meterID string) bool {
	if len(meterID) != 5 || meterID[:2] != "M-" {
		return false
	}
	for _, digit := range meterID[2:] {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func round(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func (server *Server) internalError(context *gin.Context, operation string, err error) {
	server.logger.Error("request failed", "operation", operation, "error", err)
	context.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
