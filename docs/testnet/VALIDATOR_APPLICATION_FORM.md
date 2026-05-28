## Operator Information

| Field                  | Your Response                                |
| ---------------------- | -------------------------------------------- |
| Operator Name          | UTSA                                         |
| Organisation (if any)  | UTSA (Chain House)                           |
| Country / Jurisdiction | Armenia (AM)                                 |
| Contact Email          | validator@utsa.tech / lesnik13utsa@yandex.ru |
| Discord Handle         | lesnik_utsa                                  |
| Telegram Handle        | @lesnik13utsa                                |
| GitHub Handle          | lesnikutsa                                   |
| Website (optional)     | https://utsa.gitbook.io/services             |

## Validator Information

| Field                        | Your Response            |
| ---------------------------- | ------------------------ |
| Validator Moniker            | UTSA                     |
| Intended Commission Rate     | 5% (0.05)                |
| Intended Max Commission Rate | 20% (0.20)               |
| Intended Max Change Rate     | 5% (0.05)                |
| Self-Delegation Amount       | 500,000,000 unxrl        |

## Infrastructure

| Field                      | Your Response                                                                 |
| -------------------------- | ----------------------------------------------------------------------------- |
| Hosting Provider           | Hetzner / Mevspase                                                            |
| Operating System           | Ubuntu 22.04 LTS                                                              |
| CPU Cores                  | 32 vCPU                                                                       |
| RAM (GB)                   | 128 GB                                                                        |
| Disk Size & Type           | 2000 GB NVMe                                                                  |
| Network Speed              | 1 Gbps                                                                        |
| Static IP Available?       | Yes                                                                           |
| Geographic Region          | Multiple regions (EU, US, CA)                                                 |
| Redundant Power?           | Yes                                                                           |
| Redundant Network?         | Yes                                                                           |
| Monitoring Setup           | We use Grafana, Prometheus, Node Exporter, Alertmanager, Zabbix, and custom scripts depending on the network. Alerts are sent to Telegram and/or Discord for issues such as missed blocks, low disk space, high RAM/CPU usage, service downtime, peer issues, or abnormal node behavior. For Cosmos/Tendermint-based networks, we also use tools like Tenderduty                            |
| Backup / Snapshot Strategy | There are backup servers for emergency recovery. We support public RPC/API interfaces, snapshots, and state synchronization in a publicly accessible location at https://utsa.gitbook.io/services |

## Experience

| Field                           | Your Response                                                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Years Running Validators        | 5+                                                                                                                 |
| Chains Previously Validated     | Atomone, Celestia, Provenance, Avail, Story, Dymension, Polkadot, Kusama, Minima, Warden, Nibiru, Solana, Concordium, Realio, Dora, Haqq, Crossfi, Espresso, Uptick and other (mainnet and testnet) |
| Cosmos SDK Experience?          | Yes                                                                                                                |
| Tendermint/CometBFT Experience? | Yes                                                                                                                |
| Key Management Practice         | Validator keys are protected using encrypted backups, strict access control, and documented recovery procedures. We follow a secure key-management process with regular backup verification and key-rotation practices where needed. For critical validator setups, we also support remote-signer             |
| Incident Response Experience    | 24/7 monitoring with automatic alerts, including via Telegram and Discord, rapid response to incidents, and prompt coordination with the team        |

## Commitments

Please confirm each statement with your initials or a check:

| #  | Statement                                                                              | Confirmed |
| --- | -------------------------------------------------------------------------------------- | ---------|
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
