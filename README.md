# Order Service

Este projeto é um serviço de pedidos (Order Service) desenvolvido em Go, com arquitetura modular, integração com banco de dados relacional (CockroachDB/PostgreSQL), mensageria (Kafka), observabilidade (OpenTelemetry), e preparado para execução em ambientes Docker.

## Funcionalidades
- Gerenciamento de pedidos (criação, atualização, consulta)
- Validação de clientes via serviço externo
- Processamento assíncrono de eventos via Kafka
- Observabilidade com métricas, logs e traces
- Migrações de banco de dados automatizadas
- Mock Server para testes de integração
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

### Opção 1: Rodar tudo com Docker (Recomendado)

```bash
# Ver todos os comandos disponíveis
make help

# Subir apenas a infraestrutura (DB, Kafka, Observabilidade)
make infra-up

# Subir tudo (Infraestrutura + Aplicações)
make up

# Ver status dos serviços
make ps

# Ver logs dos serviços
make logs-all           # Todos os logs
make logs-api           # Logs da API
make logs-consumer      # Logs do Consumer
make logs-worker        # Logs do Worker

# Parar os serviços
make down
```

### Opção 2: Rodar localmente (Desenvolvimento)

```bash
# 1. Subir apenas a infraestrutura
make infra-up

# 2. Configurar variáveis de ambiente
make dotenv

# 3. Executar as aplicações localmente
go run ./cmd/main.go api        # API REST
go run ./cmd/main.go consumers  # Consumer Kafka
go run ./cmd/main.go workers    # Worker
```

### Serviços Disponíveis

Após iniciar com `make up`, os seguintes serviços estarão disponíveis:

| Serviço | URL | Descrição |
|---------|-----|-----------|
| Order API | http://localhost:8001 | API REST do serviço de pedidos |
| CockroachDB UI | http://localhost:8080 | Interface do banco de dados |
| Kafka UI (Redpanda) | http://localhost:8085 | Interface para gerenciar Kafka |
| Jaeger | http://localhost:16686 | Visualização de traces |
| Prometheus | http://localhost:9090 | Métricas do sistema |
| Grafana | http://localhost:3000 | Dashboards de observabilidade |

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
make up

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

## 🧪 Mock Server - Client Service

Este projeto inclui um **Mock Server Postman** completo para simular a API de consulta de clientes, facilitando testes e desenvolvimento sem dependência de serviços externos.

### 📋 Recursos do Mock Server

- ✅ **5 Cenários de Teste**: Cliente ativo, inativo, não encontrado, serviço indisponível e timeout
- ✅ **Collection Postman Importável**: Pronta para importar e usar
- ✅ **Testes Automatizados**: Validação automática de responses
- ✅ **Scripts de Exemplo**: cURL, Python, Go, JavaScript
- ✅ **Documentação Completa**: Guias passo a passo

### 🚀 Quick Start

```bash
# 1. Importe a collection no Postman
docs/postman_collection.json

# 2. Crie o Mock Server no Postman:
#    Collection → Mock Collection → Create Mock Server

# 3. Use a URL gerada nos seus testes
export CLIENT_SERVICE_BASE_URL="https://xxxxx.mock.pstmn.io"
```

### 📚 Documentação do Mock Server

- **[Setup Guide](docs/POSTMAN_MOCK_SETUP.md)** - Guia completo de configuração
- **[Exemplos de Uso](docs/POSTMAN_MOCK_EXAMPLES.md)** - Scripts e código de exemplo
- **[Collection JSON](docs/postman_collection.json)** - Arquivo para importar no Postman

### 🎯 Cenários Disponíveis

| ClientId | Status | Comportamento |
|----------|--------|---------------|
| `client-123` | 200 | Retorna cliente ativo |
| `inactive-123` | 200 | Retorna cliente inativo |
| `notfound-123` | 404 | Cliente não encontrado |
| `unavailable-123` | 503 | Serviço indisponível |
| `timeout-123` | 504 | Simula timeout |

### Exemplo de Uso

```bash
# Testar cliente ativo
curl -X GET "https://xxxxx.mock.pstmn.io/clients/client-123" \
  -H "Content-Type: application/json"

# Response:
# {
#   "id": "client-123",
#   "name": "João Silva",
#   "email": "joao.silva@example.com",
#   "active": true
# }
```

## Contribuição
Pull requests são bem-vindos! Siga o padrão de código e mantenha os testes atualizados.