import { lazy, Suspense, useEffect, useMemo, useState } from 'react'
import { Activity, ArrowLeft, Gauge, TrendingUp, Zap } from 'lucide-react'
import { AiInsightCard } from '../../components/AiInsightCard'
import { buildFriendlyInsight } from '../../insight'
import { ChartLoading, InfoTooltip, MetricPanel, PanelTitle, StatusBadge } from '../../components/ui'
import { buildHourWeekHeatmap, buildLoadDurationCurve } from '../../charts-data'
import { HeatmapChart, LoadDurationCurveChart } from '../../charts'
import { getCostByMeter, updateMeterConfig } from '../../api'
import type { HistoryPoint } from '../../charts'
import type { CostByMeter, MeterDetail } from '../../types'
import { formatDateRange } from '../../utils/format'

const MeterHistoryChart = lazy(() => import('../../charts').then((module) => ({ default: module.MeterHistoryChart })))
type ChartRange = 'hour' | 'day' | 'week' | 'month' | 'quarter'

const chartRanges: Array<{ value: ChartRange; label: string }> = [
    { value: 'hour', label: 'Hora' },
    { value: 'day', label: 'Día' },
    { value: 'week', label: 'Semana' },
    { value: 'month', label: 'Mes' },
    { value: 'quarter', label: 'Trimestre' },
]

function buildChartSeries(history: MeterDetail['history'], range: ChartRange): HistoryPoint[] {
    if (history.length === 0) return []

    const sorted = [...history].sort((left, right) => new Date(left.timestamp).getTime() - new Date(right.timestamp).getTime())
    const windowStart = rangeWindowStart(new Date(sorted[sorted.length - 1].timestamp), range)
    const visible = sorted.filter((item) => new Date(item.timestamp).getTime() >= windowStart.getTime())
    if (visible.length === 0) return []

    const grains: ChartRange[] = ['hour', 'day', 'week', 'month', 'quarter']
    let grain = range
    let buckets = groupByGrain(visible, grain)
    while (buckets.size < 3 && grain !== 'hour') {
        grain = grains[grains.indexOf(grain) - 1]
        buckets = groupByGrain(visible, grain)
    }

    return Array.from(buckets.values())
        .sort((left, right) => left.time - right.time)
        .map((bucket) => ({
            time: new Date(bucket.time).toISOString(),
            label: bucket.label,
            consumption: Number((bucket.total / bucket.count).toFixed(2)),
        }))
}

function rangeWindowStart(reference: Date, range: ChartRange): Date {
    const start = new Date(reference)
    switch (range) {
        case 'hour':
            start.setHours(start.getHours() - 24)
            break
        case 'day':
            start.setDate(start.getDate() - 7)
            break
        case 'week':
            start.setDate(start.getDate() - 30)
            break
        case 'month':
            start.setMonth(start.getMonth() - 6)
            break
        case 'quarter':
            start.setFullYear(start.getFullYear() - 1)
            break
    }
    return start
}

interface ConsumptionBucket {
    time: number
    label: string
    total: number
    count: number
}

function groupByGrain(history: MeterDetail['history'], grain: ChartRange): Map<string, ConsumptionBucket> {
    const buckets = new Map<string, ConsumptionBucket>()
    for (const item of history) {
        const pointDate = new Date(item.timestamp)
        const bucketKey = bucketKeyFor(pointDate, grain)
        const current = buckets.get(bucketKey)
        if (current) {
            current.total += item.consumption_kwh
            current.count += 1
            continue
        }
        buckets.set(bucketKey, {
            time: pointDate.getTime(),
            label: formatBucketLabel(pointDate, grain),
            total: item.consumption_kwh,
            count: 1,
        })
    }
    return buckets
}

function weekMonday(date: Date): Date {
    const monday = new Date(date.getFullYear(), date.getMonth(), date.getDate())
    monday.setDate(monday.getDate() - ((monday.getDay() + 6) % 7))
    return monday
}

function bucketKeyFor(date: Date, grain: ChartRange): string {
    const day = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    switch (grain) {
        case 'hour':
            return `${day} ${String(date.getHours()).padStart(2, '0')}`
        case 'day':
            return day
        case 'week': {
            const monday = weekMonday(date)
            return `w${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`
        }
        case 'month':
            return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
        case 'quarter':
            return `${date.getFullYear()}-Q${Math.floor(date.getMonth() / 3) + 1}`
    }
}

function formatBucketLabel(date: Date, grain: ChartRange): string {
    switch (grain) {
        case 'hour':
            return new Intl.DateTimeFormat('es-CO', { day: '2-digit', month: 'short', hour: '2-digit' }).format(date)
        case 'day':
            return new Intl.DateTimeFormat('es-CO', { day: '2-digit', month: 'short' }).format(date)
        case 'week':
            return `Sem ${new Intl.DateTimeFormat('es-CO', { day: '2-digit', month: 'short' }).format(weekMonday(date))}`
        case 'month':
            return new Intl.DateTimeFormat('es-CO', { month: 'short', year: '2-digit' }).format(date)
        case 'quarter':
            return `Q${Math.floor(date.getMonth() / 3) + 1} ${String(date.getFullYear()).slice(2)}`
    }
}

export function MeterDetailPage({ meter, onBack }: { meter: MeterDetail; onBack: () => void }) {
    const [range, setRange] = useState<ChartRange>('day')
    const [showMonetaryPanel, setShowMonetaryPanel] = useState(false)
    const [meterCosts, setMeterCosts] = useState<CostByMeter[]>([])
    const [provider, setProvider] = useState(meter.provider ?? '')
    const [region, setRegion] = useState(meter.region ?? '')
    const [rateType, setRateType] = useState(meter.rate_type ?? 'industrial')
    const [savingConfig, setSavingConfig] = useState(false)
    const [configMessage, setConfigMessage] = useState<{ tone: 'ok' | 'error'; text: string } | null>(null)
    const chartData = useMemo(() => buildChartSeries(meter.history, range), [meter.history, range])
    const heatmapData = useMemo(() => buildHourWeekHeatmap(meter.history), [meter.history])
    const loadDurationData = useMemo(() => buildLoadDurationCurve(meter.history).map((item) => item.demand), [meter.history])
    const costPeriod = useMemo(() => {
        const moments = meter.history
            .map((item) => new Date(item.timestamp).getTime())
            .filter((moment) => Number.isFinite(moment))
        if (moments.length === 0) {
            const to = new Date()
            const from = new Date()
            from.setDate(to.getDate() - 30)
            return { from: from.toISOString(), to: to.toISOString() }
        }
        return { from: new Date(Math.min(...moments)).toISOString(), to: new Date(Math.max(...moments)).toISOString() }
    }, [meter.history])

    useEffect(() => {
        let active = true
        getCostByMeter(costPeriod.from, costPeriod.to).then((rows) => {
            if (!active) return
            setMeterCosts(rows.rows ?? [])
        }).catch(() => {
            if (!active) return
            setMeterCosts([])
        })

        return () => {
            active = false
        }
    }, [costPeriod, meter.meter_id])

    async function handleSaveConfig() {
        setSavingConfig(true)
        setConfigMessage(null)
        try {
            const updated = await updateMeterConfig(meter.meter_id, {
                provider: provider.trim(),
                region: region.trim(),
                rate_type: rateType.trim() || 'industrial',
            })
            setProvider(updated.provider ?? '')
            setRegion(updated.region ?? '')
            setRateType(updated.rate_type ?? 'industrial')
            setConfigMessage({ tone: 'ok', text: 'Configuración guardada. Aplica a las próximas estimaciones de costo.' })
        } catch (cause) {
            setConfigMessage({ tone: 'error', text: cause instanceof Error ? cause.message : 'No se pudo guardar la configuración.' })
        } finally {
            setSavingConfig(false)
        }
    }

    const currentMeterCost = useMemo(
        () => meterCosts.find((item) => item.meter_id === meter.meter_id) ?? null,
        [meterCosts, meter.meter_id],
    )

    const meterAiInsight = buildFriendlyInsight(
        `${meter.meter_id} está ${meter.change_percent >= 0 ? 'por encima' : 'por debajo'} de su referencia habitual. En la última jornada el consumo se ubicó en ${meter.current_consumption_kwh.toFixed(1)} kWh frente a una base de ${meter.baseline_kwh.toFixed(1)} kWh, por lo que ${meter.change_percent >= 0 ? 'hay un aumento relevante que conviene revisar' : 'hay una reducción útil para validar si es operativa o temporal'}.`,
        `La comparación se hace entre el consumo promedio reciente y la referencia histórica del mismo medidor. Un cambio del ${meter.change_percent.toFixed(1)}% ayuda a priorizar si requiere revisión manual, pero la decisión final debe combinarse con voltaje, corriente y eventos operativos del sitio.`,
        [
            {
                question: '¿Qué está pasando con este medidor?',
                answer: `Tiene un consumo ${meter.change_percent >= 0 ? 'superior' : 'inferior'} a la referencia habitual. El indicador más importante es la variación de ${Math.abs(meter.change_percent).toFixed(1)}%, y el consumo actual es ${meter.current_consumption_kwh.toFixed(1)} kWh.`,
            },
            {
                question: '¿Debe preocuparme?',
                answer: `Depende del contexto operativo. Si el cambio es sostenido y coincide con esfuerzo extra o caídas de factor de potencia, sí hay que revisarlo; si es puntual, conviene validar primero la lectura y los eventos.`,
            },
            {
                question: '¿Qué debería revisar ahora?',
                answer: `Revise el historial del medidor, el factor de potencia y cualquier cambio reciente en producción o mantenimiento. En general, la recomendación es validar la causa antes de ajustar la línea base o escalar la alarma.`,
            },
        ],
    )

    return (
        <>
            <button className="back-button" onClick={onBack}><ArrowLeft size={16} />Medidores</button>
            <div className="page-heading detail-heading"><div><div className="eyebrow">MEDIDORES / DETALLE</div><h1>{meter.meter_id}</h1><p>Histórico de consumo y variables eléctricas.</p></div><div className="detail-heading-actions"><StatusBadge status={meter.status} /><button className="button button-secondary" onClick={() => setShowMonetaryPanel((current) => !current)}><TrendingUp size={15} />{showMonetaryPanel ? 'Ocultar análisis' : 'Ver análisis monetario'}</button></div></div>
            <section className="detail-metrics">
                <div className="metric-block"><span>Consumo actual <InfoTooltip label="Consumo actual" text="Promedio de consumo durante las últimas 24 horas, útil para ver la demanda reciente del medidor." /></span><strong>{meter.current_consumption_kwh.toFixed(1)} <small>kWh</small></strong><em>Promedio últimas 24 h</em></div>
                <div className="metric-block"><span>Referencia <InfoTooltip label="Referencia" text="Es el nivel habitual del medidor según el historial previo. Sirve como línea base para detectar cambios inusuales." /></span><strong>{meter.baseline_kwh.toFixed(1)} <small>kWh</small></strong><em>Primeros 7 días</em></div>
                <div className="metric-block"><span>Variación <InfoTooltip label="Variación" text="Mide cuánto cambia el consumo actual con respecto a la referencia. Cuanto mayor es, más relevante puede ser la variación." /></span><strong className={meter.change_percent >= 0 ? 'metric-up' : 'metric-down'}>{meter.change_percent >= 0 ? '+' : ''}{meter.change_percent.toFixed(1)}%</strong><em>Vs. referencia</em></div>
                <div className="metric-block"><span>Factor de potencia <InfoTooltip label="Factor de potencia" text="Mide la eficiencia energética del sistema. Valores bajos suelen indicar pérdidas o condiciones eléctricas problemáticas." /></span><strong>{meter.power_factor.toFixed(3)}</strong><em>Últimas 24 h</em></div>
            </section>
            {showMonetaryPanel && (
                <section className="panel monetary-panel">
                    <div className="panel-heading">
                        <div>
                            <div className="eyebrow">ANÁLISIS MONETARIO</div>
                            <h2>Coste estimado del medidor</h2>
                        </div>
                    </div>
                    {currentMeterCost ? (
                        <div className="monetary-grid">
                            <div className="metric-block monetary-metric"><span>Coste diario</span><strong>{new Intl.NumberFormat('es-CO', { style: 'currency', currency: currentMeterCost.currency ?? 'COP', maximumFractionDigits: 0 }).format(currentMeterCost.daily_cost)}</strong></div>
                            <div className="metric-block monetary-metric"><span>Coste semanal</span><strong>{new Intl.NumberFormat('es-CO', { style: 'currency', currency: currentMeterCost.currency ?? 'COP', maximumFractionDigits: 0 }).format(currentMeterCost.weekly_cost)}</strong></div>
                            <div className="metric-block monetary-metric"><span>Coste del período</span><strong>{new Intl.NumberFormat('es-CO', { style: 'currency', currency: currentMeterCost.currency ?? 'COP', maximumFractionDigits: 0 }).format(currentMeterCost.period_cost)}</strong></div>
                            <div className="metric-block monetary-metric"><span>Consumo</span><strong>{currentMeterCost.kwh.toFixed(1)} <small>kWh</small></strong></div>
                        </div>
                    ) : (
                        <p className="chart-help-text">Todavía no hay estimación de coste para este medidor en el rango activo. Cuando exista tarifa asociada, esta sección mostrará el impacto económico real.</p>
                    )}
                </section>
            )}
            <AiInsightCard insight={meterAiInsight} title={`Diagnóstico de ${meter.meter_id}`} />
            <section className="panel config-panel">
                <div className="panel-heading">
                    <div>
                        <div className="eyebrow">CONFIGURACIÓN</div>
                        <h2>Tarifa del medidor</h2>
                    </div>
                </div>
                <p className="chart-help-text">Proveedor, región y tipo de tarifa vigentes. Se usan para resolver la tarifa activa en las estimaciones de costo.</p>
                <div className="config-grid">
                    <label className="config-field"><span>Proveedor</span><input value={provider} onChange={(event) => setProvider(event.target.value)} placeholder="EPM" autoComplete="off" /></label>
                    <label className="config-field"><span>Región</span><input value={region} onChange={(event) => setRegion(event.target.value)} placeholder="Medellín" autoComplete="off" /></label>
                    <label className="config-field"><span>Tipo de tarifa</span><input value={rateType} onChange={(event) => setRateType(event.target.value)} placeholder="industrial" autoComplete="off" /></label>
                </div>
                <div className="config-actions">
                    <button className="button button-primary" onClick={handleSaveConfig} disabled={savingConfig}>{savingConfig ? 'Guardando…' : 'Guardar configuración'}</button>
                    {configMessage && <span className={`config-message ${configMessage.tone}`} role={configMessage.tone === 'error' ? 'alert' : 'status'}>{configMessage.text}</span>}
                </div>
            </section>
            <section className="panel history-panel">
                <div className="panel-heading"><div><div className="eyebrow">SERIE TEMPORAL</div><h2>Consumo por {chartRanges.find((item) => item.value === range)?.label.toLowerCase() ?? 'día'}</h2></div><div className="chart-range-tools"><div className="chart-range-toggle">{chartRanges.map((item) => <button key={item.value} className={range === item.value ? 'selected' : ''} onClick={() => setRange(item.value)}>{item.label}</button>)}</div><span className="chart-range">{meter.history.length > 0 ? formatDateRange(chartData[0]?.time ?? meter.history[0].timestamp, chartData[chartData.length - 1]?.time ?? meter.history[meter.history.length - 1].timestamp) : 'Sin datos'}</span></div></div>
                <div className="chart-wrap history-chart"><Suspense fallback={<ChartLoading />}><MeterHistoryChart data={chartData} baseline={meter.baseline_kwh} /></Suspense></div>
                <div className="chart-caption"><span><i className="legend-dot aqua-dot" />Consumo</span><span><i className="legend-line" />Referencia {meter.baseline_kwh.toFixed(1)} kWh</span></div>
            </section>
            <section className="analytics-grid">
                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="PATRÓN OPERATIVO" title="Consumo por hora y día de la semana" infoText="Identifica qué horas y días concentran más consumo, algo clave para detectar cargas elevadas fuera de los patrones normales de operación." />
                    </div>
                    <p className="chart-help-text">Este mapa identifica qué horas del día y qué días de la semana concentran más consumo. Es muy útil para detectar consumo elevado durante la madrugada o el fin de semana, cuando la operación normalmente debería estar en mínimos.</p>
                    <HeatmapChart data={heatmapData} labels={['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom']} />
                </div>
                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="DEMANDA Y PICO" title="Curva de duración de carga" infoText="La curva acumula la demanda por nivel de intensidad para mostrar cuántas horas opera cada medidor en cada rango de carga." />
                    </div>
                    <p className="chart-help-text">La curva ordena la demanda de mayor a menor. Así se ve con claridad cuándo el medidor está operando en picos innecesarios y ayuda a dimensionar mejor la potencia contratada.</p>
                    <div className="chart-wrap load-duration-chart"><LoadDurationCurveChart data={loadDurationData} /></div>
                </div>
            </section>
            <section className="electrical-row">
                <MetricPanel icon={<Activity size={17} />} label="Voltaje" value={`${meter.voltage_v.toFixed(1)} V`} note="Media últimas 24 h" />
                <MetricPanel icon={<Zap size={17} />} label="Corriente" value={`${meter.current_a.toFixed(1)} A`} note="Media últimas 24 h" />
                <MetricPanel icon={<Gauge size={17} />} label="Factor de potencia" value={meter.power_factor.toFixed(3)} note="Media últimas 24 h" />
            </section>
            {meter.anomaly && <div className="detail-anomaly panel"><div className="anomaly-callout-icon"><Activity size={18} /></div><div><div className="eyebrow">HALLAZGO ASOCIADO</div><strong>{meter.anomaly.type.replaceAll('_', ' ')} · {meter.anomaly.severity}</strong><p>{meter.anomaly.reason}</p></div></div>}
        </>
    )
}