# 📚 Índice da Documentação - Observabilidade com Coralogix

Bem-vindo à documentação de observabilidade do Order Service!

## 🚀 Início Rápido

Novo no projeto? Comece aqui:

1. **[QUICKSTART.md](./QUICKSTART.md)** 
   - ⏱️ Tempo estimado: 10 minutos
   - Setup em 5 passos simples
   - Validação rápida

## 📖 Documentação

### Guias Principais

| Documento | Descrição | Quando Usar |
|-----------|-----------|-------------|
| [QUICKSTART.md](./QUICKSTART.md) | Guia rápido de início | Primeira configuração |
| [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) | Documentação completa | Referência detalhada |
| [CHECKLIST.md](./CHECKLIST.md) | Checklist de configuração | Validar setup |

### Scripts Úteis

| Script | Descrição | Comando |
|--------|-----------|---------|
| `verify-coralogix-config.sh` | Verifica configuração | `./deployment/verify-coralogix-config.sh` |
| `test-telemetry.sh` | Gera tráfego de teste | `./deployment/test-telemetry.sh` |

### Arquivos de Configuração

| Arquivo | Descrição |
|---------|-----------|
| `.env.example` | Template de variáveis de ambiente |
| `observability/collector/otel-collector.yml` | Config do OpenTelemetry Collector |
| `docker-compose.yml` | Definição de todos os serviços |

## 🎯 Fluxo de Trabalho Recomendado

```
1. Ler QUICKSTART.md
   ↓
2. Executar setup-coralogix
   ↓
3. Configurar .env
   ↓
4. Executar verify-coralogix-config.sh
   ↓
5. Iniciar serviços (make start_docker)
   ↓
6. Executar test-telemetry.sh
   ↓
7. Validar no Coralogix Dashboard
   ↓
8. Marcar itens no CHECKLIST.md
```

## 🏗️ Arquitetura de Observabilidade

```
┌─────────────────────────────────────────────────────────────┐
│                      Order Services                          │
│  ┌──────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   API    │  │  Consumer   │  │   Worker    │            │
│  └────┬─────┘  └──────┬──────┘  └──────┬──────┘            │
│       │               │                 │                    │
│       └───────────────┴─────────────────┘                    │
│                       │                                      │
│              Logs, Metrics, Traces                          │
│                       ↓                                      │
│         ┌─────────────────────────┐                         │
│         │  OpenTelemetry          │                         │
│         │  Collector              │                         │
│         └────┬──────────────┬─────┘                         │
└──────────────┼──────────────┼──────────────────────────────┘
               │              │
        Local  │              │  Cloud
               ↓              ↓
    ┌──────────────┐   ┌─────────────┐
    │   Jaeger     │   │             │
    │   Prometheus │   │  Coralogix  │
    │   Grafana    │   │             │
    │   Loki       │   │             │
    └──────────────┘   └─────────────┘
```

## 📊 Serviços Monitorados

### order-api (Porta 8000)
- **Função**: API REST principal
- **Telemetria**:
  - ✅ HTTP request/response traces
  - ✅ Latência por endpoint
  - ✅ Taxa de erros
  - ✅ Logs estruturados

### order-consumer
- **Função**: Consumidor Kafka
- **Telemetria**:
  - ✅ Traces de processamento
  - ✅ Consumer lag
  - ✅ Mensagens processadas
  - ✅ Erros de processamento

### order-worker
- **Função**: Worker de tarefas agendadas
- **Telemetria**:
  - ✅ Execução de jobs
  - ✅ Duração dos jobs
  - ✅ Status de execução
  - ✅ Falhas e retries

## 🛠️ Comandos Makefile

```bash
# Setup
make setup-coralogix        # Criar arquivo .env

# Gestão
make start_docker           # Iniciar todos os serviços
make stop_docker            # Parar todos os serviços

# Logs
make logs-otel             # Logs do OpenTelemetry Collector
make logs-api              # Logs do Order API
make logs-consumer         # Logs do Order Consumer
make logs-worker           # Logs do Order Worker
make logs-all              # Logs de todos os serviços

# Operações
make restart-otel          # Reiniciar OpenTelemetry Collector
make health-check          # Verificar saúde dos serviços
```

## 🌍 Regiões Suportadas

| Região | Domínio |
|--------|---------|
| 🇪🇺 EU1 (Ireland) | `ingress.coralogix.com` |
| 🇪🇺 EU2 (Stockholm) | `ingress.eu2.coralogix.com` |
| 🇺🇸 US1 (Ohio) | `ingress.coralogix.us` |
| 🇺🇸 US2 (Oregon) | `ingress.cx498.coralogix.com` |
| 🇸🇬 AP1 (Singapore) | `ingress.coralogixsg.com` |
| 🇮🇳 AP2 (Mumbai) | `ingress.app.coralogix.in` |

## 🔗 Links Úteis

### Dashboards
- **Coralogix**: https://dashboard.coralogix.com
- **Jaeger**: http://localhost:16686
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Order API**: http://localhost:8000/health

### Documentação Externa
- [Coralogix Docs](https://coralogix.com/docs/)
- [OpenTelemetry Docs](https://opentelemetry.io/docs/)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Go OpenTelemetry](https://opentelemetry.io/docs/languages/go/)

## 🆘 Troubleshooting

### Problemas Comuns

| Problema | Documento | Seção |
|----------|-----------|-------|
| Configuração inicial | [QUICKSTART.md](./QUICKSTART.md) | Setup em 5 Passos |
| Private key inválida | [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) | Troubleshooting |
| Dados não aparecem | [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) | Troubleshooting |
| Erro de conexão | [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) | Troubleshooting |

### Scripts de Diagnóstico

```bash
# Verificar configuração
./deployment/verify-coralogix-config.sh

# Verificar saúde dos serviços
make health-check

# Ver logs de erro
docker compose -f deployment/docker-compose.yml logs | grep -i error

# Verificar conectividade
nc -zv ingress.eu2.coralogix.com 443
```

## 📈 Métricas Importantes

### Logs
- Volume: ~100+ linhas/minuto (todos os serviços)
- Severidades: DEBUG, INFO, WARN, ERROR
- Campos: timestamp, service, message, context

### Traces
- Volume: ~10+ traces/minuto
- Duração média: < 500ms
- Taxa de erro: < 5%

### Métricas
- Request rate: requisições/segundo
- Latência: P50, P95, P99
- Error rate: erros/total
- Consumer lag: mensagens pendentes

## 🎓 Conceitos

### OpenTelemetry
Framework de observabilidade que padroniza coleta de telemetria

### Coralogix
Plataforma SaaS de observabilidade com análise em tempo real

### Traces
Caminho completo de uma requisição através dos serviços

### Spans
Unidade individual de trabalho dentro de um trace

### Metrics
Medições numéricas ao longo do tempo

### Logs
Registros textuais de eventos da aplicação

## 🔄 Ciclo de Vida da Telemetria

```
1. Instrumentação
   ↓ (SDK OpenTelemetry no código)
2. Coleta
   ↓ (OpenTelemetry Collector)
3. Processamento
   ↓ (Batch, Filters, Transform)
4. Exportação
   ↓ (OTLP Protocol)
5. Armazenamento
   ↓ (Coralogix / Local)
6. Visualização
   ↓ (Dashboards, Queries)
7. Alertas
   (Notificações baseadas em thresholds)
```

## 📝 Próximos Passos

Após setup completo:

1. ✅ Configurar alertas no Coralogix
2. ✅ Criar dashboards customizados
3. ✅ Definir SLOs (Service Level Objectives)
4. ✅ Adicionar métricas de negócio
5. ✅ Configurar log parsing rules
6. ✅ Integrar com Slack/PagerDuty
7. ✅ Documentar runbooks

## 📞 Suporte

- **Issues do Projeto**: Abra uma issue no repositório
- **Coralogix Support**: https://coralogix.com/support/
- **OpenTelemetry Community**: https://opentelemetry.io/community/

---

**Última Atualização**: Dezembro 2025  
**Versão da Documentação**: 1.0.0
