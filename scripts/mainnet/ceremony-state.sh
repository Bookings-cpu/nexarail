#!/usr/bin/env bash
# The Cut-Over — current-state simulator.
#
# Prints what the Cut-Over arithmetic would look like if it fired right
# now under the "skin without ceding" model: coord self-bond reduction
# + delegation (NOT transfer) of freed coord stake and custodied rewards
# to non-coord active validators.
#
# Read-only. Does not mutate chain state.

set -Eeuo pipefail

BIN="${BIN:-/Users/bradleyjohnston/workspace/nexarail/build/nexaraild}"

# Coord self-bond reduction target (locked 2026-06-21).
COORD_TARGET_NXRL="${COORD_TARGET_NXRL:-2500000}"

# Distribution formula (locked 2026-06-21).
EQ_WEIGHT="${EQ_WEIGHT:-0.70}"
PROP_WEIGHT="${PROP_WEIGHT:-0.30}"

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
echo "  THE CUT-OVER — STATE PROBE"
echo "  $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "  chain height: $BEST_H"
echo "  target coord self-bond: $COORD_TARGET_NXRL NXRL each"
echo "=========================================="

VALS_JSON=$($BIN query staking validators --node "$RPC" --output json --limit 200 2>/dev/null)

python3 <<PY
import json, os

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

COORD_TARGET_NXRL = int(os.environ.get('COORD_TARGET_NXRL', 2_500_000))
EQ_WEIGHT = float(os.environ.get('EQ_WEIGHT', 0.70))
PROP_WEIGHT = float(os.environ.get('PROP_WEIGHT', 0.30))

print(f"  Active validators: {len(bonded)}  ({len(coord)} coord + {len(noncoord)} non-coord)")
print(f"  Coord bonded:      {nxrl(coord_total):>15,.2f} NXRL  ({coord_total/total*100:.4f}%)")
print(f"  Non-coord bonded:  {nxrl(noncoord_total):>15,.2f} NXRL  ({noncoord_total/total*100:.4f}%)")
print(f"  Total bonded:      {nxrl(total):>15,.2f} NXRL")
print()

if len(noncoord) < 95:
    print(f"  Cut-Over trigger requires >= 95 non-coord active validators.")
    print(f"  Currently at {len(noncoord)}. Need {95-len(noncoord)} more.")
    print()
    print(f"  At current onboarding pace (~15/day), trigger arrives in ~{max(0,95-len(noncoord))/15:.1f} days.")
    print()

# Coord self-bond reduction → freed back to coord wallets
coord_target_total = COORD_TARGET_NXRL * len(coord)
freed_from_coord = max(0, nxrl(coord_total) - coord_target_total)

# Custodied rewards: 0 today (proposal 1 hasn't executed yet)
custody_nxrl = 0  # placeholder — read from coord wallets post-mint enablement

# Delegation pool (NOT transfer pool)
pool = freed_from_coord + custody_nxrl

print(f"  -- Cut-Over arithmetic (DELEGATION model — coord retains ownership) --")
print(f"  Coord stake freed:     {freed_from_coord:>15,.2f} NXRL  (returns to coord wallets as liquid)")
print(f"  Custodied rewards:     {custody_nxrl:>15,.2f} NXRL  (proposal 1 not yet executed)")
print(f"  Delegation pool:       {pool:>15,.2f} NXRL  (delegated to externals, NOT transferred)")
print()

if len(noncoord) > 0:
    eq_share = pool * EQ_WEIGHT / len(noncoord)
    print(f"  Equal share (70%):     {pool*EQ_WEIGHT:>15,.2f} NXRL across {len(noncoord)} validators")
    print(f"      = {eq_share:,.2f} NXRL per validator (delegated)")
    if noncoord_total > 0:
        print(f"  Proportional (30%):    {pool*PROP_WEIGHT:>15,.2f} NXRL by stake")
        print(f"      avg per validator: {pool*PROP_WEIGHT/len(noncoord):,.2f} NXRL (delegated)")
    print()

    print(f"  -- Per-validator delegation received (sorted by current bond) --")
    for v in sorted(noncoord, key=lambda v: -int(v.get('tokens','0'))):
        m = v['description']['moniker'][:24]
        s = int(v['tokens'])
        prop_share = (pool * PROP_WEIGHT) * (s / noncoord_total) if noncoord_total else 0
        delegation = eq_share + prop_share
        new_total = s/1_000_000 + delegation
        print(f"    {m:24s}  own={s/1_000_000:>10,.0f}  +delegated={delegation:>12,.2f}  -> total={new_total:>12,.2f}")
    print()

# Post-ceremony power shares (delegations count as bonded power for that validator)
new_coord_total = coord_target_total * 1_000_000  # back to unxrl
new_noncoord_total = noncoord_total + int(pool * 1_000_000)
new_total = new_coord_total + new_noncoord_total

print(f"  -- Post-Cut-Over power shares (block production) --")
print(f"  Coord self-bonded:                    {new_coord_total/1_000_000:>15,.2f} NXRL  ({new_coord_total/new_total*100:.2f}%)")
print(f"  External (own + Bradley delegations): {new_noncoord_total/1_000_000:>15,.2f} NXRL  ({new_noncoord_total/new_total*100:.2f}%)")
print()

# Token ownership shares (governance vote weight follows ownership, not signing validator)
bradley_coord_bonded = new_coord_total
bradley_delegated = int(pool * 1_000_000)
external_own_bonds = noncoord_total
bradley_total_votable = bradley_coord_bonded + bradley_delegated
print(f"  -- Post-Cut-Over governance voting weight (delegators vote with their stake) --")
print(f"  Bradley votable (coord self + delegations): {bradley_total_votable/1_000_000:>15,.2f} NXRL  ({bradley_total_votable/new_total*100:.2f}%)")
print(f"  External operators' own stake:              {external_own_bonds/1_000_000:>15,.2f} NXRL  ({external_own_bonds/new_total*100:.2f}%)")
print()

# Decentralisation framing (block production)
byz_block = new_noncoord_total > new_total / 3
maj_block = new_noncoord_total > new_total / 2
crit_block = new_noncoord_total > new_total * 2 / 3
print(f"  -- Block-production decentralisation milestones --")
print(f"    Liveness independent (external >33%): {'YES' if byz_block else 'no'}")
print(f"    Coord-minority (external >50%):       {'YES' if maj_block else 'no'}")
print(f"    Coord-irrelevant (external >67%):     {'YES' if crit_block else 'no'}")
print()

# Bradley governance veto/control floors
bradley_govt_pct = bradley_total_votable/new_total*100
print(f"  -- Bradley governance protection --")
print(f"    Veto floor (Bradley >33.4%):    {'PROTECTED' if bradley_govt_pct > 33.4 else 'EXPOSED'}  ({bradley_govt_pct:.2f}%)")
print(f"    Majority pass (Bradley >50%):   {'YES' if bradley_govt_pct > 50 else 'no'}")
print(f"    Supermajority (Bradley >67%):   {'YES' if bradley_govt_pct > 67 else 'no'}")
PY
