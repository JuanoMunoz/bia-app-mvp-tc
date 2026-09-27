import { useState } from 'react'
import { ArrowDownRight, ArrowUpRight, Bolt, ChevronRight, Search } from 'lucide-react'
import { EmptyState, StatusBadge } from '../../components/ui'
import { ExportActions } from '../../components/ExportActions'
import type { MeterSummary } from '../../types'
import { anomalyTypeLabel, formatDate, severityOrder } from '../../utils/format'

export function MetersPage({ meters, periodEnd, onOpenMeter }: { meters: MeterSummary[]; periodEnd?: string; onOpenMeter: (meterId: string) => void }) {
    const [search, setSearch] = useState('')
    const [filter, setFilter] = useState('all')
    const [sortBy, setSortBy] = useState('severity')
    const filtered = meters.filter((item) => {
        const matchesSearch = item.meter_id.toLowerCase().includes(search.trim().toLowerCase())
        const matchesStatus = filter === 'all' || item.status === filter
        return matchesSearch && matchesStatus
    }).slice().sort((left, right) => {
        if (sortBy === 'consumption') return right.current_consumption_kwh - left.current_consumption_kwh
        if (sortBy === 'variation') return Math.abs(right.change_percent) - Math.abs(left.change_percent)
        return severityOrder(left.severity) - severityOrder(right.severity)
    })
    const exportRows = filtered.map((item) => ({
        Medidor: item.meter_id,
        'Consumo actual (kWh)': item.current_consumption_kwh.toFixed(1),
        'Referencia (kWh)': item.baseline_kwh.toFixed(1),
        'Variación (%)': item.change_percent.toFixed(1),
        Estado: item.status,
        Alerta: item.anomaly ? anomalyTypeLabel(item.anomaly.type) : 'Sin alertas',
    }))

    return (
        <>
            <div className="page-heading"><div><div className="eyebrow">MEDIDORES / {meters.length} REGISTRADOS</div><h1>Medidores</h1><p>Consumo, variación y estado operativo.</p></div><div className="heading-meta">{meters.length} conectados</div></div>
            <div className="toolbar panel">
                <label className="search-field"><Search size={17} /><input aria-label="Buscar medidor" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Buscar medidor…" /></label>
                <div className="segmented-control" aria-label="Filtrar medidores">
                    {[['all', 'Todos'], ['NORMAL', 'Normales'], ['ALERT', 'Alertas'], ['CRITICAL', 'Críticos']].map(([value, label]) => <button key={value} className={filter === value ? 'selected' : ''} onClick={() => setFilter(value)}>{label}</button>)}
                </div>
                <label className="sort-field">Ordenar por<select value={sortBy} onChange={(event) => setSortBy(event.target.value)}><option value="severity">Severidad</option><option value="consumption">Consumo</option><option value="variation">Variación</option></select></label>
            </div>
            <div className="table-panel panel">
                <div className="table-meta"><span>{filtered.length} medidores</span><span>{periodEnd ? `Últimos datos · ${formatDate(periodEnd)}` : 'Últimos datos disponibles'}</span><ExportActions filename="medidores" title="Medidores" columns={Object.keys(exportRows[0] ?? { Medidor: '', Estado: '' })} rows={exportRows} /></div>
                <div className="table-scroll"><table><thead><tr><th>Medidor</th><th>Consumo actual</th><th>Referencia</th><th>Variación</th><th>Estado</th><th>Alerta asociada</th><th /></tr></thead><tbody>
                    {filtered.map((item) => <tr key={item.meter_id}>
                        <td><button className="table-id" onClick={() => onOpenMeter(item.meter_id)}><span className="meter-glyph"><Bolt size={15} /></span><strong>{item.meter_id}</strong></button></td>
                        <td><strong>{item.current_consumption_kwh.toFixed(1)} kWh</strong></td>
                        <td>{item.baseline_kwh.toFixed(1)} kWh</td>
                        <td><span className={`variation ${item.change_percent >= 0 ? 'positive' : 'negative'}`}>{item.change_percent >= 0 ? <ArrowUpRight size={14} /> : <ArrowDownRight size={14} />}{Math.abs(item.change_percent).toFixed(1)}%</span></td>
                        <td><StatusBadge status={item.status} /></td>
                        <td>{item.anomaly ? <span className="anomaly-kind">{anomalyTypeLabel(item.anomaly.type)}</span> : <span className="muted">Sin alertas</span>}</td>
                        <td><button className="icon-button" onClick={() => onOpenMeter(item.meter_id)} aria-label={`Abrir medidor ${item.meter_id}`}><ChevronRight size={16} /></button></td>
                    </tr>)}
                    {filtered.length === 0 && <tr><td colSpan={7}><EmptyState text="No hay medidores que coincidan con estos filtros." /></td></tr>}
                </tbody></table></div>
            </div>
        </>
    )
}