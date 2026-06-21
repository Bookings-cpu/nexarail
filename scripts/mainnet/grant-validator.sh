#!/usr/bin/env bash
# Send a mainnet validator grant from ecosystem_grants, log to tracker.
#
# Usage:
#   grant-validator.sh <moniker> <operator_address> [amount_nxrl]
#   grant-validator.sh "Foo Node" nxr1abc... 600
#
# Defaults:
#   amount_nxrl = 600 (standard grant: 500 self-bond + 100 gas)
#   source bucket = ecosystem_grants
#
# Policy locked 2026-06-19. See docs/mainnet/NEW_VALIDATOR_ONBOARDING.md.

set -Eeuo pipefail

BIN="${BIN:-/Users/bradleyjohnston/workspace/nexarail/build/nexaraild}"
# Multi-coord fallback — picks the freshest non-lagging RPC so a single
# overloaded coord (e.g. alpha drowning in external p2p) doesn't block sends.
if [ -z "${RPC:-}" ]; then
  BEST_RPC=""; BEST_HEIGHT=-1
  for r in 32657 32667 32677 32687 32697; do
    h=$(curl -s -m 2 "http://127.0.0.1:$r/status" 2>/dev/null | python3 -c "
import sys,json
try: print(int(json.load(sys.stdin)['result']['sync_info']['latest_block_height']))
except: print(-1)" 2>/dev/null || echo -1)
    if [ "$h" -gt "$BEST_HEIGHT" ]; then BEST_HEIGHT=$h; BEST_RPC="http://127.0.0.1:$r"; fi
  done
  RPC="${BEST_RPC:-http://127.0.0.1:32657}"
fi
HOME_="${HOME_:-$HOME/.nexarail-mainnet-keys}"
SOURCE_KEY="${SOURCE_KEY:-ecosystem_grants}"
CHAIN_ID="${CHAIN_ID:-nexarail-mainnet-1}"
TRACKER="${TRACKER:-/Users/bradleyjohnston/workspace/nexarail/coordination/validators/mainnet-faucet-tracker.csv}"

DEFAULT_AMOUNT_NXRL="${DEFAULT_AMOUNT_NXRL:-600}"

usage() {
  cat <<EOF
Usage: $(basename "$0") <moniker> <operator_address> [amount_nxrl]

  moniker            Validator moniker (use quotes for spaces)
  operator_address   nxr1... bech32 address from operator's request
  amount_nxrl        Optional; defaults to ${DEFAULT_AMOUNT_NXRL} NXRL

Examples:
  $(basename "$0") "Foo Node" nxr1abc...
  $(basename "$0") "Special Big Bond" nxr1xyz... 5000
EOF
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage; exit 1
fi

MONIKER="$1"
TO_ADDR="$2"
AMOUNT_NXRL="${3:-$DEFAULT_AMOUNT_NXRL}"

if [[ ! "$TO_ADDR" =~ ^nxr1[0-9a-z]{38}$ ]]; then
  echo "ERROR: operator_address does not look like a NexaRail bech32 (nxr1... 42 chars total)" >&2
  echo "       got: $TO_ADDR" >&2
  exit 2
fi

if ! [[ "$AMOUNT_NXRL" =~ ^[0-9]+$ ]]; then
  echo "ERROR: amount_nxrl must be a positive integer (NXRL). got: $AMOUNT_NXRL" >&2
  exit 2
fi

AMOUNT_UNXRL=$(( AMOUNT_NXRL * 1000000 ))

echo "=== Mainnet validator grant ==="
echo "  moniker:    $MONIKER"
echo "  to:         $TO_ADDR"
echo "  amount:     ${AMOUNT_NXRL} NXRL  (${AMOUNT_UNXRL} unxrl)"
echo "  from:       $SOURCE_KEY"
echo "  chain:      $CHAIN_ID"
echo "  rpc:        $RPC"
echo

read -r -p "Send? [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "aborted."
  exit 0
fi

SRC_ADDR=$("$BIN" keys show "$SOURCE_KEY" -a --keyring-backend os --home "$HOME_")
echo "source address: $SRC_ADDR"
echo

OUTPUT=$("$BIN" tx send "$SOURCE_KEY" "$TO_ADDR" "${AMOUNT_UNXRL}unxrl" \
  --node "$RPC" \
  --chain-id "$CHAIN_ID" \
  --keyring-backend os \
  --home "$HOME_" \
  --gas auto --gas-adjustment 1.4 --gas-prices 0.025unxrl \
  --output json --yes 2>&1)

echo "$OUTPUT"
TXHASH=$(echo "$OUTPUT" | python3 -c "import sys,json,re
raw=sys.stdin.read()
m=re.search(r'\{.*\"txhash\".*\}', raw, re.S)
print(json.loads(m.group(0))['txhash'] if m else '')" 2>/dev/null || true)

CODE=$(echo "$OUTPUT" | python3 -c "import sys,json,re
raw=sys.stdin.read()
m=re.search(r'\{.*\"code\".*\}', raw, re.S)
print(json.loads(m.group(0)).get('code', 0) if m else 0)" 2>/dev/null || echo 0)

STATUS="ok"
if [[ -z "$TXHASH" ]]; then STATUS="error_no_hash"; fi
if [[ "${CODE:-0}" != "0" ]]; then STATUS="tx_code_${CODE}"; fi

TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
# CSV-escape moniker
MONIKER_ESCAPED=$(printf '%s' "$MONIKER" | sed 's/"/""/g')
printf '%s,"%s",%s,%s,%s,%s,%s,%s,"%s"\n' \
  "$TS" "$MONIKER_ESCAPED" "$TO_ADDR" "$AMOUNT_UNXRL" "$AMOUNT_NXRL" "$SOURCE_KEY" "$TXHASH" "$STATUS" "" >> "$TRACKER"

echo
echo "=== Logged ==="
echo "  txhash:   ${TXHASH:-<none>}"
echo "  status:   $STATUS"
echo "  tracker:  $TRACKER"

if [[ "$STATUS" == "ok" ]]; then
  echo
  echo "Verify on chain:"
  echo "  $BIN query bank balances $TO_ADDR --node $RPC"
fi
