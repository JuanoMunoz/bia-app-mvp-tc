# Backend

## Propósito

El backend procesa lecturas eléctricas y eventos operativos para detectar anomalías, persistir resultados y exponer una API REST para el frontend.

## Stack

- Go
- Gin
- PostgreSQL
- dotenv
- Google Gemini API (opcional)

## Estructura

```text
backend/
  cmd/
    api/
    seed/
  internal/
    ai/
    anomaly/
    dashboard/
    event/
    infrastructure/
    meter/
    reading/
  migrations/
  .env.example
  API.md
```

## Ejecución rápida

1. Copia la configuración de ejemplo:

```bash
cp .env.example .env
```

2. Completa las variables:

```env
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
PORT=8080
MIGRATIONS_SQL=migrations
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.0-flash
```

3. Ejecuta la API:

```bash
cd backend
go run ./cmd/api
```

4. Si quieres cargar el dataset inicial:

```bash
go run ./cmd/seed
```

## Funcionalidades principales

- carga de datos desde CSV
- detección matemática de anomalías
- almacenamiento en PostgreSQL
- generación de análisis con IA opcional
- endpoints REST versionados
- healthchecks y rate limiting

## Tests del backend

El backend tiene pruebas orientadas a validar el comportamiento de negocio, no solo la capa HTTP.

### Tipos de pruebas

- pruebas del detector de anomalías: validan cambios de consumo, outliers, perfiles horarios y calidad eléctrica
- pruebas del resumen por medidor: validan la severidad y reglas de escalado
- pruebas de HTTP: validan rutas, respuestas y health checks
- pruebas de carga de CSV: validan la ingestión esperada del dataset base

### Ejecutar la suite

```bash
cd backend
go test ./...
```

### Cobertura recomendada del core

Los casos críticos a cubrir son:

- consumo fuera del baseline
- eventos programados que no deben escalar
- anomalías eléctricas por calidad de datos
- picos puntuales y cambios horarios
- resumen del estado del medidor según severidad

## API

La referencia completa está en [API.md](API.md).

## Documentación relacionada

- [docs/architecture.md](../docs/architecture.md)
- [docs/integrations.md](../docs/integrations.md)
- [docs/backend.md](../docs/backend.md)

## Validación

```bash
go test ./...
go vet ./...
```
