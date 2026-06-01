# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field | Your Response |
|---|---|
| Operator Name | Noddex |
| Organisation (if any) | Noddex |
| Country / Jurisdiction | Vietnam |
| Contact Email | victor@noddex.com |
| Discord Handle | victorahale |
| Telegram Handle | victorahale |
| GitHub Handle | victorahale |
| Website (optional) | https://noddex.com |

## Validator Information

| Field | Your Response |
|---|---|
| Validator Moniker | Noddex |
| Intended Commission Rate | 10% (0.1) |
| Intended Max Commission Rate | 20% (0.2) |
| Intended Max Change Rate | 5% (0.05) |
| Self-Delegation Amount | 500,000,000 unxrl |

## Infrastructure

| Field | Your Response |
|---|---|
| Hosting Provider | Hetzner Dedicated Bare-Metal |
| Operating System | Ubuntu Linux |
| CPU Cores | 8 |
| RAM (GB) | 16 |
| Disk Size & Type | 500 GB |
| Network Speed | 1 Gbps |
| Static IP Available? | Yes |
| Geographic Region | EU / VN / US |
| Redundant Power? | Yes |
| Redundant Network? | Yes |
| Monitoring Setup | Prometheus and Grafana for system metrics, integrated with real-time incident alerting via Telegram. |
| Backup / Snapshot Strategy | Daily production-grade data snapshots stored on off-site backup servers to ensure immediate state recovery. |

## Experience

| Field | Your Response |
|---|---|
| Years Running Validators | 5+ years of dedicated blockchain validator deployment |
| Chains Previously Validated | Espresso, Tangle, Redbelly, Pactus, AR.IO, Kiichain, Axone, Waterfall, Selfchain, Kopi, Avail, Namada, Dymension, Warden, ZetaChain. |
| Cosmos SDK Experience? | Yes |
| Tendermint/CometBFT Experience? | Yes |
| Key Management Practice | TMKMS remote signer — validator key stored on a dedicated signing server with encrypted backup and SSH-key-only access. No raw priv_validator_key.json on the active node. |
| Incident Response Experience | 24/7 monitoring with automated alerts via Telegram. Incidents such as missed blocks or peer connectivity drops are typically resolved within minutes. Upgrades follow a stop-update-restart procedure to avoid double-sign risk. |

## Commitments

Please confirm each statement with your initials or a check:

| # | Statement | Confirmed |
|---|---|---|
| 1 | I understand this is a TESTNET only. No mainnet is live. | ☑ |
| 2 | I understand testnet tokens have ZERO monetary value and cannot be sold or exchanged. | ☑ |
| 3 | I understand this is NOT a token sale, investment, or financial opportunity. | ☑ |
| 4 | I will run my validator on a Linux host (not macOS Docker Desktop). | ☑ |
| 5 | I will maintain reasonable uptime and respond to coordinator communications. | ☑ |
| 6 | I understand testnet state may be wiped or reset at any time. | ☑ |
| 7 | I will report security issues through the designated reporting process. | ☑ |
| 8 | I will not make public claims about NXRL having monetary value or being an investment. | ☑ |
| 9 | I agree to the NexaRail testnet code of conduct. | ☑ |
| 10 | I understand my validator can be removed from the active set via governance. | ☑ |

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
