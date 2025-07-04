# AutoTracer con Claude Code 🤖🔍

> Instrumentación automática de lógica de negocio usando Claude Code en GitHub Actions

## ¿Qué es esto?

AutoTracer ahora usa el poder de Claude Code directamente desde GitHub Actions para:
- 🎯 Identificar automáticamente flujos de lógica de negocio en código Go
- 📊 Inyectar spans de OpenTelemetry directamente en tu código
- 🚀 Aplicar cambios automáticamente o dejarlos para revisión
- 🧠 Entender el contexto de negocio, no solo el código
- ⚡ Modo automático sin intervención humana (opcional)

## Modos de Operación

### 🔍 Modo Interactivo (Por defecto)
- Se activa en PRs con cambios en archivos `.go`
- Modifica los archivos directamente en el PR
- Te muestra los cambios para revisión
- Puedes aprobar o rechazar las modificaciones

### 🚀 Modo Automático
- Se activa en cada push (configurable)
- Aplica instrumentación sin intervención
- Valida que el código compile y pase tests
- Hace commit automático de los cambios
- Crea un issue si algo falla

### 🎯 Modo Manual
- Ejecutas el workflow cuando quieras
- Puedes especificar packages específicos
- Control total sobre el proceso

## Setup Rápido

### 1. Instalación automática (recomendado)

```bash
# Clona este repo
git clone <tu-repo>
cd <tu-repo>

# Ejecuta el script de setup
chmod +x scripts/setup_autotracer.sh
./scripts/setup_autotracer.sh
```

### 2. Instalación con Claude Code CLI

```bash
# Si tienes Claude Code instalado
claude /install-github-app
```

### 3. Configuración manual

1. Instala la [Claude GitHub App](https://github.com/apps/claude)
2. Añade estos secrets en tu repo:
   - `ANTHROPIC_API_KEY`: Tu API key de Anthropic
   - `APP_ID`: ID de la GitHub App
   - `APP_PRIVATE_KEY`: Private key de la GitHub App

## Uso

### 🔍 Modo Interactivo (En PRs)

1. **Crea un PR con cambios en archivos Go**
   ```bash
   git checkout -b feature/new-payment-method
   # ... hacer cambios en archivos .go ...
   git add . && git commit -m "Add new payment method"
   git push -u origin feature/new-payment-method
   gh pr create
   ```

2. **AutoTracer se ejecuta automáticamente**
   - Analiza los archivos Go modificados
   - Aplica instrumentación directamente
   - Comenta en el PR con un resumen

3. **Revisa los cambios**
   - Ve a la pestaña "Files changed" del PR
   - Los spans aparecerán como parte de tus cambios
   - Aprueba o modifica según necesites

### 🚀 Modo Automático (Sin revisión)

1. **Activa el workflow automático**
   ```bash
   # Copia el workflow automático
   cp .github/workflows/autotracer-auto.yml .github/workflows/
   ```

2. **Cada push será instrumentado automáticamente**
   - Sin intervención humana
   - Solo si pasa tests y validación
   - Commits automáticos con descripción detallada

### 🎯 Modo Manual

```bash
# Instrumentar todo el proyecto
gh workflow run autotracer-claude.yml

# Instrumentar un package específico
gh workflow run autotracer-claude.yml -f target_package=./internal/payments

# Sin auto-commit (deja cambios staged)
gh workflow run autotracer-claude.yml -f auto_commit=false
```

## Ejemplo de Output

### En Modo Interactivo (PR)

1. **Creas un PR con código nuevo**
2. **AutoTracer modifica tus archivos directamente**
3. **Ves los cambios en el diff del PR**:

```diff
  func ProcessPayment(ctx context.Context, order Order) error {
+     ctx, span := tracer.Start(ctx, "ProcessPayment",
+         trace.WithAttributes(
+             attribute.String("order.id", order.ID),
+             attribute.Float64("order.amount", order.Total),
+             attribute.String("payment.method", order.PaymentMethod),
+         ),
+     )
+     defer span.End()
+     
      if err := validateOrder(order); err != nil {
+         span.RecordError(err)
+         span.SetStatus(codes.Error, "Order validation failed")
          return err
      }
      // ... resto del código
  }
```

4. **AutoTracer comenta en el PR**:
   > ## 🔍 AutoTracer Instrumentation Report
   > 
   > ✅ Instrumentación aplicada directamente a los archivos.
   > 
   > ### 📊 Resumen
   > - Archivos Go modificados: 3
   > - Líneas añadidas: 45
   > - Spans añadidos: 8
   > - Business logic flows identificados: 3

### En Modo Automático

Los commits aparecen automáticamente:
```
[AutoTracer] Auto-instrumented 3 Go files

🤖 Automated instrumentation applied
📊 Business logic spans added
✅ Tests passed

Modified files:
internal/payment/processor.go
internal/order/validator.go
internal/inventory/manager.go

Co-authored-by: Claude <claude@anthropic.com>
```

## Personalización

### Modificar el análisis
Edita `.github/prompts/autotracer_prompt.md` para cambiar cómo Claude analiza tu código.

### Añadir patrones específicos
Edita `CLAUDE.md` para enseñar a Claude sobre tus patrones de negocio específicos.

### Cambiar herramientas disponibles
En el workflow, modifica `allowed_tools`:
```yaml
allowed_tools: "View,GrepTool,BatchTool,EditTool,CreateFile,Bash(go:*)"
```

## Ventajas sobre el enfoque tradicional

| Característica | AutoTracer Original | AutoTracer + Claude Code |
|----------------|-------------------|-------------------------|
| Análisis semántico | LLM local limitado | Claude con contexto completo |
| Comprensión de negocio | Heurísticas básicas | Comprensión profunda |
| Mantenimiento | Requiere actualización de reglas | Se adapta automáticamente |
| Integración CI/CD | Script separado | Nativo en GitHub Actions |
| Costo | Infraestructura propia | Pay-per-use |

## Testing

### Con código de ejemplo
```bash
# Copia el ejemplo
cp example/payment_service.go internal/

# Haz commit y push
git add .
git commit -m "test: add payment service"
git push

# El workflow se activará automáticamente
```

### Verificar resultados
1. Ve a Actions en GitHub
2. Busca el workflow "AutoTracer"
3. Revisa los logs de Claude Code
4. Verifica el PR creado

## Troubleshooting

### "Permission denied" al crear PR
- Verifica que la GitHub App tenga permisos de `contents: write`
- Asegúrate que el token se genera correctamente

### Claude no encuentra lógica de negocio
- Revisa que tu código tenga patrones de negocio claros
- Actualiza `CLAUDE.md` con ejemplos de tu dominio

### Workflow falla con "API key invalid"
- Verifica que `ANTHROPIC_API_KEY` esté configurado correctamente
- Asegúrate de usar la API key correcta (no la de Claude.ai)

## Costos

- **Claude Code API**: ~$0.01-0.05 por análisis (depende del tamaño del código)
- **GitHub Actions**: Gratis para repos públicos, minutos incluidos para privados
- **Estimado mensual**: $5-20 para un equipo activo

## Roadmap

- [ ] Soporte para otros lenguajes (Python, TypeScript)
- [ ] Integración con Datadog/New Relic APM
- [ ] Análisis incremental (solo archivos cambiados)
- [ ] Dashboard de métricas de instrumentación
- [ ] Auto-rollback si los tests fallan

## Contribuir

¡PRs bienvenidos! Especialmente para:
- Mejores prompts para dominios específicos
- Ejemplos de diferentes tipos de aplicaciones
- Integraciones con otras herramientas de observabilidad

## Licencia

MIT - Úsalo como quieras 🚀