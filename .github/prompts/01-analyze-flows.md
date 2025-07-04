# AutoTracer: Análisis de Flujos de Negocio

## Objetivo
Analizar el código Go del repositorio para identificar y documentar los flujos de lógica de negocio, clasificando las funciones por nivel de criticidad.

## Contexto del Proyecto
Este es un proyecto Go que necesita instrumentación con OpenTelemetry. Tu trabajo es identificar los flujos de lógica de negocio (no infraestructura) para posterior instrumentación.

## Tareas Específicas

### 1. Identificar Flujos de Negocio
Busca y agrupa funciones que implementen lógica de negocio:

**Patrones a buscar:**
- Funciones que procesan transacciones o datos de negocio
- Validaciones de reglas de negocio
- Cálculos financieros o de dominio
- Integraciones con APIs externas para lógica de negocio
- Flujos de autenticación/autorización
- Operaciones CRUD con lógica de negocio compleja

**Ignorar:**
- Middleware HTTP genérico
- Utilidades de logging/debugging
- Helpers simples sin lógica de negocio
- Configuración y setup

### 2. Clasificar Criticidad
Para cada función, asigna un nivel de criticidad:

- **CRÍTICA**: Transacciones financieras, autenticación, operaciones que afectan datos críticos
- **ALTA**: Validaciones importantes, integraciones externas, cálculos de negocio
- **MEDIA**: Transformaciones de datos, validaciones simples
- **BAJA**: Operaciones de lectura, formateo, utilidades de negocio

### 3. Generar Documentación
Crea:

1. **Un archivo JSON** (`flows.json`) con la estructura:
```json
{
  "flows": [
    {
      "name": "PaymentProcessing",
      "description": "Procesamiento de pagos y transacciones",
      "files": ["internal/payment/processor.go", "internal/payment/validator.go"],
      "functions": [
        {
          "name": "ProcessPayment",
          "file": "internal/payment/processor.go",
          "line": 45,
          "criticality": "CRÍTICA",
          "description": "Procesa pagos con validaciones"
        }
      ],
      "entry_points": ["POST /api/payments"],
      "criticality": "CRÍTICA"
    }
  ]
}
```

2. **Archivos de documentación** (`docs/flows/`) para cada flujo:
   - `docs/flows/payment-processing.md`
   - `docs/flows/user-authentication.md`
   - etc.

## Instrucciones de Implementación

1. **Explorar el código:** Usa las herramientas para examinar la estructura del proyecto
2. **Identificar patrones:** Busca archivos y funciones que implementen lógica de negocio
3. **Crear flows.json:** Genera el archivo JSON con todos los flujos identificados
4. **Documentar flujos:** Crea archivos markdown detallados para cada flujo
5. **Priorizar por criticidad:** Asegúrate de que la clasificación sea precisa

## Resultado Esperado
Al final debes haber creado:
- `flows.json` (archivo principal para el pipeline)
- `docs/flows/*.md` (documentación detallada de cada flujo)
- Los flujos deben estar ordenados por criticidad (CRÍTICA primero)

## Consideraciones Especiales
- Enfócate en Go pero considera otros lenguajes si existen
- Presta atención a las dependencias entre flujos
- Identifica puntos de entrada (endpoints HTTP, funciones públicas)
- Considera el contexto de negocio del proyecto