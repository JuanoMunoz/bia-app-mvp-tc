export function severityOrder(severity: string) {
    return ({ CRITICAL: 0, HIGH: 1, MEDIUM: 2, LOW: 3 } as Record<string, number>)[severity] ?? 4
}

export function compactNumber(value: number) {
    return new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 }).format(value)
}

export function timeAgo(value: string | null | undefined) {
    if (!value) return '—'
    const time = new Date(value).getTime()
    // Fecha inválida, vacía o tiempo cero de Go (0001-01-01) / epoch: no hay revisión real.
    if (Number.isNaN(time) || time <= 0) return '—'
    const minutes = Math.max(0, Math.floor((Date.now() - time) / 60_000))
    if (minutes < 1) return 'ahora'
    if (minutes < 60) return `hace ${minutes} min`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `hace ${hours} h`
    return `hace ${Math.floor(hours / 24)} d`
}

/** Devuelve el timestamp válido de una revisión (prefiere completed_at) o null si no hay revisión real. */
export function revisionTimestamp(revision: { completed_at?: string | null; started_at?: string | null } | null | undefined) {
    const raw = revision?.completed_at || revision?.started_at
    if (!raw) return null
    const time = new Date(raw).getTime()
    if (Number.isNaN(time) || time <= 0) return null
    return raw
}

export function anomalyTypeLabel(type: string) {
    return ({
        REAL_ANOMALY: 'Variación real',
        EXPLAINABLE_ANOMALY: 'Variación explicable',
        FALSE_POSITIVE: 'Alerta descartada',
        DATA_QUALITY: 'Calidad de datos',
    } as Record<string, string>)[type] ?? type.replaceAll('_', ' ').toLowerCase()
}

export function formatDate(value: string, options: Intl.DateTimeFormatOptions = {}) {
    return new Intl.DateTimeFormat('es-CO', { day: '2-digit', month: 'short', year: 'numeric', ...options }).format(new Date(value))
}

export function formatDateRange(start: string, end: string) {
    return `${formatDate(start)} — ${formatDate(end)}`
}
