#!/bin/bash

# Script de verificação da configuração do Coralogix
# Este script valida se a configuração está correta antes de iniciar os serviços

set -e

DEPLOYMENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$DEPLOYMENT_DIR/.env"

echo "🔍 Verificando configuração do Coralogix..."
echo ""

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para verificar variável
check_env_var() {
    local var_name=$1
    local var_value=$2
    local is_required=$3
    
    if [ -z "$var_value" ] || [ "$var_value" == "your-private-key-here" ]; then
        if [ "$is_required" == "true" ]; then
            echo -e "${RED}✗${NC} $var_name não configurada"
            return 1
        else
            echo -e "${YELLOW}⚠${NC} $var_name não configurada (opcional)"
            return 0
        fi
    else
        echo -e "${GREEN}✓${NC} $var_name configurada"
        return 0
    fi
}

# Verificar se o arquivo .env existe
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}✗${NC} Arquivo .env não encontrado em $ENV_FILE"
    echo ""
    echo "Execute: cp $DEPLOYMENT_DIR/.env.example $ENV_FILE"
    echo "E configure as variáveis necessárias"
    exit 1
fi

echo "✓ Arquivo .env encontrado"
echo ""

# Carregar variáveis do .env
export $(cat "$ENV_FILE" | grep -v '^#' | xargs)

# Verificar variáveis obrigatórias
echo "📋 Verificando variáveis obrigatórias:"
echo ""

ERRORS=0

check_env_var "CORALOGIX_PRIVATE_KEY" "$CORALOGIX_PRIVATE_KEY" "true" || ERRORS=$((ERRORS+1))
check_env_var "CORALOGIX_APP_NAME" "$CORALOGIX_APP_NAME" "false"
check_env_var "CORALOGIX_DOMAIN" "$CORALOGIX_DOMAIN" "true" || ERRORS=$((ERRORS+1))
check_env_var "ENVIRONMENT" "$ENVIRONMENT" "false"

echo ""

# Verificar formato da private key
if [ ! -z "$CORALOGIX_PRIVATE_KEY" ] && [ "$CORALOGIX_PRIVATE_KEY" != "your-private-key-here" ]; then
    KEY_LENGTH=${#CORALOGIX_PRIVATE_KEY}
    if [ $KEY_LENGTH -lt 32 ]; then
        echo -e "${YELLOW}⚠${NC} A private key parece muito curta (tamanho: $KEY_LENGTH)"
        echo "   Verifique se você copiou a chave completa"
        ERRORS=$((ERRORS+1))
    else
        echo -e "${GREEN}✓${NC} Private key tem tamanho adequado"
    fi
fi

# Verificar domínio válido
if [ ! -z "$CORALOGIX_DOMAIN" ]; then
    case "$CORALOGIX_DOMAIN" in
        "ingress.coralogix.com"|"ingress.eu2.coralogix.com"|"ingress.coralogix.us"|"ingress.cx498.coralogix.com"|"ingress.coralogixsg.com"|"ingress.app.coralogix.in")
            echo -e "${GREEN}✓${NC} Domínio válido: $CORALOGIX_DOMAIN"
            ;;
        *)
            echo -e "${YELLOW}⚠${NC} Domínio não reconhecido: $CORALOGIX_DOMAIN"
            echo "   Domínios válidos:"
            echo "   - ingress.coralogix.com (EU1)"
            echo "   - ingress.eu2.coralogix.com (EU2)"
            echo "   - ingress.coralogix.us (US1)"
            echo "   - ingress.cx498.coralogix.com (US2)"
            echo "   - ingress.coralogixsg.com (AP1)"
            echo "   - ingress.app.coralogix.in (AP2)"
            ;;
    esac
fi

echo ""

# Verificar conectividade com Coralogix
if [ ! -z "$CORALOGIX_DOMAIN" ]; then
    echo "🌐 Testando conectividade com Coralogix..."
    
    if command -v nc &> /dev/null; then
        if nc -zv -w3 "$CORALOGIX_DOMAIN" 443 2>&1 | grep -q succeeded; then
            echo -e "${GREEN}✓${NC} Conectividade com $CORALOGIX_DOMAIN:443 OK"
        else
            echo -e "${RED}✗${NC} Não foi possível conectar com $CORALOGIX_DOMAIN:443"
            echo "   Verifique sua conexão de rede e firewall"
            ERRORS=$((ERRORS+1))
        fi
    elif command -v telnet &> /dev/null; then
        if timeout 3 telnet "$CORALOGIX_DOMAIN" 443 2>&1 | grep -q Connected; then
            echo -e "${GREEN}✓${NC} Conectividade com $CORALOGIX_DOMAIN:443 OK"
        else
            echo -e "${RED}✗${NC} Não foi possível conectar com $CORALOGIX_DOMAIN:443"
            echo "   Verifique sua conexão de rede e firewall"
            ERRORS=$((ERRORS+1))
        fi
    else
        echo -e "${YELLOW}⚠${NC} nc ou telnet não disponível, pulando teste de conectividade"
    fi
fi

echo ""

# Verificar arquivos de configuração
echo "📁 Verificando arquivos de configuração:"
echo ""

CONFIG_FILES=(
    "$DEPLOYMENT_DIR/observability/collector/otel-collector.yml"
    "$DEPLOYMENT_DIR/docker-compose.yml"
)

for file in "${CONFIG_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $(basename $file) existe"
    else
        echo -e "${RED}✗${NC} $(basename $file) não encontrado"
        ERRORS=$((ERRORS+1))
    fi
done

echo ""

# Resultado final
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ Todas as verificações passaram!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Você pode iniciar os serviços com:"
    echo "  cd $DEPLOYMENT_DIR"
    echo "  docker compose up -d"
    echo ""
    echo "Para verificar os logs:"
    echo "  docker compose logs -f otel-collector"
    echo ""
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗ $ERRORS erro(s) encontrado(s)${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo "Corrija os erros acima antes de iniciar os serviços"
    echo ""
    exit 1
fi
