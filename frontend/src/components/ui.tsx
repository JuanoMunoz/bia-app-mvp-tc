import { useId, type ReactNode } from 'react'
import { Activity, Info } from 'lucide-react'

export function InfoTooltip({ label, text }: { label: string; text: string }) {
    const tooltipId = useId()

    return (
        <span className="info-tooltip">
            <span
                className="info-trigger"
                aria-label={`Más información sobre ${label}`}
                aria-describedby={tooltipId}
                tabIndex={0}
                title={text}
            >
                <Info size={12} />
            </span>
            <span id={tooltipId} className="info-tooltip-text" role="tooltip">
                {text}
            </span>
        </span>
    )
}

export function KpiCard({ label, value, note, icon, tone }: { label: string; value: string | number; note: string; icon: ReactNode; tone: string }) {
    return <div className="kpi-card"><div className={`kpi-icon ${tone}`}>{icon}</div><span className="kpi-label">{label}</span><strong className="kpi-value">{value}</strong><span className="kpi-note">{note}</span></div>
}

export function PanelTitle({ eyebrow, title, infoText }: { eyebrow: string; title: string; infoText?: string }) {
    return (
        <div className="panel-heading-copy">
            <div className="eyebrow">{eyebrow}</div>
            <h2>
                {title}
                {infoText ? <InfoTooltip label={title} text={infoText} /> : null}
            </h2>
        </div>
    )
}

export function MetricPanel({ icon, label, value, note }: { icon: ReactNode; label: string; value: string; note: string }) {
    return <div className="metric-panel panel"><span className="metric-panel-icon">{icon}</span><span className="metric-panel-label">{label}</span><strong>{value}</strong><small>{note}</small></div>
}

export function EvidenceStat({ label, value }: { label: string; value: string }) {
    return <div><span>{label}</span><strong>{value}</strong></div>
}

export function StatusBadge({ status }: { status: string }) {
    const label = status === 'NORMAL' ? 'Normal' : status === 'CRITICAL' ? 'Crítico' : 'Alerta'
    return <span className={`status-badge ${status.toLowerCase()}`}><i />{label}</span>
}

export function SeverityBadge({ severity }: { severity: string }) {
    const label = severity === 'CRITICAL' ? 'Crítica' : severity === 'HIGH' ? 'Alta' : severity === 'MEDIUM' ? 'Media' : 'Baja'
    return <span className={`severity-badge ${severity.toLowerCase()}`}><i />{label}</span>
}

export function EmptyState({ text }: { text: string }) {
    return <div className="empty-state"><span><Activity size={17} /></span><p>{text}</p></div>
}

export function ChartLoading() {
    return <div className="chart-loading" aria-label="Cargando gráfico" />
}

function SkeletonLine({ className = '' }: { className?: string }) {
    return <span className={`skeleton-line ${className}`} aria-hidden="true" />
}

export function WorkspaceSkeleton({ section }: { section: 'dashboard' | 'meters' | 'costs' | 'anomalies' }) {
    if (section === 'dashboard') {
        return <div className="workspace-skeleton" aria-label="Cargando resumen" role="status">
            <div className="skeleton-heading"><div><SkeletonLine className="skeleton-eyebrow" /><SkeletonLine className="skeleton-title" /><SkeletonLine className="skeleton-copy" /></div><SkeletonLine className="skeleton-button" /></div>
            <div className="skeleton-kpi-grid">{Array.from({ length: 6 }, (_, index) => <div className="skeleton-card" key={index}><SkeletonLine className="skeleton-icon" /><SkeletonLine className="skeleton-label" /><SkeletonLine className="skeleton-value" /><SkeletonLine className="skeleton-note" /></div>)}</div>
            <div className="skeleton-dashboard-grid"><div className="skeleton-card skeleton-chart"><SkeletonLine className="skeleton-label" /><SkeletonLine className="skeleton-panel-title" /><SkeletonLine className="skeleton-chart-area" /></div><div className="skeleton-card skeleton-chart"><SkeletonLine className="skeleton-label" /><SkeletonLine className="skeleton-panel-title" /><SkeletonLine className="skeleton-chart-area" /></div></div>
        </div>
    }

    return <div className="workspace-skeleton" aria-label={`Cargando ${section === 'meters' ? 'medidores' : 'anomalías'}`} role="status">
        <div className="skeleton-heading"><div><SkeletonLine className="skeleton-eyebrow" /><SkeletonLine className="skeleton-title" /><SkeletonLine className="skeleton-copy" /></div><SkeletonLine className="skeleton-meta" /></div>
        <div className="skeleton-toolbar"><SkeletonLine className="skeleton-search" /><SkeletonLine className="skeleton-filters" /><SkeletonLine className="skeleton-select" /></div>
        <div className="skeleton-table skeleton-card">{Array.from({ length: 7 }, (_, index) => <div className="skeleton-row" key={index}><SkeletonLine className="skeleton-cell wide" /><SkeletonLine className="skeleton-cell" /><SkeletonLine className="skeleton-cell" /><SkeletonLine className="skeleton-cell" /><SkeletonLine className="skeleton-cell narrow" /></div>)}</div>
    </div>
}

