#!/usr/bin/env bash
#
# Publish DNS-AID (DNS for AI Discovery) records for infolinks.app via Cloudflare API v4.
# Zonefile mirror: dns/dns-aid.zone
#
# Records (idempotent upsert — create or update):
#   _index._agents                SVCB + HTTPS + TXT
#   _a2a._agents                  SVCB + HTTPS
#   _mcp._agents                  SVCB + HTTPS
#   info-links._a2a._agents       SVCB + HTTPS
#   info-links._mcp._agents       SVCB + HTTPS
#
# Usage:
#   CLOUDFLARE_API_TOKEN=... ./scripts/publish-dns-aid.sh
#   CLOUDFLARE_API_TOKEN=... ENABLE_DNSSEC=1 ./scripts/publish-dns-aid.sh
#
# Required token scopes: Zone → DNS → Edit on infolinks.app
# Optional (DNSSEC): Zone → DNSSEC → Edit (+ registrar DS paste still manual)
#
# If you have no API token, paste dns/dns-aid.zone into the Cloudflare dashboard
# and follow the DNSSEC checklist in that file's header comments.

set -euo pipefail

ZONE="${ZONE:-infolinks.app}"
API="https://api.cloudflare.com/client/v4"
TTL=3600
INDEX_URL="https://${ZONE}/.well-known/agents-index.json"

if [[ -z "${CLOUDFLARE_API_TOKEN:-}" ]]; then
  cat <<EOF >&2
CLOUDFLARE_API_TOKEN is not set.

Publish options:
  1. Export a Cloudflare API token with Zone → DNS → Edit for ${ZONE}, then re-run:
       CLOUDFLARE_API_TOKEN=... ./scripts/publish-dns-aid.sh
  2. Manually create the records from dns/dns-aid.zone in the Cloudflare DNS UI.

DNSSEC (for DoH AD=1):
  Cloudflare → DNS → Settings → DNSSEC → Enable, then add the DS record at the
  .app registrar. See comments in dns/dns-aid.zone.
EOF
  exit 1
fi

H_AUTH="Authorization: Bearer ${CLOUDFLARE_API_TOKEN}"
H_TYPE="Content-Type: application/json"

echo "→ Looking up zone ID for ${ZONE}…"
ZONE_ID=$(curl -sS -H "${H_AUTH}" "${API}/zones?name=${ZONE}" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); r=d.get('result') or [];
print(r[0]['id']) if r else sys.exit('zone not found or token lacks Zone:Read')")
echo "  zone_id=${ZONE_ID}"

# Upsert SVCB or HTTPS ServiceMode record.
# Args: type name value  (type is SVCB or HTTPS; name is relative owner label)
upsert_svcb_like() {
  local rrtype=$1 name=$2 value=$3
  local fqdn="${name}.${ZONE}"
  local target="${ZONE}"

  local existing_id
  existing_id=$(curl -sS -H "${H_AUTH}" \
    "${API}/zones/${ZONE_ID}/dns_records?type=${rrtype}&name=${fqdn}" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); r=d.get("result") or []; print(r[0]["id"] if r else "")')

  # Cloudflare wants structured data:{priority,target,value} for SVCB/HTTPS.
  local payload
  payload=$(python3 -c 'import json,sys; print(json.dumps({
    "type": sys.argv[1],
    "name": sys.argv[2],
    "data": {"priority": 1, "target": sys.argv[3], "value": sys.argv[4]},
    "ttl": int(sys.argv[5]),
    "comment": "DNS-AID — agent discovery",
  }))' "${rrtype}" "${fqdn}" "${target}" "${value}" "${TTL}")

  local method url verb
  if [[ -n "${existing_id}" ]]; then
    method=PUT
    url="${API}/zones/${ZONE_ID}/dns_records/${existing_id}"
    verb="Updating"
  else
    method=POST
    url="${API}/zones/${ZONE_ID}/dns_records"
    verb="Creating"
  fi
  echo "→ ${verb} ${rrtype} ${fqdn}…"
  curl -sS -X "${method}" "${url}" \
    -H "${H_AUTH}" -H "${H_TYPE}" \
    --data "${payload}" \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
if d.get('success'):
    r=d['result']
    print(f\"  ✓ {r['name']} {r['type']} {r.get('content') or r.get('data')}\")
else:
    for e in d.get('errors', []):
        print(f\"  ✗ [{e.get('code')}] {e.get('message')}\")
    sys.exit(1)
"
}

upsert_txt() {
  local name=$1 content=$2
  local fqdn="${name}.${ZONE}"

  local existing_id
  existing_id=$(curl -sS -H "${H_AUTH}" \
    "${API}/zones/${ZONE_ID}/dns_records?type=TXT&name=${fqdn}" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); r=d.get("result") or []; print(r[0]["id"] if r else "")')

  local payload
  payload=$(python3 -c 'import json,sys; print(json.dumps({
    "type": "TXT",
    "name": sys.argv[1],
    "content": sys.argv[2],
    "ttl": int(sys.argv[3]),
    "comment": "DNS-AID — agent index",
  }))' "${fqdn}" "${content}" "${TTL}")

  local method url verb
  if [[ -n "${existing_id}" ]]; then
    method=PUT
    url="${API}/zones/${ZONE_ID}/dns_records/${existing_id}"
    verb="Updating"
  else
    method=POST
    url="${API}/zones/${ZONE_ID}/dns_records"
    verb="Creating"
  fi
  echo "→ ${verb} TXT ${fqdn}…"
  curl -sS -X "${method}" "${url}" \
    -H "${H_AUTH}" -H "${H_TYPE}" \
    --data "${payload}" \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
if d.get('success'):
    r=d['result']
    print(f\"  ✓ {r['name']} TXT {r['content']}\")
else:
    for e in d.get('errors', []):
        print(f\"  ✗ [{e.get('code')}] {e.get('message')}\")
    sys.exit(1)
"
}

INDEX_VALUE="alpn=\"h2\" port=443 key65400=\"${INDEX_URL}\" key65409=\"agents-index.json\" mandatory=alpn,port"
A2A_VALUE='alpn="h2" port=443 key65402="a2a" key65409="agent-card.json" mandatory=alpn,port'
MCP_VALUE='alpn="h2" port=443 key65402="mcp" key65409="mcp/server-card.json" mandatory=alpn,port'

for rrtype in SVCB HTTPS; do
  upsert_svcb_like "${rrtype}" "_index._agents" "${INDEX_VALUE}"
  upsert_svcb_like "${rrtype}" "_a2a._agents" "${A2A_VALUE}"
  upsert_svcb_like "${rrtype}" "_mcp._agents" "${MCP_VALUE}"
  upsert_svcb_like "${rrtype}" "info-links._a2a._agents" "${A2A_VALUE}"
  upsert_svcb_like "${rrtype}" "info-links._mcp._agents" "${MCP_VALUE}"
done

upsert_txt "_index._agents" "agents=info-links:a2a,info-links:mcp"

if [[ "${ENABLE_DNSSEC:-0}" == "1" ]]; then
  echo "→ Enabling DNSSEC on the zone…"
  curl -sS -X PATCH "${API}/zones/${ZONE_ID}/dnssec" \
    -H "${H_AUTH}" -H "${H_TYPE}" \
    --data '{"status":"active"}' \
  | python3 -c "
import json,sys
d=json.load(sys.stdin)
if not d.get('success'):
    for e in d.get('errors', []):
        print(f\"  ✗ [{e.get('code')}] {e.get('message')}\")
    sys.exit(1)
r=d['result']
print(f\"  ✓ DNSSEC status: {r.get('status')}\")
ds=r.get('ds')
if ds:
    print('')
    print('  DS record (add at the .app registrar if not using Cloudflare Registrar):')
    print(f'    {ds}')
"
fi

cat <<EOF

Done. Verify with:
  curl -sS 'https://cloudflare-dns.com/dns-query?name=_index._agents.${ZONE}&type=SVCB&do=1' -H 'Accept: application/dns-json'
  curl -sS -X POST https://isitagentready.com/api/scan -H 'Content-Type: application/json' \\
    -d '{"url":"https://${ZONE}"}' | python3 -c \"import json,sys; print(json.load(sys.stdin)['checks']['discoverability']['dnsAid'])\"

Deploy the app (/.well-known/agents-index.json) before or with the DNS key65400 URL.
EOF
