package application

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"learning/internal/billing"
)

const (
	ReportFormatPDF   = "pdf"
	ReportFormatExcel = "excel"
)

type CostSummaryResult struct {
	TotalCost         float64
	PreviousTotalCost float64
	ChangePercent     float64
	TopMeterID        string
	AverageDailyCost  float64
	Currency          string
	PeriodStart       time.Time
	PeriodEnd         time.Time
}

type CostTimePoint struct {
	Label     string
	Date      time.Time
	TotalCost float64
	KWh       float64
}

type MeterCostRow struct {
	MeterID     string
	KWh         float64
	TariffName  string
	DailyCost   float64
	WeeklyCost  float64
	PeriodCost  float64
	Trend       float64
	Calculated  bool
	Currency    string
}

type ReportFile struct {
	Content     []byte
	FileName    string
	ContentType string
}

type CreateTariffInput struct {
	Provider    string
	Region      string
	RateType    string
	PricePerKWh float64
	Currency    string
	ValidFrom   time.Time
	ValidTo     *time.Time
}

type EstimateConsumptionResult struct {
	TotalKWh    float64
	PricePerKWh float64
	TotalCost   float64
	Currency    string
}

type CreateTariffUseCase struct {
	repository billing.TariffRepositoryPort
}

func NewCreateTariffUseCase(repository billing.TariffRepositoryPort) *CreateTariffUseCase {
	return &CreateTariffUseCase{repository: repository}
}

func (useCase *CreateTariffUseCase) Execute(ctx context.Context, input CreateTariffInput) (billing.Tariff, error) {
	tariff, err := billing.NewTariff(input.Provider, input.Region, input.RateType, input.PricePerKWh, input.Currency, input.ValidFrom, input.ValidTo)
	if err != nil {
		return billing.Tariff{}, err
	}
	return useCase.repository.Save(ctx, tariff)
}

func (useCase *CreateTariffUseCase) Create(ctx context.Context, input CreateTariffInput) (billing.Tariff, error) {
	return useCase.Execute(ctx, input)
}

type ListTariffsUseCase struct {
	repository billing.TariffRepositoryPort
}

func NewListTariffsUseCase(repository billing.TariffRepositoryPort) *ListTariffsUseCase {
	return &ListTariffsUseCase{repository: repository}
}

func (useCase *ListTariffsUseCase) Execute(ctx context.Context, provider, region string, active bool) ([]billing.Tariff, error) {
	return useCase.repository.List(ctx, provider, region, active)
}

func (useCase *ListTariffsUseCase) List(ctx context.Context, provider, region string, active bool) ([]billing.Tariff, error) {
	return useCase.Execute(ctx, provider, region, active)
}

type GetActiveTariffForMeterUseCase struct {
	tariffRepository billing.TariffRepositoryPort
	meterRepository  billing.MeterProfileRepositoryPort
}

func NewGetActiveTariffForMeterUseCase(tariffRepository billing.TariffRepositoryPort, meterRepository billing.MeterProfileRepositoryPort) *GetActiveTariffForMeterUseCase {
	return &GetActiveTariffForMeterUseCase{tariffRepository: tariffRepository, meterRepository: meterRepository}
}

func (useCase *GetActiveTariffForMeterUseCase) Execute(ctx context.Context, meterID string) (billing.Tariff, error) {
	profile, err := useCase.meterRepository.GetMeterProfile(ctx, meterID)
	if err != nil {
		return billing.Tariff{}, err
	}
	if strings.TrimSpace(profile.Provider) == "" || strings.TrimSpace(profile.Region) == "" || strings.TrimSpace(profile.RateType) == "" {
		return billing.Tariff{}, fmt.Errorf("%w: meter %s is missing provider, region or rate_type configuration", billing.ErrMeterNotConfigured, meterID)
	}
	return useCase.tariffRepository.GetActiveForMeter(ctx, profile.Provider, profile.Region, profile.RateType)
}

func (useCase *GetActiveTariffForMeterUseCase) GetForMeter(ctx context.Context, meterID string) (billing.Tariff, error) {
	return useCase.Execute(ctx, meterID)
}

type EstimateMeterConsumptionCostUseCase struct {
	billingRepository billing.BillingRepositoryPort
	meterRepository   billing.MeterProfileRepositoryPort
	tariffRepository  billing.TariffRepositoryPort
	consumption       billing.ConsumptionEstimationPort
}

func NewEstimateMeterConsumptionCostUseCase(
	billingRepository billing.BillingRepositoryPort,
	meterRepository billing.MeterProfileRepositoryPort,
	tariffRepository billing.TariffRepositoryPort,
	consumption billing.ConsumptionEstimationPort,
) *EstimateMeterConsumptionCostUseCase {
	return &EstimateMeterConsumptionCostUseCase{billingRepository: billingRepository, meterRepository: meterRepository, tariffRepository: tariffRepository, consumption: consumption}
}

func (useCase *EstimateMeterConsumptionCostUseCase) Execute(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (EstimateConsumptionResult, error) {
	if meterID == "" {
		return EstimateConsumptionResult{}, fmt.Errorf("%w: meter_id is required", billing.ErrInvalidPeriod)
	}
	if periodEnd.Before(periodStart) || periodEnd.Equal(periodStart) {
		return EstimateConsumptionResult{}, fmt.Errorf("%w: period_end must be greater than period_start", billing.ErrInvalidPeriod)
	}
	profile, err := useCase.meterRepository.GetMeterProfile(ctx, meterID)
	if err != nil {
		return EstimateConsumptionResult{}, fmt.Errorf("resolve meter profile for %s: %w", meterID, err)
	}
	if strings.TrimSpace(profile.Provider) == "" || strings.TrimSpace(profile.Region) == "" || strings.TrimSpace(profile.RateType) == "" {
		return EstimateConsumptionResult{}, fmt.Errorf("%w: meter %s is missing provider, region or rate_type configuration", billing.ErrMeterNotConfigured, meterID)
	}
	tariff, err := useCase.tariffRepository.GetActiveForMeter(ctx, profile.Provider, profile.Region, profile.RateType)
	if err != nil {
		return EstimateConsumptionResult{}, fmt.Errorf("resolve active tariff for meter %s: %w", meterID, err)
	}
	totalKWh, err := useCase.consumption.SumConsumption(ctx, meterID, periodStart, periodEnd)
	if err != nil {
		return EstimateConsumptionResult{}, fmt.Errorf("sum consumption for meter %s: %w", meterID, err)
	}
	bill, err := billing.NewBilling(meterID, &tariff.TariffID, periodStart, periodEnd, totalKWh, tariff.PricePerKWh, tariff.Currency)
	if err != nil {
		return EstimateConsumptionResult{}, fmt.Errorf("build billing record: %w", err)
	}
	if _, err := useCase.billingRepository.Save(ctx, bill); err != nil {
		return EstimateConsumptionResult{}, fmt.Errorf("save billing record: %w", err)
	}
	slog.Info("billing estimate generated",
		"meter_id", meterID,
		"provider", profile.Provider,
		"region", profile.Region,
		"rate_type", profile.RateType,
		"period_start", periodStart.Format(time.RFC3339),
		"period_end", periodEnd.Format(time.RFC3339),
		"consumption_kwh", totalKWh,
		"price_per_kwh", tariff.PricePerKWh,
		"total_cost", bill.TotalCost,
		"currency", bill.Currency,
		"tariff_id", bill.TariffID,
	)
	result := EstimateConsumptionResult{TotalKWh: bill.TotalKWh, PricePerKWh: bill.PricePerKWh, TotalCost: bill.TotalCost, Currency: bill.Currency}
	return result, nil
}

func (useCase *EstimateMeterConsumptionCostUseCase) Estimate(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (EstimateConsumptionResult, error) {
	return useCase.Execute(ctx, meterID, periodStart, periodEnd)
}

type GetBillingHistoryByMeterUseCase struct {
	repository billing.BillingRepositoryPort
}

func NewGetBillingHistoryByMeterUseCase(repository billing.BillingRepositoryPort) *GetBillingHistoryByMeterUseCase {
	return &GetBillingHistoryByMeterUseCase{repository: repository}
}

func (useCase *GetBillingHistoryByMeterUseCase) Execute(ctx context.Context, meterID string) ([]billing.Billing, error) {
	if meterID == "" {
		return nil, fmt.Errorf("%w: meter_id is required", billing.ErrInvalidPeriod)
	}
	return useCase.repository.ListByMeter(ctx, meterID)
}

func (useCase *GetBillingHistoryByMeterUseCase) History(ctx context.Context, meterID string) ([]billing.Billing, error) {
	return useCase.Execute(ctx, meterID)
}

type GetBillingSummaryUseCase struct {
	repository billing.BillingRepositoryPort
}

func NewGetBillingSummaryUseCase(repository billing.BillingRepositoryPort) *GetBillingSummaryUseCase {
	return &GetBillingSummaryUseCase{repository: repository}
}

func (useCase *GetBillingSummaryUseCase) Execute(ctx context.Context, from, to time.Time) (billing.BillingSummary, error) {
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return billing.BillingSummary{}, fmt.Errorf("%w: from cannot be greater than to", billing.ErrInvalidPeriod)
	}
	return useCase.repository.SummaryByPeriod(ctx, from, to)
}

func (useCase *GetBillingSummaryUseCase) Summary(ctx context.Context, from, to time.Time) (billing.BillingSummary, error) {
	return useCase.Execute(ctx, from, to)
}

type GetCostSummaryUseCase struct {
	repository billing.BillingRepositoryPort
}

func NewGetCostSummaryUseCase(repository billing.BillingRepositoryPort) *GetCostSummaryUseCase {
	return &GetCostSummaryUseCase{repository: repository}
}

func (useCase *GetCostSummaryUseCase) Execute(ctx context.Context, from, to time.Time) (CostSummaryResult, error) {
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return CostSummaryResult{}, fmt.Errorf("%w: from cannot be greater than to", billing.ErrInvalidPeriod)
	}
	current, err := useCase.repository.SummaryByPeriod(ctx, from, to)
	if err != nil {
		return CostSummaryResult{}, err
	}
	periodDays := 1
	if !from.IsZero() && !to.IsZero() {
		periodDays = int(to.Sub(from).Hours()/24) + 1
		if periodDays < 1 {
			periodDays = 1
		}
	}
	previousFrom := from.AddDate(0, 0, -periodDays)
	previousTo := to.AddDate(0, 0, -periodDays)
	if from.IsZero() || to.IsZero() {
		previousFrom, previousTo = time.Time{}, time.Time{}
	}
	previous, err := useCase.repository.SummaryByPeriod(ctx, previousFrom, previousTo)
	if err != nil && !from.IsZero() && !to.IsZero() {
		previous = billing.BillingSummary{TotalCost: 0, Currency: current.Currency}
	}
	rows, err := useCase.repository.ListByPeriod(ctx, from, to)
	if err != nil {
		return CostSummaryResult{}, err
	}
	topID := "—"
	topCost := 0.0
	meterTotals := map[string]float64{}
	for _, item := range rows {
		meterTotals[item.MeterID] += item.TotalCost
	}
	for meterID, meterTotal := range meterTotals {
		if topID == "—" || meterTotal > topCost {
			topID = meterID
			topCost = meterTotal
		}
	}
	change := 0.0
	if previous.TotalCost > 0 {
		change = ((current.TotalCost - previous.TotalCost) / previous.TotalCost) * 100
	}
	return CostSummaryResult{
		TotalCost:         current.TotalCost,
		PreviousTotalCost: previous.TotalCost,
		ChangePercent:     change,
		TopMeterID:        topID,
		AverageDailyCost:  current.TotalCost / float64(periodDays),
		Currency:          current.Currency,
		PeriodStart:       from,
		PeriodEnd:         to,
	}, nil
}

type GetCostByMeterUseCase struct {
	billingRepository billing.BillingRepositoryPort
	tariffRepository  billing.TariffRepositoryPort
}

func NewGetCostByMeterUseCase(billingRepository billing.BillingRepositoryPort, tariffRepository billing.TariffRepositoryPort) *GetCostByMeterUseCase {
	return &GetCostByMeterUseCase{billingRepository: billingRepository, tariffRepository: tariffRepository}
}

func (useCase *GetCostByMeterUseCase) Execute(ctx context.Context, from, to time.Time) ([]MeterCostRow, error) {
	rows, err := useCase.billingRepository.ListByPeriod(ctx, from, to)
	if err != nil {
		return nil, err
	}
	periodDays := measuredPeriodDays(from, to)
	grouped := map[string]MeterCostRow{}
	for _, item := range rows {
		entry := grouped[item.MeterID]
		entry.MeterID = item.MeterID
		entry.KWh += item.TotalKWh
		entry.PeriodCost += item.TotalCost
		entry.Currency = item.Currency
		entry.Calculated = true
		if item.TariffID != nil {
			tariff, err := useCase.tariffRepository.GetByID(ctx, *item.TariffID)
			if err == nil {
				entry.TariffName = tariff.Provider + " / " + tariff.RateType
			}
		}
		if entry.TariffName == "" {
			entry.TariffName = "Tarifa vigente"
		}
		grouped[item.MeterID] = entry
	}
	results := make([]MeterCostRow, 0, len(grouped))
	for _, item := range grouped {
		dailyAverage := item.PeriodCost / float64(periodDays)
		item.DailyCost = dailyAverage
		item.WeeklyCost = dailyAverage * 7
		results = append(results, item)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].PeriodCost > results[j].PeriodCost })
	return results, nil
}

func measuredPeriodDays(from, to time.Time) int {
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return 1
	}
	days := int(to.Sub(from).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

type GetCostTimeSeriesUseCase struct {
	repository billing.BillingRepositoryPort
}

func NewGetCostTimeSeriesUseCase(repository billing.BillingRepositoryPort) *GetCostTimeSeriesUseCase {
	return &GetCostTimeSeriesUseCase{repository: repository}
}

func (useCase *GetCostTimeSeriesUseCase) Execute(ctx context.Context, from, to time.Time, granularity string) ([]CostTimePoint, error) {
	rows, err := useCase.repository.ListByPeriod(ctx, from, to)
	if err != nil {
		return nil, err
	}
	points := map[string]CostTimePoint{}
	for _, item := range rows {
		key := item.PeriodStart.Format(time.DateOnly)
		switch granularity {
		case "weekly":
			_, week := item.PeriodStart.ISOWeek()
			key = fmt.Sprintf("Semana %02d", week)
		case "monthly":
			key = item.PeriodStart.Format("Jan 2006")
		}
		entry := points[key]
		entry.Label = key
		entry.Date = item.PeriodStart
		entry.TotalCost += item.TotalCost
		entry.KWh += item.TotalKWh
		points[key] = entry
	}
	result := make([]CostTimePoint, 0, len(points))
	for _, point := range points {
		result = append(result, point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date.Before(result[j].Date) })
	return result, nil
}

type GenerateBillingReportUseCase struct {
	reportGenerator billing.ReportGeneratorPort
	billingRepo     billing.BillingRepositoryPort
	tariffRepo      billing.TariffRepositoryPort
}

func NewGenerateBillingReportUseCase(reportGenerator billing.ReportGeneratorPort, billingRepo billing.BillingRepositoryPort, tariffRepo billing.TariffRepositoryPort) *GenerateBillingReportUseCase {
	return &GenerateBillingReportUseCase{reportGenerator: reportGenerator, billingRepo: billingRepo, tariffRepo: tariffRepo}
}

func (useCase *GenerateBillingReportUseCase) Execute(ctx context.Context, from, to time.Time, format string, meterID *string) (ReportFile, error) {
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return ReportFile{}, fmt.Errorf("%w: from cannot be greater than to", billing.ErrInvalidPeriod)
	}
	rows, err := useCase.billingRepo.ListByPeriod(ctx, from, to)
	if err != nil {
		return ReportFile{}, err
	}
	filtered := make([]billing.Billing, 0, len(rows))
	for _, item := range rows {
		if meterID != nil && item.MeterID != *meterID {
			continue
		}
		filtered = append(filtered, item)
	}
	reportRows := make([]billing.MeterBillingReportRow, 0, len(filtered))
	totalCost := 0.0
	for _, item := range filtered {
		row := billing.MeterBillingReportRow{MeterID: item.MeterID, KWh: item.TotalKWh, Cost: item.TotalCost, Currency: item.Currency, Calculated: true}
		totalCost += item.TotalCost
		if item.TariffID != nil {
			tariff, err := useCase.tariffRepo.GetByID(ctx, *item.TariffID)
			if err == nil {
				row.TariffName = tariff.Provider + " / " + tariff.RateType
			}
		}
		if row.TariffName == "" {
			row.TariffName = "Tarifa vigente"
		}
		reportRows = append(reportRows, row)
	}
	data := billing.BillingReportData{
		From:        from,
		To:          to,
		Currency:    "COP",
		TotalCost:   totalCost,
		Rows:        reportRows,
		PeriodLabel: fmt.Sprintf("%s a %s", formatPeriodLabel(from), formatPeriodLabel(to)),
	}
	if format == "" {
		format = ReportFormatPDF
	}
	switch format {
	case ReportFormatPDF:
		content, err := useCase.reportGenerator.GenerateBillingReportPDF(data)
		if err != nil {
			return ReportFile{}, err
		}
		return ReportFile{Content: content, FileName: "billing-report.pdf", ContentType: "application/pdf"}, nil
	case ReportFormatExcel:
		content, err := useCase.reportGenerator.GenerateBillingReportExcel(data)
		if err != nil {
			return ReportFile{}, err
		}
		return ReportFile{Content: content, FileName: "billing-report.xlsx", ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}, nil
	default:
		return ReportFile{}, fmt.Errorf("%w: unsupported report format %q", billing.ErrInvalidPeriod, format)
	}
}

func formatPeriodLabel(value time.Time) string {
	if value.IsZero() {
		return "todo el período"
	}
	return value.Format(time.DateOnly)
}
