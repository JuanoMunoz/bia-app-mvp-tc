export type Severity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
export type Section = 'dashboard' | 'meters' | 'costs' | 'anomalies'

export type AnomalyType =
    | 'REAL_ANOMALY'
    | 'EXPLAINABLE_ANOMALY'
    | 'FALSE_POSITIVE'
    | 'DATA_QUALITY'

export interface Evidence {
    baseline_kwh: number
    current_kwh: number
    change_percent: number
    voltage_v: number
    current_a: number
    power_factor: number
    current_variation_percent: number
    related_events: string[]
    abnormal_hours: number[]
    outlier_count: number
}

export interface Anomaly {
    id: string
    meter_id: string
    anomaly: string
    type: AnomalyType
    severity: Severity
    confidence: number
    reason: string
    recommended_action: string
    detected_at: string
    evidence: Evidence
}

export interface Reading {
    timestamp: string
    consumption_kwh: number
    voltage_v: number
    current_a: number
    power_factor: number
    status: string
}

export interface SitePoint {
    timestamp: string
    consumption_kwh: number
}

export interface MeterSummary {
    meter_id: string
    provider?: string
    region?: string
    rate_type?: string
    current_consumption_kwh: number
    period_consumption_kwh: number
    baseline_kwh: number
    change_percent: number
    status: 'NORMAL' | 'ALERT' | 'CRITICAL'
    severity: Severity
    anomaly?: Anomaly
}

export interface MeterDetail extends MeterSummary {
    voltage_v: number
    current_a: number
    power_factor: number
    history: Reading[]
}

export interface Analysis {
    id: string
    status: 'RUNNING' | 'COMPLETED' | 'FAILED'
    started_at: string
    completed_at?: string
    anomalies: Anomaly[]
    error?: string
    insight?: AnalysisInsight
    insight_error?: string
}

export interface AnalysisInsight {
    answer: string
    explanation: string
    suggested_questions: SuggestedQuestion[]
}

export interface SuggestedQuestion {
    question: string
    answer: string
}

export interface DashboardSummary {
    meter_count: number
    total_consumption_kwh: number
    anomaly_count: number
    high_priority_count: number
    aggregate_confidence: number
    latest_analysis: Analysis | null
    period_start: string
    period_end: string
}

export interface BillingCostSummary {
    total_cost: number
    currency: string
    change_percent?: number
    previous_total_cost?: number
    average_daily_cost?: number
    top_meter_id?: string
    period_start?: string
    period_end?: string
}

export interface BillingTimePoint {
    label: string
    date?: string
    total_cost: number
    kwh: number
}

export interface CostByMeter {
    meter_id: string
    kwh: number
    tariff_name?: string
    daily_cost: number
    weekly_cost: number
    period_cost: number
    trend?: number
    calculated?: boolean
    currency?: string
}