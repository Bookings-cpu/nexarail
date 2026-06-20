# Mainnet launch messaging — 2026-06-18

All paste-ready. Use in order: NodeSync/UTSA DM → Discord pre-launch → Discord post-launch → X.

---

## 1. NodeSync + UTSA DM — send now / first thing tomorrow

```
NexaRail mainnet-1 genesis is published.

Release:   https://github.com/Bookings-cpu/nexarail/releases/tag/mainnet-genesis-nexarail-mainnet-1
SHA256:    f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32
Genesis:   2026-06-18T18:00:00Z (chain starts here)
Binary:    same validator-recovery-hotfix you used on testnet

Your mainnet gentx is merged. At launch your node will come up with:
  - 7,142,857 NXRL allocated to your delegator address (vesting linearly over 1 year)
  - 500 NXRL self-bonded (active set, block 1)
  - ~7,142,357 NXRL liquid for post-launch use

Drop the genesis at:
  ~/.nexarail-mainnet/config/genesis.json

Set minimum-gas-prices = "0.025unxrl" in app.toml.

Persistent peer string from us: I'll publish a public coord endpoint within
a few hours of launch — there's a router-side piece I'm finishing this
morning. You can start your node now and it will sit at the
"waiting for genesis time" screen. At T-0 (18:00 UTC) it'll wake; reload
persistent_peers once I post them and it'll connect.

Block 1 fires when 2/3+ voting power is online. Coord set alone is
99.99% — chain will start with or without you in the first minutes, but
your validator is already in the active set per the genesis.

— Bradley
```

---

## 2. Discord #announcements — pre-launch (send T-6 to T-1 hours, ~12:00-17:00 UTC)

```
🚀 NexaRail mainnet-1 launches today at 18:00 UTC

Genesis package:   https://github.com/Bookings-cpu/nexarail/releases/tag/mainnet-genesis-nexarail-mainnet-1
Genesis SHA256:    f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32
Binary:            https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64

Chain ID:          nexarail-mainnet-1
Denom:             unxrl  (1 NXRL = 1,000,000 unxrl)
Total supply:      1,000,000,000 NXRL (fixed-cap)
Validator set @ genesis:  5 coordinator + NodeSync + UTSA
Live product flags:       all OFF, flip via governance

Note for new operators: mainnet is OPEN to additional validators
post-launch via standard `tx staking create-validator`. You'll need
NXRL — testnet unxrl is NOT redeemable. We're not running a token sale
or airdrop right now; community/airdrop allocation exists in genesis
and will be distributed on a later schedule.

Slashing is ON:
  • downtime jail at >50% missed in last 10,000 blocks, 600s jail, 0.01% slash
  • double-sign 5% slash + permanent tombstone

Status page (will go live at T-0):  <add URL here once GH Pages is set up>

Block 1 produces at 18:00 UTC. See you on chain.
```

---

## 3. Discord #announcements — post-launch (send after block 1, confirm with `nexaraild status`)

```
🟢 NexaRail mainnet-1 is LIVE

Block 1 signed at <timestamp UTC>
Active validators at launch: <fill in count, expect 5-7>
Genesis hash confirmed: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

What's working tonight:
  • Block production
  • Standard staking (delegate, redelegate, unbond)
  • Gov module (you can submit proposals against placeholder min-deposit)
  • Slashing (don't get jailed)

What's NOT working tonight (intentional):
  • Live product flags (payout / settlement / escrow / treasury) — off
  • These flip on by governance, not at launch

Coming next:
  • Public RPC endpoint
  • Public status page
  • First governance proposal (live-flags vote schedule)
  • New validator onboarding open via standard cosmos staking flow

Drop your `nexaraild status` output here so we can confirm you're on the
right chain.

GG everyone who shipped this.
```

---

## 4. X / Twitter — post-launch (send 30 min after block 1, once chain has produced ~150 blocks cleanly)

```
NexaRail mainnet is live.

Chain ID: nexarail-mainnet-1
Block 0 timestamp: 2026-06-18T18:00:00Z

Fixed supply: 1 000 000 000 NXRL
Validator set at genesis: 5 coordinator + 2 external (NodeSync, UTSA)
Live product flags: all off, flipped by governance only.

Genesis package + SHA256 in the repo. No token sale; no airdrop snapshot
ran tonight; testnet unxrl is not redeemable.

Validators wanting in — DM, or stake into the existing set.

🍀
```

---

## 5. Optional — personal X post (Bradley, founder voice)

```
1/  Eighteen months from "what if rails for X" sketched on a flight to
    "block one signed".

2/  NexaRail mainnet started today at 18:00 UTC. Fixed-cap NXRL, no token
    sale, live product flags off and gated to governance from the first
    block. Validator set is professional from day one.

3/  We're a public chain now. The hard part isn't launching, it's the
    next twelve months of governance, security, and operator hygiene.

4/  To everyone who validated testnet, ran nodes, found bugs, sent the
    gentxs that got bounced because I'd written the wrong amount in the
    onboarding pack — thank you. You're the chain.

5/  🍀
```

---

## Sequencing — for the day of launch

| Time UTC | Action |
|---|---|
| 06:00 | Verify mainnet validators still running (`launchctl list | grep nexarail.mainnet`) |
| 06:30 | Resolve router port-forward TCP 32656 → Mac mini #1 LAN IP. Get public IP, publish peer string. |
| 09:00 | Send NodeSync + UTSA the DM (message 1 above), include the peer string. |
| 12:00 | Send Discord pre-launch (message 2) to #announcements. |
| 17:00 | Final pre-launch checks: validators still alive, network reachable, eyes on terminal. |
| 18:00 | Block 1 fires automatically. No action needed. Watch the logs. |
| 18:05 | Confirm block production with a node status query. |
| 18:15 | Post message 3 to Discord #announcements once block ~150 lands. |
| 18:30 | Post X message 4. |
| 19:00 | Founder thread message 5 (optional). |

---

## If something goes wrong

- If only some coords come up at 18:00: chain still produces blocks as long as the alive coords have 2/3+ power (they do — any 4 of 5 = 80% bonded).
- If NO coords come up: there's an infra problem at your end. Unload LaunchAgents, debug, restart. Until 2/3 voting power is online the chain doesn't progress past height 0.
- If gentx hashes don't match externally: someone tampered with the genesis. Hash verification: `shasum -a 256 ~/.nexarail-mainnet/config/genesis.json` must equal `f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32`.
- If a coord double-signs (catastrophic — 5% slash + permanent tombstone): immediately unload the duplicate's plist, do NOT bring it back, do NOT recover the same priv_validator_key. Spin up a fresh coord operator key via Ledger to replace it, with a `tx staking create-validator` once chain is live.

---

Stay calm tomorrow. The work is done.
