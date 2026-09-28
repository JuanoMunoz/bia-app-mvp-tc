import { lazy, Suspense } from 'react'
import {
    Activity,
    AlertTriangle,
    ArrowDownRight,
    ArrowUpRight,
    Bolt,
    CheckCircle2,
    ChevronRight,
    Gauge,
    Sparkles,
} from 'lucide-react'
import { HeatmapChart, LoadDurationCurveChart } from '../../charts'
import { buildHourWeekHeatmap } from '../../charts-data'
import type { Analysis, Anomaly, DashboardSummary, MeterSummary, SitePoint } from '../../types'
import { ChartLoading, InfoTooltip, KpiCard, PanelTitle } from '../../components/ui'
import { InsightCard } from '../../components/InsightCard'
import { anomalyTypeLabel, compactNumber, formatDate, formatDateRange, revisionTimestamp, timeAgo } from '../../utils/format'

const ConsumptionChart = lazy(() => import('../../charts').then((module) => ({ default: module.ConsumptionChart })))

export function DashboardPage({
    summary,
    meters,
    siteHistory,
    anomalies,
    analysis,
    analyzing,
    onAnalyze,
    onOpenMeter,
    onOpenAnomaly,
}: {
    summary: DashboardSummary | null
    meters: MeterSummary[]
    siteHistory: SitePoint[]
    anomalies: Anomaly[]
    analysis: Analysis | null
    analyzing: boolean
    onAnalyze: () => void
    onOpenMeter: (meterId: string) => void
    onOpenAnomaly: (anomaly: Anomaly) => void
}) {
    const highPriority = anomalies.filter((item) => item.severity === 'HIGH' || item.severity === 'CRITICAL')
    const chartData = meters
        .slice()
        .sort((left, right) => right.current_consumption_kwh - left.current_consumption_kwh)
        .slice(0, 6)
        .map((item) => ({ meter: item.meter_id, consumption: Number(item.current_consumption_kwh.toFixed(1)) }))
    const heatmapData = buildHourWeekHeatmap(siteHistory)
    const loadCurveData = siteHistory.map((item) => item.consumption_kwh)
    const confidence = summary ? Math.round(summary.aggregate_confidence * 100) : 0
    // La revisión más reciente entre el análisis recién ejecutado y el del resumen.
    // Evita fechas inválidas/cero (mostraban "hace NNN días") y prioriza el resultado
    // fresco aunque el resumen recargado venga desactualizado.
    let lastRevisionDate: string | null = null
    let lastRevisionStatus: string | null = null
    for (const item of [analysis, summary?.latest_analysis ?? null]) {
        const stamp = revisionTimestamp(item)
        if (stamp && (!lastRevisionDate || new Date(stamp).getTime() > new Date(lastRevisionDate).getTime())) {
            lastRevisionDate = stamp
            lastRevisionStatus = item?.status ?? null
        }
    }

    return (
        <>
            <div className="page-heading dashboard-heading">
                <div><div className="eyebrow">{summary ? formatDate(summary.period_end, { weekday: 'long' }).toUpperCase() : 'PERÍODO DE CONSUMO'}</div></div>
                <button className="button button-primary" onClick={onAnalyze} disabled={analyzing}>
                    <Sparkles size={17} />{analyzing ? 'Revisando consumo…' : 'Revisar consumo'}
                </button>
            </div>

            {analyzing && <PipelineProgress />}
            {analysis?.status === 'COMPLETED' && !analyzing && (
                <div className="analysis-result"><CheckCircle2 size={17} /><span><strong>Análisis completado</strong> · {analysis.anomalies.length} anomalías detectadas · {analysis.anomalies.filter((item) => item.severity === 'HIGH' || item.severity === 'CRITICAL').length} requieren atención prioritaria</span></div>
            )}
            {analysis?.insight && !analyzing && <InsightCard insight={analysis.insight} />}

            <section className="kpi-grid" aria-label="Indicadores principales">
                <KpiCard label="Medidores totales" value={summary?.meter_count ?? '—'} note="En operación" icon={<Gauge size={17} />} tone="aqua" />
                <KpiCard label="Consumo del período" value={summary ? `${compactNumber(summary.total_consumption_kwh)} kWh` : '—'} note={summary ? formatDateRange(summary.period_start, summary.period_end) : 'Sin período'} icon={<Bolt size={17} />} tone="blue" />
                <KpiCard label="Alertas detectadas" value={summary?.anomaly_count ?? '—'} note="Última revisión" icon={<Activity size={17} />} tone="violet" />
                <KpiCard label="Alta prioridad" value={summary?.high_priority_count ?? '—'} note="Requieren atención" icon={<AlertTriangle size={17} />} tone="coral" />
                <KpiCard label="Confianza del análisis" value={summary ? `${confidence}%` : '—'} note="Promedio de hallazgos" icon={<Sparkles size={17} />} tone="green" />
                <KpiCard label="Última revisión" value={lastRevisionDate ? timeAgo(lastRevisionDate) : '—'} note={lastRevisionStatus === 'COMPLETED' ? 'Completada' : lastRevisionStatus === 'RUNNING' ? 'En curso' : lastRevisionStatus === 'FAILED' ? 'Fallida' : 'Sin revisión'} icon={<CheckCircle2 size={17} />} tone="neutral" />
            </section>

            <section className="context-help-list" aria-label="Glosario de ayuda contextual">
                <div className="context-help-item">
                    <span>Base histórica</span>
                    <InfoTooltip label="Base histórica" text="Es el consumo habitual del medidor en un periodo previo, usado como referencia para comparar si el consumo actual es inusual." />
                </div>
                <div className="context-help-item">
                    <span>Confianza</span>
                    <InfoTooltip label="Confianza" text="Indica qué tan segura es la detección. Un valor alto significa que la anomalía coincide con el patrón y la evidencia disponible." />
                </div>
                <div className="context-help-item">
                    <span>Factor de potencia</span>
                    <InfoTooltip label="Factor de potencia" text="Mide la eficiencia con la que la energía eléctrica se convierte en trabajo útil. Un valor bajo puede señalar pérdidas o condiciones anómalas." />
                </div>
            </section>

            <section className="dashboard-grid">
                <div className="panel consumption-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="DEMANDA POR MEDIDOR" title="Consumo actual" infoText="Promedio de consumo reciente por medidor para comparar la demanda actual con el patrón habitual del sitio." />
                        <button className="text-button" onClick={() => onOpenMeter(meters[0]?.meter_id ?? '')}>Ver medidores <ChevronRight size={15} /></button>
                    </div>
                    {chartData.length > 0 ? (
                        <div className="chart-wrap consumption-chart"><Suspense fallback={<ChartLoading />}><ConsumptionChart data={chartData} /></Suspense></div>
                    ) : <div className="empty-state"><span><Activity size={17} /></span><p>Carga los datos y ejecuta un análisis para ver el consumo.</p></div>}
                    <div className="chart-caption"><span><i className="legend-dot aqua-dot" /> Promedio últimas 24 h</span><span>kWh</span></div>
                </div>

                <div className="panel attention-panel">
                    <div className="panel-heading"><div><div className="eyebrow">ALERTAS PRIORITARIAS</div><h2>Atención requerida</h2></div><span className="count-mark">{highPriority.length.toString().padStart(2, '0')}</span></div>
                    {highPriority.length > 0 ? <div className="priority-list">
                        {highPriority.slice(0, 4).map((item) => (
                            <button className="priority-row" key={item.id} onClick={() => onOpenAnomaly(item)}>
                                <span className={`priority-icon ${item.severity.toLowerCase()}`}><AlertTriangle size={16} /></span>
                                <span className="priority-copy"><strong>{item.meter_id}<span className="priority-delta">{item.evidence.change_percent > 0 ? '+' : ''}{item.evidence.change_percent.toFixed(1)}%</span></strong><small>{item.reason}</small></span>
                                <ChevronRight size={16} className="muted-icon" />
                            </button>
                        ))}
                    </div> : <div className="empty-state"><span><Activity size={17} /></span><p>No hay alertas prioritarias.</p></div>}
                    <button className="panel-footer-link" onClick={() => onOpenAnomaly(highPriority[0] ?? anomalies[0])} disabled={anomalies.length === 0}>Ver todas las alertas <ChevronRight size={15} /></button>
                </div>
            </section>

            <section className="analytics-grid">
                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="PATRÓN OPERATIVO" title="Consumo por hora y día de la semana" infoText="Este mapa muestra en qué horas y días el consumo suele ser más alto, ayudando a detectar desperdicios o cargas fuera del patrón operativo esperado." />
                    </div>
                    <p className="chart-help-text">Muestra en qué horas y días el consumo suele ser más alto. Es útil para detectar desperdicio cuando la planta debería estar apagada o en condiciones de menor carga.</p>
                    {siteHistory.length > 0 ? <HeatmapChart data={heatmapData} labels={['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom']} /> : <div className="empty-state"><span><Activity size={17} /></span><p>No hay datos suficientes para el mapa de consumo.</p></div>}
                </div>

                <div className="panel analytics-panel">
                    <div className="panel-heading">
                        <PanelTitle eyebrow="DEMANDA Y PICO" title="Curva de duración de carga" infoText="Ordena la demanda de mayor a menor para identificar cuánto tiempo se mantiene cada nivel y detectar picos innecesarios o sobrecargas recurrentes." />
                    </div>
                    <p className="chart-help-text">Ordena la demanda de mayor a menor para ver cuánto tiempo se mantiene en cada nivel y detectar picos innecesarios o sobredimensionamiento.</p>
                    {loadCurveData.length > 0 ? <div className="chart-wrap load-duration-chart"><LoadDurationCurveChart data={loadCurveData} /></div> : <div className="empty-state"><span><Activity size={17} /></span><p>No hay demanda suficiente para la curva.</p></div>}
                </div>
            </section>

            <section className="meter-strip panel">
                <div className="panel-heading"><div><div className="eyebrow">ESTADO DE MEDIDORES</div><h2>Medidores destacados</h2></div><span className="subtle-label">Ordenados por variación</span></div>
                <div className="strip-list">
                    {meters.slice().sort((left, right) => Math.abs(right.change_percent) - Math.abs(left.change_percent)).slice(0, 5).map((item) => (
                        <button className="strip-item" key={item.meter_id} onClick={() => onOpenMeter(item.meter_id)}>
                            <span className={`strip-status ${item.status.toLowerCase()}`} />
                            <span className="strip-name"><strong>{item.meter_id}</strong><small>{item.status === 'NORMAL' ? 'Normal' : item.anomaly ? anomalyTypeLabel(item.anomaly.type) : 'Revisar'}</small></span>
                            <span className="strip-value">{item.current_consumption_kwh.toFixed(1)}<small> kWh</small></span>
                            <span className={`variation ${item.change_percent >= 0 ? 'positive' : 'negative'}`}>{item.change_percent >= 0 ? <ArrowUpRight size={14} /> : <ArrowDownRight size={14} />}{Math.abs(item.change_percent).toFixed(1)}%</span>
                        </button>
                    ))}
                    {meters.length === 0 && <div className="empty-state"><span><Activity size={17} /></span><p>No hay medidores disponibles.</p></div>}
                </div>
            </section>
        </>
    )
}

function PipelineProgress() {
    const steps = ['Lecturas', 'Referencia', 'Detección', 'Comparación', 'Eventos', 'Explicación', 'Recomendación']
    return <div className="pipeline-progress"><div className="pipeline-top"><span className="analysis-spinner" /><strong>Revisión en ejecución</strong><span>Proceso de 7 pasos</span></div><div className="pipeline-track">{steps.map((step, index) => <div className="pipeline-step" key={step}><span className={index === 0 ? 'pipeline-node active' : 'pipeline-node'}>{index === 0 ? <span /> : index + 1}</span><span>{step}</span></div>)}</div></div>
}