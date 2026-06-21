# Coord Staking Rewards — Custody Commitment

**Date:** 2026-06-21
**Chain:** `nexarail-mainnet-1`
**Related governance:** Proposal #1 (Enable NXRL staking yield via mint module)
**Related roadmap:** `docs/mainnet/MINT_ENABLEMENT_SCOPING_2026-06-21.md`

## The commitment

The five NexaRail coordinator validators — **alpha, bravo, charlie, delta, echo** — operated by Bradley Johnston, commit publicly that:

**All NXRL staking rewards accrued to the five coord validators between proposal #1 enabling staking yield and the decentralisation ceremony at 100 active validators will be pooled, custodied, and redistributed during the ceremony.**

Coord operators will not withdraw, sell, restake to coord validators, or otherwise consume coord-attributable staking rewards for personal or operational use during this window.

## Why

At the moment proposal #1 enables the mint module, the bonded power
distribution looks like this:

| Cohort | Bonded NXRL | Power share |
|---|---|---|
| Coord (5 × 5M) | 25,000,000 | 99.97% |
| External (21 active) | ~8,400 | 0.03% |

Staking rewards are paid pro-rata to bonded power. Without an explicit
commitment, ~all new mint goes to coord validators. That is mechanically
correct, but optically corrosive — it makes the "open mainnet"
positioning look like a coord-side yield farm with extra steps.

This commitment converts the day-1 coord concentration from a problem
into a feature: coord operators are publicly custodying yield on behalf
of the future ecosystem rather than scooping it for themselves.

## How

### Accrual

- Each coord validator's `WithdrawDelegatorReward` flow remains
  functional (it must, for accounting). Coord operators may execute
  withdrawals technically — but the withdrawn NXRL is held untouched
  in the coord operator wallets, segregated from any other balance
  movements.
- Coord operator wallets in use:
  - `nxr1aatswzp7m2c9udj6rz4jdvq44dr8mck3fx4g72` (initial_validator_alpha)
  - `nxr1k2t7v2jcakqm3fcf7a6dwpk2q4dxn9dh9f3vfn` (initial_validator_bravo)
  - `nxr18k9f5dpg0tex4yk0ujnn0drtnlgn0chl8xjkzd` (initial_validator_charlie)
  - `nxr1t22amsmmql46luwqxzp933djw4yk7nd0ryk3nx` (initial_validator_delta)
  - `nxr1n4g70g0jm9g3xd7chz3adwngkafl0ge6c05nvr` (initial_validator_echo)

### Transparency

- The status page (`https://bookings-cpu.github.io/nexarail-status/`)
  will surface a "coord reward accrual" line item showing the running
  total of withdrawn but uncommitted rewards.
- Anyone can verify the totals by querying
  `nexaraild query distribution rewards <coord-valoper>` and
  `nexaraild query bank balances <coord-operator-wallet>`.

### Redistribution event

At the decentralisation ceremony triggered when the chain reaches
**100 active validators**:

1. All accumulated coord rewards are aggregated.
2. Coord self-delegations are reduced (target TBD, likely 5M → 1M each
   per `docs/...decentralisation-ceremony` plan).
3. The aggregated reward pool plus the released self-delegation tokens
   are redistributed to **all non-coord active validators present at
   the ceremony block**, weighted by:
   - 70% — equal weight (rewards participation, not stake size)
   - 30% — proportional to current bonded stake (rewards skin in the game)
4. Specific weighting formula and exact amounts will be finalised in a
   ceremony-prep governance proposal at least 14 days before the
   ceremony itself.

## What this is not

- **Not a contractual obligation.** This is a public commitment by the
  coord operators on behalf of the project. It is documented on-chain
  via metadata reference and on GitHub. There is no smart-contract
  escrow yet because the chain modules required for that (e.g.
  programmable escrow with time-locks tied to a validator-count
  trigger) are not configured for this use case. If trustless
  enforcement matters to a future stakeholder, the on-chain mechanism
  can be designed before the ceremony.
- **Not a guarantee of any specific yield amount.** Coord rewards
  accrual depends on actual inflation rate (floating between 5%-12%
  based on bonded ratio), bonded power changes over time, and the time
  to reach 100 validators.
- **Not an offer or distribution.** Future redistribution is to
  validators actively securing the chain at the ceremony block, not to
  third parties, holders, or applicants.

## Verifiable record

This document is committed to the public NexaRail repository at
`https://github.com/Bookings-cpu/nexarail` with a timestamped git
history. Proposal #1's metadata field also references the scoping doc
that mentions this commitment.

A signed message asserting this commitment, broadcast from each coord
operator wallet, will be added below once the proposal passes (2026-06-26).
