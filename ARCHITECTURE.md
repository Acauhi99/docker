# 🏗️ Arquitetura do Sistema

## Diagrama de Componentes

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENTE / USUÁRIO                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP POST /events
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    NGINX (Load Balancer)                         │
│                    - Port: 80                                    │
│                    - Algorithm: least_conn                       │
│                    - Health checks                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Round-robin
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Producer-1  │    │  Producer-2  │    │  Producer-3  │
│   (Go API)   │    │   (Go API)   │    │   (Go API)   │
│  Port: 8080  │    │  Port: 8080  │    │  Port: 8080  │
│  CPU: 0.5    │    │  CPU: 0.5    │    │  CPU: 0.5    │
│  Mem: 256M   │    │  Mem: 256M   │    │  Mem: 256M   │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                           │ Publish Message
                           │
                           ▼
                  ┌─────────────────┐
                  │    RabbitMQ     │
                  │  (Message Bus)  │
                  │  Port: 5672     │
                  │  UI: 15672      │
                  │  CPU: 1.0       │
                  │  Mem: 512M      │
                  │  Queue: events  │
                  └────────┬────────┘
                           │
                           │ Consume Message
                           │
                           ▼
                  ┌─────────────────┐
                  │    Consumer     │
                  │    (Go API)     │
                  │  Port: 8081     │
                  │  CPU: 0.5       │
                  │  Mem: 256M      │
                  └────────┬────────┘
                           │
                           │ Insert Document
                           │
                           ▼
                  ┌─────────────────┐
                  │    MongoDB      │
                  │   (Database)    │
                  │  Port: 27017    │
                  │  CPU: 1.0       │
                  │  Mem: 512M      │
                  │  DB: events_db  │
                  └─────────────────┘
```

## Fluxo de Dados Detalhado

### 1. Recebimento do Evento
```
Cliente → NGINX → Producer (1, 2 ou 3)
```
- Cliente envia POST /events com JSON
- NGINX distribui usando least_conn
- Producer valida e aceita

### 2. Publicação na Fila
```
Producer → RabbitMQ (Queue: events_queue)
```
- Producer serializa evento
- Publica na fila com persistência
- Retorna 202 Accepted ao cliente

### 3. Processamento Assíncrono
```
RabbitMQ → Consumer → MongoDB
```
- Consumer consome mensagem
- Adiciona timestamp
- Insere no MongoDB
- Confirma processamento (ACK)

## Network Topology

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network: backend                   │
│                    Subnet: 172.20.0.0/16                     │
│                    Driver: bridge                            │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Producer │  │ Producer │  │ Producer │  │ Consumer │   │
│  │    1     │  │    2     │  │    3     │  │          │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       │             │             │             │          │
│       └─────────────┼─────────────┼─────────────┘          │
│                     │             │                        │
│              ┌──────┴─────┐  ┌────┴─────┐                 │
│              │  RabbitMQ  │  │ MongoDB  │                 │
│              └────────────┘  └──────────┘                 │
│                                                            │
│  ┌──────────┐                                             │
│  │  NGINX   │ (Exposed: 0.0.0.0:80)                       │
│  └──────────┘                                             │
└─────────────────────────────────────────────────────────────┘
```

## Volumes e Persistência

```
┌─────────────────────────────────────────────────────────────┐
│                      Docker Volumes                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  mongodb_data        → /data/db                             │
│  mongodb_config      → /data/configdb                       │
│  rabbitmq_data       → /var/lib/rabbitmq                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Health Check Strategy

```
┌──────────────┐
│   Service    │
└──────┬───────┘
       │
       │ Every 10-30s
       │
       ▼
┌──────────────┐
│ Health Check │
└──────┬───────┘
       │
       ├─ Success → healthy
       ├─ Fail (1-3x) → unhealthy
       └─ Timeout → unhealthy
```

### Health Check Endpoints

| Service | Method | Endpoint | Interval |
|---------|--------|----------|----------|
| MongoDB | CMD | mongosh ping | 10s |
| RabbitMQ | CMD | rabbitmq-diagnostics | 10s |
| Producer | HTTP | GET /health | 30s |
| Consumer | HTTP | GET /health | 30s |
| NGINX | HTTP | GET /health | 10s |

## Resource Allocation

```
Total Resources:
├─ CPU: 3.75 cores
│  ├─ Producer (3x): 1.5 cores (0.5 each)
│  ├─ Consumer: 0.5 cores
│  ├─ MongoDB: 1.0 cores
│  ├─ RabbitMQ: 1.0 cores
│  └─ NGINX: 0.25 cores
│
└─ Memory: 2.25 GB
   ├─ Producer (3x): 768 MB (256 each)
   ├─ Consumer: 256 MB
   ├─ MongoDB: 512 MB
   ├─ RabbitMQ: 512 MB
   └─ NGINX: 128 MB
```

## Security Layers

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Network Isolation                                   │
│ - Isolated bridge network                                    │
│ - No direct external access to internal services            │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Container Security                                  │
│ - Non-root user (UID 65534)                                 │
│ - Read-only filesystem                                       │
│ - No new privileges                                          │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Image Security                                      │
│ - Minimal base images (distroless)                           │
│ - Multi-stage builds                                         │
│ - Specific versions                                          │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: Application Security                                │
│ - Input validation                                           │
│ - Secrets via environment                                    │
│ - Graceful shutdown                                          │
└─────────────────────────────────────────────────────────────┘
```

## Scaling Strategy

### Horizontal Scaling

```
Current:
┌──────────┐  ┌──────────┐  ┌──────────┐
│Producer 1│  │Producer 2│  │Producer 3│
└──────────┘  └──────────┘  └──────────┘

Scale Up:
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│Producer 1│  │Producer 2│  │Producer 3│  │Producer 4│  │Producer 5│
└──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘
```

### Load Distribution

```
NGINX Load Balancer (least_conn algorithm)
│
├─ Producer 1: 10 connections → Selected ✓
├─ Producer 2: 15 connections
└─ Producer 3: 12 connections
```

## Message Flow

```
┌─────────┐
│ Request │
└────┬────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 1. Validate JSON                    │
│    - device (required)              │
│    - os (required)                  │
│    - tipo (required)                │
└────┬────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 2. Publish to RabbitMQ              │
│    - Queue: events_queue            │
│    - Persistent: true               │
└────┬────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 3. Return 202 Accepted              │
└─────────────────────────────────────┘

     (Async Processing)

┌─────────────────────────────────────┐
│ 4. Consumer receives message        │
└────┬────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 5. Add timestamp                    │
└────┬────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 6. Insert into MongoDB              │
│    - Database: events_db            │
│    - Collection: events             │
└────┬────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────┐
│ 7. ACK message to RabbitMQ          │
└─────────────────────────────────────┘
```

## Deployment Architecture

```
Development:
└─ docker-compose up -d

Production (Future):
└─ Kubernetes
   ├─ Deployment: producer (replicas: 3-10)
   ├─ Deployment: consumer (replicas: 1-5)
   ├─ StatefulSet: mongodb (replicas: 3)
   ├─ StatefulSet: rabbitmq (replicas: 3)
   └─ Ingress: nginx
```

## Monitoring Points

```
┌─────────────────────────────────────────────────────────────┐
│                     Monitoring Stack                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Health Checks → Docker Compose                             │
│  Logs → docker-compose logs                                 │
│  Metrics → docker stats                                     │
│  RabbitMQ → Management UI (port 15672)                      │
│  NGINX → /nginx_status                                      │
│                                                              │
│  Future:                                                     │
│  - Prometheus (metrics)                                     │
│  - Grafana (dashboards)                                     │
│  - ELK Stack (logs)                                         │
│  - Jaeger (tracing)                                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Failure Scenarios

### Producer Failure
```
Producer-1 fails
    ↓
NGINX detects (health check)
    ↓
Routes to Producer-2 and Producer-3
    ↓
System continues operating
```

### RabbitMQ Failure
```
RabbitMQ fails
    ↓
Producers return 500 error
    ↓
Messages lost (no persistence)
    ↓
Restart RabbitMQ
    ↓
System recovers
```

### MongoDB Failure
```
MongoDB fails
    ↓
Consumer cannot insert
    ↓
Messages requeued (NACK)
    ↓
Restart MongoDB
    ↓
Consumer processes queued messages
```

## Performance Characteristics

```
Throughput:
├─ NGINX: ~10,000 req/s
├─ Producer: ~5,000 req/s (per instance)
├─ RabbitMQ: ~20,000 msg/s
└─ Consumer: ~3,000 msg/s

Latency:
├─ NGINX → Producer: <5ms
├─ Producer → RabbitMQ: <10ms
├─ RabbitMQ → Consumer: <5ms
└─ Consumer → MongoDB: <20ms

Total: ~40ms (async processing)
```
