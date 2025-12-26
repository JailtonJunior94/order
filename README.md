# Order Service

Este projeto é um serviço de pedidos (Order Service) desenvolvido em Go, com arquitetura modular, integração com banco de dados relacional (CockroachDB/PostgreSQL), mensageria (Kafka), observabilidade (OpenTelemetry), e preparado para execução em ambientes Docker.

## Funcionalidades
- Gerenciamento de pedidos (criação, atualização, consulta)
- Processamento assíncrono de eventos via Kafka
- Observabilidade com métricas, logs e traces
- Migrações de banco de dados automatizadas
- Estrutura pronta para testes e escalabilidade

## Estrutura de Pastas

```
├── bin/                  # Binários gerados
├── cmd/                  # Entrypoints da aplicação
│   ├── main.go           # Main principal
│   ├── consumer/         # Entrypoint do consumidor Kafka
│   ├── server/           # Entrypoint do servidor HTTP/API
│   └── worker/           # Entrypoint de workers/background jobs
├── configs/              # Configurações da aplicação
├── database/             # Scripts e migrações do banco de dados
│   └── migrations/       # Migrações SQL
├── deployment/           # Arquivos de deploy (Docker, Compose, Observabilidade)
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── observability/    # Configuração de Prometheus, Grafana, Otel
├── internal/             # Domínio e regras de negócio
│   └── order/            # Módulo de pedidos
│       ├── domain/       # Entidades e regras de domínio
│       ├── infrastructure/ # Integrações externas
│       └── usecase/      # Casos de uso
├── pkg/                  # Pacotes reutilizáveis
│   ├── bundle/           # Container de injeção de dependências
│   ├── database/         # Utilitários de banco de dados
│   ├── entity/           # Entidades genéricas
│   ├── events/           # Event dispatcher
│   ├── http-client/      # Cliente HTTP
│   ├── messaging/        # Integração com Kafka
│   ├── responses/        # Respostas padrão
│   └── vos/              # Value Objects (UUID, ULID, datas)
├── go.mod                # Dependências Go
├── go.sum                # Hashes das dependências
├── Makefile              # Comandos de build, test, run
└── coverage.out          # Relatório de cobertura de testes
```

## Como rodar o projeto

1. **Suba os serviços necessários:**
   ```sh
   make start_docker
   ```
2. **Execute a aplicação:**
   ```sh
   go run ./cmd/main.go
   ```

## Observabilidade

Este projeto está totalmente integrado com **Coralogix** para observabilidade em produção, além de ferramentas locais para desenvolvimento.

### 🎯 Stack de Observabilidade

#### Produção (Coralogix)
- **Logs**: Logs estruturados com contexto completo
- **Traces**: Distributed tracing com OpenTelemetry
- **Métricas**: Métricas customizadas e de sistema

#### Desenvolvimento (Local)
- **Jaeger**: Visualização de traces (http://localhost:16686)
- **Prometheus**: Coleta de métricas (http://localhost:9090)
- **Grafana**: Dashboards e visualizações (http://localhost:3000)
- **Loki**: Agregação de logs

### 🚀 Quick Start com Coralogix

```bash
# 1. Configure o Coralogix
make setup-coralogix

# 2. Edite deployment/.env com sua Private Key
# CORALOGIX_PRIVATE_KEY=sua-chave-aqui

# 3. Inicie os serviços
make start_docker

# 4. Verifique os logs
make logs-otel
```

### 📚 Documentação Detalhada

- **[Quick Start](deployment/QUICKSTART.md)** - Guia rápido de início
- **[Setup Completo](deployment/CORALOGIX_SETUP.md)** - Documentação detalhada
- **Script de Verificação**: `./deployment/verify-coralogix-config.sh`

### 🔍 Serviços Instrumentados

Todos os serviços enviam telemetria para o Coralogix:

1. **order-api** (porta 8000) - API REST
   - Traces de requisições HTTP
   - Métricas de performance
   - Logs estruturados

2. **order-consumer** - Consumidor Kafka
   - Traces de processamento de mensagens
   - Métricas de lag e throughput
   - Logs de eventos

3. **order-worker** - Worker de tarefas
   - Traces de execução de jobs
   - Métricas de processamento
   - Logs de execução

### 📊 Acessar Dashboards

```bash
# Coralogix Dashboard
open https://dashboard.coralogix.com

# Ferramentas Locais
open http://localhost:16686  # Jaeger
open http://localhost:9090  # Prometheus
open http://localhost:3000  # Grafana
```

## Banco de Dados
- Migrações SQL em `database/migrations/`
- Suporte a CockroachDB e PostgreSQL

## Mensageria
- Integração com Kafka para eventos de pedidos

## Contribuição
Pull requests são bem-vindos! Siga o padrão de código e mantenha os testes atualizados.