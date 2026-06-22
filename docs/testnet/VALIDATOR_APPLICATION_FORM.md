## Operator Information

| Field                  | Your Response                                |
| ---------------------- | -------------------------------------------- |
| Operator Name          | owlstake                                     |
| Organisation (if any)  | owlstake.com                                 |
| Country / Jurisdiction | Vietnam (VN)                                 |
| Contact Email          | work@owlstake.com                            |
| Discord Handle         | owlstake                                     |
| Telegram Handle        | @bettermanzZ                                 |
| GitHub Handle          | owlstake                                     |
| Website (optional)     | [Hompage](https://owlstake.com), [Analytics](https://analytics.owlstake.com), [IBC Explorer](https://compass.owlstake.com)                         |

## Validator Information

| Field                        | Your Response            |
| ---------------------------- | ------------------------ |
| Validator Moniker            | owlstake                 |
| Intended Commission Rate     | 5% (0.05)                |
| Intended Max Commission Rate | 20% (0.20)               |
| Intended Max Change Rate     | 5% (0.05)                |
| Self-Delegation Amount       | 500,000,000 unxrl        |

## Infrastructure

| Field                      | Your Response                                                                 |
| -------------------------- | ----------------------------------------------------------------------------- |
| Hosting Provider           | BizFlyCloud                                                                   |
| Operating System           | Ubuntu 24.04 LTS                                                              |
| CPU Cores                  | 128 vCPU                                                                      |
| RAM (GB)                   | 256 GB                                                                        |
| Disk Size & Type           | 4000 GB NVMe                                                                  |
| Network Speed              | 10 Gbps                                                                        |
| Static IP Available?       | Yes                                                                           |
| Geographic Region          | VN                                                                            |
| Redundant Power?           | Yes                                                                           |
| Redundant Network?         | Yes                                                                           |
| Monitoring Setup           | We deploy Grafana, Prometheus, Node Exporter, Alertmanager, Zabbix, and network-specific custom scripts for comprehensive observability. Alerts are delivered via Telegram and/or Discord for critical events including missed blocks, low disk space, high CPU/RAM usage, service outages, peer connection issues, and anomalous node behavior. For Cosmos/Tendermint-based chains, Tenderduty is additionally utilized for validator monitoring.                            |
| Backup / Snapshot Strategy | Backup servers are maintained for rapid emergency recovery. We provide public RPC/API endpoints, regular snapshots, and state-sync support, all accessible at: https://services.owlstake.com |

## Experience

| Field                           | Your Response                                                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Years Running Validators        | 5+                                                                                                                 |
| Chains Previously Validated     | Terra, Near, AtomOne, Juno, Starknet, Seda, Shentu, Esxpresso (mainnet and testnet) |
| Cosmos SDK Experience?          | Yes                                                                                                                |
| Tendermint/CometBFT Experience? | Yes                                                                                                                |
| Key Management Practice         | Validator Key Management:Validator keys are secured through encrypted backups, strict access controls, and well-documented recovery procedures. We implement a robust key management process with regular backup verification and scheduled key rotation where applicable. For high-security setups, we also support remote signer solutions (e.g., TMKMS).             |
| Incident Response Experience    | Incident Response:We provide 24/7 real-time monitoring with automatic alerts via Telegram and Discord, rapid incident response, and immediate coordination with the team.        |

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
