import { useState } from 'react'
import { AlertTriangle, Bolt, ChevronRight, Search } from 'lucide-react'
import { EmptyState, SeverityBadge } from '../../components/ui'
import { ExportActions } from '../../components/ExportActions'
import type { Anomaly } from '../../types'
import { InvestigationPage } from './InvestigationPage'
import { anomalyTypeLabel } from '../../utils/format'

export function AnomaliesPage({ anomalies, selected, onSelect, onBack }: { anomalies: Anomaly[]; selected: Anomaly | null; onSelect: (anomaly: Anomaly) => void; onBack: () => void }) {
    const [type, setType] = useState('all')
    const [severity, setSeverity] = useState('all')
    const [search, setSearch] = useState('')
    if (selected) return <InvestigationPage anomaly={selected} onBack={onBack} />
    const filtered = anomalies.filter((item) => {
        const searchValue = search.trim().toLowerCase()
        return (type === 'all' || item.type === type) && (severity === 'all' || item.severity === severity) && (!searchValue || item.meter_id.toLowerCase().includes(searchValue) || item.reason.toLowerCase().includes(searchValue))
    })
    const attention = anomalies.filter((item) => item.severity === 'HIGH' || item.severity === 'CRITICAL').length
    const exportRows = filtered.map((item) => ({
        Medidor: item.meter_id,
        Tipo: anomalyTypeLabel(item.type),
        Severidad: item.severity,
        'Confianza (%)': Math.round(item.confidence * 100),
        'Variación (%)': item.evidence.change_percent.toFixed(1),
        'Acción recomendada': item.recommended_action,
    }))
    return (
        <>
            <div className="page-heading"><div><div className="eyebrow">ALERTAS / HALLAZGOS</div><h1>Alertas de consumo</h1><p>Variaciones detectadas, evidencia y acciones recomendadas.</p></div><div className="heading-meta"><AlertTriangle size={16} />{attention} de alta prioridad</div></div>
            <div className="anomaly-summary-row"><div className="summary-chip"><strong>{anomalies.length}</strong><span>Total alertas</span></div><div className="summary-chip"><strong className="text-coral">{attention}</strong><span>Alta prioridad</span></div><div className="summary-chip"><strong>{anomalies.filter((item) => item.type === 'FALSE_POSITIVE').length}</strong><span>Alertas descartadas</span></div>
                <label className="search-field anomaly-search"><Search size={15} /><input aria-label="Buscar alerta" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Buscar medidor o alerta" /></label>
                <label className="sort-field">Tipo<select value={type} onChange={(event) => setType(event.target.value)}><option value="all">Todos los tipos</option><option value="REAL_ANOMALY">Real</option><option value="EXPLAINABLE_ANOMALY">Explicable</option><option value="FALSE_POSITIVE">Falso positivo</option><option value="DATA_QUALITY">Calidad de datos</option></select></label>
                <label className="sort-field">Prioridad<select value={severity} onChange={(event) => setSeverity(event.target.value)}><option value="all">Todas</option><option value="CRITICAL">Crítica</option><option value="HIGH">Alta</option><option value="MEDIUM">Media</option><option value="LOW">Baja</option></select></label>
            </div>
            <div className="table-panel panel"><div className="table-meta"><span>{filtered.length} alertas</span><span>Última revisión · {anomalies[0] ? new Date(anomalies[0].detected_at).toLocaleDateString('es-CO') : '—'}</span><ExportActions filename="alertas-de-consumo" title="Alertas de consumo" columns={Object.keys(exportRows[0] ?? { Medidor: '', Tipo: '' })} rows={exportRows} /></div><div className="table-scroll"><table><thead><tr><th>Medidor</th><th>Tipo</th><th>Severidad</th><th>Confianza del análisis</th><th>Variación</th><th>Acción recomendada</th><th /></tr></thead><tbody>
                {filtered.map((item) => <tr key={item.id}>
                    <td><button className="table-id" onClick={() => onSelect(item)}><span className="meter-glyph"><Bolt size={15} /></span><strong>{item.meter_id}</strong></button></td>
                    <td><span className={`type-marker ${item.type.toLowerCase()}`} />{anomalyTypeLabel(item.type)}</td>
                    <td><SeverityBadge severity={item.severity} /></td>
                    <td><span className="confidence-cell"><span className="confidence-track"><i style={{ width: `${Math.round(item.confidence * 100)}%` }} /></span>{Math.round(item.confidence * 100)}%</span></td>
                    <td className={item.evidence.change_percent >= 0 ? 'text-coral' : 'text-blue'}>{item.evidence.change_percent >= 0 ? '+' : ''}{item.evidence.change_percent.toFixed(1)}%</td>
                    <td className="action-cell">{item.recommended_action}</td>
                    <td><button className="icon-button" onClick={() => onSelect(item)} aria-label={`Investigar ${item.meter_id}`}><ChevronRight size={16} /></button></td>
                </tr>)}
                {filtered.length === 0 && <tr><td colSpan={7}><EmptyState text="No hay hallazgos para este tipo." /></td></tr>}
            </tbody></table></div></div>
        </>
    )
}