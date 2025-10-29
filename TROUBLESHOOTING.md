# 🔧 Troubleshooting Guide

## Problemas Comuns e Soluções

### 1. Container não inicia

#### Sintoma
```bash
docker-compose ps
# Status: Exit 1 ou Restarting
```

#### Diagnóstico
```bash
# Ver logs do container
docker-compose logs <service-name>

# Ver últimas 50 linhas
docker-compose logs --tail=50 <service-name>

# Inspecionar container
docker inspect <container-name>
```

#### Soluções Comuns
- Verificar variáveis de ambiente no .env
- Verificar se portas estão disponíveis
- Verificar dependências (depends_on)
- Verificar health checks

---

### 2. Producer não conecta ao RabbitMQ

#### Sintoma
```
Failed to connect to RabbitMQ
```

#### Diagnóstico
```bash
# Verificar se RabbitMQ está healthy
docker-compose ps rabbitmq

# Testar conectividade
docker-compose exec producer-1 ping rabbitmq

# Ver logs do RabbitMQ
docker-compose logs rabbitmq
```

#### Soluções
```bash
# Reiniciar RabbitMQ
docker-compose restart rabbitmq

# Verificar credenciais no .env
RABBITMQ_DEFAULT_USER=<seu_usuario>
RABBITMQ_DEFAULT_PASS=<sua_senha>

# Aguardar health check
docker-compose up -d
sleep 30
```

---

### 3. Consumer não conecta ao MongoDB

#### Sintoma
```
Failed to connect to MongoDB
```

#### Diagnóstico
```bash
# Verificar MongoDB
docker-compose ps mongodb

# Testar conexão
docker-compose exec consumer ping mongodb

# Acessar MongoDB shell
docker-compose exec mongodb mongosh -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD
```

#### Soluções
```bash
# Reiniciar MongoDB
docker-compose restart mongodb

# Verificar credenciais
MONGO_INITDB_ROOT_USERNAME=<seu_usuario>
MONGO_INITDB_ROOT_PASSWORD=<sua_senha>

# Verificar volumes
docker volume ls
docker volume inspect docker_mongodb_data
```

---

### 4. NGINX retorna 502 Bad Gateway

#### Sintoma
```bash
curl http://localhost/events
# 502 Bad Gateway
```

#### Diagnóstico
```bash
# Verificar producers
docker-compose ps | grep producer

# Ver logs do NGINX
docker-compose logs nginx

# Testar producer diretamente
docker-compose exec producer-1 wget -O- http://localhost:8080/health
```

#### Soluções
```bash
# Verificar se producers estão rodando
docker-compose up -d producer-1 producer-2 producer-3

# Verificar configuração do NGINX
docker-compose exec nginx cat /etc/nginx/nginx.conf

# Reiniciar NGINX
docker-compose restart nginx
```

---

### 5. Mensagens não chegam ao MongoDB

#### Sintoma
- Producer aceita mensagens (202)
- MongoDB não recebe dados

#### Diagnóstico
```bash
# Verificar fila do RabbitMQ
docker-compose exec rabbitmq rabbitmqctl list_queues

# Ver logs do consumer
docker-compose logs consumer

# Acessar RabbitMQ UI
open http://localhost:15672
```

#### Soluções
```bash
# Verificar se consumer está rodando
docker-compose ps consumer

# Reiniciar consumer
docker-compose restart consumer

# Verificar se há mensagens na fila
# RabbitMQ UI → Queues → events_queue

# Purgar fila (se necessário)
docker-compose exec rabbitmq rabbitmqctl purge_queue events_queue
```

---

### 6. Erro de permissão (Permission Denied)

#### Sintoma
```
Permission denied: /data/db
```

#### Diagnóstico
```bash
# Verificar volumes
docker volume inspect docker_mongodb_data

# Verificar permissões
docker-compose exec mongodb ls -la /data/db
```

#### Soluções
```bash
# Remover volumes e recriar
docker-compose down -v
docker-compose up -d

# Ou ajustar permissões
docker-compose exec --user root mongodb chown -R mongodb:mongodb /data/db
```

---

### 7. Container consome muita CPU/Memória

#### Sintoma
```bash
docker stats
# CPU > 100% ou Memory > limit
```

#### Diagnóstico
```bash
# Ver estatísticas em tempo real
docker stats

# Ver logs para identificar problema
docker-compose logs --tail=100 <service>
```

#### Soluções
```bash
# Ajustar limites no .env
PRODUCER_CPU_LIMIT=1.0
PRODUCER_MEMORY_LIMIT=512M

# Reiniciar com novos limites
docker-compose down
docker-compose up -d

# Escalar horizontalmente
docker-compose up -d --scale producer=5
```

---

### 8. Build falha

#### Sintoma
```
ERROR: failed to solve
```

#### Diagnóstico
```bash
# Ver logs completos do build
docker buildx bake --progress=plain

# Verificar Dockerfile
cat producer/Dockerfile
```

#### Soluções
```bash
# Limpar cache
docker builder prune -a

# Build sem cache
docker buildx bake --no-cache

# Verificar go.mod e go.sum
cd producer
go mod tidy
```

---

### 9. Network issues

#### Sintoma
```
Cannot resolve hostname
```

#### Diagnóstico
```bash
# Listar networks
docker network ls

# Inspecionar network
docker network inspect docker_backend

# Testar DNS
docker-compose exec producer-1 nslookup rabbitmq
```

#### Soluções
```bash
# Recriar network
docker-compose down
docker network prune
docker-compose up -d

# Verificar se containers estão na mesma network
docker network inspect docker_backend | grep Name
```

---

### 10. Health check sempre unhealthy

#### Sintoma
```bash
docker-compose ps
# Status: unhealthy
```

#### Diagnóstico
```bash
# Ver detalhes do health check
docker inspect <container> | grep -A 20 Health

# Testar health check manualmente
docker-compose exec producer-1 wget -O- http://localhost:8080/health
```

#### Soluções
```bash
# Aumentar timeout no docker-compose.yml
healthcheck:
  timeout: 10s
  start_period: 30s

# Verificar se serviço está realmente rodando
docker-compose logs <service>

# Desabilitar health check temporariamente (debug)
# Comentar seção healthcheck no docker-compose.yml
```

---

## Comandos Úteis para Debug

### Logs
```bash
# Todos os serviços
docker-compose logs -f

# Serviço específico
docker-compose logs -f producer-1

# Últimas N linhas
docker-compose logs --tail=100 consumer

# Desde timestamp
docker-compose logs --since 2024-01-01T10:00:00
```

### Inspecionar
```bash
# Container
docker inspect producer-1

# Network
docker network inspect docker_backend

# Volume
docker volume inspect docker_mongodb_data

# Imagem
docker inspect producer:latest
```

### Executar comandos
```bash
# Shell no container (se disponível)
docker-compose exec producer-1 sh

# Comando específico
docker-compose exec mongodb mongosh

# Como root
docker-compose exec --user root mongodb sh
```

### Recursos
```bash
# Estatísticas em tempo real
docker stats

# Uso de disco
docker system df

# Processos
docker-compose top
```

### Limpeza
```bash
# Parar tudo
docker-compose down

# Parar e remover volumes
docker-compose down -v

# Limpar sistema
docker system prune -a

# Limpar volumes órfãos
docker volume prune
```

---

## Checklist de Troubleshooting

Quando algo não funciona, siga esta ordem:

1. ✅ Verificar se todos os containers estão rodando
   ```bash
   docker-compose ps
   ```

2. ✅ Verificar logs de erro
   ```bash
   docker-compose logs
   ```

3. ✅ Verificar health checks
   ```bash
   docker-compose ps
   ```

4. ✅ Verificar conectividade de rede
   ```bash
   docker network inspect docker_backend
   ```

5. ✅ Verificar variáveis de ambiente
   ```bash
   docker-compose config
   ```

6. ✅ Verificar recursos disponíveis
   ```bash
   docker stats
   ```

7. ✅ Reiniciar serviços problemáticos
   ```bash
   docker-compose restart <service>
   ```

8. ✅ Recriar tudo se necessário
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

---

## Logs de Erro Comuns

### "bind: address already in use"
**Causa**: Porta já está sendo usada

**Solução**:
```bash
# Encontrar processo usando a porta
lsof -i :80
# ou
netstat -tulpn | grep :80

# Matar processo
kill -9 <PID>

# Ou mudar porta no .env
NGINX_PORT=8080
```

### "no space left on device"
**Causa**: Disco cheio

**Solução**:
```bash
# Ver uso de disco
docker system df

# Limpar
docker system prune -a
docker volume prune
```

### "connection refused"
**Causa**: Serviço não está pronto

**Solução**:
```bash
# Aguardar health checks
sleep 30

# Verificar se serviço iniciou
docker-compose logs <service>
```

### "context deadline exceeded"
**Causa**: Timeout na operação

**Solução**:
```bash
# Aumentar timeouts no código
# Verificar performance do sistema
docker stats
```

---

## Suporte

Se o problema persistir:

1. Coletar informações:
```bash
docker-compose ps > debug.txt
docker-compose logs >> debug.txt
docker stats --no-stream >> debug.txt
docker system df >> debug.txt
```

2. Verificar documentação:
- [README.md](README.md)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [SECURITY.md](SECURITY.md)

3. Abrir issue no repositório com:
- Descrição do problema
- Logs relevantes
- Passos para reproduzir
- Ambiente (OS, Docker version)
