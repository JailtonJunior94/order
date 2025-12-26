# Configuração Coralogix - Observabilidade

Este guia descreve como configurar o envio de logs, métricas e traces para o Coralogix.

## Arquitetura

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────┐
│  order-api  │────▶│                 │────▶│              │
├─────────────┤     │                 │     │              │
│order-consumer│────▶│ otel-collector │────▶│  Coralogix   │
├─────────────┤     │                 │     │              │
│order-worker │────▶│                 │     │              │
└─────────────┘     └─────────────────┘     └──────────────┘
                           │
                           ├──▶ Jaeger (traces local)
                           ├──▶ Prometheus (metrics local)
                           └──▶ Loki (logs local)
```

## Serviços Instrumentados

Os seguintes serviços enviam telemetria para o Coralogix:

1. **order-api** - API REST principal
   - Traces de requisições HTTP
   - Métricas de performance
   - Logs de aplicação

2. **order-consumer** - Consumidor Kafka
   - Traces de processamento de mensagens
   - Métricas de consumo
   - Logs de processamento

3. **order-worker** - Worker de tarefas agendadas
   - Traces de execução de jobs
   - Métricas de processamento
   - Logs de execução

## Configuração

### 1. Obter Credenciais do Coralogix

1. Acesse [Coralogix Dashboard](https://dashboard.coralogix.com)
2. Vá em **Settings** → **Send Your Data** → **API Key**
3. Copie sua **Private Key**

### 2. Configurar Variáveis de Ambiente

Copie o arquivo `.env.example` para `.env` no diretório `deployment/`:

```bash
cd deployment
cp .env.example .env
```

Edite o arquivo `.env` e configure:

```env
# Sua chave privada do Coralogix
CORALOGIX_PRIVATE_KEY=your-private-key-here

# Nome da aplicação (será exibido no Coralogix)
CORALOGIX_APP_NAME=orders

# Domínio baseado na sua região
CORALOGIX_DOMAIN=ingress.eu2.coralogix.com

# Ambiente
ENVIRONMENT=development
```

### 3. Domínios por Região

Escolha o domínio correto baseado na região do seu account Coralogix:

| Região | Domínio |
|--------|---------|
| EU1 (Ireland) | `ingress.coralogix.com` |
| EU2 (Stockholm) | `ingress.eu2.coralogix.com` |
| US1 (Ohio) | `ingress.coralogix.us` |
| US2 (Oregon) | `ingress.cx498.coralogix.com` |
| AP1 (Singapore) | `ingress.coralogixsg.com` |
| AP2 (Mumbai) | `ingress.app.coralogix.in` |

### 4. Iniciar os Serviços

```bash
# Iniciar todos os serviços
docker-compose up -d

# Verificar logs do OpenTelemetry Collector
docker-compose logs -f otel-collector

# Verificar logs dos serviços
docker-compose logs -f order-api
docker-compose logs -f order-consumer
docker-compose logs -f order-worker
```

## Validação

### Verificar Conectividade

1. **OpenTelemetry Collector**: Verifique os logs do collector para confirmar conexão com Coralogix:
   ```bash
   docker-compose logs otel-collector | grep -i coralogix
   ```

2. **Serviços**: Verifique se os serviços estão enviando telemetria:
   ```bash
   docker-compose logs order-api | grep -i "otlp"
   ```

### Visualizar Dados no Coralogix

1. Acesse o [Coralogix Dashboard](https://dashboard.coralogix.com)
2. Vá para as seções:
   - **Logs** → Filtre por `cx.application.name:orders`
   - **Tracing** → Visualize traces dos serviços
   - **Metrics** → Dashboard de métricas customizadas

### Filtros Úteis no Coralogix

```
# Ver logs de um serviço específico
cx.application.name:orders AND cx.subsystem.name:order-api

# Ver todos os erros
cx.application.name:orders AND level:error

# Ver traces de um endpoint específico
cx.application.name:orders AND http.route:/orders
```

## Atributos Enviados

Cada telemetria inclui os seguintes atributos:

```yaml
cx.application.name: orders              # Nome da aplicação
cx.subsystem.name: order-api|order-consumer|order-worker  # Nome do serviço
deployment.environment: development      # Ambiente
service.name: order-api                 # Nome do serviço OpenTelemetry
service.version: 1.0.0                  # Versão da aplicação
service.namespace: orders               # Namespace
```

## Observabilidade Local

Além do Coralogix, a telemetria também é enviada para ferramentas locais:

- **Jaeger**: http://localhost:16686 (Traces)
- **Prometheus**: http://localhost:9090 (Metrics)
- **Grafana**: http://localhost:3000 (Dashboards)

## Troubleshooting

### Problema: Dados não aparecem no Coralogix

1. Verifique se a `CORALOGIX_PRIVATE_KEY` está correta
2. Confirme que está usando o domínio correto para sua região
3. Verifique os logs do otel-collector:
   ```bash
   docker-compose logs otel-collector | grep -i error
   ```

### Problema: Erros de autenticação

```
Error: Permanent error: rpc error: code = Unauthenticated
```

- Verifique se a private key está correta e não expirou

### Problema: Timeout de conexão

```
Error: context deadline exceeded
```

- Verifique se o domínio do Coralogix está correto
- Confirme que sua rede permite conexões HTTPS na porta 443

## Métricas Customizadas

Para adicionar métricas customizadas, use a biblioteca OpenTelemetry no seu código:

```go
// Exemplo de contador
counter, _ := meter.Int64Counter("orders.created",
    metric.WithDescription("Number of orders created"))
counter.Add(ctx, 1)
```

## Dashboards Sugeridos no Coralogix

1. **API Performance**
   - Request rate por endpoint
   - Latência P50, P95, P99
   - Taxa de erros

2. **Kafka Consumer**
   - Mensagens processadas
   - Lag de consumo
   - Erros de processamento

3. **Worker Jobs**
   - Execuções do worker
   - Duração dos jobs
   - Falhas

## Alertas Recomendados

Configure alertas no Coralogix para:

- Taxa de erros > 5%
- Latência P95 > 1s
- Kafka consumer lag > 1000 mensagens
- Worker failures consecutivas > 3

## Referências

- [Coralogix Documentation](https://coralogix.com/docs/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
