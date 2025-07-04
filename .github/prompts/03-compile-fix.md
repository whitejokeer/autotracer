# AutoTracer: Corrección de Errores de Compilación

## Objetivo
Identificar y corregir errores de compilación introducidos durante la instrumentación con OpenTelemetry.

## Contexto
Este prompt se ejecuta cuando la compilación falla después de la instrumentación. Tu trabajo es identificar y corregir todos los errores hasta que el código compile correctamente.

## Errores Comunes a Corregir

### 1. Imports Faltantes
```go
// Error: undefined: tracer
// Solución: Agregar imports
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("autotracer")
```

### 2. Contexto No Propagado
```go
// Error: function signature mismatch
// Antes:
func ProcessPayment(payment *Payment) error

// Después:
func ProcessPayment(ctx context.Context, payment *Payment) error
```

### 3. Variables No Utilizadas
```go
// Error: span declared but not used
// Solución: Usar _ si el span no se necesita
_, span := tracer.Start(ctx, "operation")
defer span.End()
```

### 4. Tipos Incorrectos
```go
// Error: cannot use string as attribute.Value
// Solución: Usar las funciones correctas de attribute
attribute.String("key", stringValue)
attribute.Int("key", intValue)
attribute.Bool("key", boolValue)
```

## Estrategia de Corrección

### 1. Análisis de Errores
1. **Ejecutar compilación:** `go build ./...` o `go vet ./...`
2. **Categorizar errores:**
   - Import errors
   - Type errors
   - Unused variable errors
   - Function signature mismatches
   - Package errors

### 2. Corrección Sistemática

#### Imports y Declaraciones
```go
// Verificar que todos los archivos tengan los imports necesarios
import (
    "context"
    "fmt"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

// Verificar que el tracer esté declarado
var tracer = otel.Tracer("autotracer")
```

#### Firmas de Funciones
```go
// Asegurar que todas las funciones que usan spans tengan contexto
func OriginalFunction(param1 string, param2 int) error {
    // Cambiar a:
    func OriginalFunction(ctx context.Context, param1 string, param2 int) error {
        ctx, span := tracer.Start(ctx, "OriginalFunction")
        defer span.End()
        // ... resto del código
    }
}
```

#### Llamadas a Funciones
```go
// Actualizar todas las llamadas para incluir contexto
// Antes:
result := someFunction(param1, param2)

// Después:
result := someFunction(ctx, param1, param2)
```

### 3. Casos Especiales

#### Funciones HTTP Handlers
```go
// Extraer contexto de la request
func (h *Handler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx, span := tracer.Start(ctx, "ProcessPayment")
    defer span.End()
    
    // ... resto del código
}
```

#### Goroutines
```go
// Propagar contexto a goroutines
go func() {
    ctx, span := tracer.Start(ctx, "async_operation")
    defer span.End()
    
    // ... operación asíncrona
}()
```

#### Interfaces
```go
// Actualizar interfaces si es necesario
type PaymentProcessor interface {
    ProcessPayment(ctx context.Context, payment *Payment) error
}
```

## Instrucciones de Implementación

1. **Ejecutar build:** Ejecuta `go build ./...` para identificar errores
2. **Analizar output:** Categoriza los errores por tipo
3. **Corregir sistemáticamente:**
   - Imports faltantes
   - Firmas de funciones
   - Llamadas a funciones
   - Variables no utilizadas
   - Problemas de tipos
4. **Verificar:** Ejecuta `go build ./...` nuevamente
5. **Repetir:** Hasta que no haya errores

## Comandos de Verificación

```bash
# Compilación completa
go build ./...

# Verificación de código
go vet ./...

# Ejecución de tests (si existen)
go test ./...

# Verificación de formato
go fmt ./...
```

## Resultado Esperado
- Código que compile sin errores
- Todos los imports necesarios agregados
- Funciones con firmas correctas
- Contexto propagado apropiadamente
- Tests pasando (si existen)

## Consideraciones Especiales
- **No cambiar** la lógica de negocio original
- **Mantener** la funcionalidad existente
- **Minimizar** cambios en APIs públicas
- **Documentar** cambios significativos en signatures
- **Preservar** tests existentes

## Logging de Correcciones
Mantén un registro de los cambios realizados:
```
- Agregado contexto a función ProcessPayment en payment.go:45
- Corregido import faltante en validator.go
- Actualizada interfaz PaymentProcessor para incluir contexto
```