# NexaRail Network Litepaper

**Version:** 2.0
**Date:** 2026-06-21  (v1.0 was 2026-05-27, pre-mainnet)
**Chain:** `nexarail-mainnet-1` — LIVE since 2026-06-18 18:00 UTC
**Framework:** Cosmos SDK v0.47.17 + CometBFT v0.37.18
**Native Coin:** NXRL (base denom: `unxrl`, 1 NXRL = 1,000,000 unxrl)
**Address prefix:** `nxr`

---

## 1. Status Disclaimer

**Please read carefully.** This litepaper describes a live network under active development.

- **Mainnet is live.** `nexarail-mainnet-1` launched 2026-06-18 18:00 UTC. ~5s blocks, BFT consensus across an open validator set. Status: https://bookings-cpu.github.io/nexarail-status/
- **External validator onboarding is OPEN.** Onboarding runbook at `docs/mainnet/NEW_VALIDATOR_ONBOARDING.md`. The validator set grew from 7 at genesis (5 coord + 2 external) to 23+ external validators within the first week, and continues to grow. New validators receive a 600 NXRL bootstrap grant from `ecosystem_grants` after a sync-proof gate.
- **No token sale.** NXRL has not been offered for sale through any mechanism — no ICO, IEO, IDO, private sale, or public sale. There is no way to purchase NXRL. Any claim to the contrary is fraudulent.
- **No DEX listing yet.** NXRL has no listed price, no liquidity pools, no market quote. DEX listing is on the roadmap but has not occurred. Anyone claiming to sell NXRL is operating fraudulently.
- **Live-funds modules disabled by default.** All modules capable of moving tokens (`x/escrow`, `x/payout`, `x/settlement`, `x/treasury`) have their `LiveEnabled` flags set to `false` by default. These flags require governance approval to activate.
- **Coord-side power concentration is high.** As of mainnet launch, the five coordinator validators (alpha/bravo/charlie/delta/echo) hold approximately 99.97% of bonded voting power. A planned "Cut-Over" ceremony at 100 active validators will rebalance toward >50% external block-production share via founder-side delegation (not token transfer). See `docs/mainnet/DECENTRALISATION_CEREMONY_PLAN.md`.
- **No investment.** Participation in NexaRail mainnet validation is not an investment. No financial returns are promised, expected, or implied beyond standard staking yield from the chain's own inflation parameters.
- **Legal review pending.** Formal independent legal review has not been completed.
- **External security audit pending.** A formal third-party security audit has not been completed.

---

## 2. Executive Summary

NexaRail is a sovereign Layer 1 blockchain built on the Cosmos SDK and CometBFT consensus engine, designed to serve as payment and settlement infrastructure. It targets a set of real-world financial workflows that remain fragmented, slow, or trust-dependent in existing systems.

The network provides purpose-built modules for merchant payment processing, settlement with programmable fee routing, escrow custody, automated payouts, and treasury management. All fund-moving functionality is gated behind governance-controlled flags that default to disabled — live movement of tokens is not active by default.

The native coin is NXRL, with a base denom of `unxrl`. The address prefix is `nxr`.

**Mainnet (`nexarail-mainnet-1`) launched 2026-06-18 18:00 UTC** with a 7-validator genesis set (5 coordinator + 2 external) and has since onboarded additional external validators via an open application process. As of v2.0 of this litepaper, the active set is 23 bonded validators and growing at approximately 15 onboardings per day.

A governance proposal to enable the standard Cosmos SDK mint module (5%-12% floating inflation, 67% bonded target) was submitted on 2026-06-21 and is currently in the voting period. Once executed, this begins continuous validator staking yield.

The decentralisation path is staged: open external validator onboarding (live now), staking yield activation (in vote), a planned "Cut-Over" ceremony at 100 active validators that rebalances voting share via founder-side delegation, external security audit, legal review, and DEX listing.

---

## 3. Problem Statement

### Fragmented merchant payment infrastructure

Merchants accepting payments across multiple channels face a fragmented landscape of payment processors, settlement timelines, fee structures, and reconciliation requirements. Each channel introduces its own settlement cadence, dispute process, and reporting surface.

### Settlement delays

Payment settlement in traditional and crypto-adjacent systems takes hours to days. Funds remain in transit, earning no yield and creating operational friction for businesses that need predictable cash flow.

### Escrow trust issues

Escrow arrangements typically require a trusted third party to hold and release funds. This introduces counterparty risk, manual oversight, and settlement latency. Programmatic escrow on public infrastructure can eliminate the trust requirement but needs clear enforcement mechanisms.

### Payout and treasury opacity

Automated payouts to multiple recipients (affiliates, partners, contractors, team members) are often managed through bespoke scripts, manual transfers, or third-party payout services — each with its own trust model, fee structure, and audit trail. Treasuries managing protocol funds need transparent controls: budgets, grant milestones, spend approval workflows, and execution audit trails.

### Lack of programmable settlement rails

General-purpose blockchains can support payment logic through smart contracts, but the programming overhead, gas costs, and contract security risks create barriers for merchant-focused payment infrastructure. A purpose-built chain with native payment modules can expose these capabilities through deterministic, auditable, governance-controlled module logic rather than general-purpose contract code.

---

## 4. Vision

NexaRail aims to become a **payment and settlement infrastructure chain** — a sovereign L1 where payment flows, escrow arrangements, payout schedules, and treasury controls are first-class protocol primitives rather than application-layer afterthoughts.

The design targets:

- **Merchant-aware blockchain.** Native modules for merchant registration, fee parameters, settlement routing, and rebate structures, exposed through Cosmos SDK keeper APIs and governance-controlled parameters.
- **Governance-controlled live funds.** All fund-moving capabilities are gated behind on-chain governance flags. No single party can authorise token movement without passing a governance proposal.
- **Transparent treasury and payout flows.** Treasury module accounts, budget allocations, grant milestones, and spend execution are all on-chain operations with full audit trails. Payouts are deterministic and governance-authorised.
- **Validator-secured network.** Consensus and block production are secured by a CometBFT validator set. The target model moves from current agent validators to an external validator cohort, then to a permissioned public set, and ultimately toward progressive decentralisation.

NexaRail does not aim to compete with general-purpose smart contract platforms. It targets a narrower vertical: payment and settlement infrastructure where protocol-level guarantees — deterministic fee splits, governance-gated fund movement, module-level auditability — matter more than general programmability.

---

## 5. Network Overview

| Parameter | Value |
|---|---|
| **Chain** | NexaRail Network |
| **Framework** | Cosmos SDK v0.47.17 + CometBFT v0.37.18 |
| **Coin / Ticker** | NXRL |
| **Base denom** | `unxrl` |
| **Display precision** | 1 NXRL = 1,000,000 unxrl |
| **Address prefix** | `nxr` |
| **Binary** | `nexaraild` |
| **Language** | Go 1.22+ |
| **Mainnet chain ID** | `nexarail-mainnet-1` (LIVE since 2026-06-18 18:00 UTC) |
| **Devnet chain ID** | `nexarail-devnet-1` (local development) |
| **Retired testnet chain IDs** | `nexarail-testnet-1`, `nexarail-agent-testnet-1` (retired 2026-06-20) |
| **Total supply** | 1,000,000,000 NXRL (fixed, plus mint module inflation post-vote) |
| **Block time** | ~5s |

### Networking

- CometBFT RPC: port 26657 (default) — coord home overrides per-node
- Cosmos SDK REST API: port 1317 (default)
- gRPC: port 9090 (default)
- P2P: port 26656 (default)

### Mainnet endpoints

- Genesis file: https://github.com/Bookings-cpu/nexarail/releases/tag/mainnet-genesis-nexarail-mainnet-1
- Live status page: https://bookings-cpu.github.io/nexarail-status/
- IPv6 public peer: `96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656`
- IPv4 public peer (via bore.pub relay): `96e659f9a87723304dcd614e3ca89d9b6daf26cc@159.223.110.159:32656` (no SLA; full sentry deployment pending)
- Slashing: live from block 1. Downtime >50% over 10,000 blocks → 600s jail + 0.01% slash. Double-sign → 5% slash + permanent tombstone.

### Standard SDK Modules

All standard Cosmos SDK v0.47.17 modules are wired: `auth`, `bank`, `staking`, `slashing`, `gov`, `distribution`, `mint`, `params`, `crisis`, `upgrade`, `evidence`, `feegrant`, `authz`, `capability`, `vesting`, `genutil`.

### Custom NexaRail Modules

Six purpose-built modules: `x/fees`, `x/merchant`, `x/settlement`, `x/escrow`, `x/payout`, `x/treasury`.

---

## 6. Core Modules

### x/fees

**Purpose:** Defines and manages fee split parameters for the network. Default fee split is 60/20/20 (validator rewards / treasury reserve / burn). The module stores fee parameters as governance-controlled KV state.

**Current status:** Implemented. Metadata-only — no coin routing. Parameter changes require a `MsgUpdateParams` governance proposal.

**Live funds enabled by default:** No (policy parameters only — no fund movement capability in this module).

---

### x/merchant

**Purpose:** Merchant registration and rebate tier management. Merchants register on-chain with profile metadata, category, and rebate tier. The module tracks merchant status, fee rebate eligibility, and registration parameters.

**Current status:** Implemented. Full lifecycle (register, update, deactivate, reactivate) with governance-controlled parameters for registration deposit and rebate tiers.

**Live funds enabled by default:** N/A (registration metadata only — no fund movement).

---

### x/settlement

**Purpose:** Payment settlement with programmable fee routing. Records settlement metadata, calculates fee splits, manages settlement status transitions (pending, confirmed, failed). Supports three routing flags: `LiveEnabled` (merchant-net transfers), `TreasuryRoutingEnabled` (treasury share), `BurnRoutingEnabled` (burn share).

**Current status:** Implemented. All three routing flags exist and default to `false`. Live transfer tests passing.

**Live funds enabled by default:** No. Three separate governance flags are all `false`:
- `LiveEnabled` — default `false`
- `TreasuryRoutingEnabled` — default `false`
- `BurnRoutingEnabled` — default `false`

---

### x/escrow

**Purpose:** Payment escrow custody lifecycle. Supports creation, funding, release, refund, and dispute resolution. Funds move through the escrow module account only when `LiveEnabled` is `true`.

**Current status:** Implemented. Metadata-only lifecycle always available. Live custody logic implemented and tested but gated behind `LiveEnabled` flag.

**Live funds enabled by default:** No — `LiveEnabled` defaults to `false`.

---

### x/payout

**Purpose:** Automated payout execution. Records payout instructions, supports approval workflows, manages payout status transitions.

**Current status:** Implemented. Metadata-only lifecycle always available. Live disbursement logic implemented and tested but gated behind `LiveEnabled` flag.

**Live funds enabled by default:** No — `LiveEnabled` defaults to `false`.

---

### x/treasury

**Purpose:** Protocol treasury management. Manages treasury accounts, budget allocations, grant milestones, and spend request workflows. Supports budget tracking, milestone completion tracking, and spend execution.

**Current status:** Implemented. Metadata-only lifecycle always available. Live spend execution implemented and tested but gated behind `LiveEnabled` flag.

**Live funds enabled by default:** No — `LiveEnabled` defaults to `false`.

---

## 7. Live Funds Safety Model

The live funds safety model is designed to prevent accidental or premature fund movement on any running network.

### Default state: all flags false

Every module capable of moving tokens has a `LiveEnabled` (or equivalent) boolean flag that defaults to `false`. When `false`, the module operates in metadata-only mode: it records lifecycle state transitions, keeps audit logs, enforces business rules, but never initiates bank sends, module account transfers, or burn operations.

### Governance-controlled enablement

Each flag can only be changed through an on-chain governance proposal (`MsgUpdateParams`) with a voting period, deposit, and quorum. No single key or authority can enable live fund movement.

### Flag inventory

| Module | Flag | Default | Effect when `true` |
|---|---|---|---|
| x/escrow | `LiveEnabled` | `false` | Escrow custody: buyer funds locked in escrow module account; release sends to seller; refund returns to buyer |
| x/treasury | `LiveEnabled` | `false` | Spend execution: treasury module account transfers to recipients |
| x/payout | `LiveEnabled` | `false` | Payout execution: treasury-to-recipient transfers |
| x/settlement | `LiveEnabled` | `false` | Merchant-net transfer: payer sends coins to merchant |
| x/settlement | `TreasuryRoutingEnabled` | `false` | Treasury share routing to module account (depends on `LiveEnabled`) |
| x/settlement | `BurnRoutingEnabled` | `false` | Burn share: supply reduction via `BurnCoins` (depends on `LiveEnabled` and `TreasuryRoutingEnabled`) |

### Module accounts

The following module accounts would participate in live fund movement when enabled:

- **`escrow`** — holds buyer funds during escrow lifecycle
- **`treasury`** — holds protocol treasury reserves
- **`fee_collector`** — standard Cosmos SDK fee collection
- **`fee_router`** — temporary holding during fee splitting (if implemented)
- **Burn** — implemented via `bank.BurnCoins` (supply reduction)

All module account addresses are added to the bank module's blocked recipients list to prevent direct deposits outside approved message paths.

### No live funds by default

To be explicit: **no live funds can move on any NexaRail network without a governance proposal passing first.** Testnet tokens have no monetary value. Mainnet does not exist. Live fund flags are disabled by default.

---

## 8. Validator and Consensus Model

### Consensus engine

NexaRail uses CometBFT v0.37.18 (a fork of Tendermint) for Byzantine Fault Tolerant consensus. Block production requires >2/3 validator voting power to sign each block. The validator set is defined in genesis and managed through staking and governance.

### Current validator set: open, mainnet, growing

Mainnet (`nexarail-mainnet-1`) launched 2026-06-18 18:00 UTC with a 7-validator genesis set: five development-operated coordinator validators (`nxrl-controlled-alpha/bravo/charlie/delta/echo`) and two external validators (NODESYNC, UTSA). External validator onboarding has been open from launch and the active set has grown to 23+ validators within the first week.

### External validator onboarding: live

The onboarding process is documented in `docs/mainnet/NEW_VALIDATOR_ONBOARDING.md` and produces:

1. Operator self-generates an operator key and a sync node
2. Operator syncs from genesis until `catching_up:false`
3. Operator submits a request with their operator address, consensus pubkey, moniker, and contact
4. Sync-proof gate: coordinator verifies the operator is running an actual synced node
5. Coordinator issues a 600 NXRL bootstrap grant from `ecosystem_grants` (500 NXRL self-bond + 100 NXRL gas runway)
6. Operator submits `create-validator` from their newly-funded wallet
7. Validator joins active set within one block

Pace: approximately 15 onboardings per day during the first week. Active set max is currently 100 validators.

### Target validator cohort

| Cohort | Size | Status |
|---|---|---|
| Coord (launched) | 5 | Live |
| Genesis external (NODESYNC + UTSA) | 2 | Live |
| External (additional) | growing — 16+ as of v2.0 of litepaper | Live, open queue |
| Active set cap (current) | 100 | Configurable via governance |

### Coordinator power concentration & The Cut-Over

Because coordinator validators self-bonded 5M NXRL each at genesis (vs 500 NXRL standard for external validators), the coord cohort holds approximately 99.97% of bonded voting power at the time of v2.0 of this litepaper.

A planned ceremony codenamed "**The Cut-Over**" will rebalance voting power when the chain reaches 100 active validators. The mechanism is **delegation, not transfer**: coord wallets reduce self-bonds from 5M → 2.5M each, then delegate the freed 12.5M (plus any custodied coord staking rewards accumulated post-mint-enablement) across non-coord active validators using a 70/30 formula. This brings block-production share to approximately 50/50 coord/external while preserving founder economic ownership and governance voting rights (delegators vote with their stake on Cosmos).

Full plan: `docs/mainnet/DECENTRALISATION_CEREMONY_PLAN.md`. Founder commitment doc: `docs/mainnet/COORD_REWARDS_COMMITMENT_2026-06-21.md`.

---

## 9. Governance

### Governance framework

NexaRail uses the standard Cosmos SDK `gov` module (v1 proposal pathway) for on-chain governance. Governance is the sole authority for:

- Parameter changes via `MsgUpdateParams` for each custom module
- Enabling/disabling live fund flags
- Software upgrade proposals
- Text proposals (non-binding signalling)

### Current status

Governance transactions (submit proposal, deposit, vote, pass) have been executed at the transaction and event level on the agent testnet. The full governance lifecycle works end-to-end.

Phase 9T validated the escrow live-flag lifecycle with state readback: proposal `1` enabled `escrow.live_enabled`, proposal `2` disabled it, and final live flags returned to `false`. Broader public/external testnet validation remains pending until external validators and gentxs exist.

### Governance and live flags

Live fund flags cannot be changed by any entity other than governance. There is no backdoor, no admin key, and no emergency override that bypasses governance for parameter changes. (Emergency stop or circuit breaker mechanisms, if implemented in future, would be separate from the governance model.)

---

## 10. Current Technical Status

### Module implementation

- **Six custom modules** in production-ready state:
  - `x/fees`, `x/merchant`, `x/settlement`, `x/escrow`, `x/payout`, `x/treasury`
- Each module has: keeper, MsgServer, QueryServer, CLI, proto definitions, and app wiring
- **Sixteen standard Cosmos SDK modules** wired and functional

### API surfaces

- REST API (Cosmos SDK LCD)
- gRPC (Cosmos SDK gRPC server)
- CometBFT RPC
- CLI (`nexaraild` binary)

### Tests

- Approximately 500+ tests across all custom module packages and app integration tests
- Tests include: unit tests, keeper tests, integration/app tests, invariant tests, fuzz tests (where applicable)
- All tests pass on `go test ./...`

### Mainnet (`nexarail-mainnet-1`)

- Genesis: 2026-06-18 18:00 UTC, 7 validators (5 coord + 2 external)
- Active set as of v2.0 of this litepaper: 23 bonded validators
- Block production: ~5 second blocks, no halts since launch-day localhost-mesh fix (see incident log in `memory/2026-06-19.md`)
- Slashing active from block 1
- Governance: Proposal 1 (enable mint module for staking yield) submitted 2026-06-21, in voting period
- Faucet operational: 16 grants distributed in first 24 hours of v2.0 cycle, ~9,600 NXRL out of 150M `ecosystem_grants` reserve

### Retired testnets

- `nexarail-testnet-1` retired 2026-06-20 (validator readiness, slashing rehearsal, faucet calibration completed)
- `nexarail-agent-testnet-1` retired (precursor agent testnet, runtime hardening completed)

### Tooling

- `nexaraild` binary — full node and CLI
- Devnet initialisation (`make init-devnet`, `make start-devnet`)
- Docker rehearsal environment
- Genesis coordination tooling (gentx collection, validation, genesis assembly)
- Governance transaction builder (`tools/govtxbuilder`)
- Store inspector (`tools/storeinspector`)

### Limitations to be clear about

- **Coord-side power is concentrated.** ~99.97% of bonded voting power is held by the 5 coordinator validators at the time of v2.0. The Cut-Over ceremony at 100 validators will rebalance to ~50/50 block-production share via founder-side delegation, with founder retaining governance voting rights.
- **Not audited.** No formal third-party security audit.
- **Not legally reviewed.** No formal independent legal review.
- **Live-funds modules still disabled.** All six payment-vertical modules (`x/fees`, `x/merchant`, `x/settlement`, `x/escrow`, `x/payout`, `x/treasury`) have their LiveEnabled flags `false`. Governance proposals to enable them are sequenced behind security audit completion.
- **No bridge or stablecoin registry.** Deferred.
- **No token sale.** NXRL has not been offered for sale.
- **No DEX listing.** No price discovery venue exists yet. Anyone advertising NXRL purchase opportunities is fraudulent.
- **Public RPC infrastructure pending.** The mainnet RPC, REST, and explorer surfaces require a dedicated sentry node (Oracle Cloud free-tier VM in provisioning queue) before public-facing endpoints are stable.

---

## 11. Roadmap

### Phases A–C completed pre-launch

- Phase A: Controlled agent testnet hardening — **COMPLETE**
- Phase B: External validator cohort onboarding — **COMPLETE** (genesis NodeSync + UTSA, then open queue from mainnet launch)
- Phase C: Public testnet (`nexarail-testnet-1`) — **COMPLETE** (retired 2026-06-20 after serving its calibration purpose)

### Phase 1 (current — mainnet operations, week 1+)

- Open external validator queue (live)
- Faucet operational from `ecosystem_grants` (live)
- Mainnet status page (live)
- Validator coordination Discord (live)
- Onboarding runbook (live, hardened with self-bond decimal warning after first two operator typos)

### Phase 2 (immediate — staking yield activation)

- Mint module enablement proposal (Proposal 1, in voting period) — executes 2026-06-26
- Coord rewards custody commitment (documented, on-chain attestation pending vote execution)

### Phase 3 (next 30 days — sentry + explorer + reach)

- Oracle Cloud Always Free Ampere sentry deployment (blocked on capacity)
- Public RPC + REST endpoints surfacing via sentry
- Block explorer (ping.pub fork) deployment
- Validator recruitment outreach (target: 100 active validators)

### Phase 4 (at 100 active validators — The Cut-Over)

- Coord self-bond reduction 5M → 2.5M each
- Delegation of 12.5M (plus custodied rewards) across non-coord active validators using 70/30 formula
- Block-production share rebalances to ~50/50
- Marketing moment positioned as the ecosystem's transition from coord-bootstrapped to community-secured

### Phase 5 (Q3-Q4 2026 — utility activation + listings)

- Third-party security audit
- Governance proposals enabling LiveEnabled flags per payment-vertical module (escrow, payout, settlement, treasury)
- Independent legal review covering token classification, regulatory positioning, jurisdictional risk
- DEX listings (Osmosis pool, others TBD)
- First merchant onboardings via the live `x/merchant` and `x/settlement` modules

---

## 12. Limitations

This section is a consolidated list of limitations that apply to NexaRail in its current state.

1. **Mainnet is live but early.** First week of operation. Onboarding and infrastructure surfaces are still being built out.
2. **Coord-side power concentration.** ~99.97% of bonded voting power is held by 5 coordinator validators at the time of v2.0. Decentralisation via The Cut-Over is scheduled at 100 active validators.
3. **No external audit.** A formal third-party security audit has not been completed.
4. **No token sale.** NXRL has not been offered for sale through any mechanism. There is no way to purchase NXRL. Testnet tokens never had monetary value; mainnet tokens have no listed price yet.
5. **No DEX listing.** No liquidity venue exists yet.
6. **Bridge and stablecoin registry deferred.** IBC integration and stablecoin registry are on the deferred list. The network operates as an isolated sovereign chain.
7. **Live funds disabled by default.** All live fund movement flags default to `false`. No funds move through `x/escrow`, `x/payout`, `x/settlement`, or `x/treasury` payment paths without governance approval.
8. **Legal review pending.** Formal independent legal review has not been completed.
9. **Public RPC + explorer infrastructure pending.** Sentry deployment is in the Oracle Cloud Free Tier queue.
10. **Roadmap is provisional.** All phases, dates, and targets are subject to change. No timeline commitments are made.

---

## 13. Security and Audit Posture

### Threat register

A comprehensive threat register exists at `docs/security/THREAT_REGISTER.md` covering:

- Module-level threats for each custom module
- Cross-module fund flow threats
- Governance attack vectors
- Validator compromise scenarios
- Network-level threats (eclipse, sybil, DDoS)

Additional threat models exist for specific subsystems: settlement live transfers (`docs/security/SETTLEMENT_LIVE_THREAT_MODEL.md`), treasury/fee routing (`docs/security/SETTLEMENT_TREASURY_FEE_THREAT_MODEL.md`), burn routing (`docs/security/SETTLEMENT_BURN_THREAT_MODEL.md`), and validator distribution (`docs/security/VALIDATOR_DISTRIBUTION_THREAT_MODEL.md`).

### Audit package

An audit preparation package is available at `docs/audit/` containing:

- Audit package index (`AUDIT_PACKAGE_INDEX.md`)
- Final audit package (`PHASE_8D_AUDIT_PACKAGE_FINAL.md`)
- Audit-specific security review (`PHASE_8D_SECURITY_REVIEW.md`)
- Live funds audit preparation (`PHASE_5_AUDIT_PREP.md`)
- Phase 3 threat review (`PHASE_3_THREAT_REVIEW.md`)

### Predeployment checks

Predeployment release checklists exist at `docs/release/` covering:

- Controlled testnet release checklist (`CONTROLLED_TESTNET_RELEASE_CHECKLIST.md`)
- Pre-launch sign-off process (`PRE_LAUNCH_SIGN_OFF.md`)
- Pre-launch freeze checklist (`docs/testnet/PRE_LAUNCH_FREEZE_CHECKLIST.md`)

### Release and change-control

A formal change-control policy exists at `docs/release/CHANGE_CONTROL_POLICY.md`, with a release process runbook (`docs/release/RELEASE_PROCESS_RUNBOOK.md`), release tagging and checksums guide (`docs/release/RELEASE_TAGGING_AND_CHECKSUMS.md`), and reproducible build notes (`docs/release/REPRODUCIBLE_BUILD_NOTES.md`).

### External audit still required

None of the above replaces a formal third-party security audit. An external audit by a recognised blockchain security firm is required before any consideration of mainnet. Current documentation supports that future audit — it does not substitute for it.

---

## 14. Conclusion

NexaRail is a Cosmos SDK sovereign L1 built for payment and settlement infrastructure. The network has six purpose-built modules for merchant payments, settlement, escrow, payouts, and treasury management, currently gated behind governance-controlled live-funds flags that default to disabled pending external security audit.

**Mainnet has been live since 2026-06-18 18:00 UTC.** The validator set has grown from 7 at genesis to 23+ active validators, with continuous external onboarding via an open application process. A governance proposal to enable staking yield via the standard mint module is in voting and executes 2026-06-26. Coord-side voting concentration will rebalance via "The Cut-Over" delegation ceremony when the chain reaches 100 active validators.

**The honest summary:** mainnet is live with active block production, BFT consensus, slashing, and an open external validator queue. Validator economics are about to activate via the in-vote mint proposal. Payment-vertical live functionality is sequenced behind security audit completion. There is no DEX listing, no token sale, no investment offer. If you are a validator operator, onboarding is open and documented. If you are looking to acquire NXRL through any market, that does not exist yet — anyone advertising otherwise is fraudulent.

---

**NexaRail Network — Payment Infrastructure L1. Live mainnet. Open validator queue.**
