package billing

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidTariff      = errors.New("invalid tariff")
	ErrInvalidPeriod      = errors.New("invalid billing period")
	ErrNotFound           = errors.New("billing record not found")
	ErrMeterNotConfigured = errors.New("meter not configured")
)

/*
*
Tariff representa la estructura tarifaria aplicada a un medidor según proveedor,
región y tipo de tarifa.

Es la entidad de negocio que define el precio por kWh y el periodo de vigencia
con el que se calcula la facturación y la estimación del costo energético.
*/
type Tariff struct {
	TariffID    int
	Provider    string
	Region      string
	RateType    string
	PricePerKWh float64
	Currency    string
	ValidFrom   time.Time
	ValidTo     *time.Time
	CreatedAt   time.Time
}

// NewTariff valida y crea una tarifa con su vigencia y precio asociados.
func NewTariff(provider, region, rateType string, pricePerKWh float64, currency string, validFrom time.Time, validTo *time.Time) (Tariff, error) {
	if provider == "" {
		return Tariff{}, fmt.Errorf("%w: provider is required", ErrInvalidTariff)
	}
	if rateType == "" {
		rateType = "industrial"
	}
	if pricePerKWh <= 0 {
		return Tariff{}, fmt.Errorf("%w: price_per_kwh must be greater than zero", ErrInvalidTariff)
	}
	if validFrom.IsZero() {
		return Tariff{}, fmt.Errorf("%w: valid_from is required", ErrInvalidTariff)
	}
	if validTo != nil && validTo.Before(validFrom) {
		return Tariff{}, fmt.Errorf("%w: valid_to must be after valid_from", ErrInvalidTariff)
	}
	if currency == "" {
		currency = "COP"
	}
	return Tariff{Provider: provider, Region: region, RateType: rateType, PricePerKWh: pricePerKWh, Currency: currency, ValidFrom: validFrom, ValidTo: validTo}, nil
}

// IsActiveOn indica si una tarifa está vigente para una fecha determinada.
func (t Tariff) IsActiveOn(date time.Time) bool {
	if date.Before(t.ValidFrom) {
		return false
	}
	if t.ValidTo != nil && date.After(*t.ValidTo) {
		return false
	}
	return true
}

/*
*
Billing representa la facturación generada para un medidor en un periodo concreto.

Es la entidad que consolida consumo, tarifa y costo para dar trazabilidad a la
factura y permitir comparaciones por medidor, periodo y tendencia de gasto.
*/
type Billing struct {
	BillingID   int
	MeterID     string
	TariffID    *int
	PeriodStart time.Time
	PeriodEnd   time.Time
	TotalKWh    float64
	PricePerKWh float64
	TotalCost   float64
	Currency    string
	CreatedAt   time.Time
}

// NewBilling valida y crea un registro de facturación con su período y tarifa asociados.
func NewBilling(meterID string, tariffID *int, periodStart, periodEnd time.Time, totalKWh, pricePerKWh float64, currency string) (Billing, error) {
	if meterID == "" {
		return Billing{}, fmt.Errorf("%w: meter_id is required", ErrInvalidPeriod)
	}
	if periodStart.IsZero() || periodEnd.IsZero() {
		return Billing{}, fmt.Errorf("%w: period_start and period_end are required", ErrInvalidPeriod)
	}
	if periodEnd.Before(periodStart) || periodEnd.Equal(periodStart) {
		return Billing{}, fmt.Errorf("%w: period_end must be greater than period_start", ErrInvalidPeriod)
	}
	if totalKWh < 0 {
		return Billing{}, fmt.Errorf("%w: total_kwh cannot be negative", ErrInvalidPeriod)
	}
	if pricePerKWh <= 0 {
		return Billing{}, fmt.Errorf("%w: price_per_kwh must be greater than zero", ErrInvalidPeriod)
	}
	if currency == "" {
		currency = "COP"
	}
	return Billing{MeterID: meterID, TariffID: tariffID, PeriodStart: periodStart, PeriodEnd: periodEnd, TotalKWh: totalKWh, PricePerKWh: pricePerKWh, TotalCost: totalKWh * pricePerKWh, Currency: currency}, nil
}

/*
*
Money modela una cantidad monetaria con su moneda de referencia.

Es útil en el dominio para encapsular el manejo del dinero y asegurar que los
cálculos de costo se operen en el contexto correcto de moneda y redondeo.
*/
type Money struct {
	Amount   float64
	Currency string
}

// Rounded normaliza el valor monetario a dos decimales según la convención del dominio.
func (m Money) Rounded() Money {
	return Money{Amount: round(m.Amount), Currency: m.Currency}
}

func round(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

/*
*
BillingSummary es el agregado financiero de un periodo.

Responde a la pregunta: ¿cuánto costó el consumo total en este intervalo? y se
usa para resumir dashboards, comparativas y decisiones en tiempo real.
*/
type BillingSummary struct {
	TotalCost float64
	Currency  string
}

type TariffRepositoryPort interface {
	Save(ctx context.Context, tariff Tariff) (Tariff, error)
	List(ctx context.Context, provider string, region string, active bool) ([]Tariff, error)
	GetByID(ctx context.Context, tariffID int) (Tariff, error)
	GetActiveForMeter(ctx context.Context, provider string, region string, rateType string) (Tariff, error)
}

type BillingRepositoryPort interface {
	Save(ctx context.Context, billing Billing) (Billing, error)
	ListByMeter(ctx context.Context, meterID string) ([]Billing, error)
	ListByPeriod(ctx context.Context, from, to time.Time) ([]Billing, error)
	SummaryByPeriod(ctx context.Context, from, to time.Time) (BillingSummary, error)
}

/*
*
MeterProfile representa el perfil tarifario del medidor dentro del negocio.

Se usa para saber qué proveedor, región y tipo de tarifa aplica al medidor,
lo cual evita decisiones ambiguas al resolver la tarifa activa correcta.
*/
type MeterProfile struct {
	MeterID  string
	Provider string
	Region   string
	RateType string
}

type MeterProfileRepositoryPort interface {
	GetMeterProfile(ctx context.Context, meterID string) (MeterProfile, error)
	GetProviderAndRegion(ctx context.Context, meterID string) (string, string, error)
	UpdateMeterProfile(ctx context.Context, meterID string, provider, region, rateType string) error
}

type ConsumptionEstimationPort interface {
	SumConsumption(ctx context.Context, meterID string, periodStart, periodEnd time.Time) (float64, error)
}

type BillingReportData struct {
	From        time.Time
	To          time.Time
	Currency    string
	TotalCost   float64
	Rows        []MeterBillingReportRow
	PeriodLabel string
}

type MeterBillingReportRow struct {
	MeterID    string
	KWh        float64
	TariffName string
	Cost       float64
	Currency   string
	Trend      float64
	Calculated bool
}

type ReportGeneratorPort interface {
	GenerateBillingReportPDF(data BillingReportData) ([]byte, error)
	GenerateBillingReportExcel(data BillingReportData) ([]byte, error)
}
