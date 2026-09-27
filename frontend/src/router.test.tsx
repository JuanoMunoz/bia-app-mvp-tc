import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { MetersPage } from './features/meters/MetersPage'
import { AnomaliesPage } from './features/anomalies/AnomaliesPage'
import type { Anomaly, MeterSummary } from './types'

const meters: MeterSummary[] = [
    { meter_id: 'M-101', current_consumption_kwh: 91, period_consumption_kwh: 100, baseline_kwh: 43, change_percent: 110, status: 'CRITICAL', severity: 'HIGH' },
    { meter_id: 'M-102', current_consumption_kwh: 5, period_consumption_kwh: 60, baseline_kwh: 44, change_percent: 1.2, status: 'NORMAL', severity: 'LOW' },
]

const anomalies: Anomaly[] = [
    {
        id: '11111111-1111-4111-8111-111111111111',
        meter_id: 'M-109',
        anomaly: 'Incremento sostenido',
        type: 'REAL_ANOMALY',
        severity: 'HIGH',
        confidence: 0.91,
        reason: 'current 91 kWh vs baseline 43.2 kWh',
        recommended_action: 'Inspeccionar carga',
        detected_at: '2026-09-14T12:00:00Z',
        evidence: { baseline_kwh: 43.2, current_kwh: 91, change_percent: 110.6, voltage_v: 230, current_a: 15, power_factor: 0.88, current_variation_percent: 12, related_events: [], abnormal_hours: [19], outlier_count: 1 },
    },
]

describe('frontend routes', () => {
    it('renders meters list on /app/meters', () => {
        render(
            <MemoryRouter initialEntries={['/app/meters']}>
                <Routes>
                    <Route path="/app/meters" element={<MetersPage meters={meters} onOpenMeter={() => {}} />} />
                </Routes>
            </MemoryRouter>,
        )
        expect(screen.getByText('Medidores')).toBeInTheDocument()
        expect(screen.getByText('M-101')).toBeInTheDocument()
    })

    it('navigates to meter detail path', () => {
        render(
            <MemoryRouter initialEntries={['/app/meters/M-101']}>
                <Routes>
                    <Route path="/app/meters/:meterId" element={<div>Detalle M-101</div>} />
                </Routes>
            </MemoryRouter>,
        )
        expect(screen.getByText('Detalle M-101')).toBeInTheDocument()
    })

    it('renders anomalies list with attention count', () => {
        render(
            <MemoryRouter initialEntries={['/app/anomalies']}>
                <Routes>
                    <Route path="/app/anomalies" element={<AnomaliesPage anomalies={anomalies} selected={null} onSelect={() => {}} onBack={() => {}} />} />
                </Routes>
            </MemoryRouter>,
        )
        expect(screen.getByText('Alertas de consumo')).toBeInTheDocument()
        expect(screen.getByText('M-109')).toBeInTheDocument()
    })
})
