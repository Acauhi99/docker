# Docker POC - Infraestrutura Gerenciada

Projeto de prova de conceito para orquestração de containers Docker com foco em segurança, performance e escalabilidade.



## 🏗️ Arquitetura

```
Cliente → NGINX (Load Balancer) → Producer API (3 instâncias) → RabbitMQ → Consumer API → MongoDB
```

### Componentes

- **NGINX**: Load balancer com algoritmo least_conn
- **Producer API**: 3 instâncias em Go que recebem eventos via POST
- **RabbitMQ**: Message broker para comunicação assíncrona
- **Consumer API**: Processa mensagens e persiste no MongoDB
- **MongoDB**: Banco de dados NoSQL para armazenamento

## 📋 Pré-requisitos

- Docker Engine 24.0+
- Docker Compose 2.20+
- Docker Buildx (para SBOM e multi-platform)

## 🚀 Como Usar

### 1. Configurar Variáveis de Ambiente

Copie o arquivo `.env` e ajuste conforme necessário:

```bash
cp .env .env.local
```

### 2. Build com Docker Bake (Recomendado)

```bash
# Build com SBOM
docker buildx bake --load

# Gerar SBOM separadamente
docker buildx bake producer-sbom consumer-sbom
```

### 3. Iniciar Infraestrutura

```bash
docker-compose up -d
```

### 4. Verificar Status

```bash
docker-compose ps
docker-compose logs -f
```

### 5. Testar API

```bash
# Enviar evento
curl -X POST http://localhost/events \
  -H "Content-Type: application/json" \
  -d '{
    "device": "smartphone",
    "os": "android",
    "tipo": "click",
    "valor": "100",
    "ip": "192.168.1.1",
    "region": "us-east-1"
  }'

# Verificar health
curl http://localhost/health
```

### 6. Acessar RabbitMQ Management

```
URL: http://localhost:15672
User: conforme .env (RABBITMQ_DEFAULT_USER)
Pass: conforme .env (RABBITMQ_DEFAULT_PASS)
```

## 🔒 Segurança

### Implementações

- ✅ Containers rodando como usuário não-root (UID 65534)
- ✅ Filesystem read-only onde possível
- ✅ Security options: no-new-privileges
- ✅ Imagens multi-stage build (distroless para Go)
- ✅ SBOM gerado automaticamente
- ✅ Sem credenciais hardcoded (tudo via .env)
- ✅ Networks isoladas

### Docker Scout

```bash
# Analisar vulnerabilidades
docker scout cves producer:latest
docker scout cves consumer:latest

# Recomendações
docker scout recommendations producer:latest
```

## 📊 Monitoramento

### Health Checks

Todos os serviços possuem healthcheck configurado:

- MongoDB: `mongosh ping`
- RabbitMQ: `rabbitmq-diagnostics ping`
- Producer: endpoint `/health`
- Consumer: endpoint `/health`
- NGINX: endpoint `/health`

### Logs

```bash
# Todos os serviços
docker-compose logs -f

# Serviço específico
docker-compose logs -f producer-1
docker-compose logs -f consumer
```

## ⚙️ Recursos e Limites

Cada serviço possui limites configurados via .env:

| Serviço | CPU Limit | Memory Limit |
|---------|-----------|--------------|
| Producer | 0.5 | 256M |
| Consumer | 0.5 | 256M |
| MongoDB | 1.0 | 512M |
| RabbitMQ | 1.0 | 512M |
| NGINX | 0.25 | 128M |

## 🔧 Escalabilidade

### Escalar Producer

Edite `docker-compose.yml` e adicione mais instâncias:

```yaml
producer-4:
  # ... mesma configuração
```

Atualize `nginx/nginx.conf`:

```nginx
upstream producer_backend {
    server producer-4:8080;
    # ...
}
```

### Escalar Consumer

```bash
docker-compose up -d --scale consumer=3
```

## 🧪 Testes de Carga

```bash
# Instalar hey
go install github.com/rakyll/hey@latest

# Teste de carga
hey -n 10000 -c 100 -m POST \
  -H "Content-Type: application/json" \
  -d '{"device":"test","os":"linux","tipo":"load","valor":"1","ip":"127.0.0.1","region":"local"}' \
  http://localhost/events
```

## 🛠️ Troubleshooting

### Container não inicia

```bash
docker-compose logs <service-name>
docker inspect <container-name>
```

### Problemas de conectividade

```bash
docker network inspect docker_backend
docker-compose exec producer-1 ping rabbitmq
```

### Limpar tudo

```bash
docker-compose down -v
docker system prune -a
```

## 📁 Estrutura do Projeto

```
.
├── .env                    # Variáveis de ambiente
├── .gitignore
├── .dockerignore
├── docker-compose.yml      # Orquestração
├── docker-bake.hcl        # Build configuration
├── README.md              # Este arquivo
├── producer/
│   ├── Dockerfile
│   ├── main.go
│   └── go.mod
├── consumer/
│   ├── Dockerfile
│   ├── main.go
│   └── go.mod
└── nginx/
    └── nginx.conf
```

## 🎯 Boas Práticas Implementadas

1. **Multi-stage builds**: Reduz tamanho das imagens
2. **Layer caching**: Otimiza tempo de build
3. **Health checks**: Garante disponibilidade
4. **Resource limits**: Previne resource starvation
5. **Network isolation**: Segurança por camadas
6. **Non-root users**: Princípio do menor privilégio
7. **Read-only filesystem**: Imutabilidade
8. **SBOM generation**: Rastreabilidade de dependências
9. **Environment variables**: Configuração flexível
10. **Graceful shutdown**: Finalização limpa

## 📝 Licença

MIT
