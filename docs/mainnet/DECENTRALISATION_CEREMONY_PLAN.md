# Decentralisation Ceremony — Plan

**Status:** Pre-work / planning. Final execution proposal must be
submitted to governance ≥14 days before the ceremony block per
`COORD_REWARDS_COMMITMENT_2026-06-21.md`.

**Trigger:** When `count(active validators where coord!=true) >= 95`,
the ceremony is **scheduled** for a specific block height ≥14 days out
via a governance proposal.

**Aim:** In a single coordinated block window, simultaneously:
1. Reduce coord self-delegations from 5,000,000 NXRL each → target TBD
   (likely 1,000,000 NXRL each)
2. Redistribute the freed coord stake AND all custodied coord staking
   rewards to non-coord active validators
3. Convert the moment into a marketed ecosystem milestone

## Why a coordinated block window

If coord stake is reduced over many days, the optics are "stealth
rebalance". If it happens in one block alongside external validator
boosts, the optics are "ceremony" — a marketing moment that
strengthens the chain's positioning and reinforces the on-chain
narrative of progressive decentralisation.

## Distribution formula (locked 2026-06-21 in commitment doc)

For the pool of (freed coord stake + custodied coord rewards):

- **70%** distributed equally to all non-coord active validators at
  the ceremony block (rewards being there, not stake size)
- **30%** distributed proportionally to non-coord bonded stake
  (rewards skin in the game)

This formula is symmetric and easy to verify on chain.

## Concrete example (illustrative — actual values calculated at ceremony block)

Assumptions for example only:
- 95 non-coord active validators at ceremony block
- Each at 500 NXRL self-bond (matching today's distribution)
- Custodied coord rewards: ~3M NXRL (assumes mint at 12% for 90 days
  on 25M coord-bonded stake)
- Coord reduction: 5M → 1M = 4M × 5 = 20M NXRL freed
- Total ceremony pool: 23M NXRL

Distribution:
- 70% equal: 16.1M NXRL / 95 validators = 169,474 NXRL each
- 30% proportional: 6.9M NXRL × (validator's stake / 47,500 total) = ~72,632 NXRL each (assuming equal stakes)
- Per-validator ceremony bonus: ~242,106 NXRL

Post-ceremony state (illustrative):
- Each non-coord validator: 500 + 242,106 = 242,606 NXRL self-bond
- Coord: 1M each × 5 = 5M NXRL bonded
- Non-coord total: 242,606 × 95 = 23,047,570 NXRL bonded
- Total bonded: 28,047,570 NXRL
- Coord power share: 17.83%
- External power share: 82.17%

**Result: chain crosses the >67% external power threshold —
"coord-irrelevant" by the definition in the strategic brief.**

## Execution mechanism

### Approach 1 — Multiple individual txs in a single block (simplest)

Submit ~100 separate txs in one bundle:
- 5× `MsgUndelegate` (coord wallets reducing self-bond)
- 5× `MsgWithdrawDelegatorReward` (coord pulling custodied rewards)
- ~95× `MsgSend` (custody wallet → non-coord operator wallets)
- ~95× `MsgDelegate` (non-coord operators → their own valoper)

Pros: standard cosmos primitives, fully auditable
Cons: blocks have tx limits — would need ~3-4 blocks not 1
strictly. "Ceremony window" of 2-3 blocks is fine narratively.

### Approach 2 — Custom x/treasury distribution (cleaner)

Use the existing `x/treasury` module's spend execution path. Create a
single "milestone disbursement" approved by governance proposal that
fans out to all recipients atomically.

Pros: single proposal, single approval, atomic execution
Cons: requires `x/treasury.LiveEnabled = true` first (currently false
per litepaper). Could be combined: enable treasury then run
ceremony in the same governance window.

**Recommendation: Approach 2.** Lines up the live-flag flip with a
high-stakes use case, plus simpler optics.

### Approach 3 — Custom ceremony module (overkill)

Build a dedicated x/decentralisation module. Reject — single-use
infrastructure, premature.

## Inputs needed before submitting ceremony proposal

1. **Final coord reduction target** — recommend 1M each (current
  scoping); could be 2M for less drastic optics. Bradley's call.
2. **External self-bond cap raise** — currently 500 NXRL is the bar
  for new validators. Ceremony recipients will receive ~240k NXRL
  per the example above. Need to ensure non-coord wallets can accept
  delegations of that size (they can — it's just a `MsgDelegate`).
3. **Distribution recipient list** — frozen at ceremony block. The
  proposal text declares "all non-coord active validators in the
  active set at block N" where N is the ceremony block.
4. **Custodied reward total** — calculated at the day-before-ceremony
  block. Verifiable by querying each coord operator wallet.

## Open questions for Bradley

1. **Coord reduction target**: 1M each (sharp drop) or 2M each
  (gentler)?
2. **Active set during ceremony**: do nothing about the active set
  cap (currently 100)? Or raise to e.g. 150 if we reach 100 fast?
3. **Branding**: name the ceremony? ("NexaRail Devolution", "The
  Handover", "Threshold Event", etc.) — affects marketing copy.
4. **Live treasury flag**: do we enable `x/treasury.LiveEnabled` as
  part of the ceremony proposal, or in a separate governance window
  before?

## Pre-ceremony script

A status script (`scripts/mainnet/ceremony-state.sh`) prints the
current ceremony arithmetic on demand so we can see where we'd land
if the ceremony fired today. Run anytime, no chain mutation.

Outputs:
- Current active validator count
- Coord vs non-coord bonded
- Custodied coord rewards estimate
- Per-validator distribution under the 70/30 formula
- Post-ceremony coord vs external power %
