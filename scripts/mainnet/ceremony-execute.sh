#!/usr/bin/env bash
# THE CUT-OVER — execution generator.
#
# Generates the actual tx batch that fires the Cut-Over ceremony based on
# the live chain state. By default, prints a plan summary + JSON manifest
# of every tx that would fire. With --json, prints the full manifest only.
#
# This is the inverse of ceremony-state.sh — that one simulates and prints
# arithmetic; this one builds the actual tx manifest for execution.
#
# Mechanism (skin without ceding, locked 2026-06-21):
#   1. Each coord wallet undelegates from its own validator by (5M - target).
#   2. Each coord wallet withdraws all accumulated rewards.
#   3. Liquid pool is delegated across non-coord active validators using
#      the 70/30 formula. Pool is split equally across the 5 coord wallets
#      so each fans out ~equally — keeps batches small and parallelisable.
#
# Usage:
#   bash ceremony-execute.sh                          # plan summary + first 10 txs
#   bash ceremony-execute.sh --json                   # full JSON manifest only
#   bash ceremony-execute.sh --target-self-bond 2000000   # override default 2.5M
#
# Broadcasting the actual txs is a deliberately separate manual step (production
# safety): take the manifest, review, sign + broadcast each tx via the coord
# operator key, with proper sequence handling. A future hardened version will
# wrap that loop in a single command behind an explicit confirmation prompt.

set -Eeuo pipefail

BIN="${BIN:-/Users/bradleyjohnston/workspace/nexarail/build/nexaraild}"
HOME_KEYS="${HOME_KEYS:-$HOME/.nexarail-mainnet-keys}"
COORD_TARGET_NXRL="${COORD_TARGET_NXRL:-2500000}"
EQ_WEIGHT="${EQ_WEIGHT:-0.70}"
PROP_WEIGHT="${PROP_WEIGHT:-0.30}"
CHAIN_ID="${CHAIN_ID:-nexarail-mainnet-1}"
OUTPUT_MODE="summary"

while [ $# -gt 0 ]; do
  case "$1" in
    --json) OUTPUT_MODE="json"; shift;;
    --target-self-bond) COORD_TARGET_NXRL="$2"; shift 2;;
    *) echo "unknown arg: $1" >&2; exit 1;;
  esac
done

# Pick freshest coord RPC (multi-coord fallback). Defensive: pipefail aware.
set +o pipefail
BEST_RPC=""; BEST_H=-1
for r in 32657 32667 32677 32687 32697; do
  resp=$(curl -s -m 2 "http://127.0.0.1:$r/status" 2>/dev/null || true)
  if [ -z "$resp" ]; then continue; fi
  h=$(printf '%s' "$resp" | python3 -c "
import sys, json
try:
    print(int(json.load(sys.stdin)['result']['sync_info']['latest_block_height']))
except Exception:
    print(-1)
" 2>/dev/null || echo -1)
  h="${h:--1}"
  if [ "$h" -gt "$BEST_H" ] 2>/dev/null; then
    BEST_H="$h"
    BEST_RPC="http://127.0.0.1:$r"
  fi
done
set -o pipefail
RPC="${RPC:-$BEST_RPC}"

if [ -z "$RPC" ]; then
  echo "ERROR: no coord RPC reachable" >&2
  exit 1
fi

# Write validators JSON to a tmpfile. Coord RPCs intermittently EOF under load;
# retry across all 5 coord nodes until one returns a non-empty body.
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
GOT=""
for try_rpc in "$RPC" http://127.0.0.1:32667 http://127.0.0.1:32687 http://127.0.0.1:32697 http://127.0.0.1:32677 http://127.0.0.1:32657; do
  : > "$TMP"
  $BIN query staking validators --node "$try_rpc" --output json --limit 200 > "$TMP" 2>/dev/null || true
  if [ -s "$TMP" ] && head -c 12 "$TMP" | grep -q '"validators"'; then
    GOT="$try_rpc"
    break
  fi
done

if [ -z "$GOT" ]; then
  echo "ERROR: validators query failed on all coord RPCs" >&2
  exit 1
fi
RPC="$GOT"

# Generate manifest as JSON via python (reads tmpfile, prints to stdout).
PLAN_JSON=$(COORD_TARGET_NXRL="$COORD_TARGET_NXRL" EQ_WEIGHT="$EQ_WEIGHT" PROP_WEIGHT="$PROP_WEIGHT" python3 <<PYEOF
import json, os, sys

with open("$TMP") as f:
    d = json.load(f)

vs = d.get('validators', [])
bonded = [v for v in vs if v.get('status') == 'BOND_STATUS_BONDED']

def is_coord(v):
    return 'controlled' in v.get('description', {}).get('moniker', '').lower()

coord = sorted([v for v in bonded if is_coord(v)], key=lambda v: v['description']['moniker'])
noncoord = sorted([v for v in bonded if not is_coord(v)], key=lambda v: v['operator_address'])

COORD_TARGET_NXRL = int(os.environ['COORD_TARGET_NXRL'])
COORD_TARGET_UNXRL = COORD_TARGET_NXRL * 1_000_000
EQ_WEIGHT = float(os.environ['EQ_WEIGHT'])
PROP_WEIGHT = float(os.environ['PROP_WEIGHT'])

def coord_to_op_name(moniker):
    short = moniker.lower().replace('nxrl-controlled-', '')
    return f'initial_validator_{short}'

plan = {
    'meta': {
        'noncoord_count': len(noncoord),
        'coord_count': len(coord),
        'target_per_coord_unxrl': COORD_TARGET_UNXRL,
        'target_per_coord_nxrl': COORD_TARGET_NXRL,
        'eq_weight': EQ_WEIGHT,
        'prop_weight': PROP_WEIGHT,
    },
    'undelegates': [],
    'withdraw_rewards': [],
    'delegations': [],
}

# Step 1: undelegate
for v in coord:
    cur = int(v['tokens'])
    if cur > COORD_TARGET_UNXRL:
        plan['undelegates'].append({
            'from_key': coord_to_op_name(v['description']['moniker']),
            'validator': v['operator_address'],
            'amount_unxrl': cur - COORD_TARGET_UNXRL,
        })

# Step 2: withdraw rewards
for v in coord:
    plan['withdraw_rewards'].append({
        'from_key': coord_to_op_name(v['description']['moniker']),
        'validator': v['operator_address'],
    })

# Step 3: delegations (split across 5 coord wallets per recipient)
total_freed = sum(u['amount_unxrl'] for u in plan['undelegates'])
custodied = 0  # placeholder; at execution time read from coord operator wallets
pool = total_freed + custodied
noncoord_total = sum(int(v['tokens']) for v in noncoord)

if len(noncoord) > 0 and len(coord) > 0:
    eq_share = pool * EQ_WEIGHT / len(noncoord)
    for v in noncoord:
        prop = (pool * PROP_WEIGHT) * (int(v['tokens']) / noncoord_total) if noncoord_total else 0
        total_delegation = int(eq_share + prop)
        per_coord = total_delegation // len(coord)
        for cv in coord:
            plan['delegations'].append({
                'from_key': coord_to_op_name(cv['description']['moniker']),
                'validator': v['operator_address'],
                'recipient_moniker': v['description']['moniker'],
                'amount_unxrl': per_coord,
            })

print(json.dumps(plan, indent=2))
PYEOF
)

if [ "$OUTPUT_MODE" = "json" ]; then
  echo "$PLAN_JSON"
  exit 0
fi

echo "=========================================="
echo "  THE CUT-OVER — EXECUTION GENERATOR"
echo "  $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "  RPC: $RPC  (h=$BEST_H)"
echo "  target coord self-bond: $COORD_TARGET_NXRL NXRL each"
echo "=========================================="
echo

echo "$PLAN_JSON" | python3 <<'PYSUM'
import sys, json
plan = json.load(sys.stdin)
m = plan['meta']
print(f"  Plan summary")
print(f"    coord count:              {m['coord_count']}")
print(f"    non-coord active count:   {m['noncoord_count']}")
print(f"    target coord self-bond:   {m['target_per_coord_nxrl']:,} NXRL each")
print(f"    undelegate txs:           {len(plan['undelegates'])}")
print(f"    withdraw-reward txs:      {len(plan['withdraw_rewards'])}")
print(f"    delegate txs:             {len(plan['delegations'])}  (= {m['noncoord_count']} externals × {m['coord_count']} coord wallets)")
total_pool = sum(u['amount_unxrl'] for u in plan['undelegates'])
print(f"    total stake to fan out:   {total_pool/1_000_000:,.0f} NXRL  (plus accumulated rewards at execution time)")
print()
print("  Undelegates:")
for u in plan['undelegates']:
    print(f"    {u['from_key']:25s}  undelegate {u['amount_unxrl']/1_000_000:>12,.0f} NXRL from {u['validator']}")
print()
print("  Withdraw rewards:")
for w in plan['withdraw_rewards']:
    print(f"    {w['from_key']:25s}  withdraw rewards from {w['validator']}")
print(f"\n  Delegations (first 10 of {len(plan['delegations'])}):")
for de in plan['delegations'][:10]:
    print(f"    {de['from_key']:25s}  delegate {de['amount_unxrl']/1_000_000:>10,.2f} NXRL -> {de['recipient_moniker'][:22]:22s}  ({de['validator'][:30]}...)")
print(f"    ... and {len(plan['delegations'])-10} more")
PYSUM

echo
echo "  Full JSON manifest:  bash $0 --json"
echo "  Apply manifest:      [production-safety: separate hardened broadcast script TBD]"
