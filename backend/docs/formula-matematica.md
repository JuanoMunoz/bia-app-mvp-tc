# Fórmula matemática del detector de anomalías

Este proyecto detecta anomalías de consumo comparando el comportamiento reciente del medidor contra un baseline histórico y evaluando además desviaciones eléctricas y patrones horarios.

## 1) Baseline y consumo actual

Para cada medidor se ordenan las lecturas por fecha y se calcula:

- `baseline`: promedio de consumo de los primeros 7 días de la serie (168 horas, ventana de referencia)
- `current`: promedio de consumo de las últimas 24 horas

La lógica exacta está en `backend/internal/anomaly/detector.go` y se resume así:

```go
baselineWindow := 7 * 24
currentWindow := 24
baseline := meanConsumption(series[:baselineWindow])
current := series[len(series)-currentWindow:]
currentMean := meanConsumption(current)
change := percentChange(baseline, currentMean)
```

La variación porcentual se calcula con:

$$
\Delta\% = \left(\frac{consumoActual}{consumoBase} - 1\right) \times 100
$$

La función exacta es:

```go
func percentChange(baseline, current float64) float64 {
    if baseline == 0 {
        return 0
    }
    return (current/baseline - 1) * 100
}
```

Cuando `abs(change) >= 25`, el sistema considera que hubo un cambio significativo de consumo en la instalación.

## 2) Uso de la media

El promedio se calcula como media aritmética sobre la serie del medidor:

$$
\mu = \frac{\sum x_i}{n}
$$

En código:

```go
func meanField(readings []reading.Reading, field func(reading.Reading) float64) float64 {
    if len(readings) == 0 {
        return 0
    }
    total := 0.0
    for _, item := range readings {
        total += field(item)
    }
    return total / float64(len(readings))
}
```

Esto se usa para:

- consumo promedio por periodo
- tensión promedio
- corriente promedio
- factor de potencia promedio

## 3) Detección de patrones horarios anómalos

Además de comparar el promedio general, el sistema revisa si cada hora del día se comporta distinto respecto al historial anterior.

```go
for hour := range baseline {
    baselineMean := meanValues(baseline[hour])
    currentMean := meanValues(current[hour])
    if baselineMean > 0 && len(current[hour]) > 0 && math.Abs(percentChange(baselineMean, currentMean)) >= 25 {
        result = append(result, hour)
    }
}
```

Si aparecen al menos 4 horas con desviación mayor o igual al 25%, se marca un patrón horario anómalo.

Esto ayuda a detectar cambios del perfil operativo sin depender solo del consumo total.

## 4) Detección de picos puntuales

También se analiza si una lectura puntual está muy lejos de sus vecinos. El sistema compara la lecturacon la mediana de 4 lecturas vecinas:

```go
neighbors := []float64{
    series[index-2].ConsumptionKWh,
    series[index-1].ConsumptionKWh,
    series[index+1].ConsumptionKWh,
    series[index+2].ConsumptionKWh,
}
median := (neighbors[1] + neighbors[2]) / 2
if median > 0 && math.Abs(series[index].ConsumptionKWh-median)/median > 0.5 {
    // outlier
}
```

La condición es:

$$
\frac{|x - mediana|}{mediana} > 0.5
$$

Es decir, si la lectura tiene más del 50% de diferencia respecto a la mediana local, se considera un pico puntual.

## 5) Calidad de datos

El sistema también revisa problemas eléctricos:

- `PowerFactor < 0.7` (límite típico de penalización por reactiva)
- `VoltageV < 207` o `VoltageV > 253` (230 V ±10 %, banda de UNE-EN 50160)
- estado distinto de `OK`

```go
if item.PowerFactor < 0.7 || item.VoltageV < 207 || item.VoltageV > 253 || !strings.EqualFold(item.Status, "OK") {
    issueCount++
}
```

Si se detectan varios problemas en las últimas 24 lecturas y el cambio de consumo no es grande, el sistema clasifica la anomalía como inconsistencia en variables eléctricas.

## 6) Cómo se decide el tipo de anomalía

La lógica final en `detectMeter` prioriza estos casos:

1. Inconsistencia en variables eléctricas
2. Cambio de consumo significativo
3. Patrón horario anómalo
4. Pico puntual de consumo
5. Parada programada no escalable o explicada por eventos

El resultado incluye:\
- tipo (real, explicable, falso positivo, calidad de datos)
- severidad
- confianza
- evidencia
- acción recomendada

## 7) Resumen práctico

La detección no se basa en un único umbral, sino en una combinación de:

- comparación contra baseline histórico
- análisis del último día
- patrones por hora del día
- outliers locales
- calidad eléctrica
- correlación con eventos operativos

Esto permite distinguir entre:

- consumo real anómalo
- variación esperada por cambio operativo
- falsas alarmas provocadas por cortes programados
- errores de medición o calidad del dato

En otras palabras, la fórmula principal es una comparación porcentual del consumo actual frente al baseline, y sobre esa base se agregan múltiples validaciones para reducir falsos positivos.

## 8) Ventana de correlación con eventos

Un evento solo explica una anomalía si ocurre cerca de su inicio:

```
|timestamp_evento - inicio_cambio| <= eventCorrelationWindow (24 h)
```

La ventana de 24 h existe porque los eventos se registran con resolución de día ("2026-09-11 00:00") mientras las lecturas son horarias: un cambio real puede aparecer horas después del registro. `TestDetectorIgnoresDistantEvents` fija el comportamiento opuesto: un `OPERATIONAL_CHANGE` a 10 días del cambio no convierte un `REAL_ANOMALY` en `EXPLAINABLE_ANOMALY` ni aparece en `related_events`.

## 9) Confianza como prior heurístico

Los valores de `confidence` (0.86–0.94 según la rama de detección) son priors heurísticos, no probabilidades calibradas: ordenan prioridad entre hallazgos pero no deben leerse como "probabilidad de que la anomalía sea real". Para calibrarlos, contrastar cada tipo contra desenlaces etiquetados por operación (confirmada/descartada) y ajustar por tipo.
