import { useEffect, useState } from 'react'
import { getAnomalies, getAnomaly, getDashboardSummary, getMeter, getMeters, getSiteHistory, runAnalysis } from '../api'
import type { Analysis, Anomaly, DashboardSummary, MeterDetail, MeterSummary, SitePoint } from '../types'

export function useEnergyWorkspace() {
    const [summary, setSummary] = useState<DashboardSummary | null>(null)
    const [meters, setMeters] = useState<MeterSummary[]>([])
    const [siteHistory, setSiteHistory] = useState<SitePoint[]>([])
    const [anomalies, setAnomalies] = useState<Anomaly[]>([])
    const [analysis, setAnalysis] = useState<Analysis | null>(null)
    const [selectedMeter, setSelectedMeter] = useState<MeterDetail | null>(null)
    const [selectedAnomaly, setSelectedAnomaly] = useState<Anomaly | null>(null)
    const [analyzing, setAnalyzing] = useState(false)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [refreshKey, setRefreshKey] = useState(0)

    useEffect(() => {
        let current = true
        Promise.all([getDashboardSummary(), getMeters(), getAnomalies(), getSiteHistory()])
            .then(([nextSummary, nextMeters, nextAnomalies, nextHistory]) => {
                if (!current) return
                setSummary(nextSummary)
                setMeters(nextMeters)
                setAnomalies(nextAnomalies)
                setSiteHistory(nextHistory)
                setAnalysis(nextSummary.latest_analysis)
                setError(null)
                setLoading(false)
            })
            .catch((cause: unknown) => {
                if (current) {
                    setError(cause instanceof Error ? cause.message : 'No se pudo conectar con la API.')
                    setLoading(false)
                }
            })
        return () => {
            current = false
        }
    }, [refreshKey])

    async function refreshData() {
        const [nextSummary, nextMeters, nextAnomalies, nextHistory] = await Promise.all([
            getDashboardSummary(),
            getMeters(),
            getAnomalies(),
            getSiteHistory(),
        ])
        setSummary(nextSummary)
        setMeters(nextMeters)
        setAnomalies(nextAnomalies)
        setSiteHistory(nextHistory)
        setAnalysis(nextSummary.latest_analysis)
    }

    async function analyze() {
        setAnalyzing(true)
        setError(null)
        try {
            const result = await runAnalysis()
            await refreshData()
            // El resultado recién ejecutado es la revisión más nueva: se impone sobre
            // cualquier resumen desactualizado que haya traído refreshData().
            setSummary((previous) => {
                if (!previous) return previous
                const previousStamp = previous.latest_analysis?.completed_at || previous.latest_analysis?.started_at
                const nextStamp = result.completed_at || result.started_at
                if (!previousStamp) return { ...previous, latest_analysis: result }
                if (nextStamp && new Date(nextStamp).getTime() >= new Date(previousStamp).getTime()) {
                    return { ...previous, latest_analysis: result }
                }
                return previous
            })
            setAnalysis((current) => {
                if (!current) return result
                const currentStamp = current.completed_at || current.started_at
                const nextStamp = result.completed_at || result.started_at
                if (nextStamp && (!currentStamp || new Date(nextStamp).getTime() >= new Date(currentStamp).getTime())) {
                    return result
                }
                return current
            })
            setSelectedAnomaly(null)
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : 'No se pudo completar el análisis.')
        } finally {
            setAnalyzing(false)
        }
    }

    async function openMeter(meterId: string) {
        if (!meterId) return
        setError(null)
        try {
            setSelectedMeter(await getMeter(meterId))
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : 'No se pudo cargar el medidor.')
        }
    }

    async function openAnomaly(item: Anomaly) {
        setError(null)
        try {
            setSelectedAnomaly(await getAnomaly(item.id))
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : 'No se pudo cargar la investigación.')
            setSelectedAnomaly(item)
        }
    }

    function retry() {
        setError(null)
        setLoading(true)
        setRefreshKey((current) => current + 1)
    }

    return {
        summary,
        meters,
        siteHistory,
        anomalies,
        analysis,
        selectedMeter,
        selectedAnomaly,
        analyzing,
        loading,
        error,
        analyze,
        openMeter,
        openAnomaly,
        retry,
        refreshData,
        setSelectedMeter,
        setSelectedAnomaly,
    }
}