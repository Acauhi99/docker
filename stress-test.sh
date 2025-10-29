#!/bin/bash

set -e

echo "=== Stress Test - Docker POC ==="
echo ""

# Configurações
TOTAL_REQUESTS=10000
CONCURRENT=100
URL="http://localhost/events"
RESULTS_FILE="stress-test-results.txt"

# Verificar se o sistema está rodando
echo "1. Verificando se o sistema está rodando..."
if ! curl -s http://localhost/health > /dev/null; then
    echo "Error: Sistema não está respondendo. Execute 'make up' primeiro."
    exit 1
fi
echo "OK - Sistema está rodando"
echo ""

# Verificar se hey está instalado
if ! command -v hey &> /dev/null; then
    echo "Instalando hey..."
    go install github.com/rakyll/hey@latest
fi

# Limpar resultados anteriores
rm -f $RESULTS_FILE

echo "2. Iniciando monitoramento de recursos..."
docker stats --no-stream > docker-stats-before.txt
echo ""

echo "3. Executando stress test..."
echo "   - Total de requisições: $TOTAL_REQUESTS"
echo "   - Requisições concorrentes: $CONCURRENT"
echo "   - Endpoint: $URL"
echo ""

# Executar stress test
hey -n $TOTAL_REQUESTS -c $CONCURRENT -m POST \
  -H "Content-Type: application/json" \
  -d '{"device":"stress-test","os":"linux","tipo":"load","valor":"1","ip":"127.0.0.1","region":"local"}' \
  $URL | tee $RESULTS_FILE

echo ""
echo "4. Coletando estatísticas finais..."
sleep 5
docker stats --no-stream > docker-stats-after.txt

echo ""
echo "5. Verificando processamento..."

# Verificar fila do RabbitMQ
QUEUE_SIZE=$(docker compose exec -T rabbitmq rabbitmqctl list_queues 2>/dev/null | grep events_queue | awk '{print $2}')
echo "   - Mensagens na fila: $QUEUE_SIZE"

# Verificar MongoDB
MONGO_COUNT=$(docker compose exec -T mongodb mongosh -u admin -p secure_password_123 --authenticationDatabase admin --quiet --eval "db.getSiblingDB('events_db').events.countDocuments()" 2>/dev/null)
echo "   - Eventos no MongoDB: $MONGO_COUNT"

# Verificar logs de erro
ERROR_COUNT=$(docker compose logs 2>/dev/null | grep -i error | wc -l)
echo "   - Erros nos logs: $ERROR_COUNT"

echo ""
echo "=== Resumo do Teste ==="
echo ""

# Extrair métricas do resultado
if [ -f $RESULTS_FILE ]; then
    echo "Performance:"
    grep "Requests/sec:" $RESULTS_FILE
    grep "Total:" $RESULTS_FILE | head -1
    grep "Slowest:" $RESULTS_FILE
    grep "Fastest:" $RESULTS_FILE
    grep "Average:" $RESULTS_FILE
    
    echo ""
    echo "Status Codes:"
    grep -A 5 "Status code distribution:" $RESULTS_FILE | tail -5
fi

echo ""
echo "Recursos (Antes vs Depois):"
echo "ANTES:"
cat docker-stats-before.txt | grep -E "producer|consumer|nginx|mongodb|rabbitmq"
echo ""
echo "DEPOIS:"
cat docker-stats-after.txt | grep -E "producer|consumer|nginx|mongodb|rabbitmq"

echo ""
echo "=== Análise ==="

# Calcular taxa de sucesso
SUCCESS_RATE=$(grep "200" $RESULTS_FILE | awk '{print $3}' | head -1)
if [ ! -z "$SUCCESS_RATE" ]; then
    echo "Taxa de sucesso: $SUCCESS_RATE"
fi

# Verificar se há mensagens pendentes
if [ "$QUEUE_SIZE" -gt 100 ]; then
    echo "AVISO: Muitas mensagens pendentes na fila ($QUEUE_SIZE)"
fi

# Verificar se há erros
if [ "$ERROR_COUNT" -gt 10 ]; then
    echo "AVISO: Muitos erros detectados nos logs ($ERROR_COUNT)"
fi

echo ""
echo "Arquivos gerados:"
echo "  - $RESULTS_FILE (resultados completos)"
echo "  - docker-stats-before.txt (recursos antes)"
echo "  - docker-stats-after.txt (recursos depois)"
echo ""
echo "Para ver logs detalhados: docker-compose logs"
echo "Para ver recursos em tempo real: docker stats"
