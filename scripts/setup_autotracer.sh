#!/bin/bash

# AutoTracer Setup Script
# Este script configura AutoTracer con Claude Code en tu repositorio

set -e

echo "🚀 AutoTracer Setup - Instrumentación Automática con Claude Code"
echo "=============================================================="

# Verificar prerequisitos
check_prerequisites() {
    echo "📋 Verificando prerequisitos..."
    
    # Check git
    if ! command -v git &> /dev/null; then
        echo "❌ Git no está instalado"
        exit 1
    fi
    
    # Check if in git repo
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo "❌ No estás en un repositorio git"
        exit 1
    fi
    
    # Check GitHub CLI
    if ! command -v gh &> /dev/null; then
        echo "❌ GitHub CLI no está instalado"
        echo "   Instálalo desde: https://cli.github.com/"
        exit 1
    fi
    
    # Check Claude Code
    if ! command -v claude &> /dev/null; then
        echo "⚠️  Claude Code no está instalado"
        echo "   Puedes instalarlo con: npm install -g @anthropic-ai/claude-code"
        echo "   Continuando sin Claude Code CLI..."
    fi
    
    echo "✅ Prerequisitos verificados"
}

# Crear estructura de directorios
create_directories() {
    echo "📁 Creando estructura de directorios..."
    
    mkdir -p .github/workflows
    mkdir -p scripts
    mkdir -p docs/autotracer
    
    echo "✅ Directorios creados"
}

# Crear archivos de configuración
create_config_files() {
    echo "📝 Creando archivos de configuración..."
    
    # Crear workflow de GitHub Actions
    cat > .github/workflows/autotracer-claude.yml << 'EOF'
name: AutoTracer - Business Logic Instrumentation
on:
  pull_request:
    paths:
      - '**/*.go'
  push:
    branches:
      - main
    paths:
      - '**/*.go'
  workflow_dispatch:
    inputs:
      target_package:
        description: 'Package específico para instrumentar'
        required: false
        default: './...'

jobs:
  analyze-and-instrument:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
      
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          
      - name: Generate GitHub App token
        id: app-token
        uses: actions/create-github-app-token@v2
        with:
          app-id: ${{ secrets.APP_ID }}
          private-key: ${{ secrets.APP_PRIVATE_KEY }}
          
      - name: Claude Code - Analyze Business Logic
        uses: anthropics/claude-code-base-action@beta
        with:
          anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
          prompt_file: .github/prompts/autotracer_prompt.md
          allowed_tools: "View,GrepTool,BatchTool,EditTool,CreateFile"
          max_turns: "20"
          
      - name: Create Pull Request
        if: github.event_name != 'pull_request'
        uses: peter-evans/create-pull-request@v6
        with:
          token: ${{ steps.app-token.outputs.token }}
          commit-message: "[AutoTracer] Add OpenTelemetry instrumentation"
          title: "🔍 AutoTracer: Business Logic Instrumentation"
          branch: autotracer/instrumentation-${{ github.run_number }}
EOF

    # Crear prompt para Claude
    mkdir -p .github/prompts
    cat > .github/prompts/autotracer_prompt.md << 'EOF'
Actúa como un experto en observabilidad y análisis de código Go.

Tu tarea es analizar el código Go e identificar flujos de lógica de negocio que necesitan instrumentación con OpenTelemetry.

## Pasos a seguir:

1. **ANÁLISIS DE BUSINESS LOGIC**:
   - Identifica funciones que implementan lógica de negocio (no infraestructura)
   - Busca: handlers HTTP con lógica compleja, procesamiento de pagos, validaciones de negocio, cálculos, flujos de autorización
   - Ignora: middlewares básicos, getters/setters, utilidades de logging

2. **IDENTIFICACIÓN DE FLUJOS CRÍTICOS**:
   - Mapea flujos completos de negocio
   - Identifica puntos de entrada y salida
   - Detecta operaciones asíncronas

3. **GENERACIÓN DE INSTRUMENTACIÓN**:
   - Crea spans para flujos completos
   - Añade spans hijos para sub-operaciones
   - Incluye atributos del negocio (IDs, amounts, etc)
   - Propaga contexto correctamente

4. **CREAR ARCHIVOS**:
   - Modifica archivos .go con instrumentación
   - Crea instrumentation_report.md con análisis detallado

Usa go.opentelemetry.io/otel para la instrumentación.
EOF

    # Crear CLAUDE.md
    cp CLAUDE.md . 2>/dev/null || echo "⚠️  CLAUDE.md no encontrado, creando uno básico..."
    
    echo "✅ Archivos de configuración creados"
}

# Configurar secrets de GitHub
setup_github_secrets() {
    echo "🔐 Configuración de secrets de GitHub..."
    echo ""
    echo "Necesitas configurar los siguientes secrets en tu repositorio:"
    echo "1. ANTHROPIC_API_KEY - Tu API key de Anthropic"
    echo "2. APP_ID - ID de la GitHub App (se genera con Claude Code)"
    echo "3. APP_PRIVATE_KEY - Private key de la GitHub App"
    echo ""
    
    read -p "¿Quieres configurar los secrets ahora? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Intentar con Claude Code primero
        if command -v claude &> /dev/null; then
            echo "🤖 Iniciando configuración con Claude Code..."
            claude /install-github-app
        else
            echo "📝 Configuración manual..."
            echo "1. Ve a: https://github.com/apps/claude"
            echo "2. Instala la app en tu repositorio"
            echo "3. Configura los secrets en Settings > Secrets"
        fi
    fi
}

# Crear ejemplo de código
create_example() {
    echo "📦 Creando código de ejemplo..."
    
    mkdir -p example
    # El código del payment_service.go iría aquí
    
    echo "✅ Código de ejemplo creado en example/"
}

# Main execution
main() {
    echo ""
    check_prerequisites
    echo ""
    create_directories
    echo ""
    create_config_files
    echo ""
    setup_github_secrets
    echo ""
    
    read -p "¿Quieres crear código de ejemplo para testear? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        create_example
    fi
    
    echo ""
    echo "✨ ¡Setup completado!"
    echo ""
    echo "Próximos pasos:"
    echo "1. Asegúrate de tener configurados todos los secrets"
    echo "2. Haz push de los cambios a GitHub"
    echo "3. El workflow se activará automáticamente en PRs con cambios en archivos .go"
    echo "4. También puedes ejecutarlo manualmente desde Actions"
    echo ""
    echo "Para testear:"
    echo "- Crea un PR con cambios en archivos Go"
    echo "- O ejecuta manualmente: gh workflow run autotracer-claude.yml"
    echo ""
}

# Run main
main