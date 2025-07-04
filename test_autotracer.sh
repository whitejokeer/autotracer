#!/bin/bash

# Test rápido de AutoTracer con Claude Code

echo "🧪 Test de AutoTracer con Claude Code"
echo "===================================="

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Verificar configuración
check_config() {
    echo "🔍 Verificando configuración..."
    
    if [ -f .autotracer.yml ]; then
        echo -e "${GREEN}✓${NC} Archivo de configuración encontrado"
        echo "  Modo: $(grep 'mode:' .autotracer.yml | awk '{print $2}')"
    else
        echo -e "${YELLOW}!${NC} No se encontró .autotracer.yml, usando defaults"
    fi
    
    # Verificar secrets
    if gh secret list | grep -q "ANTHROPIC_API_KEY"; then
        echo -e "${GREEN}✓${NC} ANTHROPIC_API_KEY configurado"
    else
        echo -e "${RED}✗${NC} ANTHROPIC_API_KEY no configurado"
        echo "  Ejecuta: gh secret set ANTHROPIC_API_KEY"
    fi
}

# Opción 1: Test con código de ejemplo en PR
test_interactive_mode() {
    echo "📝 Creando PR de test (Modo Interactivo)..."
    
    git checkout -b test/autotracer-interactive-$(date +%s)
    
    # Crear código de ejemplo si no existe
    # Copiar proyecto de ejemplo ecommerce para la prueba
    if [ -d example/ecommerce ]; then
        cp -R example/ecommerce ./ecommerce
        echo -e "${GREEN}✓${NC} Proyecto ecommerce copiado exitosamente"
    elif [ -f example/payment_service.go ]; then
        # Fallback al archivo único si no existe el proyecto completo
        mkdir -p internal/payment
        cp example/payment_service.go internal/payment/
        echo -e "${YELLOW}!${NC} Usando archivo único payment_service.go"
    else
        echo -e "${RED}✗${NC} No se encontró ejemplo para test"
        return 1
    fi
    
    git add .
    git commit -m "test: add payment service for autotracer testing"
    git push -u origin HEAD
    
    # Crear PR
    pr_url=$(gh pr create \
        --title "Test: AutoTracer E-Commerce Instrumentation" \
        --body "🧪 Testing AutoTracer automated instrumentation on E-Commerce microservice
        
Este PR contiene un microservicio completo de e-commerce para testear AutoTracer.
        
**Estructura del proyecto:**
- 🛒 Order Service: Gestión de pedidos
- 💳 Payment Service: Procesamiento de pagos
- 📦 Inventory Service: Control de inventario
- 👤 User Service: Autenticación y perfiles
- 📧 Notification Service: Notificaciones async
        
**Qué esperar:**
- AutoTracer analizará todos los servicios
- Identificará flujos de negocio críticos
- Añadirá spans de OpenTelemetry
- Propagará contexto entre servicios
- Un comentario resumirá todos los cambios" \
        --base main \
        --draft)
        
    echo -e "${GREEN}✅ PR creado: $pr_url${NC}"
    echo "AutoTracer se ejecutará automáticamente en unos segundos."
    echo ""
    echo "Próximos pasos:"
    echo "1. Ve al PR: $pr_url"
    echo "2. Espera a que AutoTracer termine (~1-2 min)"
    echo "3. Revisa los cambios en 'Files changed'"
    echo "4. Lee el comentario con el resumen"
}

# Opción 2: Test modo automático
test_automatic_mode() {
    echo "🚀 Activando modo automático..."
    
    # Verificar si existe el workflow automático
    if [ ! -f .github/workflows/autotracer-auto.yml ]; then
        echo -e "${RED}✗${NC} Workflow automático no encontrado"
        echo "  Copia el workflow desde los artifacts"
        return 1
    fi
    
    # Crear branch de test
    git checkout -b test/autotracer-auto-$(date +%s)
    
    # Modificar config para modo auto
    if [ -f .autotracer.yml ]; then
        sed -i.bak 's/mode: .*/mode: automatic/' .autotracer.yml
    fi
    
    # Añadir código de test
    echo "package main" > test_auto.go
    echo "func CalculateTotalPrice(items []Item) float64 {" >> test_auto.go
    echo "    var total float64" >> test_auto.go
    echo "    for _, item := range items {" >> test_auto.go
    echo "        total += item.Price * float64(item.Quantity)" >> test_auto.go
    echo "    }" >> test_auto.go
    echo "    return total" >> test_auto.go
    echo "}" >> test_auto.go
    
    git add .
    git commit -m "test: trigger automatic instrumentation"
    git push -u origin HEAD
    
    echo -e "${GREEN}✅ Push completado${NC}"
    echo "El modo automático instrumentará el código y hará commit automáticamente."
    echo ""
    echo "Monitorea el progreso en:"
    echo "https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/actions"
}

# Opción 3: Ejecutar workflow manualmente
run_workflow() {
    echo "🎯 Ejecutando workflow manualmente..."
    
    # Listar archivos Go
    echo "Archivos Go encontrados:"
    find . -name "*.go" -type f | grep -v vendor | grep -v .git | head -10
    echo ""
    
    read -p "¿Qué package quieres analizar? (default: ./...): " package
    package=${package:-./...}
    
    read -p "¿Auto-commit los cambios? (y/n, default: n): " auto_commit
    auto_commit=${auto_commit:-n}
    
    if [[ $auto_commit =~ ^[Yy]$ ]]; then
        auto_commit="true"
    else
        auto_commit="false"
    fi
    
    echo "Ejecutando workflow con:"
    echo "  Package: $package"
    echo "  Auto-commit: $auto_commit"
    
    gh workflow run autotracer-claude.yml \
        -f target_package="$package" \
        -f auto_commit="$auto_commit"
    
    echo -e "${GREEN}✅ Workflow iniciado${NC}"
    
    # Esperar un momento y mostrar el status
    sleep 5
    echo ""
    echo "Estado actual:"
    gh run list --workflow=autotracer-claude.yml --limit 1
}

# Opción 4: Ver estado de workflows
check_status() {
    echo "📊 Estado de workflows de AutoTracer:"
    echo ""
    
    # Workflows en progreso
    echo -e "${YELLOW}En progreso:${NC}"
    gh run list --workflow=autotracer-claude.yml --status in_progress --limit 5 || echo "  Ninguno"
    
    echo ""
    echo -e "${GREEN}Completados recientemente:${NC}"
    gh run list --workflow=autotracer-claude.yml --status completed --limit 5
    
    # Si hay workflow automático
    if [ -f .github/workflows/autotracer-auto.yml ]; then
        echo ""
        echo -e "${YELLOW}Modo automático:${NC}"
        gh run list --workflow=autotracer-auto.yml --limit 3
    fi
}

# Opción 5: Ver cambios de instrumentación
view_changes() {
    echo "🔍 Cambios de instrumentación recientes:"
    echo ""
    
    # Buscar commits de AutoTracer
    echo "Commits de AutoTracer:"
    git log --oneline --grep="\[AutoTracer\]" -10 || echo "  No se encontraron commits"
    
    echo ""
    echo "PRs con instrumentación:"
    gh pr list --search "AutoTracer in:title,body" --state all --limit 5
    
    # Si hay archivos instrumentados
    if [ -f .autotracer/instrumented_files.txt ]; then
        echo ""
        echo "Archivos instrumentados:"
        cat .autotracer/instrumented_files.txt | head -10
    fi
}

# Opción 6: Configuración rápida
quick_setup() {
    echo "⚡ Configuración rápida de AutoTracer"
    echo ""
    
    # Crear configuración básica
    if [ ! -f .autotracer.yml ]; then
        echo "Creando .autotracer.yml..."
        cat > .autotracer.yml << 'EOF'
mode: interactive
instrumentation:
  level: 2
  include_patterns:
    - "Process.*"
    - "Handle.*"
    - ".*Payment.*"
validation:
  run_tests: true
  verify_build: true
EOF
        echo -e "${GREEN}✓${NC} Configuración creada"
    fi
    
    # Verificar workflows
    if [ ! -f .github/workflows/autotracer-claude.yml ]; then
        echo -e "${YELLOW}!${NC} Falta el workflow principal"
        echo "  Copia los archivos de workflow desde los artifacts"
    fi
    
    # Verificar Claude app
    echo ""
    read -p "¿Instalar GitHub App de Claude? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if command -v claude &> /dev/null; then
            claude /install-github-app
        else
            echo "Instala manualmente desde: https://github.com/apps/claude"
        fi
    fi
}

# Menú principal
clear
echo "🤖 AutoTracer Test Suite"
echo "========================"
echo ""

# Verificar estado inicial
check_config
echo ""

echo "¿Qué quieres hacer?"
echo ""
echo "  1) 📝 Test Modo Interactivo (crea PR de prueba)"
echo "  2) 🚀 Test Modo Automático (push con auto-commit)"
echo "  3) 🎯 Ejecutar workflow manualmente"
echo "  4) 📊 Ver estado de workflows"
echo "  5) 🔍 Ver cambios de instrumentación"
echo "  6) ⚡ Configuración rápida"
echo "  0) Salir"
echo ""

read -p "Selecciona una opción (0-6): " -n 1 -r
echo ""

case $REPLY in
    1) test_interactive_mode ;;
    2) test_automatic_mode ;;
    3) run_workflow ;;
    4) check_status ;;
    5) view_changes ;;
    6) quick_setup ;;
    0) echo "👋 ¡Hasta luego!" ;;
    *) echo -e "${RED}Opción inválida${NC}" ;;
esac

echo ""
echo "💡 Tips:"
echo "- Los logs completos están en GitHub Actions"
echo "- Personaliza el comportamiento en .autotracer.yml"
echo "- Usa 'gh run view' para ver detalles de una ejecución"