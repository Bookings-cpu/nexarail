# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field                  | Your Response                  |
| ---------------------- | ------------------------------ |
| Operator Name          | Cumulo                         |
| Organisation (if any)  | Cumulo Pro                     |
| Country / Jurisdiction | Spain (EU)                     |
| Contact Email          | info@cumulo.pro                |
| Discord Handle         | mon.cumulo.pro / sami.cumulo.pro |
| Telegram Handle        | @monjsd / @Marisamipi          |
| GitHub Handle          | cumulo-pro                     |
| Website (optional)     | https://cumulo.pro             |

## Validator Information

| Field                        | Your Response            |
| ---------------------------- | ------------------------ |
| Validator Moniker            | Cumulo                   |
| Intended Commission Rate     | 5% (0.05)                |
| Intended Max Commission Rate | 20% (0.20)               |
| Intended Max Change Rate     | 5% (0.05)                |
| Self-Delegation Amount       | 500,000,000 unxrl        |

## Infrastructure

| Field                      | Your Response                                                                 |
| -------------------------- | ----------------------------------------------------------------------------- |
| Hosting Provider           | Velia                                                                         |
| Operating System           | Ubuntu 22.04 LTS                                                              |
| CPU Cores                  | 8 vCPU                                                                        |
| RAM (GB)                   | 16 GB                                                                         |
| Disk Size & Type           | 200 GB NVMe                                                                   |
| Network Speed              | 1 Gbps                                                                        |
| Static IP Available?       | Yes                                                                           |
| Geographic Region          | Multiple regions (EU, US, CA)                                                 |
| Redundant Power?           | Yes                                                                           |
| Redundant Network?         | Yes                                                                           |
| Monitoring Setup           | Grafana + Prometheus with active alerting; custom Front-Chain dashboards for real-time node, endpoint and validator metrics |
| Backup / Snapshot Strategy | Regular snapshots and state sync endpoints published publicly at cumulo.pro/services |

## Experience

| Field                           | Your Response                                                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Years Running Validators        | 5+                                                                                                                 |
| Chains Previously Validated     | Cosmos Hub, Celestia, Avail, Story, Dymension, SEDA, Starknet, Concordium, XRPL EVM, Warden Protocol, GenLayer, Fuel (mainnet and testnet) |
| Cosmos SDK Experience?          | Yes                                                                                                                |
| Tendermint/CometBFT Experience? | Yes                                                                                                                |
| Key Management Practice         | Horcrux (distributed MPC signing — threshold signature scheme for validator keys)                                  |
| Incident Response Experience    | Active 24/7 monitoring with automated alerts. Established incident response process with <24h coordinator response SLA. Experience coordinating upgrades and chain restarts across multiple networks. |

## Commitments

Please confirm each statement with your initials or a check:

| #  | Statement                                                                              | Confirmed |
| --- | -------------------------------------------------------------------------------------- | --------- |
| 1  | I understand this is a TESTNET only. No mainnet is live.                               | ✅        |
| 2  | I understand testnet tokens have ZERO monetary value and cannot be sold or exchanged.  | ✅        |
| 3  | I understand this is NOT a token sale, investment, or financial opportunity.           | ✅        |
| 4  | I will run my validator on a Linux host (not macOS Docker Desktop).                    | ✅        |
| 5  | I will maintain reasonable uptime and respond to coordinator communications.           | ✅        |
| 6  | I understand testnet state may be wiped or reset at any time.                          | ✅        |
| 7  | I will report security issues through the designated reporting process.                | ✅        |
| 8  | I will not make public claims about NXRL having monetary value or being an investment. | ✅        |
| 9  | I agree to the NexaRail testnet code of conduct.                                       | ✅        |
| 10 | I understand my validator can be removed from the active set via governance.           | ✅        |

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
