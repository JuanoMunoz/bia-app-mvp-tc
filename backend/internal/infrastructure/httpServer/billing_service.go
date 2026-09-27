package httpServer

import (
	"context"
	"fmt"
	"time"

	aiapp "learning/internal/ai/application"
	"learning/internal/billing"
	billingapp "learning/internal/billing/application"
)

type BillingServiceImpl struct {
	createTariffUseCase                 *billingapp.CreateTariffUseCase
	listTariffsUseCase                  *billingapp.ListTariffsUseCase
	estimateMeterConsumptionCostUseCase *billingapp.EstimateMeterConsumptionCostUseCase
	getBillingHistoryByMeterUseCase     *billingapp.GetBillingHistoryByMeterUseCase
	getBillingSummaryUseCase            *billingapp.GetBillingSummaryUseCase
	getCostSummaryUseCase               *billingapp.GetCostSummaryUseCase
	getCostTimeSeriesUseCase            *billingapp.GetCostTimeSeriesUseCase
	getCostByMeterUseCase               *billingapp.GetCostByMeterUseCase
	generateBillingReportUseCase        *billingapp.GenerateBillingReportUseCase
	getMeterOptimizationInsightUseCase  *aiapp.GetConsumptionOptimizationUseCase
}

func NewBillingService(tariffRepository billing.TariffRepositoryPort, billingRepository billing.BillingRepositoryPort, meterProfileRepository billing.MeterProfileRepositoryPort, consumptionEstimator billing.ConsumptionEstimationPort, reportGenerator billing.ReportGeneratorPort) *BillingServiceImpl {
	return &BillingServiceImpl{
		createTariffUseCase:                 billingapp.NewCreateTariffUseCase(tariffRepository),
		listTariffsUseCase:                  billingapp.NewListTariffsUseCase(tariffRepository),
		estimateMeterConsumptionCostUseCase: billingapp.NewEstimateMeterConsumptionCostUseCase(billingRepository, meterProfileRepository, tariffRepository, consumptionEstimator),
		getBillingHistoryByMeterUseCase:     billingapp.NewGetBillingHistoryByMeterUseCase(billingRepository),
		getBillingSummaryUseCase:            billingapp.NewGetBillingSummaryUseCase(billingRepository),
		getCostSummaryUseCase:               billingapp.NewGetCostSummaryUseCase(billingRepository),
		getCostTimeSeriesUseCase:            billingapp.NewGetCostTimeSeriesUseCase(billingRepository),
		getCostByMeterUseCase:               billingapp.NewGetCostByMeterUseCase(billingRepository, tariffRepository),
		generateBillingReportUseCase:        billingapp.NewGenerateBillingReportUseCase(reportGenerator, billingRepository, tariffRepository),
	}
}

func (service *BillingServiceImpl) SetOptimizationInsightUseCase(useCase *aiapp.GetConsumptionOptimizationUseCase) {
	service.getMeterOptimizationInsightUseCase = useCase
}

func (service *BillingServiceImpl) CreateTariff(ctx context.Context, input tariffRequestDTO) (tariffResponseDTO, error) {
	result, err := service.createTariffUseCase.Execute(ctx, billingapp.CreateTariffInput{
		Provider:    input.Provider,
		Region:      input.Region,
		RateType:    input.RateType,
		PricePerKWh: input.PricePerKWh,
		Currency:    input.Currency,
		ValidFrom:   input.ValidFrom,
		ValidTo:     input.ValidTo,
	})
	if err != nil {
		return tariffResponseDTO{}, err
	}
	return mapTariffResponse(result), nil
}

func (service *BillingServiceImpl) ListTariffs(ctx context.Context, provider, region string, active bool) ([]tariffResponseDTO, error) {
	results, err := service.listTariffsUseCase.Execute(ctx, provider, region, active)
	if err != nil {
		return nil, err
	}
	response := make([]tariffResponseDTO, 0, len(results))
	for _, item := range results {
		response = append(response, mapTariffResponse(item))
	}
	return response, nil
}

func (service *BillingServiceImpl) EstimateMeterConsumptionCost(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (estimateConsumptionResponseDTO, error) {
	result, err := service.estimateMeterConsumptionCostUseCase.Execute(ctx, meterID, periodStart, periodEnd)
	if err != nil {
		return estimateConsumptionResponseDTO{}, err
	}
	return estimateConsumptionResponseDTO{
		MeterID:     meterID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalKWh:    result.TotalKWh,
		PricePerKWh: result.PricePerKWh,
		TotalCost:   result.TotalCost,
		Currency:    result.Currency,
	}, nil
}

func (service *BillingServiceImpl) GetBillingHistoryByMeter(ctx context.Context, meterID string) ([]billingResponseDTO, error) {
	results, err := service.getBillingHistoryByMeterUseCase.Execute(ctx, meterID)
	if err != nil {
		return nil, err
	}
	response := make([]billingResponseDTO, 0, len(results))
	for _, item := range results {
		response = append(response, billingResponseDTO{
			BillingID:   item.BillingID,
			MeterID:     item.MeterID,
			PeriodStart: item.PeriodStart,
			PeriodEnd:   item.PeriodEnd,
			TotalKWh:    item.TotalKWh,
			PricePerKWh: item.PricePerKWh,
			TotalCost:   item.TotalCost,
			Currency:    item.Currency,
			TariffID:    item.TariffID,
		})
	}
	return response, nil
}

func (service *BillingServiceImpl) GetBillingSummary(ctx context.Context, from, to time.Time) (billingSummaryResponseDTO, error) {
	result, err := service.getBillingSummaryUseCase.Execute(ctx, from, to)
	if err != nil {
		return billingSummaryResponseDTO{}, err
	}
	return billingSummaryResponseDTO{TotalCost: result.TotalCost, Currency: result.Currency}, nil
}

func (service *BillingServiceImpl) GetCostSummary(ctx context.Context, from, to time.Time) (billingSummaryResponseDTO, error) {
	if service.getCostSummaryUseCase == nil {
		return billingSummaryResponseDTO{}, fmt.Errorf("cost summary use case unavailable")
	}
	result, err := service.getCostSummaryUseCase.Execute(ctx, from, to)
	if err != nil {
		return billingSummaryResponseDTO{}, err
	}
	return billingSummaryResponseDTO{
		TotalCost:         result.TotalCost,
		PreviousTotalCost: result.PreviousTotalCost,
		ChangePercent:     result.ChangePercent,
		AverageDailyCost:  result.AverageDailyCost,
		TopMeterID:        result.TopMeterID,
		Currency:          result.Currency,
		PeriodStart:       result.PeriodStart,
		PeriodEnd:         result.PeriodEnd,
	}, nil
}

func (service *BillingServiceImpl) GetCostTimeSeries(ctx context.Context, from, to time.Time, granularity string) ([]costTimePointResponseDTO, error) {
	if service.getCostTimeSeriesUseCase == nil {
		return nil, fmt.Errorf("cost time series use case unavailable")
	}
	results, err := service.getCostTimeSeriesUseCase.Execute(ctx, from, to, granularity)
	if err != nil {
		return nil, err
	}
	response := make([]costTimePointResponseDTO, 0, len(results))
	for _, item := range results {
		response = append(response, costTimePointResponseDTO{Label: item.Label, Date: item.Date, TotalCost: item.TotalCost, KWh: item.KWh})
	}
	return response, nil
}

func (service *BillingServiceImpl) GetCostByMeter(ctx context.Context, from, to time.Time) ([]costByMeterRowResponseDTO, error) {
	if service.getCostByMeterUseCase == nil {
		return nil, fmt.Errorf("cost by meter use case unavailable")
	}
	results, err := service.getCostByMeterUseCase.Execute(ctx, from, to)
	if err != nil {
		return nil, err
	}
	response := make([]costByMeterRowResponseDTO, 0, len(results))
	for _, item := range results {
		response = append(response, costByMeterRowResponseDTO{
			MeterID:     item.MeterID,
			KWh:         item.KWh,
			TariffName:  item.TariffName,
			DailyCost:   item.DailyCost,
			WeeklyCost:  item.WeeklyCost,
			PeriodCost:  item.PeriodCost,
			Trend:       item.Trend,
			Calculated:  item.Calculated,
			Currency:    item.Currency,
		})
	}
	return response, nil
}

func (service *BillingServiceImpl) GenerateBillingReport(ctx context.Context, format string, from, to time.Time, meterID *string) (billingReportFileResponseDTO, error) {
	if service.generateBillingReportUseCase == nil {
		return billingReportFileResponseDTO{}, fmt.Errorf("billing report use case unavailable")
	}
	result, err := service.generateBillingReportUseCase.Execute(ctx, from, to, format, meterID)
	if err != nil {
		return billingReportFileResponseDTO{}, err
	}
	return billingReportFileResponseDTO{Content: result.Content, FileName: result.FileName, ContentType: result.ContentType}, nil
}

func (service *BillingServiceImpl) GetMeterOptimizationInsight(ctx context.Context, meterID string, from, to time.Time) (optimizationInsightResponseDTO, error) {
	if service.getMeterOptimizationInsightUseCase == nil {
		return optimizationInsightResponseDTO{}, fmt.Errorf("optimization insight service unavailable")
	}
	result, err := service.getMeterOptimizationInsightUseCase.Execute(ctx, meterID, from, to)
	if err != nil {
		return optimizationInsightResponseDTO{}, err
	}
	return optimizationInsightResponseDTO{
		Summary:              result.Summary,
		Recommendations:      result.Recommendations,
		PotentialSavingsKWh:  result.PotentialSavingsKWh,
		PotentialSavingsCost: result.PotentialSavingsCost,
		Currency:             result.Currency,
		GeneratedAt:          result.GeneratedAt,
	}, nil
}

func mapTariffResponse(item billing.Tariff) tariffResponseDTO {
	return tariffResponseDTO{
		TariffID:    item.TariffID,
		Provider:    item.Provider,
		Region:      item.Region,
		RateType:    item.RateType,
		PricePerKWh: item.PricePerKWh,
		Currency:    item.Currency,
		ValidFrom:   item.ValidFrom,
		ValidTo:     item.ValidTo,
		CreatedAt:   item.CreatedAt,
	}
}

func (service *BillingServiceImpl) String() string {
	return fmt.Sprintf("billing-service[%T]", service)
}
