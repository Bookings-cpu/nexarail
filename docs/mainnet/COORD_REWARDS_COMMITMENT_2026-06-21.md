# Coord Staking Rewards — Custody Commitment (v2)

**Date:** 2026-06-21
**Chain:** `nexarail-mainnet-1`
**Related governance:** Proposal #1 (Enable NXRL staking yield via mint module)
**Related ceremony plan:** `docs/mainnet/DECENTRALISATION_CEREMONY_PLAN.md`

**v2 update 2026-06-21 evening:** Founder confirmed the ceremony mechanism is **delegation, not transfer**. Updated wording throughout to reflect that coord rewards are custodied and then **delegated** at the Cut-Over, not transferred to external validators. Token ownership stays with coord operator wallets; external validators receive commission (default 10%) on the delegated stake.

## The commitment

The five NexaRail coordinator validators — **alpha, bravo, charlie, delta, echo** — operated by Bradley Johnston, commit publicly that:

**All NXRL staking rewards accrued to the five coord validators between proposal #1 enabling staking yield (executes 2026-06-26) and The Cut-Over ceremony at 100 active validators will be pooled, custodied, and delegated to external validators during the Cut-Over.**

Coord operators will not withdraw, sell, restake to coord validators, or otherwise consume coord-attributable staking rewards for personal use during this window. Withdrawn rewards may be moved into segregated custody wallets pending the Cut-Over.

## Mechanism (the important distinction)

The commitment is to **delegate** custodied coord rewards to external validators at the Cut-Over, not to **transfer** them.

| What changes at the Cut-Over | What does NOT change |
|---|---|
| Coord self-bond: 5M → 2.5M each | Coord operator wallets retain economic ownership of all NXRL |
| Coord-attributable bonded power: 25M → 12.5M | Bradley's total token holdings (~985M NXRL) |
| External validators' bonded power: tiny → ~13M+ | Bradley's governance voting weight (delegators vote with their stake) |
| Validators receiving commission on delegations: 0 → 95 | The 21-day unbonding period (delegations can be unbonded by Bradley if needed) |

This is the **"skin without ceding"** model — block production decentralises across 100 validators, but token control and governance voting stay with the founder. Identical to how major Cosmos holders operate.

## Why

At the moment proposal #1 enables the mint module, the bonded power distribution looks like this:

| Cohort | Bonded NXRL | Power share |
|---|---|---|
| Coord (5 × 5M) | 25,000,000 | 99.97% |
| External (21 active) | ~8,400 | 0.03% |

Staking rewards are paid pro-rata to bonded power. Without the Cut-Over, ~all new mint rewards continue to accrue to coord validators indefinitely. With the Cut-Over, post-ceremony rewards flow proportionally to whoever holds the delegated stake — meaning coord wallets continue to earn yield on delegations to external validators (~90% after commission), while external validators earn commission income (~10%) on top of their own block-signing rewards.

This is honest, sustainable, and verifiable.

## Accrual mechanics (between now and Cut-Over)

- Each coord validator's `WithdrawDelegatorReward` flow remains functional.
- Coord operators may execute withdrawals technically — but withdrawn NXRL is held untouched in coord operator wallets (or moved to dedicated custody wallets), segregated from any other balance movements.
- Coord operator wallets in use:
  - `nxr1aatswzp7m2c9udj6rz4jdvq44dr8mck3fx4g72` (initial_validator_alpha)
  - `nxr1k2t7v2jcakqm3fcf7a6dwpk2q4dxn9dh9f3vfn` (initial_validator_bravo)
  - `nxr18k9f5dpg0tex4yk0ujnn0drtnlgn0chl8xjkzd` (initial_validator_charlie)
  - `nxr1t22amsmmql46luwqxzp933djw4yk7nd0ryk3nx` (initial_validator_delta)
  - `nxr1n4g70g0jm9g3xd7chz3adwngkafl0ge6c05nvr` (initial_validator_echo)

## Distribution formula at Cut-Over

For the pool of (freed coord stake + custodied coord rewards), delegated across non-coord active validators:

- **70%** delegated equally to all non-coord active validators (rewards participation, not stake size)
- **30%** delegated proportionally to non-coord bonded stake (rewards skin in the game)

Specific weighting formula and exact amounts are pre-locked here; the only variables are (a) the active set at the Cut-Over block and (b) the total custodied reward amount, both of which are read from on-chain state at execution time.

## Transparency

- The status page will surface a "coord reward accrual" line item showing the running total of withdrawn but uncommitted rewards.
- Anyone can verify the totals by querying:
  - `nexaraild query distribution rewards <coord-valoper>`
  - `nexaraild query bank balances <coord-operator-wallet>`
- The `scripts/mainnet/ceremony-state.sh` script prints the live Cut-Over arithmetic at any chain height.

## What this is not

- **Not a transfer of ownership.** Coord wallets retain economic ownership of the NXRL. External validators receive commission income (default 10%) on the delegated stake, not the stake itself.
- **Not a permanent delegation.** Standard 21-day unbonding applies. If a specific external validator misbehaves (downtime, double-sign, governance abuse), Bradley can unbond and redelegate to a different validator.
- **Not a guarantee of any specific yield amount.** Custodied rewards depend on actual inflation rate (5%-12% floating based on bonded ratio), bonded power changes over time, and the time to reach 100 validators.
- **Not an investment offer or token distribution.** Delegations are to validators actively securing the chain at the Cut-Over block, not to third parties, holders, or applicants.

## Verifiable record

This document is committed to the public NexaRail repository at https://github.com/Bookings-cpu/nexarail with a timestamped git history. Proposal #1's metadata field references the related scoping doc.

A signed message asserting this commitment, broadcast from each coord operator wallet, will be added below once the proposal passes (2026-06-26).
