import {
    Area,
    AreaChart,
    Bar,
    BarChart,
    CartesianGrid,
    ReferenceLine,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from 'recharts'
import { buildLoadDurationCurve } from './charts-data'

export interface ConsumptionPoint {
    meter: string
    consumption: number
}

export interface HistoryPoint {
    time: string
    label: string
    consumption: number
}

export interface ComparisonPoint {
    label: string
    value: number
}

export function HeatmapChart({ data, labels }: { data: number[][]; labels: string[] }) {
    const values = data.flat().filter((value) => value > 0)
    const maxValue = values.length > 0 ? Math.max(...values) : 1

    return (
        <div className="heatmap-chart" aria-label="Mapa de consumo por hora y día de la semana">
            <div className="heatmap-axis">
                <span className="heatmap-day-label" aria-hidden="true" />
                <div className="heatmap-hours">
                    {Array.from({ length: 24 }, (_, hour) => (
                        <span key={`hour-${hour}`} className="heatmap-hour-label">{hour % 2 === 0 ? `${hour}:00` : ''}</span>
                    ))}
                </div>
            </div>
            {labels.map((day, dayIndex) => (
                <div key={day} className="heatmap-row">
                    <span className="heatmap-day-label">{day}</span>
                    <div className="heatmap-cells">
                        {data[dayIndex].map((value, hourIndex) => {
                            const opacity = maxValue <= 0 ? 0 : Math.min(1, value / maxValue)
                            const tone = value === 0 ? 'rgba(148, 163, 184, 0.08)' : `hsla(${170 - opacity * 55}, 70%, ${60 - opacity * 20}%, ${0.2 + opacity * 0.8})`
                            return (
                                <span
                                    key={`${day}-${hourIndex}`}
                                    className="heatmap-cell"
                                    style={{
                                        background: tone,
                                        borderColor: value === 0 ? 'rgba(148, 163, 184, 0.14)' : 'rgba(255,255,255,0.12)',
                                        boxShadow: value > 0 ? 'inset 0 0 0 1px rgba(255,255,255,0.04)' : 'none',
                                    }}
                                    title={`${day} · ${hourIndex}:00 · ${value.toFixed(1)} kWh`}
                                />
                            )
                        })}
                    </div>
                </div>
            ))}
        </div>
    )
}

export function LoadDurationCurveChart({ data }: { data: number[] }) {
    const curve = buildLoadDurationCurve(data.map((value) => ({ consumption_kwh: value })))
    if (curve.length === 0) {
        return <div className="empty-state"><span><svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true"><path d="M4 15h16M5 18l3-4 3 2 5-8 3 6" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" /></svg></span><p>Sin demanda suficiente para esta curva.</p></div>
    }

    return (
        <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={curve} margin={{ top: 8, right: 12, bottom: 0, left: -14 }}>
                <defs>
                    <linearGradient id="loadCurveFill" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor="#5d67ff" stopOpacity={0.38} />
                        <stop offset="100%" stopColor="#5d67ff" stopOpacity={0.02} />
                    </linearGradient>
                </defs>
                <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                <XAxis dataKey="share" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} tickFormatter={(value) => `${Number(value).toFixed(0)}%`} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <Tooltip
                    cursor={{ stroke: 'rgba(148,163,184,0.45)', strokeDasharray: '4 4' }}
                    contentStyle={{ backgroundColor: 'var(--surface-strong)', border: '1px solid var(--line)', borderRadius: 10, color: 'var(--ink)', boxShadow: '0 10px 24px rgba(15, 23, 42, 0.2)' }}
                    labelStyle={{ color: 'var(--ink)', fontWeight: 600 }}
                    itemStyle={{ color: 'var(--ink)' }}
                    labelFormatter={(value) => `${Number(value).toFixed(0)}% del tiempo`}
                    formatter={(value) => [`${Number(value).toFixed(1)} kWh`, 'Demanda']}
                />
                <Area type="monotone" dataKey="demand" stroke="#5d67ff" strokeWidth={2} fill="url(#loadCurveFill)" dot={false} activeDot={{ r: 4, fill: '#5d67ff' }} />
            </AreaChart>
        </ResponsiveContainer>
    )
}

export function ConsumptionChart({ data }: { data: ConsumptionPoint[] }) {
    return (
        <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -18 }}>
                <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                <XAxis dataKey="meter" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <Tooltip
                    cursor={{ fill: 'var(--chart-hover)' }}
                    contentStyle={{ backgroundColor: 'var(--surface-strong)', border: '1px solid var(--line)', borderRadius: 10, color: 'var(--ink)' }}
                    labelStyle={{ color: 'var(--ink)', fontWeight: 600 }}
                    itemStyle={{ color: 'var(--ink)' }}
                    formatter={(value) => [`${value} kWh`, 'Consumo']}
                />
                <Bar dataKey="consumption" fill="#08a98f" radius={[3, 3, 0, 0]} maxBarSize={38} />
            </BarChart>
        </ResponsiveContainer>
    )
}

export function MeterHistoryChart({ data, baseline }: { data: HistoryPoint[]; baseline: number }) {
    return (
        <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={data} margin={{ top: 10, right: 14, bottom: 0, left: -14 }}>
                <defs><linearGradient id="consumptionFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#08a98f" stopOpacity={0.22} /><stop offset="95%" stopColor="#08a98f" stopOpacity={0} /></linearGradient></defs>
                <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                <XAxis dataKey="label" interval={Math.max(0, Math.ceil(data.length / 8) - 1)} axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <Tooltip
                    contentStyle={{ backgroundColor: 'var(--surface-strong)', border: '1px solid var(--line)', borderRadius: 10, color: 'var(--ink)' }}
                    labelStyle={{ color: 'var(--ink)', fontWeight: 600 }}
                    itemStyle={{ color: 'var(--ink)' }}
                    labelFormatter={(_, payload) => payload[0]?.payload.time ? new Date(payload[0].payload.time).toLocaleString('es-CO') : ''}
                    formatter={(value) => [`${Number(value).toFixed(2)} kWh`, 'Consumo']}
                />
                <ReferenceLine y={baseline} stroke="var(--coral)" strokeDasharray="5 5" />
                <Area type="monotone" dataKey="consumption" stroke="#087e70" strokeWidth={2} fill="url(#consumptionFill)" dot={false} activeDot={{ r: 4, fill: '#087e70' }} />
            </AreaChart>
        </ResponsiveContainer>
    )
}

export function ComparisonChart({ data }: { data: ComparisonPoint[] }) {
    return (
        <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} layout="vertical" margin={{ left: 8, right: 20, top: 6, bottom: 4 }}>
                <CartesianGrid horizontal={false} stroke="var(--chart-grid)" />
                <XAxis type="number" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 11 }} />
                <YAxis type="category" dataKey="label" axisLine={false} tickLine={false} tick={{ fill: 'var(--chart-label)', fontSize: 12 }} width={68} />
                <Tooltip
                    cursor={{ fill: 'transparent' }}
                    contentStyle={{ border: '1px solid var(--line)', borderRadius: 10, backgroundColor: 'var(--surface-strong)', color: 'var(--ink)' }}
                    labelStyle={{ color: 'var(--ink)', fontWeight: 600 }}
                    itemStyle={{ color: 'var(--ink)' }}
                    formatter={(value) => [`${Number(value).toFixed(1)} kWh`, 'Consumo']}
                />
                <Bar dataKey="value" fill="#168e81" radius={[0, 3, 3, 0]} barSize={24} />
            </BarChart>
        </ResponsiveContainer>
    )
}