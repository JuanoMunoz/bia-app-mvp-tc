# Documentación del backend

## Propósito

El backend expone la lógica de análisis de consumo, detecta anomalías y persiste la información necesaria para consultar el estado operativo del sistema.

## Tecnologías

- Go
- Gin
- PostgreSQL
- GoDotEnv
- Google Generative Language API (opcional)

## Estructura principal

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
```

## Componentes clave

### `cmd/api`

Arranca el servidor HTTP y conecta los servicios reales con la infraestructura.

### `internal/ai`

Contiene la lógica del análisis completo y la gestión del insight estructurado.

### `internal/anomaly`

Incluye el detector de anomalías, sus reglas y sus formatos de evidencia.

### `internal/infrastructure/db`

Se encarga de la persistencia en PostgreSQL: lecturas, eventos, anomalías y análisis.

### `internal/infrastructure/gemini`

Implementa el adaptador para enviar el contexto del análisis a Gemini y parsear la respuesta estructurada.

### `internal/infrastructure/httpServer`

Expone las rutas REST, DTOs y validaciones HTTP.

## Configuración

El backend lee variables desde `.env`.

Ejemplo:

```env
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
PORT=8080
MIGRATIONS_SQL=migrations/001_init.sql
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.0-flash
```

## Ejecución

### Iniciar API

```bash
cd backend
go run ./cmd/api
```

### Cargar datos iniciales

```bash
cd backend
go run ./cmd/seed
```

### Ejecutar tests

```bash
go test ./...
```

## Suite de pruebas del backend

El backend mantiene una estrategia de pruebas de negocio orientada al dominio. El objetivo no es solo comprobar que una ruta responde, sino asegurar que las reglas del detector y la lógica operativa se comportan como se espera.

### Pruebas del core del negocio

Estas pruebas validan la lógica central del sistema:

- `internal/anomaly/detector_test.go`: clasificación de anomalías, cambios de consumo, outliers y reducción de falsos positivos.
- `internal/meter/service_test.go`: resumen del estado del medidor según severidad y tipo de anomalía.

### Pruebas de integración y contrato

- `internal/infrastructure/csvloader/loader_test.go`: ingesta del dataset y casos esperados.
- `internal/infrastructure/httpServer/server_test.go`: rutas, status codes y payloads HTTP.

### Qué se valida con estas pruebas

- detección de cambios reales y explicables
- detección de calidad de datos
- anomalías de perfil horario
- picos puntuales aislados
- condiciones de parada programada y falsos positivos
- respuestas públicas del API y su estructura

## Endpoints principales

Consulta la referencia en [backend/API.md](../backend/API.md).

Los puntos más relevantes son:

- `/api/v1/dashboard/summary`
- `/api/v1/meters`
- `/api/v1/meters/:meterId`
- `/api/v1/anomalies`
- `/api/v1/ai/analyze`

## Observaciones de diseño

- El backend está construido con una separación clara entre dominio e infraestructura.
- La detección es determinista y no depende de IA.
- Gemini se usa como capa de explicación, no como motor primario del análisis.
- Todas las rutas están versionadas bajo `/api/v1`.

## Calidad y mantenimiento

Se recomienda:

- mantener el detector en funciones pequeñas y reproducibles,
- probar nuevos casos con datos reales o de regresión,
- conservar los prompts en archivos versionados,
- revisar el estado de la API y la BD con healthchecks.
