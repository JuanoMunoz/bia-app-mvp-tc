# Frontend

## Propósito

El frontend consume la API del backend para mostrar consumo, medidores, alertas y análisis operativos en una interfaz interactiva.

## Stack

- React
- Vite
- TypeScript
- Recharts
- GSAP
- xlsx / jsPDF

## Ejecución rápida

```bash
cd frontend
pnpm install
pnpm dev
```

## Configuración

Crea un archivo `.env` o usa la variable de entorno:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Estructura

```text
frontend/
  src/
    api.ts
    components/
    features/
    hooks/
    utils/
  public/
```

## Funcionalidades principales

- resumen del dashboard
- detalle de medidores
- lista de anomalías
- filtros y paginación
- exportación de datos
- transiciones visuales
- modo claro/oscuro

## Relación con el backend

El frontend no calcula análisis; consume los resultados expuestos por la API del backend.

## Documentación relacionada

- [../docs/architecture.md](../docs/architecture.md)
- [../docs/integrations.md](../docs/integrations.md)
- [../docs/frontend.md](../docs/frontend.md)

## Validación

```bash
pnpm run lint
pnpm run build
```
