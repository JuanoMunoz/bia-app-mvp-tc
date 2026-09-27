#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080}"
ORIGIN="http://localhost:5173"

fail() { echo "DEMO-CHECK FAILED: $1" >&2; exit 1; }
ok() { echo "ok: $1"; }

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "API en $API_BASE"

curl -sS --fail "$API_BASE/isalive" > /dev/null || fail "isalive sin respuesta"
ok "isalive"

curl -sS --fail "$API_BASE/health" | grep -q '"healthy"' || fail "health no saludable"
ok "health + postgres"

curl -sS --fail -X POST "$API_BASE/api/v1/ai/analyze" > "$tmpdir/analysis.json" || fail "ai/analyze falló"
python3 - "$tmpdir/analysis.json" <<'EOF'
import json, sys
data = json.load(open(sys.argv[1]))
assert data["status"] == "COMPLETED", data
want = {
    "M-109": ("REAL_ANOMALY", "HIGH"),
    "M-104": ("EXPLAINABLE_ANOMALY", "MEDIUM"),
    "M-106": ("FALSE_POSITIVE", "LOW"),
    "M-112": ("DATA_QUALITY", "HIGH"),
}
got = {a["meter_id"]: (a["type"], a["severity"]) for a in data["anomalies"]}
bad = {m: got.get(m) for m in want if got.get(m) != want[m]}
assert not bad, bad
print("anomalias:", len(got))
EOF
ok "ai/analyze + 4 casos dorados"

curl -sS --fail "$API_BASE/api/v1/meters?limit=100" > "$tmpdir/meters.json" || fail "meters falló"
python3 - "$tmpdir/meters.json" <<'EOF'
import json, sys
meters = json.load(open(sys.argv[1]))["meters"]
assert len(meters) == 12, len(meters)
prov = [m for m in meters if m.get("provider")]
assert len(prov) == 12, "perfiles sin poblar"
EOF
ok "12 medidores con perfil de tarifa"

curl -sS --fail "$API_BASE/api/v1/dashboard/summary" > "$tmpdir/summary.json" || fail "summary falló"
python3 - "$tmpdir/summary.json" <<'EOF'
import json, sys
summary = json.load(open(sys.argv[1]))
assert summary["meter_count"] == 12, summary
assert summary["period_start"] and summary["period_end"], summary
print(summary["period_start"], summary["period_end"])
EOF
ok "dashboard/summary"
PERIOD_LINE="$(python3 -c "import json,sys; s=json.load(open(sys.argv[1])); print(s['period_start'], s['period_end'])" "$tmpdir/summary.json")"
PERIOD_START="${PERIOD_LINE%% *}"
PERIOD_END="${PERIOD_LINE##* }"

curl -sS --fail "$API_BASE/api/v1/dashboard/history" > "$tmpdir/history.json" || fail "history falló"
python3 - "$tmpdir/history.json" <<'EOF'
import json, sys
points = json.load(open(sys.argv[1]))["points"]
assert len(points) > 300, len(points)
stamps = [p["timestamp"] for p in points]
assert stamps == sorted(stamps), "puntos sin ordenar"
assert all(p["consumption_kwh"] >= 0 for p in points), "consumo negativo"
print("puntos:", len(points))
EOF
ok "dashboard/history ordenado"

curl -sS --fail -X PATCH -H 'Content-Type: application/json' \
  -d '{"provider":"DEMO","region":"TEST","rate_type":"industrial"}' \
  "$API_BASE/api/v1/meters/M-101" > "$tmpdir/patched.json" || fail "PATCH falló"
python3 - "$tmpdir/patched.json" <<'EOF'
import json, sys
body = json.load(open(sys.argv[1]))
assert (body["meter_id"], body["provider"], body["region"]) == ("M-101", "DEMO", "TEST"), body
EOF
curl -sS --fail -X PATCH -H 'Content-Type: application/json' \
  -d '{"provider":"bia","region":"CUND-EAST","rate_type":"industrial"}' \
  "$API_BASE/api/v1/meters/M-101" > /dev/null || fail "restaurar perfil falló"
ok "PATCH tarifa (ida y vuelta)"

curl -sS -o /dev/null -D "$tmpdir/cors.txt" -X OPTIONS \
  -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: PATCH" \
  "$API_BASE/api/v1/meters" || fail "preflight falló"
grep -qi "access-control-allow-methods:.*PATCH" "$tmpdir/cors.txt" || fail "PATCH no permitido en CORS"
ok "CORS preflight con PATCH"

curl -sS --fail "$API_BASE/api/v1/meters/M-101/estimated-consumption?period_start=$PERIOD_START&period_end=$PERIOD_END" > "$tmpdir/estimate.json" || fail "estimate falló"
python3 - "$tmpdir/estimate.json" <<'EOF'
import json, sys
body = json.load(open(sys.argv[1]))
assert body["total_kwh"] > 0 and body["total_cost"] > 0, body
print("kwh:", round(body["total_kwh"], 1), "costo:", round(body["total_cost"], 1))
EOF
ok "estimación de costo M-101"

echo "ALL_GREEN"
