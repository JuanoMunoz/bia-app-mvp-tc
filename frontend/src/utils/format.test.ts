import { describe, expect, it } from 'vitest'
import { anomalyTypeLabel, compactNumber, formatDate, revisionTimestamp, severityOrder, timeAgo } from './format'

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

    it('shows a placeholder when there is no real revision', () => {
        expect(timeAgo(null)).toBe('—')
        expect(timeAgo(undefined)).toBe('—')
        expect(timeAgo('')).toBe('—')
        expect(timeAgo('not-a-date')).toBe('—')
        // Tiempo cero de Go: antes mostraba "hace NNN días"
        expect(timeAgo('0001-01-01T00:00:00Z')).toBe('—')
        expect(revisionTimestamp(null)).toBeNull()
        expect(revisionTimestamp({ started_at: '0001-01-01T00:00:00Z' })).toBeNull()
        expect(revisionTimestamp({ completed_at: '', started_at: '0001-01-01T00:00:00Z' })).toBeNull()
    })

    it('prefers completed_at over started_at for revisions', () => {
        const started = new Date(Date.now() - 60_000).toISOString()
        const completed = new Date().toISOString()
        expect(revisionTimestamp({ completed_at: completed, started_at: started })).toBe(completed)
        expect(revisionTimestamp({ started_at: started })).toBe(started)
    })

    it('formats dates with day and year', () => {
        const text = formatDate('2026-09-14T12:00:00Z')
        expect(text).toContain('2026')
        expect(text).toContain('14')
    })
})
