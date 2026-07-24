# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field | Your Response |
|---|---|
| Operator Name | NODERS |
| Organisation (if any) | NODERS LLC |
| Country / Jurisdiction | Caymans |
| Contact Email | office@noders.team |
| Discord Handle | noders |
| Telegram Handle | @septima_noders |
| GitHub Handle | noders-team |
| Website (optional) | https://noders.team/ |

## Validator Information

| Field | Your Response |
|---|---|
| Validator Moniker | [NODERS] |
| Intended Commission Rate | 0.1 |
| Intended Max Commission Rate | 0.1 |
| Intended Max Change Rate | 0.01 |
| Self-Delegation Amount | 500,000,000 |

## Infrastructure

| Field | Your Response |
|---|---|
| Hosting Provider | Vultr |
| Operating System | Ubuntu 24.04 |
| CPU Cores | 16 |
| RAM (GB) | 128 |
| Disk Size & Type | 4tb nvme |
| Network Speed | 1Gbps |
| Static IP Available? | Yes |
| Geographic Region | Africa/Asia/US/EU |
| Redundant Power? | Yes |
| Redundant Network? | Yes |
| Monitoring Setup | Prometheus\Grapaha stack, Alerta, Multiple notification channels, 24\7 devops on-call |
| Backup / Snapshot Strategy | Keys backuped securelly with BitWarden / Snapshots every 12hrs |

## Experience

| Field | Your Response |
|---|---|
| Years Running Validators | 5 |
| Chains Previously Validated | Solana,Sui,Monad,Story,Canton,Berachain,Polygon ,IOTA,Starknet,Celestia,Nillion,Stader,IKA,Gnosis,Fuel,Autonity,SSV,Obol,Namada,Avail,Dymension,Seda,Supra,Zetachain,Haqq,Realio,AtomOne,NYM,Aura,Aura,Stratos,HumansAI,Uptick,Stake,Jackal |
| Cosmos SDK Experience? | Yes |
| Tendermint/CometBFT Experience? | Yes |
| Key Management Practice | HSM/Horcrux |
| Incident Response Experience | We have various expertise on all levels of chain functionallity, including halts, launches, updates etc. |

## Commitments

Please confirm each statement with your initials or a check:

| # | Statement | Confirmed |
|---|---|---|
| 1 | I understand this is a TESTNET only. No mainnet is live. | [v] |
| 2 | I understand testnet tokens have ZERO monetary value and cannot be sold or exchanged. | [v] |
| 3 | I understand this is NOT a token sale, investment, or financial opportunity. | [v] |
| 4 | I will run my validator on a Linux host (not macOS Docker Desktop). | [v] |
| 5 | I will maintain reasonable uptime and respond to coordinator communications. | [v] |
| 6 | I understand testnet state may be wiped or reset at any time. | [v] |
| 7 | I will report security issues through the designated reporting process. | [v] |
| 8 | I will not make public claims about NXRL having monetary value or being an investment. | [v] |
| 9 | I agree to the NexaRail testnet code of conduct. | [v] |
| 10 | I understand my validator can be removed from the active set via governance. | [v] |

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
