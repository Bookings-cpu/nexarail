# NexaRail Network

[![CI](https://github.com/Bookings-cpu/nexarail/actions/workflows/nexarail-regression.yml/badge.svg)](https://github.com/Bookings-cpu/nexarail/actions/workflows/nexarail-regression.yml)

A sovereign Layer 1 blockchain for railway settlement and payments — built with Cosmos SDK v0.47.17 and CometBFT v0.37.18.

## Mainnet — `nexarail-mainnet-1` — LIVE

NexaRail mainnet launched on **2026-06-18 18:00 UTC**. Block production averages ~5 s, slashing is live, and the validator active set was open from launch.

| | |
|---|---|
| Chain ID | `nexarail-mainnet-1` |
| Native denom | `unxrl` (1 NXRL = 1,000,000 unxrl) |
| Block time | ~5 s |
| Genesis SHA256 | `f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32` |
| Live status page | https://bookings-cpu.github.io/nexarail-status/ |
| Genesis file | [GitHub release: `mainnet-genesis-nexarail-mainnet-1`](https://github.com/Bookings-cpu/nexarail/releases/tag/mainnet-genesis-nexarail-mainnet-1) |
| Binary (linux/amd64) | [GitHub release: `v0.1.0-rc1-validator-recovery-hotfix`](https://github.com/Bookings-cpu/nexarail/releases/tag/v0.1.0-rc1-validator-recovery-hotfix) |
| Validator onboarding runbook | [docs/mainnet/NEW_VALIDATOR_ONBOARDING.md](docs/mainnet/NEW_VALIDATOR_ONBOARDING.md) |
| Validator broadcast / pin | [coordination/outreach/mainnet-validator-broadcast-2026-06-19.md](coordination/outreach/mainnet-validator-broadcast-2026-06-19.md) |
| Mint enablement scoping | [docs/mainnet/MINT_ENABLEMENT_SCOPING_2026-06-21.md](docs/mainnet/MINT_ENABLEMENT_SCOPING_2026-06-21.md) |
| The Cut-Over (decentralisation ceremony) plan | [docs/mainnet/DECENTRALISATION_CEREMONY_PLAN.md](docs/mainnet/DECENTRALISATION_CEREMONY_PLAN.md) |
| Coord rewards custody commitment | [docs/mainnet/COORD_REWARDS_COMMITMENT_2026-06-21.md](docs/mainnet/COORD_REWARDS_COMMITMENT_2026-06-21.md) |

**Active set:** launched with 7 validators (5 coord + 2 external: NodeSync, UTSA). Now growing daily via the open external onboarding queue.

**Active governance:** Proposal #1 (enable NXRL staking yield via mint module, 5%-12% floating inflation) is in voting until **2026-06-26 15:14 UTC**. Once executed, validators and delegators begin earning continuous staking rewards.

**Slashing is live from block 1.** Downtime >50% over 10,000 blocks → 600 s jail + 0.01% slash. Double-sign → 5% slash + permanent tombstone. Never run two nodes with the same `priv_validator_key.json`.

**Standard faucet grant for new validators:** 500 NXRL self-bond + 100 NXRL gas = **600 NXRL total**, drawn from the 150M NXRL `ecosystem_grants` bucket. Request format and full step-by-step are in the onboarding runbook above. Critical: count the zeros on `--amount=500000000unxrl` — multiple validators have typo'd two zeros short and ended up with 5 NXRL self-bond instead of 500.

### Testnet — `nexarail-testnet-1` — RETIRED 2026-06-20

The controlled testnet served its purpose (validator readiness, slashing rehearsal, peer-mesh proving, faucet pipeline calibration) and was retired on 2026-06-20. See [coordination/outreach/2026-06-20-testnet-retirement-mainnet-migration.md](coordination/outreach/2026-06-20-testnet-retirement-mainnet-migration.md) for the migration notice. Historical testnet docs remain in `docs/testnet/` for reference.

---

## Developer quickstart (local devnet)

For local development, the source tag `v0.1.0-rc1-cli-hotfix` provides a single-command devnet:

```bash
git clone https://github.com/Bookings-cpu/nexarail.git
cd nexarail
git checkout v0.1.0-rc1-cli-hotfix
bash scripts/dev/run-nexarail-regression-matrix.sh --fast
```

See [Reviewer Handoff](docs/release/REVIEWER_HANDOFF.md) | [Litepaper](docs/NEXARAIL_LITEPAPER.md) | [Quickstart](docs/release/RC1_QUICKSTART.md) | [Release Verification](docs/release/GITHUB_RELEASE_V0.1.0_RC1_VERIFICATION.md)

---

## Run the Local Demo

```bash
scripts/dev/run-local-demo.sh --serve-dashboard --keep-running
```

One command: verify package → launch devnet → run smoke tests → open dashboard.

| Option | Effect |
|---|---|
| `--serve-dashboard` | Open local dashboard at http://localhost:8088 |
| `--keep-running` | Leave devnet running after demo |
| `--skip-smoke` | Skip smoke tests (faster) |

[Full demo guide →](docs/developers/LOCAL_DEMO_GUIDE.md)

**⚠️ LOCAL DEVNET ONLY — NOT MAINNET — No token sale.**

---

## Controlled Testnet RC1 — historical

> The controlled testnet was retired on 2026-06-20 in favour of mainnet. The
> table below is preserved as a historical record of where the testnet
> programme stood at the point of retirement. For live status see the
> mainnet section above.

| Item | Detail |
|---|---|
| Packaging | **Complete** — RC1 release assets packaged and verified |
| Local devnet | **Ready** — single-node and 5-agent modes available |
| Controlled external-validator launch | **Completed** — testnet ran with 5 coord + 2 external validators (NodeSync, UTSA) |
| Mainnet | **LIVE** — `nexarail-mainnet-1` launched 2026-06-18 18:00 UTC (see top of README) |
| Validator onboarding | **Open on mainnet** — see [docs/mainnet/NEW_VALIDATOR_ONBOARDING.md](docs/mainnet/NEW_VALIDATOR_ONBOARDING.md) |
| Product-flow suite | **487 pass / 0 fail** — full coverage validated |
| REST parity | **36/36 (100%)** — REST readback parity confirmed |
| Testnet status | `RETIRED_2026-06-20` |

### Quick Links

| Link | Description |
|---|---|
| [Local Demo Guide](docs/developers/LOCAL_DEMO_GUIDE.md) | One-command demo walkthrough |
| [RC1 Reviewer README](docs/release/RC1_REVIEWER_README.md) | Primary entry point for reviewers |
| [RC1 Quickstart](docs/release/RC1_QUICKSTART.md) | Command-first devnet guide |
| [RC1 Evidence Summary](docs/release/RC1_EVIDENCE_SUMMARY.md) | All evidence at a glance |
| [RC1 Review Checklist](docs/release/RC1_REVIEW_CHECKLIST.md) | Reviewer verification checklist |
| [RC1 Release Notes](docs/release/TESTNET_RC1_RELEASE_NOTES.md) | Full release notes |
| [RC1 Devnet Launch Guide](docs/release/TESTNET_RC1_DEVNET_LAUNCH_GUIDE.md) | Step-by-step devnet setup |
| [Litepaper](docs/NEXARAIL_LITEPAPER.md) | Project overview |
| [REST API Docs](docs/api/REST_READBACK_ROUTES.md) | All REST readback endpoints |
| [Evidence Manifest](docs/release/TESTNET_RC1_EVIDENCE_MANIFEST.md) | Evidence index |
| [Known Limitations](docs/release/TESTNET_RC1_KNOWN_LIMITATIONS.md) | Current limitations |
| [Controlled Testnet Launch Plan](docs/testnet/PHASE_17A_CONTROLLED_TESTNET_LAUNCH_PLAN.md) | Phase 17A external-validator launch candidate plan |
| [Validator Action Pack](docs/testnet/EXTERNAL_VALIDATOR_ACTION_PACK.md) | Source-build onboarding commands for accepted validators |
| [Controlled Testnet Status](docs/testnet/CONTROLLED_TESTNET_STATUS.md) | Current launch status and guardrails |
| [Phase 17B Intake Status](docs/testnet/PHASE_17B_VALIDATOR_INTAKE_AND_GENESIS_CANDIDATE.md) | Validator intake, gentx, peers, and genesis candidate status |
| [Phase 18A Coordinator Candidate](docs/testnet/PHASE_18A_INTERNAL_COORDINATOR_TESTNET_CANDIDATE.md) | Internal coordinator-only candidate plan and safety boundary |
| [Public Join Readiness Checklist](docs/testnet/PUBLIC_JOIN_READINESS_CHECKLIST.md) | Join-readiness gates while external gentxs remain pending |
| [Phase 18B Intake Execution](docs/testnet/PHASE_18B_EXTERNAL_VALIDATOR_INTAKE_EXECUTION.md) | External validator intake, gentx review, endpoint collection, and freeze gate |
| [Final Genesis Freeze Decision](docs/testnet/FINAL_GENESIS_FREEZE_DECISION.md) | Current freeze decision and next required action |
| [Phase 17C First External Gentx Verification](docs/testnet/PHASE_17C_FIRST_EXTERNAL_GENTX_VERIFICATION.md) | NodeSync SHA256 match, gentx verifier pass, confirmed DNS peer, and next action |
| [Phase 17C.1 NodeSync Recheck](docs/testnet/PHASE_17C1_NODESYNC_SUBMISSION_RECHECK.md) | Local evidence recheck for attachment/download handling and SHA256 proof |
| [Phase 17D Genesis Candidate](docs/testnet/PHASE_17D_EXTERNAL_VALIDATOR_GENESIS_CANDIDATE.md) | Six-validator review candidate with NodeSync plus coordinator-operated validators |
| [Phase 17E Reachability Gate](docs/testnet/PHASE_17E_NODESYNC_P2P_REACHABILITY_AND_FREEZE_GATE.md) | NodeSync DNS/TCP reachability result and freeze decision |
| [Launch Operations Dashboard](docs/testnet/CONTROLLED_TESTNET_LAUNCH_READINESS_DASHBOARD.md) | Coordinator readiness table for launch operations |
| [Incident Response Runbook](docs/testnet/CONTROLLED_TESTNET_INCIDENT_RESPONSE.md) | Halt, rollback, support, and incident response process |

---

## Developer Resources

| Resource | Description |
|---|---|
| [Developer Quickstart](docs/developers/DEVELOPER_QUICKSTART.md) | Get started building against the devnet |
| [API Examples](docs/developers/API_EXAMPLES.md) | Curl examples for all REST endpoints |
| [Product Flow Examples](docs/developers/PRODUCT_FLOW_EXAMPLES.md) | Complete merchant/settlement/escrow/treasury/payout flows |
| [Demo App Plan](docs/developers/DEMO_APP_PLAN.md) | Future lightweight dashboard plan |
| [REST Examples](examples/rest/) | Read-only REST example scripts (bash + curl) |
| [Node.js Client](examples/node-client/) | Lightweight Node.js client (no dependencies) |
| [Python Client](examples/python-client/) | Lightweight Python client (stdlib only) |
| [Local Dashboard](examples/dashboard/) | Read-only local devnet dashboard (static HTML/JS/CSS) |
| [SDK Package Preparation](docs/developers/SDK_PACKAGE_PREPARATION.md) | Local SDK packaging status — Node `@nexarail/devnet-client@0.1.0-dev`, Python `nexarail-devnet-client==0.1.0.dev` (local install only, no npm/PyPI publishing) |
| [Regression Matrix](docs/developers/DEMO_REGRESSION_MATRIX.md) | CI regression checks and expected pass counts |
| [CI Checks](docs/release/RC1_CI_CHECKS.md) | When to run each check and expected outputs |

### End-to-End Developer Demo

Run the complete local devnet workflow from launch to inspect to SDK to dashboard:

```bash
scripts/dev/run-end-to-end-demo.sh
```

See [Scenario Doc](docs/developers/END_TO_END_DEMO_SCENARIO.md) and [Summary](docs/developers/END_TO_END_DEMO_SUMMARY.md) for details.

### Developer Onboarding Bundle

Get all developer assets in one archive:

```bash
scripts/dev/prepare-developer-bundle.sh
```

The bundle includes docs, SDK archives, examples, scripts, and manifests.
Archive is saved to `releases/developer-bundles/`.

See [Onboarding Checklist](docs/developers/ONBOARDING_CHECKLIST.md) and [Contributing Guide](CONTRIBUTING.md) to get started.

See [Contributor Testing Guide](docs/developers/CONTRIBUTOR_TESTING_GUIDE.md) for expected check results.

### Developer Portal

Browse the documentation as a static website:

```bash
# Build the portal
bash scripts/dev/build-developer-portal.sh

# Serve locally on port 8090
bash scripts/dev/serve-developer-portal.sh
```

Portal source: `docs/portal/index.html`
Build output: `site/developer-portal/index.html`

### SDK Documentation

| Resource | Description |
|---|---|
| [SDK RC1 Release Notes](docs/developers/SDK_RC1_RELEASE_NOTES.md) | Release notes for the SDK RC1 — changes, installation, usage |
| [Node.js SDK API Reference](docs/developers/NODE_SDK_API_REFERENCE.md) | Comprehensive API reference for the Node.js devnet client |
| [Python SDK API Reference](docs/developers/PYTHON_SDK_API_REFERENCE.md) | Comprehensive API reference for the Python devnet client |
| [SDK Recipes](docs/developers/SDK_RECIPES.md) | Common SDK usage patterns and recipes |
| [Local Package Archive](releases/sdk-local/) | Local SDK tarballs for offline/dev use — NOT published to npm or PyPI |

### Run Regression Checks

```bash
# Fast (no devnet needed):
scripts/dev/run-nexarail-regression-matrix.sh --fast

# Full (with devnet):
scripts/dev/run-nexarail-regression-matrix.sh --full --with-devnet
```

---

**Runtime note:** Local 5-agent runtime readiness, the Phase 17A dry-run, and the Phase 18A internal coordinator-candidate dry-run are proven. Phase 18B intake execution remains open, Phase 18C coordinator launch operations are prepared, and the NodeSync gentx plus DNS peer are verified, but final genesis publication and external validator launch evidence remain pending.

**🔗 [Apply to run a validator →](docs/testnet/VALIDATOR_APPLICATION_FORM.md)**

---

## Quick Start

### Prerequisites
- Go 1.22+
- Make

### Build
```bash
make build
# Binary: build/nexaraild
```

### Local Devnet
```bash
make init-devnet   # Create 3-validator devnet under ~/.nexarail/
make start-devnet  # Launch validators
```

Devnet endpoints:
- RPC: `http://127.0.0.1:26657`
- REST: `http://127.0.0.1:1317`
- gRPC: `127.0.0.1:9090`

## Chain Configuration

| Parameter | Value |
|---|---|
| Chain ID (devnet) | `nexarail-devnet-1` |
| Chain ID (testnet) | `nexarail-testnet-1` (proposed) |
| Coin Denom | `unxrl` |
| Display Ticker | `NXRL` (1 NXRL = 1,000,000 unxrl) |
| Bech32 Prefix | `nxr` |

## Modules

### Standard Cosmos SDK Modules
auth, bank, staking, slashing, gov, distribution, mint, params, crisis, upgrade, evidence, feegrant, authz, capability, vesting, genutil

### Custom NexaRail Modules

| Module | Purpose | Live Funds? |
|---|---|---|
| x/fees | Fee split parameters (60/20/20 bps) | Policy only |
| x/merchant | Merchant registration + rebate tiers | N/A |
| x/settlement | Payment settlement + fee routing | Behind 3 flags |
| x/escrow | Payment escrow custody | Behind 1 flag |
| x/payout | Automated payouts | Behind 1 flag |
| x/treasury | Protocol treasury + spend execution | Behind 1 flag |

### Live Funds

All live fund flows are implemented but **disabled by default** behind governance-gated flags:

| Module | Flag | Default |
|---|---|---|
| x/escrow | `LiveEnabled` | false |
| x/treasury | `LiveEnabled` | false |
| x/payout | `LiveEnabled` | false |
| x/settlement | `LiveEnabled` | false |
| x/settlement | `TreasuryRoutingEnabled` | false |
| x/settlement | `BurnRoutingEnabled` | false |

See `docs/design/LIVE_FLAGS_MATRIX.md` and `docs/PHASE_5_LIVE_FUNDS_STATUS.md`.

### Deferred
- Validator distribution (design complete — see `docs/design/VALIDATOR_DISTRIBUTION_DESIGN.md`)
- Fee router / BeginBlock routing
- Stablecoin registry
- Bridge (IBC or custom)

## Verification

```bash
go mod tidy && go mod verify
go build ./...
go vet ./...
go test ./...    # ~332 tests, 14 packages, all pass
```

### Validator Registration (Phase 7A–7B)

| Document | Description |
|---|---|
| `docs/testnet/CONTROLLED_VALIDATOR_REGISTRATION.md` | Registration overview, requirements, process |
| `docs/testnet/VALIDATOR_APPLICATION_FORM.md` | Application form for validators |
| `docs/testnet/VALIDATOR_ACCEPTANCE_CHECKLIST.md` | Pre-launch checklist for accepted validators |
| `docs/testnet/GENESIS_SUBMISSION_INSTRUCTIONS.md` | Gentx creation and submission guide |
| `docs/testnet/VALIDATOR_COMMUNICATIONS_PLAN.md` | Communication channels and protocols |
| `docs/testnet/GENESIS_COORDINATOR_RUNBOOK.md` | Internal coordinator operational guide |
| `docs/testnet/VALIDATOR_INTAKE_PIPELINE.md` | Application intake pipeline and SLAs |
| `docs/testnet/VALIDATOR_SCORING_RUBRIC.md` | 7-category technical review scoring |
| `docs/testnet/VALIDATOR_EMAIL_TEMPLATES.md` | Standardised email templates |
| `docs/testnet/DISCORD_TELEGRAM_MODERATION_GUIDE.md` | Community moderation rules |
| `docs/testnet/GENESIS_GENTX_REVIEW_CHECKLIST.md` | 22-point gentx verification |
| `docs/testnet/CONTROLLED_REGISTRATION_ANNOUNCEMENT_FINAL.md` | Public announcement (ready to publish) |
| `docs/testnet/FAQ.md` | Frequently asked questions |
| `docs/testnet/VALIDATOR_REGISTRATION_TRACKER_TEMPLATE.csv` | Intake tracking spreadsheet |

### Project Documentation

| Document | Description |
|---|---|
| `docs/NEXARAIL_LITEPAPER.md` | Full public litepaper — problem, vision, modules, safety model, roadmap, limitations |
| `docs/NEXARAIL_LITEPAPER_SUMMARY.md` | One-page summary of the litepaper |

### General

## ⚠️ Disclaimers

- **Testnet only.** NexaRail has no public mainnet.
- **No token sale.** NXRL has not been offered for sale. No ICO/IEO/IDO has occurred.
- **No monetary value.** Testnet tokens have zero value and cannot be exchanged.
- **No investment.** Participation is for technical testing only. No returns are promised.
- **Resets possible.** Testnet state may be wiped at any time.
- **Legal review pending.** See `docs/legal/LEGAL_REVIEW_PACKAGE.md`.

## License

Proprietary — NexaRail Protocol
