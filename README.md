# Bia Energy Monitoring

Aplicación para analizar consumo eléctrico, detectar anomalías operativas y explicar los resultados con evidencia técnica y, opcionalmente, con IA generativa.

## ¿Qué resuelve este proyecto?

El sistema toma lecturas de medidores y eventos operativos para:

- detectar cambios anómalos en el consumo,
- correlacionarlos con eventos del sistema,
- conservar evidencia técnica y recomendaciones,
- ofrecer un dashboard operativo en frontend,
- explicar el análisis con Gemini cuando hay clave configurada.

La idea principal es simple:

1. El cálculo de anomalías es determinista y explicable.
2. Gemini solo ayuda a contextualizar y explicar esos resultados.
3. La UI consume un backend versionado y estable.

## Vista rápida

```mermaid
flowchart LR
    CSV[CSV de lecturas y eventos] --> Seed[Seed / migración]
    Seed --> DB[(PostgreSQL)]
    DB --> Backend[Backend Go + Gin]
    Backend --> Detector[Detector matemático]
    Detector --> Anomalies[Anomalías y evidencia]
    Anomalies --> Gemini[Gemini opcional]
    Backend --> API[API REST /api/v1]
    API --> Frontend[React + Vite]
```

## Requisitos

Antes de arrancar el proyecto necesitas:

- Go 1.22+ o compatible
- Node.js 20+
- pnpm
- PostgreSQL disponible
- opcional: clave API de Gemini para activar la explicación generativa

## Estructura del repositorio

```text
.
├── backend/                # API Go, lógica de análisis y persistencia
├── frontend/               # app React + Vite
├── data/                   # CSV de ejemplo
├── docs/                   # documentación técnica y onboarding
├── README.md               # este archivo
├── AGENTS.MD               # instrucciones del proyecto
├── GEMINI.MD               # referencia de configuración Gemini
└── TASKS.md                # backlog / tareas
```

## Arranque rápido

### 1) Preparar entorno

Crea un archivo `.env` dentro de `backend` con:

```env
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
PORT=8080
MIGRATIONS_SQL=migrations/001_init.sql
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.0-flash
```

Y en `frontend`:

```env
VITE_API_BASE_URL=http://localhost:8080
```

### 2) Iniciar backend

```bash
cd backend
go run ./cmd/api
```

### 3) Cargar datos de ejemplo

```bash
cd backend
go run ./cmd/seed
```

### 4) Iniciar frontend

```bash
cd frontend
pnpm install
pnpm dev
```

Tras esto: 

- backend: http://localhost:8080
- frontend: http://localhost:5173

## Flujo de trabajo recomendado

### Para empezar a usar la app

1. Levanta PostgreSQL.
2. Genera la base con la migración automática al arrancar la API.
3. Ejecuta el seed para cargar `data/readings.csv` y `data/events.csv`.
4. Abre el frontend y revisa dashboard, medidores y anomalías.
5. Ejecuta `POST /api/v1/ai/analyze` para generar un análisis reforzado con contexto IA.

### Para desarrollar

- backend: mantiene la lógica de detección y API REST
- frontend: consume endpoints REST y renderiza dashboard y vistas
- docs: actualiza siempre que cambie el flujo de usuario o de datos

## API principal

La API usa rutas versionadas bajo `/api/v1`.

Algunas rutas importantes:

```bash
GET /api/v1/health
GET /api/v1/isalive
GET /api/v1/ping
GET /api/v1/dashboard/summary
GET /api/v1/meters
GET /api/v1/meters/:meterId
GET /api/v1/anomalies
POST /api/v1/ai/analyze
GET /api/v1/ai/analysis/:id
```

Más detalle: [backend/API.md](backend/API.md)

## Ejemplos de uso

### Consultar salud del sistema

```bash
curl http://localhost:8080/api/v1/health
```

### Ver resumen del dashboard

```bash
curl http://localhost:8080/api/v1/dashboard/summary
```

### Ejecutar análisis completo

```bash
curl -X POST http://localhost:8080/api/v1/ai/analyze
```

### Ver el último análisis persistido

```bash
curl http://localhost:8080/api/v1/ai/analysis/:id
```

## Documentación útil

- [docs/onboarding.md](docs/onboarding.md): guía de arranque para nuevos miembros
- [docs/examples.md](docs/examples.md): ejemplos de uso en consola y API
- [docs/architecture.md](docs/architecture.md): arquitectura del sistema
- [docs/integrations.md](docs/integrations.md): integraciones internas y externas
- [docs/backend.md](docs/backend.md): backend en detalle
- [docs/frontend.md](docs/frontend.md): frontend en detalle

## Validación

### Backend

```bash
cd backend
go test ./...
go vet ./...
```

### Frontend

```bash
cd frontend
pnpm run lint
pnpm run build
```

## Decisiones clave del diseño

- La detección de anomalías es local y reproducible.
- Gemini es opcional y complementario.
- La API está versionada para evitar rotura de clientes.
- El frontend no duplica la lógica de negocio: solo consume la API.
- Los datos se persisten en PostgreSQL para mantener trazabilidad.

## Contribución

Si quieres extender la solución:

- mantén la lógica de detección reproducible,
- documenta cambios de API y contratos,
- sigue separando dominio, infraestructura y presentación,
- actualiza los ejemplos cuando cambien rutas o payloads.

## Resumen

Este proyecto combina análisis operativo, evidencia basada en datos y apoyo explicativo con IA para dar una visión más útil de los consumos energéticos. El objetivo es que la alerta no sea un simple valor, sino un resultado con contexto, explicación y acción recomendada.
