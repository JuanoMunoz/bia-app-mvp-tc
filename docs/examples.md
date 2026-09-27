# Ejemplos de uso

## 1. Backend

### Health check

```bash
curl http://localhost:8080/api/v1/health
```

Respuesta esperada:

```json
{
  "status": "ok",
  "database": "connected"
}
```

### Obtener dashboard

```bash
curl http://localhost:8080/api/v1/dashboard/summary
```

### Ejecutar análisis

```bash
curl -X POST http://localhost:8080/api/v1/ai/analyze
```

### Consultar un análisis concreto

```bash
curl http://localhost:8080/api/v1/ai/analysis/<analysis-id>
```

## 2. Frontend

La app consume la API desde el cliente en `frontend/src/api.ts`.

Ejemplo de llamada:

```ts
const summary = await getDashboardSummary()
console.log(summary)
```

## 3. Seed y carga inicial

```bash
cd backend
go run ./cmd/seed
```

Esto permite llenar PostgreSQL con las series y eventos del CSV base.

## 4. Flujo completo de prueba

```bash
cd backend
go run ./cmd/seed

go run ./cmd/api
```

En otra terminal:

```bash
cd frontend
pnpm install
pnpm dev
```

## 5. Validación rápida

```bash
cd backend
go test ./...

cd ../frontend
pnpm run build
```

## 6. Variable opcional de Gemini

Cuando `GEMINI_API_KEY` está vacía, el sistema sigue funcionando con el análisis local. Cuando existe, el backend guarda también un insight generado por IA.

```env
GEMINI_API_KEY=tu_clave
GEMINI_MODEL=gemini-2.0-flash
```

## 7. Casos de uso del negocio

- revisar anomalías por medidor,
- ver la evidencia de consumo y calidad eléctrica,
- investigar eventos relacionados,
- ejecutar análisis planificado o manual,
- mostrar el resultado a un operador sin pedirle que entienda el cálculo interno.

## 8. Buenas prácticas

- no ejecutes el seed en entornos productivos sin intención,
- documenta cambios de endpoints,
- valida siempre la base de datos antes de probar flujo de UI,
- usa las rutas versionadas al desarrollar clientes nuevos.
