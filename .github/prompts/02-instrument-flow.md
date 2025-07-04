# AutoTracer: Instrumentación de Flujo Específico

## Objetivo
Instrumentar un flujo específico de lógica de negocio con OpenTelemetry basándose en el nivel de criticidad de cada función.

## Contexto
Este prompt se ejecuta para cada flujo identificado en la fase anterior. La variable `FLOW_NAME` contiene el nombre del flujo a instrumentar.

## Flujo a Instrumentar
**Flujo:** `${FLOW_NAME}`

## Tareas Específicas

### 1. Leer la Configuración del Flujo
- Carga `flows.json` para obtener los detalles del flujo `${FLOW_NAME}`
- Identifica todas las funciones y su nivel de criticidad
- Revisa la documentación en `docs/flows/${FLOW_NAME}.md`

### 2. Aplicar Instrumentación por Criticidad

#### Funciones CRÍTICAS
```go
// Instrumentación completa con contexto, errores y eventos
ctx, span := tracer.Start(ctx, "operation_name",
    trace.WithAttributes(
        attribute.String("business.entity.id", entityID),
        attribute.String("business.entity.type", entityType),
        attribute.String("business.operation", "critical_operation"),
    ),
)
defer span.End()

// Registrar eventos importantes
span.AddEvent("validation.started")
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    span.AddEvent("operation.failed", trace.WithAttributes(
        attribute.String("error.type", reflect.TypeOf(err).String()),
    ))
    return err
}
span.AddEvent("operation.completed")
```

#### Funciones ALTA
```go
// Instrumentación con contexto y manejo de errores
ctx, span := tracer.Start(ctx, "operation_name",
    trace.WithAttributes(
        attribute.String("business.entity.id", entityID),
        attribute.String("business.operation", "high_priority_operation"),
    ),
)
defer span.End()

if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

#### Funciones MEDIA
```go
// Instrumentación básica con contexto
ctx, span := tracer.Start(ctx, "operation_name")
defer span.End()

if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

#### Funciones BAJA
```go
// Instrumentación mínima (solo si está en un flujo crítico)
ctx, span := tracer.Start(ctx, "operation_name")
defer span.End()
```

### 3. Patrones de Instrumentación

#### Propagación de Contexto
```go
// Asegurar que el contexto se pase a todas las funciones
func ProcessPayment(ctx context.Context, payment *Payment) error {
    ctx, span := tracer.Start(ctx, "ProcessPayment")
    defer span.End()
    
    // Propagar contexto a subfunciones
    if err := validatePayment(ctx, payment); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return processPaymentInternal(ctx, payment)
}
```

#### Instrumentación de Loops y Batches
```go
// Para operaciones en lote
for i, item := range items {
    _, span := tracer.Start(ctx, "process_item",
        trace.WithAttributes(
            attribute.Int("batch.index", i),
            attribute.String("item.id", item.ID),
        ),
    )
    
    if err := processItem(ctx, item); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        span.End()
        continue
    }
    span.End()
}
```

### 4. Imports Necesarios
Asegúrate de agregar los imports necesarios al inicio de cada archivo:
```go
import (
    "context"
    "reflect"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

// Inicializar tracer (si no existe)
var tracer = otel.Tracer("autotracer")
```

## Instrucciones de Implementación

1. **Cargar configuración:** Lee `flows.json` y filtra por `${FLOW_NAME}`
2. **Priorizar funciones:** Comienza por las CRÍTICAS, luego ALTA, etc.
3. **Instrumentar sistemáticamente:** Aplica los patrones según criticidad
4. **Validar contexto:** Asegúrate de que el contexto se propague correctamente
5. **Optimizar rendimiento:** No sobre-instrumentar funciones BAJA

## Resultado Esperado
- Todas las funciones del flujo instrumentadas según su criticidad
- Contexto propagado correctamente a través del flujo
- Imports agregados donde sea necesario
- Nombres de spans descriptivos y consistentes

## Consideraciones Especiales
- **NO instrumentar** funciones que no estén en el flujo especificado
- **Respetar** los niveles de criticidad definidos
- **Mantener** la legibilidad del código
- **Evitar** instrumentación redundante
- **Considerar** el impacto en performance para funciones de alto volumen