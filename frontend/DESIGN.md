# Design System — Bia

Documento de referencia de la paleta de colores de marca. Úsalo como fuente única de verdad para tokens de color en frontend (CSS variables, Tailwind config, etc).

## Paleta de color primaria

Distribución de uso sugerida dentro del producto (porcentaje de aparición).

| Nombre | Uso | HEX | RGB | CMYK |
|---|---|---|---|---|
| Black Neutral Bia | 40% | `#20222E` | R:32 G:34 B:46 | C:87 M:76 Y:52 K:68 |
| Azul Bia | 20% | `#001035` | R:0 G:16 B:53 | C:98 M:98 Y:54 K:41 |
| Morado Bia | 30% | `#472BEF` | R:71 G:43 B:239 | C:86 M:85 Y:0 K:0 |
| Neutral Bia | 5% | `#F0F1FA` | R:240 G:241 B:250 | C:7 M:5 Y:0 K:0 |
| Aqua Green | 5% | `#08DDBC` | R:8 G:221 B:188 | C:80 M:0 Y:47 K:0 |

## Paleta de color secundaria

Escalas tonales derivadas de los colores primarios (de más saturado a más claro), organizadas por familia.

### Morado / Neutro frío

| Tono | HEX |
|---|---|
| 100 | `#CAC0F7` |
| 200 | `#CAC0F7` |
| 300 | `#A08FF1` |
| 400 | `#6F56ED` |
| 500 | `#4929E9` |

### Morado

| Tono | HEX |
|---|---|
| 100 | `#513CCE` |
| 200 | `#675AE5` |
| 300 | `#7E7CED` |
| 400 | `#B9B9FC` |
| 500 | `#DDDDFA` |

### Azul

| Tono | HEX |
|---|---|
| 100 | `#00327C` |
| 200 | `#0048AD` |
| 300 | `#4087E7` |
| 400 | `#78A8ED` |
| 500 | `#C6DDF4` |

### Azul Marino / Navy

| Tono | HEX |
|---|---|
| 100 | `#0F133B` |
| 200 | `#171D51` |
| 300 | `#202767` |
| 400 | `#676D9B` |
| 500 | `#A3AACC` |

### Teal / Aqua

| Tono | HEX |
|---|---|
| 100 | `#4929E9` |
| 200 | `#6F56ED` |
| 300 | `#A08FF1` |
| 400 | `#CAC0F7` |
| 500 | `#CAC0F7` |

## Tipografía

Fuente detectada vía Network tab (request a `fonts.gstatic.com/s/inter/v20/...`).

| Propiedad | Valor |
|---|---|
| Familia | **Inter** |
| Proveedor | Google Fonts |
| Versión | v20 |
| Fallback stack | `Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif` |

### Import

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
```

```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap');
```

### Uso observado en el diseño

Por la captura del landing: headline en negrita fuerte (bold/extrabold ~700-800), subtítulos en peso regular/medium (~400-500), sobre fondo oscuro (`--color-black-neutral` / `--color-azul-bia`) con texto blanco/neutral claro (`--color-neutral-bia`).

| Uso | Peso sugerido | Tamaño aprox. |
|---|---|---|
| Headline principal | 800 (ExtraBold) | 40-48px |
| Subtítulo | 400 (Regular) | 16-18px |
| Botones / CTAs | 600 (SemiBold) | 16px |
| Texto de marcas/footer | 400 (Regular) | 14px |

## Tokens CSS (referencia rápida)

```css
:root {
  /* Tipografía */
  --font-family-base: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-weight-regular: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;
  --font-weight-extrabold: 800;

  /* Primarios */
  --color-black-neutral: #20222E;
  --color-azul-bia: #001035;
  --color-morado-bia: #472BEF;
  --color-neutral-bia: #F0F1FA;
  --color-aqua-green: #08DDBC;

  /* Escala morado/neutro frío */
  --color-purple-cool-100: #CAC0F7;
  --color-purple-cool-300: #A08FF1;
  --color-purple-cool-400: #6F56ED;
  --color-purple-cool-500: #4929E9;

  /* Escala morado */
  --color-purple-100: #513CCE;
  --color-purple-200: #675AE5;
  --color-purple-300: #7E7CED;
  --color-purple-400: #B9B9FC;
  --color-purple-500: #DDDDFA;

  /* Escala azul */
  --color-blue-100: #00327C;
  --color-blue-200: #0048AD;
  --color-blue-300: #4087E7;
  --color-blue-400: #78A8ED;
  --color-blue-500: #C6DDF4;

  /* Escala navy */
  --color-navy-100: #0F133B;
  --color-navy-200: #171D51;
  --color-navy-300: #202767;
  --color-navy-400: #676D9B;
  --color-navy-500: #A3AACC;
}
```

## Notas de uso

- **Black Neutral Bia** y **Azul Bia**: fondos oscuros, base del modo dark de la UI (dado el peso de uso del 40% y 20%).
- **Morado Bia**: color de acento principal / CTA (30% de uso).
- **Neutral Bia** y **Aqua Green**: acentos puntuales, indicadores de estado o resaltados (5% cada uno).
- Las escalas secundarias sirven para estados hover/active, fondos de tarjetas, bordes y variaciones de profundidad dentro de cada familia de color.