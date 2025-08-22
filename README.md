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
- Configuração de Prometheus, Grafana e OpenTelemetry em `deployment/observability/`

## Banco de Dados
- Migrações SQL em `database/migrations/`
- Suporte a CockroachDB e PostgreSQL

## Mensageria
- Integração com Kafka para eventos de pedidos

## Contribuição
Pull requests são bem-vindos! Siga o padrão de código e mantenha os testes atualizados.