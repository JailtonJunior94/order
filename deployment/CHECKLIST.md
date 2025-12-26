# ✅ Checklist de Configuração - Coralogix

Use este checklist para garantir que tudo está configurado corretamente.

## 📋 Pré-Setup

- [ ] Docker e Docker Compose instalados
- [ ] Make instalado (opcional)
- [ ] Conta no Coralogix criada
- [ ] Private Key do Coralogix obtida

## 🔧 Configuração

### 1. Arquivo de Environment
- [ ] Arquivo `deployment/.env` criado
- [ ] `CORALOGIX_PRIVATE_KEY` configurada
- [ ] `CORALOGIX_APP_NAME` definido (ex: orders)
- [ ] `CORALOGIX_DOMAIN` correto para sua região
- [ ] `ENVIRONMENT` definido (development/staging/production)

### 2. Arquivos de Configuração
- [ ] `deployment/observability/collector/otel-collector.yml` atualizado
- [ ] `deployment/docker-compose.yml` com os 3 serviços
- [ ] Variáveis de ambiente configuradas no docker-compose

### 3. Verificação
- [ ] Script `verify-coralogix-config.sh` executado
- [ ] Todas as verificações passaram ✓
- [ ] Conectividade com Coralogix OK

## 🚀 Execução

### 1. Iniciar Serviços
- [ ] `make start_docker` executado com sucesso
- [ ] Todos os containers iniciaram corretamente
- [ ] Logs dos containers não mostram erros críticos

### 2. Verificar Logs
- [ ] OpenTelemetry Collector conectou ao Coralogix
- [ ] order-api está enviando telemetria
- [ ] order-consumer está enviando telemetria
- [ ] order-worker está enviando telemetria

### 3. Gerar Tráfego de Teste
- [ ] Script `test-telemetry.sh` executado
- [ ] Requisições de teste bem-sucedidas
- [ ] Logs mostram processamento

## 🔍 Validação no Coralogix

### 1. Acesso ao Dashboard
- [ ] Dashboard Coralogix acessível
- [ ] Login realizado com sucesso

### 2. Logs
- [ ] Logs aparecem na seção "Explore → Logs"
- [ ] Filtro `cx.application.name:orders` funciona
- [ ] Logs dos 3 serviços visíveis
- [ ] Níveis de log corretos (info, error, debug)

### 3. Traces
- [ ] Traces aparecem na seção "Explore → Tracing"
- [ ] Traces de requisições HTTP visíveis
- [ ] Traces do consumer Kafka visíveis
- [ ] Traces do worker visíveis
- [ ] Trace spans conectados corretamente

### 4. Métricas
- [ ] Métricas aparecem na seção "Dashboards"
- [ ] Métricas de sistema visíveis
- [ ] Métricas customizadas (se houver)

## 📊 Ferramentas Locais

### 1. Jaeger
- [ ] Acessível em http://localhost:16686
- [ ] Serviços aparecem na lista
- [ ] Traces visíveis

### 2. Prometheus
- [ ] Acessível em http://localhost:9090
- [ ] Targets estão "UP"
- [ ] Métricas podem ser consultadas

### 3. Grafana
- [ ] Acessível em http://localhost:3000
- [ ] Data sources configurados
- [ ] Dashboards carregam

## 🔧 Troubleshooting

Se algo falhar, verifique:

### Conectividade
- [ ] Firewall não está bloqueando porta 443
- [ ] Proxy configurado corretamente (se aplicável)
- [ ] DNS resolve o domínio do Coralogix

### Autenticação
- [ ] Private Key está correta
- [ ] Chave não expirou
- [ ] Chave tem permissões corretas

### Serviços
- [ ] Todos os containers estão rodando
- [ ] Logs não mostram erros de conexão
- [ ] Variáveis de ambiente corretas

## 📈 Métricas de Sucesso

Após configuração completa, você deve ver:

### Logs
- [x] Pelo menos 100+ linhas de log por minuto
- [x] Logs de todos os 3 serviços
- [x] Diferentes níveis de severidade

### Traces
- [x] Pelo menos 10+ traces por minuto
- [x] Spans conectados corretamente
- [x] Latências sendo capturadas

### Métricas
- [x] Métricas de sistema (CPU, memória)
- [x] Métricas de aplicação (requests, latency)
- [x] Métricas de Kafka (lag, throughput)

## 🎯 Configurações Avançadas (Opcional)

- [ ] Alertas configurados no Coralogix
- [ ] Dashboards customizados criados
- [ ] SLOs (Service Level Objectives) definidos
- [ ] Integração com Slack/PagerDuty
- [ ] Recording Rules no Prometheus
- [ ] Log parsing rules no Coralogix

## 📝 Notas

### Região Configurada
**Região**: _______________
**Domínio**: _______________

### Credenciais
**App Name**: _______________
**Environment**: _______________

### Endpoints
- Order API: http://localhost:8000
- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- Coralogix: https://dashboard.coralogix.com

## ✅ Status Final

Data de configuração: _______________
Configurado por: _______________
Status: [ ] Completo [ ] Pendente [ ] Com problemas

### Observações:
```
_______________________________________________
_______________________________________________
_______________________________________________
```

---

**Documentação**: Consulte [CORALOGIX_SETUP.md](./CORALOGIX_SETUP.md) para mais detalhes.
