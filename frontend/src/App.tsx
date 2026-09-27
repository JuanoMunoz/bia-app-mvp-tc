import { useEffect, useState } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Navigate, Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom'
import { AppShell } from './components/AppShell'
import { AnomaliesPage } from './features/anomalies/AnomaliesPage'
import { InvestigationPage } from './features/anomalies/InvestigationPage'
import { LoginPage } from './features/auth/LoginPage'
import { DashboardPage } from './features/dashboard/DashboardPage'
import { CostsPage } from './features/costs/CostsPage'
import { MeterDetailPage, MetersPage } from './features/meters'
import { useEnergyWorkspace } from './hooks/useEnergyWorkspace'
import { useTheme } from './hooks/useTheme'
import type { Anomaly, Section } from './types'
import { LandingPage } from './features/marketing/LandingPage'
import { GsapPageTransition } from './components/GsapPageTransition'
import { WorkspaceSkeleton } from './components/ui'
import { NotFoundPage } from './features/errors/NotFoundPage'
import { startViewTransition } from './utils/viewTransition'
import './App.css'
import './theme.css'

function sectionFromPath(pathname: string): Section {
    if (pathname.startsWith('/app/meters')) return 'meters'
    if (pathname.startsWith('/app/costs')) return 'costs'
    if (pathname.startsWith('/app/anomalies')) return 'anomalies'
    return 'dashboard'
}

function titleFromPath(pathname: string, hasMeter: boolean, hasAnomaly: boolean): string {
    if (hasMeter) return 'Detalle de medidor'
    if (hasAnomaly) return 'Detalle de alerta'
    const section = sectionFromPath(pathname)
    if (section === 'meters') return 'Medidores'
    if (section === 'costs') return 'Consumo y gastos'
    if (section === 'anomalies') return 'Alertas'
    return 'Resumen'
}

function App() {
    const location = useLocation()

    useEffect(() => {
        if (location.pathname.startsWith('/app')) {
            document.title = 'Gestión de energía | Bia Energy'
            document.querySelector('meta[name="robots"]')?.setAttribute('content', 'noindex, nofollow, noarchive')
        }
    }, [location.pathname])

    return (
        <Routes>
            <Route path="/" element={<GsapPageTransition><LandingPage /></GsapPageTransition>} />
            <Route path="/app" element={<Navigate to="/app/dashboard" replace />} />
            <Route path="/app/*" element={<ConsoleApp />} />
            <Route path="*" element={<GsapPageTransition><NotFoundPage /></GsapPageTransition>} />
        </Routes>
    )
}

function ConsoleApp() {
    const [authenticated, setAuthenticated] = useState(() => localStorage.getItem('bia-session') === 'active')
    const { theme, toggleTheme } = useTheme()

    const handleLogin = () => {
        localStorage.setItem('bia-session', 'active')
        startViewTransition(() => setAuthenticated(true))
    }

    const handleSignOut = () => {
        localStorage.removeItem('bia-session')
        startViewTransition(() => setAuthenticated(false))
    }

    if (!authenticated) return <LoginPage onLogin={handleLogin} />

    return <Workspace theme={theme} onToggleTheme={toggleTheme} onSignOut={handleSignOut} />
}

function Workspace({ theme, onToggleTheme, onSignOut }: { theme: 'dark' | 'light'; onToggleTheme: () => void; onSignOut: () => void }) {
    const workspace = useEnergyWorkspace()
    const location = useLocation()
    const navigate = useNavigate()
    const section = sectionFromPath(location.pathname)
    const title = titleFromPath(location.pathname, workspace.selectedMeter !== null, workspace.selectedAnomaly !== null)

    const goSection = (next: Section) => {
        startViewTransition(() => {
            workspace.setSelectedMeter(null)
            workspace.setSelectedAnomaly(null)
            navigate(`/app/${next}`)
        })
    }

    const openMeter = (meterId: string) => {
        if (!meterId) return
        startViewTransition(() => navigate(`/app/meters/${encodeURIComponent(meterId)}`))
    }

    const openAnomaly = (anomalyId: string) => {
        startViewTransition(() => navigate(`/app/anomalies/${encodeURIComponent(anomalyId)}`))
    }

    return (
        <AppShell section={section} title={title} theme={theme} meterCount={workspace.summary?.meter_count ?? 0} periodEnd={workspace.summary?.period_end} connectionState={workspace.error ? 'disconnected' : workspace.summary ? 'connected' : 'connecting'} onToggleTheme={onToggleTheme} onSection={goSection} onSignOut={onSignOut}>
            {workspace.error && <div className="error-banner" role="alert"><AlertCircle size={17} /><span>{workspace.error}</span><button onClick={workspace.retry} aria-label="Reintentar conexión"><RefreshCw size={15} /></button></div>}
            {workspace.loading ? <WorkspaceSkeleton section={section} /> : !workspace.error ? (
                <Routes>
                    <Route index element={<Navigate to="dashboard" replace />} />
                    <Route path="dashboard" element={<DashboardPage summary={workspace.summary} meters={workspace.meters} siteHistory={workspace.siteHistory} anomalies={workspace.anomalies} analysis={workspace.analysis} analyzing={workspace.analyzing} onAnalyze={workspace.analyze} onOpenMeter={openMeter} onOpenAnomaly={(item) => openAnomaly(item.id)} />} />
                    <Route path="meters" element={<MetersPage meters={workspace.meters} periodEnd={workspace.summary?.period_end} onOpenMeter={openMeter} />} />
                    <Route path="meters/:meterId" element={<MeterDetailRoute selectedMeter={workspace.selectedMeter} openMeter={workspace.openMeter} onBack={() => startViewTransition(() => navigate('/app/meters'))} />} />
                    <Route path="costs" element={<CostsPage />} />
                    <Route path="anomalies" element={<AnomaliesPage anomalies={workspace.anomalies} selected={null} onSelect={(item) => openAnomaly(item.id)} onBack={() => startViewTransition(() => navigate('/app/anomalies'))} />} />
                    <Route path="anomalies/:anomalyId" element={<AnomalyDetailRoute anomalies={workspace.anomalies} openAnomaly={workspace.openAnomaly} selectedAnomaly={workspace.selectedAnomaly} />} />
                    <Route path="*" element={<NotFoundPage />} />
                </Routes>
            ) : null}
        </AppShell>
    )
}

function MeterDetailRoute({ selectedMeter, openMeter, onBack }: { selectedMeter: ReturnType<typeof useEnergyWorkspace>['selectedMeter']; openMeter: (meterId: string) => Promise<void>; onBack: () => void }) {
    const { meterId = '' } = useParams()
    const decoded = decodedId(meterId)

    useEffect(() => {
        if (decoded && selectedMeter?.meter_id !== decoded) void openMeter(decoded)
    }, [decoded, selectedMeter?.meter_id, openMeter])

    if (!selectedMeter || selectedMeter.meter_id !== decoded) return <WorkspaceSkeleton section="meters" />
    return <MeterDetailPage key={selectedMeter.meter_id} meter={selectedMeter} onBack={onBack} />
}

function AnomalyDetailRoute({ anomalies, openAnomaly, selectedAnomaly }: { anomalies: Anomaly[]; openAnomaly: (item: Anomaly) => Promise<void>; selectedAnomaly: Anomaly | null }) {
    const { anomalyId = '' } = useParams()
    const navigate = useNavigate()
    const decoded = decodedId(anomalyId)
    const cached = anomalies.find((item) => item.id === decoded) ?? selectedAnomaly

    useEffect(() => {
        if (!cached && decoded) void openAnomaly({ id: decoded } as Anomaly)
    }, [cached, decoded, openAnomaly])

    if (!cached) return <WorkspaceSkeleton section="anomalies" />
    return <InvestigationPage anomaly={cached} onBack={() => startViewTransition(() => navigate('/app/anomalies'))} />
}

function decodedId(value: string) {
    try {
        return decodeURIComponent(value)
    } catch {
        return value
    }
}

export default App
