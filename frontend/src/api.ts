import type {
    Analysis,
    Anomaly,
    BillingCostSummary,
    BillingTimePoint,
    CostByMeter,
    DashboardSummary,
    MeterDetail,
    MeterSummary,
    SitePoint,
} from './types'

const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
const apiVersion = '/api/v1'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${apiBase}${path}`, {
        ...init,
        headers: { 'Content-Type': 'application/json', ...init?.headers },
    })
    if (!response.ok) {
        const body = (await response.json().catch(() => null)) as { error?: string } | null
        throw new Error(body?.error ?? `Request failed (${response.status})`)
    }
    return response.json() as Promise<T>
}

export async function getDashboardSummary() {
    return request<DashboardSummary>(`${apiVersion}/dashboard/summary`)
}

export async function getMeters() {
    const response = await request<{ meters: MeterSummary[] }>(`${apiVersion}/meters`)
    return response.meters
}

export async function getMeter(meterId: string) {
    return request<MeterDetail>(`${apiVersion}/meters/${encodeURIComponent(meterId)}`)
}

export async function getSiteHistory() {
    const response = await request<{ points: SitePoint[] }>(`${apiVersion}/dashboard/history`)
    return response.points
}

export async function estimateMeterConsumptionCost(meterId: string, periodStart: string, periodEnd: string) {
    const params = new URLSearchParams({
        period_start: periodStart,
        period_end: periodEnd,
    })
    return request<{ meter_id: string; total_cost: number; total_kwh: number; price_per_kwh: number; currency: string; period_start: string; period_end: string }>(
        `${apiVersion}/meters/${encodeURIComponent(meterId)}/estimated-consumption?${params.toString()}`,
    )
}

export async function getAnomalies() {
    const response = await request<{ anomalies: Anomaly[] }>(`${apiVersion}/anomalies`)
    return response.anomalies
}

export async function getAnomaly(id: string) {
    return request<Anomaly>(`${apiVersion}/anomalies/${encodeURIComponent(id)}`)
}

export async function runAnalysis() {
    return request<Analysis>(`${apiVersion}/ai/analyze`, { method: 'POST' })
}

export async function getCostSummary(from?: string, to?: string) {
    const params = new URLSearchParams()
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    const query = params.toString() ? `?${params.toString()}` : ''
    return request<BillingCostSummary>(`${apiVersion}/billing/summary${query}`)
}

export async function getCostTimeSeries(from?: string, to?: string, granularity: 'daily' | 'weekly' | 'monthly' = 'daily') {
    const params = new URLSearchParams({ granularity })
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    return request<{ series: BillingTimePoint[] }>(`${apiVersion}/billing/history?${params.toString()}`)
}

export async function getCostByMeter(from?: string, to?: string) {
    const params = new URLSearchParams()
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    const query = params.toString() ? `?${params.toString()}` : ''
    return request<{ rows: CostByMeter[] }>(`${apiVersion}/billing/meters${query}`)
}

export async function updateMeterConfig(meterId: string, payload: { provider?: string; region?: string; rate_type?: string }) {
    return request<{ meter_id: string; provider?: string; region?: string; rate_type?: string }>(`${apiVersion}/meters/${encodeURIComponent(meterId)}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
    })
}

export async function downloadBillingReport(format: 'pdf' | 'excel', from?: string, to?: string, meterId?: string) {
    const params = new URLSearchParams({ format })
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    if (meterId) params.set('meter_id', meterId)
    const response = await fetch(`${apiBase}${apiVersion}/reports/billing?${params.toString()}`)
    if (!response.ok) {
        const body = (await response.json().catch(() => null)) as { error?: string } | null
        throw new Error(body?.error ?? `No se pudo descargar el reporte (${response.status})`)
    }
    return response.blob()
}