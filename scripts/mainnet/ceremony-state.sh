#!/usr/bin/env bash
# Decentralisation ceremony — current-state simulator.
#
# Prints what the ceremony arithmetic would look like if it fired right
# now: validator counts, coord vs non-coord bonded, custodied coord
# rewards, per-validator distribution under the 70/30 formula, and
# resulting power shares.
#
# Read-only. Does not mutate chain state.

set -Eeuo pipefail

BIN="${BIN:-/Users/bradleyjohnston/workspace/nexarail/build/nexaraild}"

# Multi-coord fallback (mirrors collector/faucet scripts).
BEST_RPC=""; BEST_H=-1
for r in 32657 32667 32677 32687 32697; do
  h=$(curl -s -m 2 "http://127.0.0.1:$r/status" 2>/dev/null | python3 -c "
import sys,json
try: print(int(json.load(sys.stdin)['result']['sync_info']['latest_block_height']))
except: print(-1)" 2>/dev/null || echo -1)
  if [ "$h" -gt "$BEST_H" ]; then BEST_H=$h; BEST_RPC="http://127.0.0.1:$r"; fi
done
RPC="${RPC:-$BEST_RPC}"

echo "=========================================="
echo "  DECENTRALISATION CEREMONY — STATE PROBE"
echo "  $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "  chain height: $BEST_H"
echo "=========================================="

# Validator set
VALS_JSON=$($BIN query staking validators --node "$RPC" --output json --limit 200 2>/dev/null)

python3 <<PY
import json
d = json.loads('''$VALS_JSON''')
vs = d.get('validators', [])
bonded = [v for v in vs if v.get('status') == 'BOND_STATUS_BONDED']

def is_coord(v):
    return 'controlled' in v.get('description', {}).get('moniker', '').lower()

coord = [v for v in bonded if is_coord(v)]
noncoord = [v for v in bonded if not is_coord(v)]

coord_total = sum(int(v.get('tokens', '0')) for v in coord)
noncoord_total = sum(int(v.get('tokens', '0')) for v in noncoord)
total = coord_total + noncoord_total

def nxrl(unxrl): return unxrl / 1_000_000

print(f"  Active validators: {len(bonded)}  ({len(coord)} coord + {len(noncoord)} non-coord)")
print(f"  Coord bonded:      {nxrl(coord_total):>15,.2f} NXRL  ({coord_total/total*100:.4f}%)")
print(f"  Non-coord bonded:  {nxrl(noncoord_total):>15,.2f} NXRL  ({noncoord_total/total*100:.4f}%)")
print(f"  Total bonded:      {nxrl(total):>15,.2f} NXRL")
print()

if len(noncoord) < 95:
    print(f"  Ceremony trigger requires ≥ 95 non-coord active validators.")
    print(f"  Currently at {len(noncoord)}. Need {95-len(noncoord)} more.")
    print()
    print("  At current onboarding pace (~15/day), ceremony trigger arrives in")
    print(f"  ~{max(0,95-len(noncoord))/15:.1f} days from today.")
    print()

# Ceremony parameters (from commitment doc + ceremony plan)
COORD_TARGET_NXRL = 1_000_000  # default per plan
EQ_WEIGHT = 0.70
PROP_WEIGHT = 0.30

# Custodied coord rewards — estimated as 0 today (proposal 1 hasn't executed yet)
# After 2026-06-26 the proposal executes; from then accrual begins.
custody_nxrl = 0  # TODO: query coord operator wallets + sum un-restaked withdrawn rewards

# Coord stake freed
coord_target_total = COORD_TARGET_NXRL * len(coord)
freed_from_coord = max(0, nxrl(coord_total) - coord_target_total)

# Total ceremony pool
pool = freed_from_coord + custody_nxrl

print(f"  -- Ceremony arithmetic (target coord self-bond: {COORD_TARGET_NXRL:,} NXRL each) --")
print(f"  Coord stake freed:     {freed_from_coord:>15,.2f} NXRL")
print(f"  Custodied rewards:     {custody_nxrl:>15,.2f} NXRL (proposal 1 not yet executed)")
print(f"  Total ceremony pool:   {pool:>15,.2f} NXRL")
print()

if len(noncoord) > 0:
    eq_share = pool * EQ_WEIGHT / len(noncoord)
    print(f"  Equal share (70%):     {pool*EQ_WEIGHT:>15,.2f} NXRL across {len(noncoord)} validators")
    print(f"      = {eq_share:,.2f} NXRL per validator")
    if noncoord_total > 0:
        print(f"  Proportional (30%):    {pool*PROP_WEIGHT:>15,.2f} NXRL by stake")
        print(f"      avg per validator: {pool*PROP_WEIGHT/len(noncoord):,.2f} NXRL")
    print()

    print(f"  -- Per-validator ceremony bonus (sorted by current bond) --")
    for v in sorted(noncoord, key=lambda v: -int(v.get('tokens','0'))):
        m = v['description']['moniker'][:24]
        s = int(v['tokens'])
        prop_share = (pool * PROP_WEIGHT) * (s / noncoord_total) if noncoord_total else 0
        bonus = eq_share + prop_share
        new_bond = s/1_000_000 + bonus
        print(f"    {m:24s}  current={s/1_000_000:>10,.0f}  +bonus={bonus:>12,.2f}  -> new={new_bond:>12,.2f}")
    print()

# Post-ceremony power shares
new_coord_total = coord_target_total * 1_000_000  # back to unxrl
new_noncoord_total = noncoord_total + int(pool * 1_000_000)
new_total = new_coord_total + new_noncoord_total
print(f"  -- Post-ceremony power shares --")
print(f"  Coord:     {new_coord_total/1_000_000:>15,.2f} NXRL  ({new_coord_total/new_total*100:.2f}%)")
print(f"  Non-coord: {new_noncoord_total/1_000_000:>15,.2f} NXRL  ({new_noncoord_total/new_total*100:.2f}%)")
print()

byz = new_noncoord_total > new_total * 1/3
maj = new_noncoord_total > new_total * 1/2
crit = new_noncoord_total > new_total * 2/3
print(f"  Decentralisation milestones reached post-ceremony:")
print(f"    Liveness independent (>33%):  {'YES' if byz else 'no'}")
print(f"    Coord-minority (>50%):        {'YES' if maj else 'no'}")
print(f"    Coord-irrelevant (>67%):      {'YES' if crit else 'no'}")
PY
