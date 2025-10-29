# 🔒 Segurança e Boas Práticas

## Implementações de Segurança

### 1. Containers

#### Non-Root User
Todos os containers Go rodam como usuário não-privilegiado (UID 65534):
```dockerfile
USER 65534:65534
```

#### Read-Only Filesystem
Containers de aplicação usam filesystem read-only:
```yaml
read_only: true
tmpfs:
  - /tmp
```

#### No New Privileges
Previne escalação de privilégios:
```yaml
security_opt:
  - no-new-privileges:true
```

### 2. Imagens Docker

#### Multi-Stage Build
Reduz superfície de ataque e tamanho:
- Build stage: golang:1.23-alpine3.20
- Runtime stage: distroless (apenas binário)

#### Versões Específicas
Nunca usar `latest`:
- Go: 1.23-alpine3.20
- MongoDB: 7.0
- RabbitMQ: 3.13-management-alpine
- NGINX: 1.25-alpine

#### SBOM (Software Bill of Materials)
Gerado automaticamente via buildx:
```bash
docker buildx bake producer-sbom consumer-sbom
```

### 3. Credenciais

#### Sem Hardcoding
Todas as credenciais via variáveis de ambiente:
```bash
MONGO_INITDB_ROOT_USERNAME=${MONGO_INITDB_ROOT_USERNAME}
```

#### Arquivo .env
Nunca commitar .env no git:
```gitignore
.env
```

### 4. Network

#### Isolamento
Network bridge isolada:
```yaml
networks:
  backend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

#### Exposição Mínima
Apenas NGINX e RabbitMQ Management expostos:
```yaml
ports:
  - "80:80"      # NGINX
  - "15672:15672" # RabbitMQ UI
```

### 5. Recursos

#### Limites Configurados
Previne resource exhaustion:
```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 256M
```

### 6. Health Checks

Todos os serviços monitorados:
```yaml
healthcheck:
  test: ["CMD", "wget", "--spider", "http://localhost/health"]
  interval: 10s
  timeout: 3s
  retries: 3
```

## Checklist de Segurança

- [x] Containers rodam como non-root
- [x] Filesystem read-only onde possível
- [x] Security options configuradas
- [x] Versões específicas (não latest)
- [x] Multi-stage builds
- [x] SBOM gerado
- [x] Credenciais via .env
- [x] Networks isoladas
- [x] Limites de recursos
- [x] Health checks implementados
- [x] Logs estruturados
- [x] Graceful shutdown
- [x] TLS/HTTPS (produção)
- [x] Rate limiting (NGINX)
- [x] Input validation

## Docker Scout

### Análise de Vulnerabilidades

```bash
# Scan completo
docker scout cves producer:latest

# Apenas críticas
docker scout cves --only-severity critical producer:latest

# Comparar com base image
docker scout compare --to golang:1.23-alpine3.20 producer:latest
```

### Recomendações

```bash
docker scout recommendations producer:latest
```

## Melhorias para Produção

### 1. Secrets Management
Usar Docker Secrets ou Vault:
```yaml
secrets:
  mongo_password:
    external: true
```

### 2. TLS/HTTPS
Configurar certificados no NGINX:
```nginx
listen 443 ssl http2;
ssl_certificate /etc/nginx/ssl/cert.pem;
ssl_certificate_key /etc/nginx/ssl/key.pem;
```

### 3. Rate Limiting
Já implementado no NGINX:
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
```

### 4. Monitoring
Integrar com Prometheus/Grafana:
```yaml
prometheus:
  image: prom/prometheus
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

### 5. Backup Automatizado
MongoDB backup periódico:
```bash
docker-compose exec mongodb mongodump --out /backup
```

### 6. Log Aggregation
Enviar logs para ELK/CloudWatch:
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Compliance

### OWASP Top 10
- ✅ A01: Broken Access Control
- ✅ A02: Cryptographic Failures
- ✅ A03: Injection
- ✅ A04: Insecure Design
- ✅ A05: Security Misconfiguration
- ✅ A06: Vulnerable Components
- ✅ A07: Authentication Failures
- ✅ A08: Software and Data Integrity
- ✅ A09: Security Logging
- ✅ A10: Server-Side Request Forgery

### CIS Docker Benchmark
Seguindo recomendações do CIS:
- Non-root containers
- Read-only filesystems
- Resource limits
- Health checks
- Minimal base images
- No secrets in images

## Auditoria

### Verificar Configurações
```bash
# Inspecionar container
docker inspect producer-1

# Verificar usuário
docker-compose exec producer-1 whoami

# Verificar filesystem
docker-compose exec producer-1 touch /test
```

### Logs de Segurança
```bash
# Eventos do Docker
docker events

# Logs de acesso NGINX
docker-compose logs nginx | grep -E "POST|GET"
```

## Contato

Para reportar vulnerabilidades de segurança, entre em contato via issue no repositório.
