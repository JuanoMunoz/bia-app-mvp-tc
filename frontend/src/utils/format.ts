export function severityOrder(severity: string) {
    return ({ CRITICAL: 0, HIGH: 1, MEDIUM: 2, LOW: 3 } as Record<string, number>)[severity] ?? 4
}

export function compactNumber(value: number) {
    return new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 }).format(value)
}

export function timeAgo(value: string) {
    const minutes = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 60_000))
    if (minutes < 1) return 'ahora'
    if (minutes < 60) return `hace ${minutes} min`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `hace ${hours} h`
    return `hace ${Math.floor(hours / 24)} d`
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
