# NexaRail Network — Litepaper Summary

**Chain:** NexaRail Network | **Framework:** Cosmos SDK v0.47 + CometBFT v0.37 | **Coin:** NXRL / `unxrl` | **Prefix:** `nxr`

*Last reconciled against live chain state: 2026-10-04.*

---

## What is NexaRail?

A sovereign Layer 1 blockchain purpose-built for payment and settlement infrastructure: merchant onboarding, programmable settlement, escrow custody, automated payouts, and governance-controlled treasury management.

## Current Status

- **Mainnet is live.** Chain ID `nexarail-mainnet-2`, live since 2026-09-15. ~1.4s block time. Status page: <https://bookings-cpu.github.io/nexarail-status/mainnet.html> · Explorer: <https://bookings-cpu.github.io/nexarail-status/explorer/#/nexarail>
- **19 bonded validators** (5 coordinator-operated, 14 external), total ~25M NXRL bonded. External validator onboarding is open — see `docs/mainnet/NEW_VALIDATOR_ONBOARDING_MAINNET2.md`.
- **Core payment modules are live.** Governance Proposal #1 (passed 2026-09-24) enabled `x/escrow`, `x/payout`, `x/settlement`, and `x/treasury` for real fund movement. See Modules table below for exact flag state.
- **Validator stake is still concentrated.** The five coordinator validators hold the large majority of bonded stake; external validators collectively hold a small fraction. This chain is **not yet meaningfully decentralised** — anyone evaluating NexaRail for custody of real value should weigh this directly rather than assume otherwise. Broadening external stake distribution is an active priority, not a solved problem.
- **No IBC, no CosmWasm, no EVM** are wired yet. NXRL cannot currently move to or interact with any other chain.

## Modules

| Module | Purpose | Live funds |
|---|---|---|
| x/fees | Fee split parameters (60/20/20 validator/treasury/burn) | Parameters live; fee routing not yet intercepting real transaction fees |
| x/merchant | Merchant registration and rebate tiers | Metadata only |
| x/settlement | Payment settlement + fee routing | `live_enabled=true`; treasury-share and burn-share routing flags remain `false` |
| x/escrow | Payment escrow custody | `live_enabled=true` |
| x/payout | Automated payouts | `live_enabled=true` (single-payout execution only; batch payout execution not yet implemented) |
| x/treasury | Protocol treasury + spend execution | `live_enabled=true` |

## Live Funds Safety Model

Each fund-moving module is gated by its own governance-controlled flag. Flags only change via a passed on-chain governance proposal — see Proposal #1 for the one that activated the four modules above. Module accounts (escrow, treasury) remain in the bank blocked-recipients list, so funds in those accounts can only move through the module's own gated logic, never a direct transfer.

## Validator & Consensus

CometBFT validator set, 19 bonded validators as of this reconciliation (5 coordinator, 14 external, growing). Network produces blocks every ~1.4s. Full validator set and live voting power: <https://bookings-cpu.github.io/nexarail-status/explorer/#/nexarail/staking>

## Roadmap

| Item | Status |
|---|---|
| Mainnet-2 launch, core modules live | **Done** |
| Broaden external validator stake distribution | In progress |
| Chain-registry listing, wallet/explorer recognition | In progress |
| IBC | Not started |
| Governance proposal on staking inflation and real fee routing | In progress |
| External security audit | Not started |
| External legal review | Not started |
| DEX listing / liquidity | Not started |

## Critical Disclaimers

- **A** NXRL has never been offered for sale — there has been no public or private token sale.
- **B** Participation is not an investment. No financial returns are promised, expected, or implied beyond standard staking yield set by the chain's own on-chain parameters, which may be zero.
- **C** Decentralisation is limited today — see "Current Status" above. Do not treat validator count alone as a measure of how distributed control of this chain actually is.
- **D** No external security audit has been completed on any custom module (`x/escrow`, `x/payout`, `x/settlement`, `x/treasury`, `x/merchant`, `x/fees`).
- **E** No formal independent legal review has been completed.
- **F** Roadmap items above are provisional — no timeline commitments.

## Full litepaper

`docs/NEXARAIL_LITEPAPER.md` — **note: as of this reconciliation (2026-10-04), the full litepaper still describes the retired `nexarail-mainnet-1` chain and has not yet been rewritten for `nexarail-mainnet-2`. This summary is the current reference; the full document needs a dedicated rewrite before being treated as authoritative.**
