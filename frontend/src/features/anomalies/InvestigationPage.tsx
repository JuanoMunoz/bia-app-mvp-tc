import { lazy, Suspense } from 'react'
import { ArrowLeft, CheckCircle2, CircleHelp, Sparkles } from 'lucide-react'
import { AiInsightCard } from '../../components/AiInsightCard'
import { buildFriendlyInsight } from '../../insight'
import { ChartLoading, EvidenceStat, SeverityBadge } from '../../components/ui'
import type { ComparisonPoint } from '../../charts'
import type { Anomaly } from '../../types'
import { anomalyTypeLabel } from '../../utils/format'

const ComparisonChart = lazy(() => import('../../charts').then((module) => ({ default: module.ComparisonChart })))

export function InvestigationPage({ anomaly, onBack }: { anomaly: Anomaly; onBack: () => void }) {
    const evidence = anomaly.evidence
    const comparison: ComparisonPoint[] = [
        { label: 'Referencia', value: Number(evidence.baseline_kwh.toFixed(1)) },
        { label: 'Consumo actual', value: Number(evidence.current_kwh.toFixed(1)) },
    ]
    const aiInsight = buildFriendlyInsight(
        `El medidor ${anomaly.meter_id} muestra una situación ${anomaly.severity.toLowerCase()} con una variación de ${evidence.change_percent.toFixed(1)}% frente a la referencia. La señal más clara es que el consumo actual es ${evidence.current_kwh.toFixed(1)} kWh contra una base de ${evidence.baseline_kwh.toFixed(1)} kWh, por lo que ${anomaly.severity === 'CRITICAL' || anomaly.severity === 'HIGH' ? 'requiere atención inmediata' : 'debe revisarse con prioridad'}.`,
        `La anomalía se evalúa comparando consumo, voltaje, corriente, factor de potencia y eventos relacionados. La evidencia disponible indica que la desviación no parece casual, pero la interpretación sigue siendo una señal de riesgo que debe confirmarse con la operación del sitio y con la historia del medidor.`,
        [
            {
                question: '¿Qué explica esta alarma?',
                answer: `La alarma se activa porque el consumo actual difiere de la referencia histórica y la evidencia disponible mostrada en ${anomaly.meter_id} sugiere una desviación significativa.`,
            },
            {
                question: '¿Cuánto riesgo hay?',
                answer: `La severidad es ${anomaly.severity.toLowerCase()} y la confianza del análisis es ${Math.round(anomaly.confidence * 100)}%, lo que indica que la señal es útil para actuar con prioridad pero aún requiere confirmación de operación.`,
            },
            {
                question: '¿Qué hago ahora?',
                answer: `Revisa la recomendación operativa, valida si hay un evento asociado y confirma si el medidor está teniendo una condición real o un ajuste temporal antes de cerrar la alerta.`,
            },
        ],
    )
    return (
        <>
            <button className="back-button" onClick={onBack}><ArrowLeft size={16} />Alertas de consumo</button>
            <div className="page-heading detail-heading"><div><div className="eyebrow">INVESTIGACIÓN / {anomaly.id.slice(0, 8).toUpperCase()}</div><h1>{anomaly.meter_id}</h1><p>Investigación de anomalía · {new Date(anomaly.detected_at).toLocaleString('es-CO')}</p></div><SeverityBadge severity={anomaly.severity} /></div>
            <section className="investigation-layout">
                <div className="investigation-main">
                    <div className="finding-banner"><div className={`finding-symbol ${anomaly.type.toLowerCase()}`}><Sparkles size={18} /></div><div><div className="eyebrow">QUÉ ENCONTRAMOS</div><h2>{anomalyTypeLabel(anomaly.type)}</h2><p>{anomaly.reason}</p></div><span className="confidence-score">{Math.round(anomaly.confidence * 100)}<small>%</small></span></div>
                    <section className="panel evidence-panel"><div className="panel-heading"><div><div className="eyebrow">COMPARACIÓN CONTRA REFERENCIA</div><h2>Evidencia de consumo</h2></div><span className={`evidence-delta ${evidence.change_percent >= 0 ? 'positive' : 'negative'}`}>{evidence.change_percent >= 0 ? '+' : ''}{evidence.change_percent.toFixed(1)}%</span></div>
                        <div className="chart-wrap evidence-chart"><Suspense fallback={<ChartLoading />}><ComparisonChart data={comparison} /></Suspense></div>
                        <div className="evidence-stats"><EvidenceStat label="Voltaje medio" value={`${evidence.voltage_v.toFixed(1)} V`} /><EvidenceStat label="Corriente media" value={`${evidence.current_a.toFixed(1)} A`} /><EvidenceStat label="Factor de potencia" value={evidence.power_factor.toFixed(3)} /><EvidenceStat label="Variación de corriente" value={`${evidence.current_variation_percent.toFixed(1)}%`} /><EvidenceStat label="Horas atípicas" value={`${evidence.abnormal_hours.length} / 24`} /><EvidenceStat label="Lecturas fuera de rango" value={String(evidence.outlier_count)} /></div>
                    </section>
                    {anomaly.type === 'DATA_QUALITY' && <div className="quality-note"><CircleHelp size={17} /><span>El consumo y la referencia son similares; algunas variables eléctricas presentan lecturas fuera de rango.</span></div>}
                    <section className="recommendation"><div className="recommendation-icon"><CheckCircle2 size={18} /></div><div><div className="eyebrow">ACCIÓN RECOMENDADA</div><p>{anomaly.recommended_action}</p></div></section>
                    <AiInsightCard insight={aiInsight} title={`Análisis de ${anomaly.meter_id}`} />
                </div>
                <aside className="investigation-aside">
                    <div className="panel investigation-meta"><div className="eyebrow">DETALLES DE LA ALERTA</div><div className="meta-row"><span>Tipo</span><strong>{anomalyTypeLabel(anomaly.type)}</strong></div><div className="meta-row"><span>Severidad</span><SeverityBadge severity={anomaly.severity} /></div><div className="meta-row"><span>Confianza del análisis</span><strong>{Math.round(anomaly.confidence * 100)}%</strong></div></div>
                    <div className="panel related-panel"><div className="eyebrow">EVENTOS RELACIONADOS</div>{evidence.related_events.length ? evidence.related_events.map((event) => <div className="event-item" key={event}><span className="event-dot" /><span>{event}</span></div>) : <p className="muted event-empty">No se encontraron eventos asociados.</p>}</div>
                    <div className="panel investigation-meter"><div className="eyebrow">MEDIDOR</div><strong>{anomaly.meter_id}</strong><span>{new Date(anomaly.detected_at).toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric' })}</span></div>
                </aside>
            </section>
        </>
    )
}