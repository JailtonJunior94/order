# 🚀 K6 Load Tests - Quick Start Guide

## 📦 Estrutura de Arquivos

```
k6/
├── README.md                          # Documentação completa
├── QUICK_START.md                     # Este guia rápido
├── Makefile                           # Comandos automatizados
├── config.example.env                 # Configurações de exemplo
├── run-ci.sh                          # Script para CI/CD
├── .github-workflows-example.yml      # Exemplo GitHub Actions
├── .gitignore                         # Ignora resultados
│
├── Scripts de Teste:
│   ├── complete-flow.js               # ⭐ Recomendado - Fluxo completo
│   ├── create-orders.js               # Apenas criação de orders
│   ├── mark-as-paid.js                # Apenas marcar como paga
│   ├── spike-test.js                  # Teste de pico
│   └── soak-test.js                   # Teste de longa duração
│
└── results/                           # Resultados (git ignored)
```

## ⚡ Começar em 3 Passos

### 1. Instalar k6
```bash
make install
# ou
brew install k6
```

### 2. Iniciar seu serviço
```bash
# Certifique-se que seu serviço está rodando em http://localhost:8080
# ou configure a URL com BASE_URL
```

### 3. Executar teste
```bash
# Teste rápido (1 usuário, 1 iteração)
make test-quick

# Teste completo (recomendado)
make test-complete

# Ver todos os comandos
make help
```

## 🎯 Casos de Uso Comuns

### Desenvolvimento Local
```bash
# Smoke test antes de commitar
make test-quick

# Teste de carga moderado
make test-complete
```

### Antes de Deploy
```bash
# Suite completa de benchmark
make benchmark

# Ou teste de spike para verificar escalabilidade
make test-spike
```

### CI/CD Pipeline
```bash
# Use o script CI
BASE_URL=http://staging-api:8080 TEST_TYPE=complete ./run-ci.sh
```

### Performance Regression Testing
```bash
# Execute e compare resultados
make test-complete
make results
```

## 📊 Comandos Mais Usados

| Comando | O que faz | Quando usar |
|---------|-----------|-------------|
| `make test-quick` | Teste rápido (1 VU) | Validação rápida, smoke test |
| `make test-complete` | Fluxo completo (100 VUs) | Teste de carga padrão |
| `make test-spike` | Pico súbito (200 VUs) | Testar escalabilidade |
| `make benchmark` | Suite completa | Antes de releases importantes |
| `make test-custom VUS=X DURATION=Ym` | Teste personalizado | Casos específicos |

## 🎨 Exemplos de Uso

### Teste Rápido para Validar API
```bash
cd k6
make test-quick
# ✅ Testa se endpoints estão funcionando
# ⚡ Rápido: ~10 segundos
```

### Teste de Carga Realista
```bash
cd k6
make test-complete
# 📊 Simula 100 usuários concorrentes
# ⏱️ Duração: ~6 minutos
# 📈 Métricas detalhadas de performance
```

### Teste de Estresse
```bash
cd k6
make test-spike
# 🚀 Pico súbito: 10 → 200 usuários
# 🔍 Identifica gargalos
# ⏱️ Duração: ~2 minutos
```

### Teste de Estabilidade
```bash
cd k6
make test-soak
# ⌛ Execução longa: 30 minutos
# 🔍 Detecta memory leaks
# 📊 20 usuários constantes
```

### Teste Customizado
```bash
cd k6
make test-custom VUS=75 DURATION=10m BASE_URL=http://staging:8080
# 🎛️ 75 usuários por 10 minutos
# 🌐 Apontando para staging
```

## 📈 Interpretando Resultados

Após executar um teste, você verá:

```
✓ create: status is 201
✓ create: response has id
✓ mark_as_paid: status is 200

checks.........................: 100.00% ✓ 3000  ✗ 0
data_received..................: 1.2 MB  20 kB/s
data_sent......................: 800 kB  13 kB/s
http_req_duration..............: avg=145ms min=50ms med=120ms max=890ms p(95)=350ms p(99)=580ms
http_req_failed................: 0.00%   ✓ 0     ✗ 3000
iterations.....................: 1000    16.6/s
```

### ✅ Sinais Positivos
- ✓ checks: 100% (todos os checks passaram)
- ✓ http_req_failed: 0% (sem erros)
- ✓ p(95) < 500ms (95% das requisições abaixo de 500ms)
- ✓ p(99) < 1000ms (99% das requisições abaixo de 1s)

### ⚠️ Sinais de Alerta
- ⚠️ checks < 95% (muitas falhas)
- ⚠️ http_req_failed > 5% (taxa de erro alta)
- ⚠️ p(95) > 1000ms (respostas lentas)
- ⚠️ Requests: tempo crescente ao longo do teste

## 🔧 Troubleshooting

### Erro: "k6 not found"
```bash
make install
# ou instale manualmente
```

### Erro: "Service not responding"
```bash
# Verifique se o serviço está rodando
curl http://localhost:8080/api/v1/orders

# Ou configure a URL correta
make test-complete BASE_URL=http://seu-servico:porta
```

### Muitos erros no teste
```bash
# 1. Verifique logs do serviço
# 2. Reduza a carga
make test-custom VUS=10 DURATION=1m
# 3. Verifique recursos (CPU, memória)
```

### Resultados inconsistentes
```bash
# Execute o teste algumas vezes
make test-complete
# Sistemas podem ter variação natural
# Procure por padrões, não valores absolutos
```

## 💡 Dicas de Boas Práticas

1. **Sempre faça warm-up**: Use `test-quick` antes de testes pesados
2. **Monitore recursos**: CPU, memória, disco durante os testes
3. **Compare resultados**: Mantenha histórico para detectar regressões
4. **Use em staging primeiro**: Nunca rode em produção sem testes
5. **Interprete com contexto**: Um p(95)=300ms pode ser bom ou ruim dependendo do caso
6. **Execute múltiplas vezes**: Uma execução pode ter variação, execute 3-5 vezes

## 🎓 Próximos Passos

1. ✅ Execute `make test-quick` para validar setup
2. ✅ Execute `make test-complete` para baseline de performance
3. ✅ Leia o [README.md](README.md) completo para opções avançadas
4. ✅ Configure CI/CD com `run-ci.sh`
5. ✅ Customize os scripts para seu caso de uso

## 📚 Links Úteis

- [Documentação Completa](README.md) - Guia detalhado
- [k6 Docs](https://k6.io/docs/) - Documentação oficial do k6
- [k6 Examples](https://github.com/grafana/k6-examples) - Exemplos da comunidade

---

**Dúvidas?** Consulte o [README.md](README.md) ou execute `make help`
