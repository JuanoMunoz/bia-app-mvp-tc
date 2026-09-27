# Onboarding

## Objetivo

Esta guía sirve para que cualquier persona pueda arrancar el proyecto con el mínimo contexto y entender qué hace cada parte del repositorio.

## 1. Qué es este proyecto

Es una aplicación de monitorización energética que:

- lee lecturas de medidores,
- combina esos datos con eventos operativos,
- detecta anomalías de consumo,
- persiste el resultado,
- expone un dashboard,
- y puede explicar el análisis con Gemini.

## 2. Cómo está organizado

```text
.
├── backend/         # lógica de negocio, API y persistencia
├── frontend/        # interfaz web del usuario
├── data/            # datasets base para pruebas
├── docs/            # documentación técnica
└── README.md        # vista general del proyecto
```

## 3. Requisitos de entorno

Necesitas:

- Go instalado
- Node.js y pnpm
- PostgreSQL corriendo
- opcional: clave de Gemini

## 4. Primero: base de datos

Crea una base PostgreSQL y configura `DATABASE_URL` en `backend/.env`.

Ejemplo:

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/bia_energy
PORT=8080
MIGRATIONS_SQL=migrations/001_init.sql
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.0-flash
```

## 5. Arrancar backend

```bash
cd backend
go run ./cmd/api
```

La API se levantará con la migración automática si la carpeta `migrations` está presente.

## 6. Cargar datos de ejemplo

```bash
cd backend
go run ./cmd/seed
```

Esto carga los CSV de `data/` en PostgreSQL.

## 7. Arrancar frontend

```bash
cd frontend
pnpm install
pnpm dev
```

La app suele quedar en:

- http://localhost:5173

## 8. Flujo natural del proyecto

### Caso principal

1. El backend carga datos y eventos.
2. El detector calcula anomalías desde la serie de consumo.
3. Se persiste la respuesta con evidencia.
4. El frontend consulta el dashboard y los detalles.
5. Si un usuario ejecuta un análisis con IA, Gemini recibe un resumen del contexto.

## 9. Rutas clave

```bash
GET /api/v1/dashboard/summary
GET /api/v1/meters
GET /api/v1/anomalies
POST /api/v1/ai/analyze
```

## 10. Dónde mirar primero

Si estás empezando, revisa estos archivos en este orden:

1. [README.md](../README.md)
2. [backend/API.md](../backend/API.md)
3. [backend/cmd/api/main.go](../backend/cmd/api/main.go)
4. [backend/internal/anomaly/detector.go](../backend/internal/anomaly/detector.go)
5. [backend/internal/infrastructure/gemini/client.go](../backend/internal/infrastructure/gemini/client.go)
6. [frontend/src/api.ts](../frontend/src/api.ts)

## 11. Errores comunes

- no existe `DATABASE_URL`
- no se cargó el seed antes de consultar el dashboard
- no hay clave de Gemini y se espera respuesta de IA
- el puerto 8080 ya está ocupado

## 12. Recomendación de desarrollo

- prueba cambios en detector con datos reales de `data/`
- usa `go test ./...` antes de entregar cambios en backend
- usa `pnpm run build` cuando modifiques frontend
- si cambias la API, actualiza también la documentación y ejemplos

## 13. Resumen

En pocas palabras: el proyecto combina datos operativos + lógica matemática + API REST + UI para generar una visión útil del consumo eléctrico y sus anomalías.
