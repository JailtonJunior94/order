# Testes de Carga com K6 - Order Service

Este diretório contém os scripts de testes de carga para o serviço de orders usando [k6](https://k6.io/).

## 📋 Pré-requisitos

1. Instalar o k6:
   ```bash
   # Usando Make (recomendado)
   make install

   # Ou manualmente:
   # macOS
   brew install k6

   # Linux
   sudo gpg -k
   sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
   echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
   sudo apt-get update
   sudo apt-get install k6

   # Windows (usando Chocolatey)
   choco install k6
   ```

2. Certifique-se de que o serviço esteja rodando (padrão: http://localhost:8080)

## 🚀 Início Rápido com Makefile

O Makefile fornece comandos simplificados para executar os testes:

```bash
# Ver todos os comandos disponíveis
make help

# Verificar se k6 está instalado
make check-k6

# Executar teste completo (recomendado)
make test-complete

# Executar teste rápido (smoke test)
make test-quick

# Executar todos os testes
make test-all

# Executar benchmark completo
make benchmark

# Ver informações de configuração
make info

# Limpar resultados
make clean
```

### Comandos Disponíveis no Makefile

| Comando | Descrição |
|---------|-----------|
| `make help` | Mostra todos os comandos disponíveis |
| `make install` | Instala o k6 (macOS/Linux) |
| `make check-k6` | Verifica se k6 está instalado |
| `make check-service` | Verifica se o serviço está rodando |
| `make test-quick` | Smoke test rápido (1 VU) |
| `make test-create` | Teste de criação de orders |
| `make test-complete` | Teste de fluxo completo (recomendado) |
| `make test-spike` | Teste de pico de carga |
| `make test-soak` | Teste de longa duração (30 min) |
| `make test-all` | Executa todos os testes sequencialmente |
| `make benchmark` | Executa suite completa de benchmark |
| `make test-custom` | Teste customizado (use VUS e DURATION) |
| `make clean` | Remove resultados de testes |
| `make results` | Lista resultados mais recentes |
| `make view-results` | Exibe resumo do último teste |
| `make info` | Mostra informações de configuração |

### Exemplos de Uso Avançado

```bash
# Executar com URL customizada
make test-complete BASE_URL=http://api.example.com:8080

# Executar teste customizado com 50 VUs por 5 minutos
make test-custom VUS=50 DURATION=5m BASE_URL=http://localhost:8080

# Executar benchmark completo
make benchmark

# Ver últimos resultados
make results
make view-results
```

## 🧪 Scripts Disponíveis

### 1. `complete-flow.js` (Recomendado)
Executa o fluxo completo de criação de order e marcação como paga.

**Uso:**
```bash
k6 run complete-flow.js
```

**Com URL customizada:**
```bash
k6 run -e BASE_URL=http://localhost:8080 complete-flow.js
```

**Características:**
- Simula o fluxo real do usuário
- Cria order e depois marca como paga
- 100 usuários virtuais no pico
- Duração total: ~6 minutos
- Métricas customizadas para orders criadas e pagas

### 2. `create-orders.js`
Foca apenas na criação de orders.

**Uso:**
```bash
k6 run create-orders.js
```

**Características:**
- 100 usuários virtuais no pico
- Duração total: ~4 minutos
- Testa apenas o endpoint de criação

### 3. `mark-as-paid.js`
Foca apenas em marcar orders como pagas.

**Uso:**
```bash
k6 run mark-as-paid.js
```

**Nota:** Requer um arquivo `order-ids.json` com IDs de orders existentes ou use o `complete-flow.js`.

**Características:**
- 50 usuários virtuais no pico
- Requer IDs pré-existentes
- Testa apenas o endpoint de pagamento

### 4. `spike-test.js`
Teste de pico para verificar comportamento sob carga súbita.

**Uso:**
```bash
k6 run spike-test.js
```

**Características:**
- Pico súbito de 10 para 200 usuários
- Verifica recuperação do sistema
- Thresholds mais relaxados
- Duração total: ~2 minutos

### 5. `soak-test.js`
Teste de longa duração para detectar vazamento de memória e degradação de performance.

**Uso:**
```bash
k6 run soak-test.js
```

**Características:**
- 20 usuários constantes
- Duração: 30 minutos
- Thresholds rigorosos
- Identifica problemas de estabilidade

## 📊 Métricas e Thresholds

Os testes incluem as seguintes métricas:

- **http_req_duration**: Tempo de resposta das requisições
  - P95 < 500ms
  - P99 < 1000ms
- **http_req_failed**: Taxa de falha das requisições
  - < 5% de erro
- **errors**: Taxa de erros customizada
  - < 5% de erro
- **orders_created**: Contador de orders criadas (complete-flow.js)
- **orders_marked_as_paid**: Contador de orders pagas (complete-flow.js)

## 🎯 Cenários de Teste

### Load Test (complete-flow.js)
Simula carga gradual até atingir o pico:
1. 30s: 0 → 10 usuários (warm-up)
2. 1m: 10 → 50 usuários (ramp-up)
3. 3m: 50 → 100 usuários (carga máxima)
4. 1m: 100 → 50 usuários (ramp-down)
5. 30s: 50 → 0 usuários (cool-down)

### Spike Test (spike-test.js)
Testa resistência a picos súbitos:
1. 10s: 0 → 10 usuários
2. 20s: 10 usuários (baseline)
3. 10s: 10 → 200 usuários (SPIKE!)
4. 30s: 200 usuários (sustentação)
5. 10s: 200 → 10 usuários (recuperação)
6. 20s: 10 usuários (verificação)
7. 10s: 10 → 0 usuários

### Soak Test (soak-test.js)
Testa estabilidade de longo prazo:
1. 2m: 0 → 20 usuários
2. 30m: 20 usuários (carga constante)
3. 2m: 20 → 0 usuários

## 📈 Gerando Relatórios

### Relatório HTML
```bash
k6 run --out json=results.json complete-flow.js
```

### Dashboard em tempo real (k6 Cloud - grátis para testes pequenos)
```bash
k6 cloud complete-flow.js
```

### Integração com InfluxDB + Grafana
```bash
k6 run --out influxdb=http://localhost:8086/k6 complete-flow.js
```

## 🔧 Customização

### Alterar URL base
```bash
k6 run -e BASE_URL=http://your-api:8080 complete-flow.js
```

### Aumentar/diminuir carga
Edite o arquivo `.js` e modifique a seção `options.stages`:
```javascript
export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Ajuste o target
    { duration: '3m', target: 200 },  // Ajuste o target
  ],
};
```

### Adicionar novos clients/produtos
Edite as funções `generateRandomClient()` e `generateRandomItems()` nos arquivos.

## 💡 Melhores Práticas

1. **Sempre faça warm-up**: Comece com poucos usuários para aquecer o sistema
2. **Monitore recursos**: Observe CPU, memória e I/O durante os testes
3. **Execute em ambiente isolado**: Não rode testes de carga em produção
4. **Análise incremental**: Comece com carga baixa e aumente gradualmente
5. **Documente resultados**: Salve os resultados para comparação futura

## 🐛 Troubleshooting

### Erro: "No order IDs available"
- Use `complete-flow.js` que cria e paga orders no mesmo fluxo
- Ou popule `order-ids.json` com IDs válidos

### Muitos erros 404/500
- Verifique se o serviço está rodando
- Confirme a URL base está correta
- Verifique os logs do serviço

### Performance ruim
- Verifique recursos da máquina (CPU, memória)
- Analise logs do serviço para gargalos
- Considere escalar o banco de dados

## 🔧 Arquivos de Configuração

### `config.example.env`
Arquivo de exemplo com todas as configurações disponíveis:
```bash
cp config.example.env config.env
# Edite config.env com suas configurações
source config.env
make test-complete
```

### `run-ci.sh`
Script para execução em pipelines CI/CD:
```bash
# Uso local
chmod +x run-ci.sh
BASE_URL=http://localhost:8080 TEST_TYPE=quick ./run-ci.sh

# Em CI/CD
export BASE_URL=http://api-service:8080
export TEST_TYPE=complete
./run-ci.sh
```

**Variáveis de ambiente:**
- `BASE_URL`: URL do serviço (padrão: http://localhost:8080)
- `TEST_TYPE`: Tipo de teste (quick, create, complete, spike, soak)

### `.github-workflows-example.yml`
Exemplo de workflow para GitHub Actions. Para usar:
```bash
mkdir -p ../.github/workflows
cp .github-workflows-example.yml ../.github/workflows/k6-load-test.yml
# Edite conforme necessário
```

## 🎯 Integração CI/CD

### GitHub Actions
Use o exemplo fornecido em `.github-workflows-example.yml`:
- Executa testes em PRs e pushes
- Upload de resultados como artifacts
- Validação de thresholds de performance

### GitLab CI
```yaml
k6-load-test:
  stage: test
  image: grafana/k6:latest
  services:
    - postgres:15
  script:
    - cd k6
    - chmod +x run-ci.sh
    - ./run-ci.sh
  artifacts:
    paths:
      - k6/results/
    expire_in: 30 days
  variables:
    BASE_URL: "http://localhost:8080"
    TEST_TYPE: "quick"
```

### Jenkins
```groovy
pipeline {
    agent any
    stages {
        stage('Load Test') {
            steps {
                sh 'cd k6 && chmod +x run-ci.sh'
                sh 'BASE_URL=http://localhost:8080 TEST_TYPE=complete ./k6/run-ci.sh'
            }
        }
    }
    post {
        always {
            archiveArtifacts artifacts: 'k6/results/**/*', fingerprint: true
        }
    }
}
```

## 📚 Recursos

- [Documentação k6](https://k6.io/docs/)
- [k6 Examples](https://github.com/grafana/k6-examples)
- [Load Testing Best Practices](https://k6.io/docs/testing-guides/load-testing/)
- [k6 Cloud](https://k6.io/cloud/) - Plataforma para visualização de resultados
