# 🚀 Quick Start - Coralogix Integration

## Pré-requisitos

- Docker e Docker Compose instalados
- Conta no Coralogix (https://coralogix.com)
- Make instalado (opcional, mas recomendado)

## Setup em 5 Passos

### 1️⃣ Configure o Coralogix

```bash
# Configurar arquivo de ambiente
make setup-coralogix

# OU manualmente
cd deployment
cp .env.example .env
```

Edite `deployment/.env` e adicione sua **Private Key** do Coralogix:

```env
CORALOGIX_PRIVATE_KEY=sua-chave-aqui
CORALOGIX_APP_NAME=orders
CORALOGIX_DOMAIN=ingress.eu2.coralogix.com  # Ajuste para sua região
ENVIRONMENT=development
```

### 2️⃣ Obtenha sua Private Key

1. Acesse: https://dashboard.coralogix.com/settings/send-your-data
2. Copie a **Private Key**
3. Cole no arquivo `.env`

### 3️⃣ Verifique a Configuração

```bash
# Execute o script de verificação
./deployment/verify-coralogix-config.sh
```

Se tudo estiver OK, você verá: ✅ **Todas as verificações passaram!**

### 4️⃣ Inicie os Serviços

```bash
# Iniciar todos os serviços
make start_docker

# OU
cd deployment
docker compose up -d
```

### 5️⃣ Verifique os Logs

```bash
# Ver logs do OpenTelemetry Collector
make logs-otel

# Ver logs de todos os serviços
make logs-all

# Verificar saúde dos serviços
make health-check
```

## 📊 Acessar Dashboards

### Coralogix (Produção)
- **Dashboard**: https://dashboard.coralogix.com
- **Logs**: Explore → Logs → Filtrar por `cx.application.name:orders`
- **Traces**: Explore → Tracing
- **Metrics**: Dashboards → Metrics

### Ferramentas Locais (Desenvolvimento)
- **Jaeger** (Traces): http://localhost:16686
- **Prometheus** (Metrics): http://localhost:9090
- **Grafana** (Dashboards): http://localhost:3000
- **Order API**: http://localhost:8000/health

## ✅ Validar Envio de Dados

### 1. Verifique os logs do collector
```bash
make logs-otel
```

Procure por linhas como:
```
Exporter	{"kind": "exporter", "data_type": "traces", "name": "otlp/coralogix-traces"}
```

### 2. Gere tráfego de teste
```bash
# Criar um pedido (vai gerar traces)
curl -X POST http://localhost:8000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "123",
    "items": [{"product_id": "456", "quantity": 2}]
  }'
```

### 3. Verifique no Coralogix
- Aguarde 1-2 minutos para os dados aparecerem
- Acesse o dashboard do Coralogix
- Filtre por `cx.application.name:orders`

## 🔍 Filtros Úteis no Coralogix

```
# Ver logs de um serviço específico
cx.application.name:orders AND cx.subsystem.name:order-api

# Ver todos os erros
cx.application.name:orders AND level:error

# Ver requisições lentas (> 1s)
cx.application.name:orders AND http.duration:>1000

# Ver eventos de um consumidor Kafka
cx.application.name:orders AND cx.subsystem.name:order-consumer
```

## 🛠️ Comandos Make Úteis

```bash
# Setup
make setup-coralogix        # Criar arquivo .env

# Gestão de containers
make start_docker           # Iniciar todos os serviços
make stop_docker            # Parar todos os serviços

# Logs
make logs-otel             # Logs do OpenTelemetry Collector
make logs-api              # Logs do Order API
make logs-consumer         # Logs do Order Consumer
make logs-worker           # Logs do Order Worker
make logs-all              # Logs de todos os serviços

# Operações
make restart-otel          # Reiniciar o OpenTelemetry Collector
make health-check          # Verificar saúde dos serviços
```

## 🌍 Regiões do Coralogix

Escolha o domínio correto baseado na sua região:

| Região | Domínio |
|--------|---------|
| 🇪🇺 EU1 (Ireland) | `ingress.coralogix.com` |
| 🇪🇺 EU2 (Stockholm) | `ingress.eu2.coralogix.com` |
| 🇺🇸 US1 (Ohio) | `ingress.coralogix.us` |
| 🇺🇸 US2 (Oregon) | `ingress.cx498.coralogix.com` |
| 🇸🇬 AP1 (Singapore) | `ingress.coralogixsg.com` |
| 🇮🇳 AP2 (Mumbai) | `ingress.app.coralogix.in` |

## ⚠️ Troubleshooting

### Problema: "Não foi possível conectar com Coralogix"
```bash
# Teste conectividade
nc -zv ingress.eu2.coralogix.com 443

# Verifique firewall/proxy
curl -v https://ingress.eu2.coralogix.com
```

### Problema: "Private key inválida"
- Verifique se copiou a chave completa
- Não inclua espaços ou quebras de linha
- A chave deve ter pelo menos 32 caracteres

### Problema: "Dados não aparecem no Coralogix"
1. Aguarde 1-2 minutos (pode haver delay)
2. Verifique logs do collector: `make logs-otel`
3. Verifique se os serviços estão rodando: `make health-check`
4. Gere novo tráfego de teste

### Problema: "Serviço não inicia"
```bash
# Ver logs detalhados
docker compose -f deployment/docker-compose.yml logs order-api

# Verificar status
docker compose -f deployment/docker-compose.yml ps
```

## 📖 Documentação Completa

Para mais detalhes, consulte:
- [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) - Documentação completa
- [Coralogix Docs](https://coralogix.com/docs/)
- [OpenTelemetry Docs](https://opentelemetry.io/docs/)

## 🎯 Próximos Passos

1. ✅ Configurar alertas no Coralogix
2. ✅ Criar dashboards customizados
3. ✅ Configurar SLOs (Service Level Objectives)
4. ✅ Adicionar métricas de negócio customizadas

---

**Precisa de ajuda?** Consulte a [documentação completa](./CORALOGIX_SETUP.md) ou abra uma issue.
