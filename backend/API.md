# API v1

Base URL: `/api/v1`

Qué es: API REST del backend Go + Gin para consumo eléctrico, anomalías, tarifas y costos. El cálculo de anomalías es determinista y local; Gemini solo agrega explicación cuando `GEMINI_API_KEY` está configurada.

Compatibilidad: las mismas rutas existen sin versionar (`/health`, `/meters`, …) como alias mientras los clientes migran a `/api/v1`. Usa siempre `/api/v1`.

## Convenciones

- Fechas: RFC3339 UTC (`2026-09-01T00:00:00Z`).
- `meter_id`: formato `M-NNN` (`M-101`…`M-112`). Otro formato → `400 {"error":"invalid meter id"}`.
- IDs de análisis/anomalía: UUID v4. UUID inválido → `400 {"error":"invalid analysis id"}` / `{"error":"invalid anomaly id"}`.
- Auth: no hay. El login del frontend es demo local (`bia-session` en `localStorage`).
- `Content-Type: application/json` en respuestas JSON. `PATCH`/`POST` con body exigen `Content-Type: application/json`.

## Errores

| Código | Cuándo |
|---|---|
| `400` | `meter_id` inválido, `period_start`/`period_end`/`from`/`to` no RFC3339, body JSON inválido, tarifa inválida (`provider` vacío, `price_per_kwh <= 0`, `valid_to < valid_from`), `format` de reporte no soportado |
| `404` | medidor, anomalía o análisis inexistente |
| `429` | rate limit superado |
| `500` | `{"error":"internal server error"}` (detalle solo en logs) |
| `503` | `GET /health` con DB caída → `{"status":"unhealthy","service":"api"}`; `PATCH /meters/:id` sin store → `{"error":"meter profile storage unavailable"}` |

Ejemplo error:

```json
{ "error": "invalid meter id" }
```

## Rate limit y paginación

Negocio: 60 req/min por IP. Exentos: `health`, `isalive`, `ping`. Al exceder: `429` + `Retry-After` + `X-RateLimit-*`.

Listas (`/meters`, `/meters/:id/readings`, `/anomalies`): `offset` (def. 0) y `limit` (def. 50, máx. 100).

```json
{ "pagination": { "offset": 0, "limit": 50, "total": 12, "has_more": false } }
```

## Estado del servicio

### `GET /health`

Chequea API + PostgreSQL. `200` sano, `503` si la DB no responde.

```bash
curl http://localhost:8080/api/v1/health
# {"status":"healthy","service":"api","timestamp":"2026-09-14T12:00:00Z"}
```

### `GET /isalive`

Vida del proceso, sin tocar DB. `200 {"status":"alive"}`.

### `GET /ping`

Conectividad rápida. `200 {"message":"pong"}`.

## Dashboard

### `GET /dashboard/summary`

Agregado de consumo, conteos y último análisis.

```bash
curl http://localhost:8080/api/v1/dashboard/summary
```

```json
{
  "meter_count": 12,
  "total_consumption_kwh": 1050.5,
  "anomaly_count": 4,
  "high_priority_count": 2,
  "aggregate_confidence": 0.89,
  "latest_analysis": { "id": "f310f1fb-…", "status": "COMPLETED", "anomalies": [] },
  "period_start": "2026-09-01T00:00:00Z",
  "period_end": "2026-09-14T23:00:00Z"
}
```

`latest_analysis` es `null` si aún no hay análisis.

### `GET /dashboard/history`

Serie horaria agregada del sitio, ascendente. Usada por gráficos de patrón operativo.

```bash
curl http://localhost:8080/api/v1/dashboard/history
# {"points":[{"timestamp":"2026-09-01T00:00:00Z","consumption_kwh":250.5}, …]}
```

Garantías: `>300` puntos en dataset demo, ordenados, sin consumos negativos.

## Medidores

### `GET /meters`

Lista resúmenes. Query: `filter=all|normal|alert|critical` (def. `all`), `meter_id` o `q` (búsqueda parcial, ej. `M-10`), `sort=severity|consumption|variation` (def. `severity`), `order=asc|desc` (def. `desc`), `offset`, `limit`.

```bash
curl "http://localhost:8080/api/v1/meters?filter=critical&meter_id=M-109&sort=consumption&limit=10"
```

```json
{
  "meters": [{
    "meter_id": "M-109",
    "provider": "bia",
    "region": "CUND-EAST",
    "rate_type": "industrial",
    "current_consumption_kwh": 91,
    "period_consumption_kwh": 2100.5,
    "baseline_kwh": 43.2,
    "change_percent": 110.6,
    "status": "CRITICAL",
    "severity": "HIGH",
    "anomaly": { "id": "…", "type": "REAL_ANOMALY", "severity": "HIGH" }
  }],
  "pagination": { "offset": 0, "limit": 10, "total": 1, "has_more": false }
}
```

### `GET /meters/:meterId`

Detalle + métricas eléctricas + historial reciente.

```bash
curl http://localhost:8080/api/v1/meters/M-101
```

```json
{
  "meter_id": "M-101",
  "status": "NORMAL",
  "severity": "LOW",
  "voltage_v": 230.1,
  "current_a": 12.4,
  "power_factor": 0.92,
  "history": [{
    "timestamp": "2026-09-14T10:00:00Z",
    "consumption_kwh": 5.2,
    "voltage_v": 230.1,
    "current_a": 12.4,
    "power_factor": 0.92,
    "status": "normal"
  }]
}
```

`400` si el id no es `M-NNN`; `404 {"error":"meter not found"}` si no existe.

### `PATCH /meters/:meterId`

Actualiza perfil tarifario del medidor. Al menos `provider` o `region`. `rate_type` por defecto `industrial`.

```bash
curl -X PATCH http://localhost:8080/api/v1/meters/M-101 \
  -H 'Content-Type: application/json' \
  -d '{"provider":"EPM","region":"Medellín","rate_type":"industrial"}'
```

```json
{
  "meter_id": "M-101",
  "provider": "EPM",
  "region": "Medellín",
  "rate_type": "industrial",
  "updated_at": "2026-09-27T17:00:00Z"
}
```

Errores: `400 {"error":"at least provider or region must be provided"}`, `400` con id malformado.

### `GET /meters/:meterId/readings`

Serie paginada de lecturas. Query: `offset`, `limit`.

```bash
curl "http://localhost:8080/api/v1/meters/M-101/readings?limit=5"
# {"meter_id":"M-101","readings":[{…}],"pagination":{…}}
```

### `GET /meters/:meterId/estimated-consumption`

Estima costo de un periodo. Query requerida: `period_start`, `period_end` (RFC3339).

```bash
curl "http://localhost:8080/api/v1/meters/M-101/estimated-consumption?period_start=2026-09-01T00:00:00Z&period_end=2026-09-14T23:00:00Z"
```

```json
{
  "meter_id": "M-101",
  "period_start": "2026-09-01T00:00:00Z",
  "period_end": "2026-09-14T23:00:00Z",
  "total_kwh": 125.5,
  "price_per_kwh": 225.5,
  "total_cost": 28300.25,
  "currency": "COP"
}
```

`400 {"error":"invalid period_start"}` si falta o no es RFC3339. Requiere perfil + tarifa activa o devuelve `400` con `meter not configured`.

### `GET /meters/:meterId/optimization-insight`

Insight de optimización IA para un medidor. Query opcional: `from`, `to` (RFC3339; vacías = periodo completo).

```bash
curl "http://localhost:8080/api/v1/meters/M-101/optimization-insight?from=2026-09-01T00:00:00Z&to=2026-09-14T23:00:00Z"
```

```json
{
  "summary": "Pico nocturno recurrente 19-22h",
  "recommendations": ["Desplazar carga fuera de pico", "Revisar PF < 0.8"],
  "potential_savings_kwh": 18.4,
  "potential_savings_cost": 4149.2,
  "currency": "COP",
  "generated_at": "2026-09-14T12:00:00Z"
}
```

### `GET /meters/:meterId/billing-history`

Historial de facturación persistida del medidor.

```bash
curl http://localhost:8080/api/v1/meters/M-101/billing-history
# {"billing_history":[{"billing_id":1,"meter_id":"M-101","period_start":"…","period_end":"…","total_kwh":125.5,"price_per_kwh":225.5,"total_cost":28300.25,"currency":"COP","tariff_id":3}]}
```

## Anomalías

Tipos: `REAL_ANOMALY`, `EXPLAINABLE_ANOMALY`, `FALSE_POSITIVE`, `DATA_QUALITY`. Severidad: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`.

### `GET /anomalies`

Lista del último análisis completado. Query: `type`, `severity`, `meter_id` o `q`, `sort=severity|meter_id`, `offset`, `limit`.

```bash
curl "http://localhost:8080/api/v1/anomalies?type=REAL_ANOMALY&severity=HIGH&limit=10"
```

```json
{
  "anomalies": [{
    "id": "f310f1fb-f500-44c3-a6a6-2d4ef4a60a43",
    "meter_id": "M-109",
    "anomaly": "Incremento sostenido +110%",
    "type": "REAL_ANOMALY",
    "severity": "HIGH",
    "confidence": 0.91,
    "reason": "current 91 kWh vs baseline 43.2 kWh",
    "recommended_action": "Inspeccionar carga aguas abajo",
    "detected_at": "2026-09-14T12:00:00Z",
    "evidence": {
      "baseline_kwh": 43.2,
      "current_kwh": 91,
      "change_percent": 110.6,
      "voltage_v": 230.1,
      "current_a": 15.2,
      "power_factor": 0.88,
      "current_variation_percent": 12.5,
      "related_events": ["Mantenimiento programado 2026-09-10"],
      "abnormal_hours": [19, 20, 21, 22],
      "outlier_count": 2
    }
  }],
  "pagination": { "offset": 0, "limit": 10, "total": 1, "has_more": false }
}
```

### `GET /anomalies/:id`

Una anomalía con evidencia y acción. `:id` UUID.

```bash
curl http://localhost:8080/api/v1/anomalies/f310f1fb-f500-44c3-a6a6-2d4ef4a60a43
```

`400` UUID inválido; `404 {"error":"anomaly not found"}`.

## Análisis IA

El detector es local. `POST /ai/analyze` corre el pipeline y persiste `RUNNING → COMPLETED|FAILED`. Si Gemini falla, el análisis se completa igual con `insight_error`.

### `POST /ai/analyze`

```bash
curl -X POST http://localhost:8080/api/v1/ai/analyze
```

```json
{
  "id": "f310f1fb-f500-44c3-a6a6-2d4ef4a60a43",
  "status": "COMPLETED",
  "started_at": "2026-09-14T12:00:00Z",
  "completed_at": "2026-09-14T12:00:05Z",
  "anomalies": [],
  "insight": {
    "answer": "…",
    "explanation": "…",
    "suggested_questions": [
      { "question": "…?", "answer": "…" },
      { "question": "…?", "answer": "…" },
      { "question": "…?", "answer": "…" }
    ]
  }
}
```

Casos dorados del dataset demo: `M-109 REAL_ANOMALY/HIGH`, `M-104 EXPLAINABLE_ANOMALY/MEDIUM`, `M-106 FALSE_POSITIVE/LOW`, `M-112 DATA_QUALITY/HIGH`.

### `GET /ai/analysis/:id`

Análisis persistido. `:id` UUID. `404 {"error":"analysis not found"}`.

```bash
curl http://localhost:8080/api/v1/ai/analysis/f310f1fb-f500-44c3-a6a6-2d4ef4a60a43
```

## Tarifas

### `POST /tariffs`

Crea tarifa. `provider`, `rate_type`, `price_per_kwh > 0` y `valid_from` requeridos. `region` opcional, `currency` def. `COP`, `rate_type` def. `industrial`, `valid_to` opcional (> `valid_from`).

```bash
curl -X POST http://localhost:8080/api/v1/tariffs \
  -H 'Content-Type: application/json' \
  -d '{"provider":"bia","region":"CUND-EAST","rate_type":"industrial","price_per_kwh":690,"currency":"COP","valid_from":"2026-09-01T00:00:00Z"}'
```

`201`:

```json
{
  "tariff_id": 3,
  "provider": "bia",
  "region": "CUND-EAST",
  "rate_type": "industrial",
  "price_per_kwh": 690,
  "currency": "COP",
  "valid_from": "2026-09-01T00:00:00Z",
  "created_at": "2026-09-27T17:00:00Z"
}
```

`400` con el mensaje del dominio (`invalid tariff: provider is required`, …).

### `GET /tariffs`

Lista tarifas. Query: `provider`, `region`, `active=true|false`.

```bash
curl "http://localhost:8080/api/v1/tariffs?provider=bia&active=true"
# {"tariffs":[{…}]}
```

## Costos y facturación

Todas aceptan `from`/`to` opcionales (RFC3339). Vacías = periodo completo.

### `GET /billing/summary`

Resumen de costo del periodo.

```bash
curl "http://localhost:8080/api/v1/billing/summary?from=2026-09-01T00:00:00Z&to=2026-09-14T23:00:00Z"
```

```json
{
  "total_cost": 184050.75,
  "previous_total_cost": 172000,
  "change_percent": 7.0,
  "average_daily_cost": 13146.48,
  "top_meter_id": "M-109",
  "currency": "COP",
  "period_start": "2026-09-01T00:00:00Z",
  "period_end": "2026-09-14T23:00:00Z"
}
```

### `GET /billing/history`

Serie temporal de costo. Query: `granularity=daily|weekly|monthly` (def. `daily`).

```bash
curl "http://localhost:8080/api/v1/billing/history?granularity=weekly&from=2026-09-01T00:00:00Z&to=2026-09-14T23:00:00Z"
# {"series":[{"label":"2026-W36","date":"2026-09-01T00:00:00Z","total_cost":92025.4,"kwh":410.2}]}
```

### `GET /billing/meters`

Costo desagregado por medidor.

```bash
curl "http://localhost:8080/api/v1/billing/meters?from=2026-09-01T00:00:00Z&to=2026-09-14T23:00:00Z"
```

```json
{ "rows": [{
  "meter_id": "M-101",
  "kwh": 125.5,
  "tariff_name": "bia industrial",
  "daily_cost": 2021.45,
  "weekly_cost": 14150.2,
  "period_cost": 28300.25,
  "trend": 2.4,
  "calculated": true,
  "currency": "COP"
}]}
```

## Reportes

### `GET /reports/billing`

Descarga reporte. Query: `format=pdf|excel` (requerido), `from`/`to` opcionales, `meter_id` opcional (filtra un medidor).

```bash
curl -OJ "http://localhost:8080/api/v1/reports/billing?format=pdf&from=2026-09-01T00:00:00Z&to=2026-09-14T23:00:00Z"
curl -OJ "http://localhost:8080/api/v1/reports/billing?format=excel&meter_id=M-101"
```

Respuesta `200` con `Content-Disposition: attachment; filename="…"` y `Content-Type: application/pdf` o `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`. `format` inválido → `400`.

## SDK frontend

El cliente está en `frontend/src/api.ts`:

```ts
import { getDashboardSummary, getMeters, estimateMeterConsumptionCost } from './api'

const summary = await getDashboardSummary()
const meters = await getMeters()
const estimate = await estimateMeterConsumptionCost('M-101', '2026-09-01T00:00:00Z', '2026-09-14T23:00:00Z')
```

Más ejemplos: `docs/examples.md`. Arquitectura: `docs/architecture.md`.
