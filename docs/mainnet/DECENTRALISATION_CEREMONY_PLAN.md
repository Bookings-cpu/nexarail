# The Cut-Over — Decentralisation Ceremony Plan

**Codename:** "The Cut-Over" (railway/infrastructure term for transition between systems)
**Locked by founder:** 2026-06-21
**Status:** Pre-work / planning. Final execution proposal must be submitted to governance ≥14 days before the ceremony block per `COORD_REWARDS_COMMITMENT_2026-06-21.md`.

**Trigger:** When `count(active validators where coord!=true) >= 95`, the ceremony is **scheduled** for a specific block height ≥14 days out via a governance proposal.

## Architectural principle: skin without ceding

The Cut-Over decentralises **block production** without transferring token ownership. Coord operator wallets retain economic ownership; voting weight is distributed across external validators via **delegation** (a standard Cosmos primitive), not transfer.

This is identical to how major Cosmos token holders (a16z, Polychain, Binance Labs) operate on every chain — they delegate their treasury across many validators to support security and decentralisation, retain token ownership and governance voting rights, and pay commission (typically 10%) to the validators that sign on their behalf.

The framing is honest and verifiable: anyone querying the chain sees Bradley's coord wallets delegated to external validators — no tokens leave Bradley's wallets.

## Cut-Over execution — single block window

In one coordinated block:

1. **Reduce coord self-bonds** — 5,000,000 → 2,500,000 NXRL each. Freed tokens (5 × 2,500,000 = 12,500,000 NXRL) return to coord operator wallets as liquid balance. **No tokens leave Bradley's control.**
2. **Withdraw custodied coord rewards** — pull all accumulated staking rewards from coord validators into coord operator wallets.
3. **Delegate freed coord stake + custodied rewards** — coord operator wallets delegate the 12.5M+ pool across the 95+ non-coord active validators, using the 70/30 distribution formula.

## Distribution formula (locked)

For the pool of delegated stake:

- **70%** distributed equally to all non-coord active validators at the Cut-Over block (rewards participation, not stake size)
- **30%** distributed proportionally to non-coord bonded stake (rewards skin in the game from external operators)

This formula was locked in `COORD_REWARDS_COMMITMENT_2026-06-21.md`.

## Concrete example (illustrative — actual values calculated at ceremony block)

Assumptions for example only:
- 95 non-coord active validators at Cut-Over block
- Each at 500 NXRL self-bond (matching today's distribution)
- Custodied coord rewards: ~700,000 NXRL (estimated ~3 months at 12% inflation on 25M coord-bonded)
- Coord reduction: 5M → 2.5M each = 12,500,000 NXRL freed
- Total delegation pool: 13,200,000 NXRL

Per-validator delegation:
- 70% equal: 9,240,000 NXRL / 95 = 97,263 NXRL each
- 30% proportional: 3,960,000 NXRL × (validator's stake / 47,500 total) = ~41,684 NXRL each (assuming equal stakes)
- Per-validator delegation amount: ~138,947 NXRL

Post-Cut-Over power shares:
- Coord self-bonded: 2.5M × 5 = 12,500,000 NXRL (49.9%)
- Non-coord total: 47,500 (own bonds) + 13,200,000 (Bradley's delegations) = 13,247,500 NXRL (50.1%)
- Coord power share: 49.9%
- External power share: 50.1%

**Crosses the >50% external power threshold — coord is no longer the majority.**

## Bradley's position post-Cut-Over

- Tokens owned: **985M+ NXRL — unchanged**
- Tokens bonded under coord validators: 12.5M (down from 25M)
- Tokens delegated to external validators: 13.2M (new)
- **Total votable stake: 25.7M NXRL — increased by ~700k** (the custodied rewards that didn't exist before mint enablement)
- Validators receiving commission on Bradley's delegations: 95
- Commission flowing to validators per year: ~150,000 NXRL/year (10% of ~1.5M annual yield on 12.5M delegated at 12% inflation), spread across 95 validators ≈ 1,579 NXRL/year per validator (in addition to their own self-bond yield)
- Governance voting power: 100% retained (delegators vote with their stake)

## What external validators get

- Block signing rewards: continue earning as before
- Commission income: 10% of yield on Bradley's delegated stake = ~1,500 NXRL/year per validator (additional)
- Higher bonded power: their validator entries on the explorer/status page show much larger bonded amounts, attracting future external delegators
- A real reason to keep their nodes up: stake from a major holder

## Public framing (honest, defensible)

> "On the day NexaRail reached 100 validators, we ran The Cut-Over — the founder delegated 12.5M NXRL of coord-attributable stake across the 95 external validators, halving coord voting share and doubling external. Coord stake remains in coord operator custody as is standard for major token holders on Cosmos chains; the founder retains governance voting rights via the delegated stake. Over time, as external delegators arrive and the ecosystem matures, the founder's relative voting share will dilute naturally."

This framing is true, verifiable on chain, and matches how every Cosmos chain with a major founder allocation actually works (Cosmos Hub, Osmosis, Juno, Stride, Sei — all have founders/foundations delegating significant treasury across the validator set).

## What the Cut-Over does NOT do

- **Does not transfer ownership of NXRL to external validators.** Tokens stay in coord operator wallets. Validators earn commission only.
- **Does not surrender Bradley's governance voting power.** Delegated stake votes with the delegator (Bradley), not the validator.
- **Does not preclude unbonding.** The standard 21-day unbonding window applies to delegations. If a specific external validator misbehaves, Bradley can unbond from them and redelegate to another.

## Execution mechanism

### Approach 1 — Multiple individual txs in 2-3 blocks (simplest, recommend)

Submit a coordinated tx batch:
- 5× `MsgUndelegate` (coord wallets reducing self-bond by 2.5M each)
- 5× `MsgWithdrawDelegatorReward` (coord pulling custodied rewards from their own validators)
- ~95× `MsgDelegate` per coord wallet × 5 coord wallets = 475 delegation txs total (parallelisable)

Block tx limits cap this to 2-3 blocks of execution — fine narratively as a "Cut-Over window" rather than single block.

Pros: standard cosmos primitives, fully auditable, no module changes
Cons: 475 delegation txs is a lot — could optimise by batching into one larger redelegation tx per coord, but cosmos doesn't natively support 1-to-many MsgMultiDelegate. Multi-tx is the path of least resistance.

### Approach 2 — Custom x/treasury orchestration (cleaner if treasury live)

If `x/treasury.LiveEnabled = true` by Cut-Over time, the treasury module can orchestrate the delegation as a single approved disbursement, fanning out to all recipient validators atomically.

Pros: single proposal, single approval, atomic execution
Cons: requires `x/treasury.LiveEnabled = true` first. Could be combined: enable treasury then run Cut-Over in the same governance window.

**Recommendation: Approach 2 if treasury is live, otherwise Approach 1.**

## Inputs needed at Cut-Over block

1. **Active set snapshot** — frozen at the Cut-Over block. The proposal text declares "all non-coord active validators in the active set at block N".
2. **Custodied reward total** — calculated at the day-before-Cut-Over block. Verifiable by querying each coord operator wallet's `WithdrawDelegatorReward` history minus any already-restaked.
3. **Coord reduction target** — locked at 2,500,000 NXRL each per founder decision 2026-06-21.

## Open follow-ups (not blocking ceremony scheduling)

- **Live treasury flag**: do we enable `x/treasury.LiveEnabled` as part of the Cut-Over proposal, or in a separate governance window before? Recommend separate window 2-4 weeks before Cut-Over to validate end-to-end.
- **External cap raise**: currently 500 NXRL bar for new validators. Once Cut-Over delegations bring external validators to ~140k bonded each, this looks healthy.

## Pre-ceremony state simulator

`scripts/mainnet/ceremony-state.sh` prints the current Cut-Over arithmetic on demand. Read-only. Run anytime to see how the ceremony would play out at the current chain state.
