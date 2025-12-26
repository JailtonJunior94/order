#!/bin/bash

# Script para gerar tráfego de teste e validar envio de telemetria ao Coralogix
# Este script cria requisições de teste para gerar logs, métricas e traces

set -e

API_URL="${API_URL:-http://localhost:8001}"
REQUESTS="${REQUESTS:-10}"

echo "🚀 Gerando tráfego de teste para Order Service"
echo "================================================"
echo "API URL: $API_URL"
echo "Número de requisições: $REQUESTS"
echo ""

# Cores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Verificar se a API está respondendo
echo -n "Verificando se a API está disponível... "
if curl -s -f "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "Erro: API não está respondendo em $API_URL/health"
    echo "Inicie os serviços com: make start_docker"
    exit 1
fi

echo ""
echo "📊 Gerando tráfego de teste..."
echo ""

# Contador de requisições
SUCCESS=0
ERRORS=0

# Função para criar pedido
create_order() {
    local customer_id=$1
    local response
    local status_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/orders" \
        -H "Content-Type: application/json" \
        -H "X-Request-ID: test-$(uuidgen)" \
        -d "{
            \"customer_id\": \"customer-$customer_id\",
            \"items\": [
                {
                    \"product_id\": \"product-$((RANDOM % 100 + 1))\",
                    \"quantity\": $((RANDOM % 5 + 1)),
                    \"price\": $((RANDOM % 10000 + 1000))
                }
            ]
        }")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        return 0
    else
        return 1
    fi
}

# Gerar requisições
for i in $(seq 1 $REQUESTS); do
    echo -n "[$i/$REQUESTS] Criando pedido... "
    
    if create_order $i; then
        echo -e "${GREEN}✓${NC}"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "${RED}✗${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    # Pequeno delay entre requisições
    sleep 0.2
done

echo ""
echo "================================================"
echo "📈 Resultados:"
echo "  Total de requisições: $REQUESTS"
echo -e "  Sucesso: ${GREEN}$SUCCESS${NC}"
echo -e "  Erros: ${RED}$ERRORS${NC}"
echo ""

# Gerar alguns casos de erro (opcional)
echo "🔴 Gerando alguns casos de erro para teste..."
echo ""

# Requisição sem corpo
echo -n "  Teste 1: Requisição sem corpo... "
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/orders" \
    -H "Content-Type: application/json")
if [ "$STATUS" = "400" ] || [ "$STATUS" = "422" ]; then
    echo -e "${GREEN}✓${NC} (Status: $STATUS)"
else
    echo -e "${YELLOW}⚠${NC} (Status: $STATUS)"
fi

# Requisição com JSON inválido
echo -n "  Teste 2: JSON inválido... "
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/orders" \
    -H "Content-Type: application/json" \
    -d "{invalid json}")
if [ "$STATUS" = "400" ] || [ "$STATUS" = "422" ]; then
    echo -e "${GREEN}✓${NC} (Status: $STATUS)"
else
    echo -e "${YELLOW}⚠${NC} (Status: $STATUS)"
fi

echo ""
echo "================================================"
echo "✅ Tráfego de teste gerado com sucesso!"
echo ""
echo "🔍 Próximos passos:"
echo ""
echo "1. Verifique os logs do OpenTelemetry Collector:"
echo "   make logs-otel"
echo ""
echo "2. Acesse o Jaeger (local) para ver os traces:"
echo "   open http://localhost:16686"
echo ""
echo "3. Acesse o Coralogix Dashboard:"
echo "   open https://dashboard.coralogix.com"
echo ""
echo "4. No Coralogix, use os seguintes filtros:"
echo "   - Logs: cx.application.name:orders"
echo "   - Traces: Procure por 'POST /orders'"
echo "   - Metrics: Filtre por service.name"
echo ""
echo "⏱️  Aguarde 1-2 minutos para os dados aparecerem no Coralogix"
echo ""
