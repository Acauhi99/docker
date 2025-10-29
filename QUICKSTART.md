# 🚀 Quick Start Guide

## Início Rápido (3 passos)

### 1. Build das Imagens

```bash
./build.sh
```

Ou manualmente:

```bash
docker buildx bake --load
```

### 2. Iniciar Infraestrutura

```bash
docker-compose up -d
```

### 3. Testar

```bash
./test.sh
```

Ou manualmente:

```bash
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
```

## 🔍 Verificar Segurança

```bash
./scout.sh
```

## 📊 Monitorar

```bash
# Logs em tempo real
docker-compose logs -f

# Status dos containers
docker-compose ps

# RabbitMQ Management
open http://localhost:15672
```

## 🛑 Parar

```bash
docker-compose down
```

## 🧹 Limpar Tudo

```bash
docker-compose down -v
docker system prune -a
```

## 📝 Exemplo de Payload

```json
{
  "device": "smartphone",
  "os": "android",
  "tipo": "click",
  "valor": "100",
  "ip": "192.168.1.1",
  "region": "us-east-1"
}
```

## 🎯 Endpoints

- `POST http://localhost/events` - Enviar evento
- `GET http://localhost/health` - Health check NGINX
- `GET http://localhost:15672` - RabbitMQ Management UI

## ⚡ Comandos Úteis

```bash
# Ver logs de um serviço específico
docker-compose logs -f producer-1

# Reiniciar um serviço
docker-compose restart producer-1

# Ver recursos utilizados
docker stats

# Inspecionar network
docker network inspect docker_backend

# Acessar MongoDB
docker-compose exec mongodb mongosh -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD

# Ver mensagens no RabbitMQ
docker-compose exec rabbitmq rabbitmqctl list_queues
```
