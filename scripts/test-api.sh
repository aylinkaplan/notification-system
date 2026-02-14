#!/usr/bin/env bash
# API + queue test. First: docker-compose up (postgres, rabbitmq, api)

set -e
BASE="${BASE_URL:-http://localhost:8080}"
if command -v jq &>/dev/null; then JQ="jq ."; else JQ="cat"; fi

echo "=== 1. Health ==="
curl -s "$BASE/health" | $JQ

echo -e "\n=== 2. Ready (DB + RabbitMQ) ==="
curl -s "$BASE/ready" | $JQ

echo -e "\n=== 3. Create single notification ==="
CREATE=$(curl -s -X POST "$BASE/notifications" \
  -H "Content-Type: application/json" \
  -d '{"recipient":"+905551234567","channel":"sms","content":"Hello test!","priority":"high"}')
echo "$CREATE" | $JQ
ID=$(echo "$CREATE" | jq -r '.id' 2>/dev/null || true)
[ -z "$ID" ] || [ "$ID" = "null" ] && ID=$(echo "$CREATE" | grep -oE '"id":"[a-f0-9-]{36}"' | head -1 | sed 's/"id":"//;s/"//')
if [ -z "$ID" ]; then
  echo "Error: could not get notification id."
  exit 1
fi

echo -e "\n=== 4. Get by ID ==="
curl -s "$BASE/notifications/$ID" | $JQ

echo -e "\n=== 4b. Wait for worker (10s), then status should be delivered ==="
sleep 10
GET=$(curl -s "$BASE/notifications/$ID")
echo "$GET" | $JQ
STATUS=$(echo "$GET" | jq -r '.status' 2>/dev/null || echo "$GET" | grep -oE '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ "$STATUS" = "failed" ]; then
  echo "ERROR: Notification stayed 'failed'. Rebuild API: docker-compose up -d --build api"
  exit 1
fi
echo "OK: status=$STATUS"

echo -e "\n=== 5. List ==="
curl -s "$BASE/notifications?limit=5" | $JQ

echo -e "\n=== 6. Metrics ==="
curl -s "$BASE/metrics" | $JQ

echo -e "\n=== 7. Batch (2 items) ==="
BATCH=$(curl -s -X POST "$BASE/notifications/batch" \
  -H "Content-Type: application/json" \
  -d '{"notifications":[{"recipient":"a@t.com","channel":"email","content":"B1"},{"recipient":"b@t.com","channel":"email","content":"B2"}]}')
echo "$BATCH" | $JQ
BATCH_ID=$(echo "$BATCH" | jq -r '.batch_id' 2>/dev/null || true)
[ -z "$BATCH_ID" ] || [ "$BATCH_ID" = "null" ] && BATCH_ID=$(echo "$BATCH" | grep -oE '"batch_id":"[a-f0-9-]{36}"' | head -1 | sed 's/"batch_id":"//;s/"//')
if [ -n "$BATCH_ID" ] && [ "$BATCH_ID" != "null" ]; then
  echo -e "\n=== 8. Get batch ==="
  curl -s "$BASE/notifications/batch/$BATCH_ID" | $JQ
  echo -e "\n=== 8b. Wait for worker to process batch (10s), no failed expected ==="
  sleep 10
  BATCH_GET=$(curl -s "$BASE/notifications/batch/$BATCH_ID")
  echo "$BATCH_GET" | $JQ
  FAILED_COUNT=$(echo "$BATCH_GET" | jq '[.notifications[]? | select(.status=="failed")] | length' 2>/dev/null || echo "0")
  if [ "${FAILED_COUNT:-1}" != "0" ]; then
    echo "ERROR: Batch contains failed notifications. Rebuild API: docker-compose up -d --build api"
    exit 1
  fi
  echo "OK: batch notifications are not failed"
fi

echo -e "\n=== 9. Metrics again ==="
curl -s "$BASE/metrics" | $JQ

echo -e "\n--- Done ---"
