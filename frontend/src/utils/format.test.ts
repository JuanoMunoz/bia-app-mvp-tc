import { describe, expect, it } from 'vitest'
import { anomalyTypeLabel, compactNumber, formatDate, severityOrder, timeAgo } from './format'

describe('format utils', () => {
    it('orders severities from critical to low', () => {
        expect(severityOrder('CRITICAL')).toBe(0)
        expect(severityOrder('HIGH')).toBe(1)
        expect(severityOrder('MEDIUM')).toBe(2)
        expect(severityOrder('LOW')).toBe(3)
        expect(severityOrder('UNKNOWN')).toBe(4)
    })

    it('labels anomaly types in Spanish', () => {
        expect(anomalyTypeLabel('REAL_ANOMALY')).toBe('Variación real')
        expect(anomalyTypeLabel('EXPLAINABLE_ANOMALY')).toBe('Variación explicable')
        expect(anomalyTypeLabel('FALSE_POSITIVE')).toBe('Alerta descartada')
        expect(anomalyTypeLabel('DATA_QUALITY')).toBe('Calidad de datos')
    })

    it('formats numbers with es-CO locale', () => {
        expect(compactNumber(1050)).toBe('1.050')
    })

    it('reports recent dates as now', () => {
        expect(timeAgo(new Date().toISOString())).toBe('ahora')
        expect(timeAgo(new Date(Date.now() - 90 * 60_000).toISOString())).toBe('hace 1 h')
    })

    it('formats dates with day and year', () => {
        const text = formatDate('2026-09-14T12:00:00Z')
        expect(text).toContain('2026')
        expect(text).toContain('14')
    })
})
