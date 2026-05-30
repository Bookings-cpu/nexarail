# Validator Application Form — NexaRail Testnet

**Chain:** nexarail-testnet-1
**Phase:** Controlled Registration
**Status:** ⚠️ Testnet only — no mainnet, no token sale, no monetary value

---

Complete this form to apply as a validator on the NexaRail controlled testnet.

## Operator Information

| Field | Your Response |
|---|---|
| Operator Name | LuckyStar |
| Organisation (if any) | |
| Country / Jurisdiction | VietNam |
| Contact Email | aluckystarasia@gmail.com |
| Discord Handle | @luckystar.asia |
| Telegram Handle | @LuckyStarAsia |
| GitHub Handle | LuckyStarAsia |
| Website (optional) | |

## Validator Information

| Field | Your Response |
|---|---|
| Validator Moniker | LuckyStar |
| Intended Commission Rate | 10% |
| Intended Max Commission Rate | 20% |
| Intended Max Change Rate | 5% |
| Self-Delegation Amount | 500,000,000 unxrl |

## Infrastructure

| Field | Your Response |
|---|---|
| Hosting Provider | Homelab & Hetner |
| Operating System | Ubuntu 24.04|
| CPU Cores | 16-104 vCPU |
| RAM (GB) | 16 - 256 GB |
| Disk Size & Type | 0.4 - 4 TB NVME|
| Network Speed | 0.5 - 1 Gbps|
| Static IP Available? | Yes |
| Geographic Region | EU/ASIA |
| Redundant Power? | Yes |
| Redundant Network? | Yes |
| Monitoring Setup | My nodes, validators running on Hetzer Dedicated servers & HomeLab. Have another servers running for backup, snapshot, api, rpc…services. My validator is handled from ASia & monitoring with Node exporter, prometheus, alert manager, Grafana & alert by telegram bot & Discord. With cosmos eco, i usually use Cosmovisor for easy prepare an upgrade at upgrade block height, it will help & keep network stable even i busy at that time. I build my explorer at https://explorer.luckystar.asia to help me & other validators can monitor validator status. |
| Backup / Snapshot Strategy | Setup some services: guides, daily snapshots & enpoints.... at https://luckystar-1.gitbook.io/luckystar.asia, https://github.com/LuckyStarAsia |

## Experience

| Field | Your Response |
|---|---|
| Years Running Validators | 4+ |
| Chains Previously Validated | I have running nodes & validators on mainnet & testnet for: Cosmos, Atomone, Axone, Orai, Sunrise, Union, Seda, Espresso, Bitbadges, Arkeo, Althea, Kopi protocol, Story, dHealth, Autonity blockchain, Tanssi network, Side protocol, Airchains, Warden, Hedge, Empeiria, Hemi, Rainbow protocol, Nubit, Allora, initia network, Entangle, AlignedLayer... |
| Cosmos SDK Experience? | Yes |
| Tendermint/CometBFT Experience? | Yes |
| Key Management Practice | TMKMS & backup offline |
| Incident Response Experience | With those setup above, my system is monitoring 24/7 & i usualy response quickly in 1-2h |

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
