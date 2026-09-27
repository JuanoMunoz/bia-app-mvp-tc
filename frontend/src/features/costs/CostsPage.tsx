import { useEffect, useMemo, useState } from 'react'
import {
    ArrowDownRight,
    ArrowUpRight,
    Download,
    FileText,
    Gauge,
    TrendingUp,
    Wallet,
    Zap,
} from 'lucide-react'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { downloadBillingReport, estimateMeterConsumptionCost, getCostByMeter, getCostSummary, getCostTimeSeries, getMeters } from '../../api'
import { InfoTooltip, KpiCard, PanelTitle } from '../../components/ui'
import type { BillingCostSummary, BillingTimePoint, CostByMeter } from '../../types'

const PERIODS = [
    { id: '7d', label: '7 días', days: 7 },
    { id: '30d', label: '30 días', days: 30 },
    { id: '90d', label: '90 días', days: 90 },
] as const

function formatCurrency(value: number, currency = 'COP') {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency,
        maximumFractionDigits: 0,
    }).format(value)
}

function formatDateLabel(value: string | undefined) {
    if (!value) return 'Periodo'
    if (/^\d{4}-\d{2}-\d{2}$/.test(value)) {
        const [year, month, day] = value.split('-').map(Number)
        return new Intl.DateTimeFormat('es-CO', { day: '2-digit', month: 'short' }).format(new Date(year, month - 1, day))
    }
    return value
}

export function CostsPage() {
    const [range, setRange] = useState<(typeof PERIODS)[number]['id']>('30d')
    const [granularity, setGranularity] = useState<'daily' | 'weekly' | 'monthly'>('daily')
    const [meterIds, setMeterIds] = useState<string[]>([])
    const [summary, setSummary] = useState<BillingCostSummary | null>(null)
    const [series, setSeries] = useState<BillingTimePoint[]>([])
    const [rows, setRows] = useState<CostByMeter[]>([])
    const [loading, setLoading] = useState(true)
    const [generating, setGenerating] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [reporting, setReporting] = useState<'pdf' | 'excel' | null>(null)

    function fetchCostData(rangeId: (typeof PERIODS)[number]['id'], granularityValue: 'daily' | 'weekly' | 'monthly') {
        const days = PERIODS.find((item) => item.id === rangeId)?.days ?? 30
        const to = new Date()
        const from = new Date()
        from.setDate(to.getDate() - days)

        setLoading(true)
        setError(null)

        return Promise.all([
            getCostSummary(from.toISOString(), to.toISOString()),
            getCostTimeSeries(from.toISOString(), to.toISOString(), granularityValue),
            getCostByMeter(from.toISOString(), to.toISOString()),
        ])
            .then(([nextSummary, nextSeries, nextRows]) => {
                setSummary(nextSummary)
                setSeries(nextSeries.series ?? [])
                setRows(nextRows.rows ?? [])
            })
            .catch((cause: unknown) => {
                setError(cause instanceof Error ? cause.message : 'No se pudo cargar la facturación.')
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        fetchCostData(range, granularity)
    }, [granularity, range])

    useEffect(() => {
        getMeters().then((meters) => setMeterIds(meters.map((meter) => meter.meter_id))).catch(() => setMeterIds([]))
    }, [])

    const total = summary?.total_cost ?? 0
    const average = summary?.average_daily_cost ?? 0
    const change = summary?.change_percent ?? 0
    const hasData = (summary !== null && (summary.total_cost > 0 || summary.average_daily_cost !== undefined)) || series.length > 0 || rows.length > 0
    const costRows = useMemo(() => rows.slice().sort((left, right) => right.period_cost - left.period_cost), [rows])

    async function handleGenerateMeasurement() {
        if (meterIds.length === 0) {
            setError('No hay medidores disponibles para generar la medición.')
            return
        }

        const days = PERIODS.find((item) => item.id === range)?.days ?? 30
        const to = new Date()
        const from = new Date()
        from.setDate(to.getDate() - days)

        setGenerating(true)
        setError(null)

        try {
            await Promise.all(meterIds.map((meterId) => estimateMeterConsumptionCost(meterId, from.toISOString(), to.toISOString())))
            await fetchCostData(range, granularity)
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : 'No se pudo generar la medición.')
        } finally {
            setGenerating(false)
        }
    }

    async function handleDownload(format: 'pdf' | 'excel') {
        const days = PERIODS.find((item) => item.id === range)?.days ?? 30
        const to = new Date()
        const from = new Date()
        from.setDate(to.getDate() - days)
        setReporting(format)
        try {
            const blob = await downloadBillingReport(format, from.toISOString(), to.toISOString())
            const url = URL.createObjectURL(blob)
            const link = document.createElement('a')
            link.href = url
            link.download = `${format === 'pdf' ? 'billing-report' : 'billing-report'}.${format === 'pdf' ? 'pdf' : 'xlsx'}`
            document.body.appendChild(link)
            link.click()
            link.remove()
            URL.revokeObjectURL(url)
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : 'No se pudo descargar el reporte.')
        } finally {
            setReporting(null)
        }
    }

    return (
        <div className="costs-page">
            <div className="page-heading dashboard-heading">
                <div>
                    <div className="eyebrow">FACTURACIÓN Y COSTO</div>
                    <h1>Consumo y gastos</h1>
                </div>
                <div className="costs-actions">
                    <div className="chart-range-toggle" aria-label="Seleccionar periodo">
                        {PERIODS.map((item) => (
                            <button key={item.id} className={range === item.id ? 'selected' : ''} onClick={() => setRange(item.id)}>
                                {item.label}
                            </button>
                        ))}
                    </div>
                    <div className="chart-range-toggle" aria-label="Seleccionar agregación">
                        {(['daily', 'weekly', 'monthly'] as const).map((item) => (
                            <button key={item} className={granularity === item ? 'selected' : ''} onClick={() => setGranularity(item)}>
                                {item === 'daily' ? 'Día' : item === 'weekly' ? 'Semana' : 'Mes'}
                            </button>
                        ))}
                    </div>
                    <button className="button button-primary" onClick={handleGenerateMeasurement} disabled={generating || loading}>
                        <Zap size={15} />{generating ? 'Calculando…' : 'Generar medición'}
                    </button>
                    <button className="button button-secondary" onClick={() => handleDownload('pdf')} disabled={reporting !== null || loading || !hasData}>
                        <Download size={15} />{reporting === 'pdf' ? 'Generando PDF…' : 'PDF'}
                    </button>
                    <button className="button button-primary" onClick={() => handleDownload('excel')} disabled={reporting !== null || loading || !hasData}>
                        <FileText size={15} />{reporting === 'excel' ? 'Generando Excel…' : 'Excel'}
                    </button>
                </div>
            </div>

            {error && <div className="error-banner" role="alert">{error}</div>}

            <section className="kpi-grid costs-kpi-grid" aria-label="Indicadores de costo">
                <KpiCard label="Costo total" value={summary ? formatCurrency(total, summary.currency) : '—'} note={summary?.period_start && summary?.period_end ? 'Periodo activo' : 'Sin período'} icon={<Wallet size={17} />} tone="aqua" />
                <KpiCard label="Cambio" value={summary ? `${change >= 0 ? '+' : ''}${change.toFixed(1)}%` : '—'} note="vs. período anterior" icon={<TrendingUp size={17} />} tone="blue" />
                <KpiCard label="Promedio diario" value={summary ? formatCurrency(average, summary.currency) : '—'} note="Distribución por día" icon={<Zap size={17} />} tone="violet" />
                <KpiCard label="Top medidor" value={summary?.top_meter_id ?? '—'} note="Mayor contribución" icon={<Gauge size={17} />} tone="coral" />
            </section>

            <section className="context-help-list" aria-label="Guías de costos">
                <div className="context-help-item">
                    <span>Costo</span>
                    <InfoTooltip label="Costo" text="Monto estimado en la moneda del sitio para el período seleccionado. Permite comparar la factura sobre una base homogénea." />
                </div>
                <div className="context-help-item">
                    <span>Variación</span>
                    <InfoTooltip label="Variación" text="Compara el costo actual contra el período anterior para identificar incrementos de consumo o cambios de tarifa." />
                </div>
                <div className="context-help-item">
                    <span>Tarifa</span>
                    <InfoTooltip label="Tarifa" text="La tarifa activa del medidor se usa para estimar consumo y costo real según proveedor, región y tipo de tarifa." />
                </div>
            </section>

            <section className="dashboard-grid costs-grid">
                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="EVOLUCIÓN" title={`Costo por ${granularity === 'daily' ? 'día' : granularity === 'weekly' ? 'semana' : 'mes'}`} infoText="Muestra la tendencia del costo total agregado según el periodo elegido para detectar picos y cambios de tendencia en la facturación." />
                    </div>
                    {loading ? <div className="empty-state"><span><TrendingUp size={17} /></span><p>Cargando evolución de costos…</p></div> : series.length > 0 ? (
                        <div className="chart-wrap consumption-chart">
                            <ResponsiveContainer width="100%" height="100%">
                                <AreaChart data={series} margin={{ top: 10, right: 12, bottom: 0, left: -12 }}>
                                    <defs>
                                        <linearGradient id="costFill" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="0%" stopColor="#5b4ef6" stopOpacity={0.32} />
                                            <stop offset="100%" stopColor="#5b4ef6" stopOpacity={0.02} />
                                        </linearGradient>
                                    </defs>
                                    <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                                    <XAxis dataKey="label" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                                    <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                                    <Tooltip formatter={(value) => [formatCurrency(Number(value), summary?.currency ?? 'COP'), 'Costo']} labelFormatter={(label) => formatDateLabel(String(label))} />
                                    <Area type="monotone" dataKey="total_cost" stroke="#5b4ef6" fill="url(#costFill)" strokeWidth={2} dot={false} activeDot={{ r: 4, fill: '#5b4ef6' }} />
                                </AreaChart>
                            </ResponsiveContainer>
                        </div>
                    ) : <div className="empty-state"><span><Wallet size={17} /></span><p>No hay historial de costo para este período.</p></div>}
                </div>

                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="POR MEDIDOR" title="Costo total del período por medidor" infoText="Compara el gasto total acumulado por cada medidor en el período seleccionado para detectar los mayores contribuyentes y priorizar intervenciones." />
                    </div>
                    {loading ? <div className="empty-state"><span><Gauge size={17} /></span><p>Cargando costos por medidor…</p></div> : costRows.length > 0 ? (
                        <div className="chart-wrap consumption-chart">
                            <ResponsiveContainer width="100%" height="100%">
                                <BarChart data={costRows} margin={{ top: 12, right: 12, bottom: 0, left: -14 }}>
                                    <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                                    <XAxis dataKey="meter_id" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                                    <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                                    <Tooltip formatter={(value) => [formatCurrency(Number(value), summary?.currency ?? 'COP'), 'Costo total']} />
                                    <Bar dataKey="period_cost" fill="#08a98f" radius={[6, 6, 0, 0]} maxBarSize={26} />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    ) : <div className="empty-state"><span><Gauge size={17} /></span><p>No hay costos por medidor disponibles para el rango actual.</p></div>}
                </div>
            </section>

            <section className="panel cost-table-panel">
                <div className="panel-heading">
                    <div>
                        <div className="eyebrow">DESGLOSE</div>
                        <h2>Coste por medidor</h2>
                    </div>
                    <span className="subtle-label">{costRows.length} resultados</span>
                </div>
                <div className="table-scroll">
                    <table>
                        <thead>
                            <tr>
                                <th>Medidor</th>
                                <th>Consumo</th>
                                <th>Tarifa</th>
                                <th>Coste diario</th>
                                <th>Coste semanal</th>
                                <th>Coste del período</th>
                                <th>Variación</th>
                            </tr>
                        </thead>
                        <tbody>
                            {costRows.map((item) => (
                                <tr key={item.meter_id}>
                                    <td><strong>{item.meter_id}</strong></td>
                                    <td>{item.kwh.toFixed(1)} kWh</td>
                                    <td>{item.tariff_name || 'Tarifa vigente'}</td>
                                    <td>{formatCurrency(item.daily_cost, item.currency ?? summary?.currency ?? 'COP')}</td>
                                    <td>{formatCurrency(item.weekly_cost, item.currency ?? summary?.currency ?? 'COP')}</td>
                                    <td>{formatCurrency(item.period_cost, item.currency ?? summary?.currency ?? 'COP')}</td>
                                    <td>
                                        {typeof item.trend === 'number' && Number.isFinite(item.trend) ? (
                                            <span className={`variation ${item.trend >= 0 ? 'positive' : 'negative'}`}>
                                                {item.trend >= 0 ? <ArrowUpRight size={14} /> : <ArrowDownRight size={14} />}
                                                {Math.abs(item.trend).toFixed(1)}%
                                            </span>
                                        ) : <span className="muted">—</span>}
                                    </td>
                                </tr>
                            ))}
                            {costRows.length === 0 && (
                                <tr>
                                    <td colSpan={7}>
                                        <div className="empty-state table-empty-state"><span><Wallet size={17} /></span><p>No hay filas de costo para mostrar en este rango.</p></div>
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </section>
        </div>
    )
}
