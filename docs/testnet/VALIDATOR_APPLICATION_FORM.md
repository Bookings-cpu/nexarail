# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field | Your Response |
|---|---|
| Operator Name | Yahya | OshVanK |
| Organisation (if any) | OshVanK |
| Country / Jurisdiction | Turkey |
| Contact Email | worms_Ss@hotmail.com|
| Discord Handle | yahya6935 |
| Telegram Handle | @Edsny|
| GitHub Handle | Edsny1 |
| Website (optional) | https://oshvank.xyz, https://monitor.oshvank.xyz, https://explorer-gnoland.oshvank.xyz, https://explorer.oshvank.xyz, https://monad.oshvank.xyz, https://celestiascope.oshvank.xyz, https://monadscope.oshvank.xyz |

## Validator Information

| Field | Your Response |
|---|---|
| Validator Moniker | 🏆OshVanK🏆 |
| Intended Commission Rate | % (0.10) |
| Intended Max Commission Rate | % (0.20) |
| Intended Max Change Rate | % (0.01) |
| Self-Delegation Amount | 500,000,000 unxrl |

## Infrastructure

| Field | Your Response |
|---|---|
| Hosting Provider | Hetzner / MevSpace |
| Operating System | Ubuntu 22.04 / 24.04 LTS |
| CPU Cores | 12 / 16 Cores |
| RAM (GB) | 96 / 128 GB|
| Disk Size & Type | 1TB / 6 TB|
| Network Speed | 1 Gbps Unshared Uplink |
| Static IP Available? | Yes |
| Geographic Region | EU |
| Redundant Power? | Yes |
| Redundant Network? | Yes |
| Monitoring Setup | Full observability via Prometheus, Grafana, and Tenderduty (OshVanK Monitor). Real-time threshold alerts integrated directly with custom Telegram and Discord notification bots. |
| Backup / Snapshot Strategy | Automated daily state snapshots stored securely on an off-site location. Disaster recovery protocol managed via custom Ansible playbooks for automated node restoration. |

## Experience

| Field | Your Response |
|---|---|
| Years Running Validators | 6 Years |
| Chains Previously Validated | Celestia, AtomOne, Monad, Arkhadian, BitBadges, Dymension, Lumera, Ar-io, Massa, Pactus, Espresso, PushChain, SafroChain, CardChain, Republic, Humanode, FortyTwo, GnoLand, Drosera and other. (Mainnet and Testnet) |
| Cosmos SDK Experience? | Yes |
| Tendermint/CometBFT Experience? | Yes |
| Key Management Practice | Consensus operations utilize private peering network isolation. Keys are secured via encrypted, offline-generated backups, strict access control lists, and fully support enterprise remote-signer configurations. |
| Incident Response Experience | Node operations feature automated Tenderduty threshold alerts routed directly to custom communication bots. Proven track record handling unscheduled network upgrades and emergency consensus state recovery under a strict 30-minute RTO. |

## Commitments

Please confirm each statement with your initials or a check:

| # | Statement | Confirmed |
|---|---|---|
| 1 | I understand this is a TESTNET only. No mainnet is live. | ✅ |
| 2 | I understand testnet tokens have ZERO monetary value and cannot be sold or exchanged. | ✅ |
| 3 | I understand this is NOT a token sale, investment, or financial opportunity. | ✅ |
| 4 | I will run my validator on a Linux host (not macOS Docker Desktop). | ✅ |
| 5 | I will maintain reasonable uptime and respond to coordinator communications. | ✅ |
| 6 | I understand testnet state may be wiped or reset at any time. | ✅ |
| 7 | I will report security issues through the designated reporting process. | ✅ |
| 8 | I will not make public claims about NXRL having monetary value or being an investment. | ✅ |
| 9 | I agree to the NexaRail testnet code of conduct. | ✅ |
| 10 | I understand my validator can be removed from the active set via governance. | ✅ |

## Public Key Submission

After acceptance, you will be instructed to generate and submit:

- Node ID (from `nexaraild tendermint show-node-id`)
- Validator public key
- Signed `gentx` file

Do NOT submit private keys, mnemonics, or passwords through any channel.

## Declaration

By submitting this application, I confirm:

- All information provided is accurate and truthful
- I have read and understood the controlled validator registration document
- I am applying for testnet participation only
- I make no claim to any financial benefit, investment return, or token value
- I am legally permitted to operate a blockchain validator in my jurisdiction

---

**Submit to:** Genesis coordinator via designated communication channel.

**Questions:** Contact the coordinator through the testnet communication channel.
