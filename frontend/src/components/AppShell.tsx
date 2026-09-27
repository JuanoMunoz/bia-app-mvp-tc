import type { ReactNode } from 'react'
import {
    ChevronRight,
    Gauge,
    LayoutDashboard,
    LogOut,
    Moon,
    ShieldAlert,
    Sun,
} from 'lucide-react'
import type { Section } from '../types'
import { GsapPageTransition } from './GsapPageTransition'
import { BrandLogo } from './BrandLogo'
import { formatDate } from '../utils/format'

export function AppShell({
    section,
    title,
    theme,
    connectionState,
    onToggleTheme,
    onSection,
    onSignOut,
    meterCount,
    periodEnd,
    children,
}: {
    section: Section
    title: string
    theme: 'dark' | 'light'
    connectionState: 'connected' | 'connecting' | 'disconnected'
    onToggleTheme: () => void
    onSection: (section: Section) => void
    onSignOut: () => void
    meterCount: number
    periodEnd?: string
    children: ReactNode
}) {
    const items = [
        { id: 'dashboard' as const, label: 'Resumen', icon: LayoutDashboard },
        { id: 'meters' as const, label: 'Medidores', icon: Gauge },
        { id: 'costs' as const, label: 'Consumo y gastos', icon: Gauge },
        { id: 'anomalies' as const, label: 'Alertas', icon: ShieldAlert },
    ]
    return (
        <div className="workspace" data-theme={theme}>
            <aside className="sidebar">
                <div className="sidebar-brand"><BrandLogo className="brand-logo-app" /><span>Bia<span className="brand-light"> / energy</span></span></div>
                <div className="workspace-label">EMPRESA</div>
                <nav className="side-nav" aria-label="Navegación principal">
                    {items.map(({ id, label, icon: Icon }) => (
                        <button
                            key={id}
                            type="button"
                            className={`nav-item ${section === id ? 'active' : ''}`}
                            onClick={() => onSection(id)}
                            aria-current={section === id ? 'page' : undefined}
                            aria-label={label}
                        >
                            <Icon size={18} strokeWidth={1.8} /><span>{label}</span>
                            {section === id && <span className="nav-active-mark" aria-hidden="true" />}
                        </button>
                    ))}
                </nav>
                <div className="sidebar-bottom">
                    <div className="plant-status">
                        <span>
                            <strong>Planta principal</strong>
                            <small>{meterCount} medidores conectados</small></span></div>
                    <button type="button" className="profile-button" onClick={onSignOut} title="Cerrar sesión" aria-label="Cerrar sesión"><span className="profile-avatar" aria-hidden="true">JM</span><span className="profile-copy"><strong>Juan Muñoz</strong><small>Administrador de energía</small></span><LogOut size={16} /></button>
                </div>
            </aside>
            <div className="workspace-main">
                <header className="topbar">
                    <div className="breadcrumbs"><span>Bia Energy</span><ChevronRight size={14} /><strong>{title}</strong></div>
                    <div className="topbar-right">
                        <span className={`system-status ${connectionState}`}>{connectionState === 'connected' ? 'DATOS ACTUALIZADOS' : connectionState === 'connecting' ? 'CARGANDO DATOS' : 'SIN CONEXIÓN'}</span>
                        <span className="topbar-date">{periodEnd ? formatDate(periodEnd, { month: 'short' }).toUpperCase() : '—'}</span>
                        <button className="theme-toggle" onClick={onToggleTheme} aria-label={theme === 'dark' ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'} title={theme === 'dark' ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}>{theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}</button>
                        <span className="topbar-avatar">JM</span>
                    </div>
                </header>
                <main className="page-content"><GsapPageTransition key={title}>{children}</GsapPageTransition></main>
            </div>
        </div>
    )
}