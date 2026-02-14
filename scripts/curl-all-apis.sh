#!/usr/bin/env bash
# Try all API endpoints with curl. Default base URL: http://localhost:8080
# Usage: ./scripts/curl-all-apis.sh [BASE_URL]

set -e
BASE="${1:-http://localhost:8080}"
if command -v jq &>/dev/null; then JQ="jq ."; else JQ="cat"; fi
echo "Using BASE_URL=$BASE"
echo ""

# --- Health & readiness ---
echo "=== 1. GET /health ==="
curl -s "$BASE/health" | $JQ
echo ""

echo "=== 2. GET /ready ==="
curl -s "$BASE/ready" | $JQ
echo ""

echo "=== 3. GET /metrics ==="
curl -s "$BASE/metrics" | $JQ
echo ""

# --- Create single notification (save id for later) ---
echo "=== 4. POST /notifications (create single) ==="
CREATE=$(curl -s -X POST "$BASE/notifications" \
  -H "Content-Type: application/json" \
  -d '{"recipient":"+905551234567","channel":"sms","content":"Hello from curl","priority":"high"}')
echo "$CREATE" | $JQ
ID=$(echo "$CREATE" | jq -r '.id' 2>/dev/null || echo "$CREATE" | grep -oE '"id":"[a-f0-9-]{36}"' | head -1 | sed 's/"id":"//;s/"//')
echo "Saved id=$ID"
echo ""

echo "=== 5. GET /notifications/{id} ==="
curl -s "$BASE/notifications/$ID" | $JQ
echo ""

echo "=== 6. GET /notifications (list, limit=5) ==="
curl -s "$BASE/notifications?limit=5" | $JQ
echo ""

echo "=== 7. GET /notifications?status=pending ==="
curl -s "$BASE/notifications?status=pending&limit=3" | $JQ
echo ""

echo "=== 8. GET /notifications?channel=sms ==="
curl -s "$BASE/notifications?channel=sms&limit=3" | $JQ
echo ""

# --- Batch ---
echo "=== 9. POST /notifications/batch ==="
BATCH=$(curl -s -X POST "$BASE/notifications/batch" \
  -H "Content-Type: application/json" \
  -d '{"notifications":[
    {"recipient":"a@test.com","channel":"email","content":"Batch 1"},
    {"recipient":"b@test.com","channel":"email","content":"Batch 2"}
  ]}')
echo "$BATCH" | $JQ
BATCH_ID=$(echo "$BATCH" | jq -r '.batch_id' 2>/dev/null || echo "$BATCH" | grep -oE '"batch_id":"[a-f0-9-]{36}"' | head -1 | sed 's/"batch_id":"//;s/"//')
echo "Saved batch_id=$BATCH_ID"
echo ""

echo "=== 10. GET /notifications/batch/{batchId} ==="
curl -s "$BASE/notifications/batch/$BATCH_ID" | $JQ
echo ""

# --- Create one then cancel (worker may consume fast; cancel right away) ---
echo "=== 11. POST /notifications (for cancel) ==="
CANCEL_CREATE=$(curl -s -X POST "$BASE/notifications" \
  -H "Content-Type: application/json" \
  -d '{"recipient":"+905559999999","channel":"push","content":"To be cancelled","priority":"low"}')
CANCEL_ID=$(echo "$CANCEL_CREATE" | jq -r '.id' 2>/dev/null || echo "$CANCEL_CREATE" | grep -oE '"id":"[a-f0-9-]{36}"' | head -1 | sed 's/"id":"//;s/"//')
echo "Created id=$CANCEL_ID (will try to cancel)"
echo ""

echo "=== 12. DELETE /notifications/{id} (cancel pending) ==="
curl -s -X DELETE "$BASE/notifications/$CANCEL_ID" | $JQ
echo ""

echo "=== 13. GET /metrics (again) ==="
curl -s "$BASE/metrics" | $JQ
echo ""

echo "=== 14. GET /openapi.yaml (first 20 lines) ==="
curl -s "$BASE/openapi.yaml" | head -20
echo "..."
echo ""

echo "--- Done. All endpoints exercised. ---"
echo "Swagger UI: $BASE/docs"
