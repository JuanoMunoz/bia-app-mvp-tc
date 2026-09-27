# Integraciones del backend

Este backend conecta varias piezas para transformar datos de consumo en análisis de anomalías y en respuestas útiles para el frontend.

## 1) Flujo general

El flujo principal es:

1. Cargar lecturas y eventos desde CSV o base de datos
2. Detectar anomalías con el algoritmo matemático
3. Guardar el resultado de la ejecución en PostgreSQL
4. Opcionalmente, enviar el contexto al modelo Gemini para obtener un insight explicativo
5. Exponer los resultados por HTTP a través de la API
6. Consumo desde el frontend para mostrar dashboard, medidores y anomalías

## 2) Integración con fuentes de datos

### Carga de CSV

El proyecto incluye un loader para leer archivos CSV con datos de medidores y eventos:

- `backend/internal/infrastructure/csvloader/loader.go`
- `backend/cmd/seed/main.go`

La semilla inicial usa:

- `data/readings.csv`
- `data/events.csv`

El loader transforma esos archivos en estructuras internas y luego los guarda en PostgreSQL mediante el repositorio de base de datos.

### Base de datos PostgreSQL

La integración con PostgreSQL está en:

- `backend/internal/infrastructure/db/postgres.go`
- `backend/internal/infrastructure/db/repositories.go`

La capa de persistencia se encarga de:

- guardar análisis en ejecución
- guardar resultados completos
- guardar anomalías asociadas
- recuperar el último análisis
- consultar un análisis por id
- recuperar lecturas y eventos para análisis y dashboard

## 3) Integración con la lógica de análisis

La orquestación principal está en:

- `backend/internal/ai/service.go`

La operación `Analyze(ctx)` hace esto:

```go
readings, err := service.source.AllReadings(ctx)
events, err := service.source.AllEvents(ctx)
result.Anomalies = service.detector.Detect(readings, events)
```

Es decir, primero obtiene todo el dataset, luego ejecuta el detector matemático y finalmente genera un resultado estructurado con:

- listado de anomalías
- duración del análisis
- estado (`RUNNING`, `COMPLETED`, `FAILED`)
- insight opcional generado por IA

## 4) Integración con Gemini

La integración externa con Gemini está en:

- `backend/internal/infrastructure/gemini/client.go`

Cuando `GEMINI_API_KEY` está presente, el backend crea un cliente `Client` y lo usa como `InsightAnalyzer`.

### Cómo funciona

1. Se construye un prompt con:
   - número total de lecturas
   - número total de medidores
   - estadísticas por medidor
   - eventos asociados
   - anomalías detectadas
2. Se envía a la API de Generative Language de Google
3. El modelo responde con JSON estructurado
4. El backend valida que el JSON tenga:
   - `answer`
   - `explanation`
   - `suggestedQuestions`
5. El insight se guarda junto al análisis

La petición se hace a:

```go
https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}
```

La configuración se toma desde variables de entorno:

- `GEMINI_API_KEY`
- `GEMINI_MODEL`

Si no hay clave, la IA queda desactivada y el sistema sólo usa el detector local.

## 5) Integración con la API HTTP

La API se sirve con Gin y está definida en:

- `backend/internal/infrastructure/httpServer/server.go`

Rutas clave:

- `POST /ai/analyze`
- `GET /ai/analysis/:id`
- `GET /meters`
- `GET /meters/:meterId`
- `GET /anomalies`
- `GET /dashboard/summary`

Además, el servidor usa middleware para:

- rate limiting
- health checks
- validación de rutas

## 6) Integración con el frontend

El frontend consume la API del backend para visualizar:

- resumen del dashboard
- medidores y series temporales
- detalles por medidor
- anomalías con evidencia y recomendaciones
- último análisis generado

La capa de cliente del frontend está en:

- `frontend/src/api.ts`

El patrón es simple: el frontend llama a endpoints REST del backend y renderiza la respuesta en la UI.

## 7) Integración de eventos y contexto operativos

La detección de anomalías no solo mira consumo; también incorpora eventos operativos, por ejemplo:

- `SCHEDULED_OUTAGE`
- `OPERATIONAL_CHANGE`
- otros eventos registrados por equipo

Cuando un cambio de consumo coincide con un evento de operación, el sistema puede:

- considerarlo una anomalía explicable
- reducir su severidad
- devolver una recomendación distinta
- evitar falsos positivos

Esto ayuda a que la detección sea más útil para personal de operación y mantenimiento.

## 8) Integración completa en un vistazo

El flujo end-to-end es:

```text
CSV / DB --> Repositorios --> Detector matemático --> Anomalías
                           \
                            --> AI (Gemini) --> Insight estructurado
                                   \
                                    --> PostgreSQL
                                            --> API HTTP
                                                    --> Frontend
```

La arquitectura se mantiene modular: cada capa tiene una responsabilidad clara y puede cambiarse sin romper el resto del sistema.

## 9) Resumen

Las integraciones principales son:

- almacenamiento de datos en PostgreSQL
- lectura de CSV para inicialización
- detección matemática local
- mejora contextual con Gemini
- exposición de resultados vía API REST
- consumo del frontend para dashboards y análisis

Esto convierte el backend en un sistema de análisis operativo con capacidad de detectar anomalías, guardar evidencia y explicar resultados en lenguaje natural para usuarios o equipos de operación.
