export function buildHourWeekHeatmap(history: Array<{ timestamp: string; consumption_kwh: number }>) {
    const matrix = Array.from({ length: 7 }, () => Array.from({ length: 24 }, () => 0))
    const counts = Array.from({ length: 7 }, () => Array.from({ length: 24 }, () => 0))

    for (const reading of history) {
        const date = new Date(reading.timestamp)
        const dayIndex = (date.getDay() + 6) % 7
        const hour = date.getHours()
        matrix[dayIndex][hour] += reading.consumption_kwh
        counts[dayIndex][hour] += 1
    }

    return matrix.map((row, dayIndex) => row.map((_, hour) => {
        const count = counts[dayIndex][hour]
        return count > 0 ? Number((matrix[dayIndex][hour] / count).toFixed(2)) : 0
    }))
}

export function buildLoadDurationCurve(history: Array<{ consumption_kwh: number }>) {
    const values = history
        .map((item) => Number(item.consumption_kwh))
        .filter((value) => Number.isFinite(value) && value >= 0)
        .sort((left, right) => right - left)

    if (values.length === 0) return []

    return values.map((value, index) => ({
        share: Number((((index + 1) / values.length) * 100).toFixed(1)),
        demand: Number(value.toFixed(2)),
    }))
}
