# Documentación del frontend

## Propósito

El frontend presenta la información operativa del sistema: consumo, medidores, anomalías y análisis relacionados con cada instalación.

## Tecnologías

- React
- Vite
- TypeScript
- Recharts
- GSAP
- xlsx y jsPDF para exportación

## Estructura principal

```text
frontend/src/
  api.ts
  App.tsx
  components/
  features/
  hooks/
  utils/
```

## Capa de API

El archivo `frontend/src/api.ts` centraliza la comunicación con el backend. Encapsula:

- base URL configurable,
- versión de la API,
- manejo de errores,
- serialización JSON.

## Flujos principales

### Dashboard

El dashboard muestra el resumen general del sistema, incluidos:

- consumo agregado,
- cantidad de anomalías,
- severidad,
- confianza,
- último análisis disponible.

### Medidores

Cada medidor tiene:

- detalle técnico,
- historial de lecturas,
- métricas eléctricas,
- contexto de anomalías.

### Anomalías

La vista de anomalías permite:

- revisar cada hallazgo,
- ver evidencia,
- consultar la acción recomendada,
- entender la causa probable.

## Hooks y estado

Los hooks personalizados agrupan la lógica de consulta y la UI relacionada. Por ejemplo, `useEnergyWorkspace` centraliza:

- carga de datos,
- filtros,
- análisis actual,
- selección del medidor o anomalía,
- manejo de errores y estados de carga.

## Exportación

La aplicación puede exportar datos observables a formatos útiles para operación:

- Excel (`xlsx`)
- PDF (`jsPDF`)

La exportación se hace bajo demanda para evitar penalizar la carga inicial.

## Configuración

Se usa un valor de entorno para apuntar al backend:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Ejecución

```bash
cd frontend
pnpm install
pnpm dev
```

## Validación

```bash
pnpm run lint
pnpm run build
```

## Observaciones

- El frontend está orientado a una UX operativa y clara.
- El login es una demostración local, no una autenticación real.
- Las pantallas se usan para revisar resultados analíticos más que para una gestión administrativa compleja.

## Relación con el backend

La aplicación no calcula anomalías por sí misma: consume el resultado del detector y el insight de IA que expone el backend. Esto mantiene la lógica de negocio centralizada y una UI más simple y predecible.
