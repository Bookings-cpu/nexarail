# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field | Your Response |
|---|---|
| Operator Name | NODESYNC |
| Organisation (if any) | NODESYNC |
| Country / Jurisdiction | Vietnam (VN)|
| Contact Email | info@nodesync.top |
| Discord Handle | nodesync_top |
| Telegram Handle | @nodesync_top |
| GitHub Handle | nodesynctop |
| Website (optional) | https://nodesync.top |

## Validator Information

| Field | Your Response |
|---|---|
| Validator Moniker | NODESYNC |
| Intended Commission Rate | 10% (0.1) |
| Intended Max Commission Rate | 20% (0.20) |
| Intended Max Change Rate | 5% (0.05) |
| Self-Delegation Amount | 500,000,000 unxrl |

## Infrastructure

| Field | Your Response |
|---|---|
| Hosting Provider | Netcup/Hetzner |
| Operating System | Ubuntu 22.04 LTS |
| CPU Cores | 8 vCPU  |
| RAM (GB) | 16 GB |
| Disk Size & Type | 500 GB NVMe |
| Network Speed | 2.5 Gbps |
| Static IP Available? | Yes |
| Geographic Region | Multiple regions (EU, US, CA) |
| Redundant Power? | Yes |
| Redundant Network? | Yes |
| Monitoring Setup | We have built custom dashboards for monitoring our validators across multiple Cosmos ecosystem networks, along with a real-time Telegram bot alerting system for validator, node, and endpoint health monitoring. |
| Backup / Snapshot Strategy | Public RPC, API, snapshot, and state sync endpoints for multiple Cosmos ecosystem networks via NodeSync at https://nodesync.top/services |

## Experience

| Field | Your Response |
|---|---|
| Years Running Validators | 4+ |
| Chains Previously Validated | (mainnet and testnet) - Warden, Kiichain, Redbelly, Axone, Espresso, BeeZee, BitBadges, Pactuschain, Safrochain... |
| Cosmos SDK Experience? | Yes |
| Tendermint/CometBFT Experience? | Yes |
| Key Management Practice | Key management uses multiple secure backups of validator keys stored in geographically distributed locations for redundancy and disaster recovery. |
| Incident Response Experience | 24/7 monitoring with automated alerts, structured incident response and rapid recovery using snapshots and key restoration procedures, with experience coordinating upgrades and chain restarts across multiple blockchain networks.|

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
