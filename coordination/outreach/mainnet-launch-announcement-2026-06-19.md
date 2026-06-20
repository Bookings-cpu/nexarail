# Mainnet launch announcement — copy pack

Drafted 2026-06-19 by Clove. Bradley to review then send.

Context: `nexarail-mainnet-1` went live 2026-06-18T18:00:00Z. Quorum confirmed, 7-validator active set (5 coord + NodeSync + UTSA), block production ~5.3s avg, status page now public at https://bookings-cpu.github.io/nexarail-status/. v4 ingress live via bore.pub free TCP relay; v6 endpoint direct to alpha. See `mainnet-validator-broadcast-2026-06-19.md` for the single canonical paste-ready validator message.

---

## 1. X / Twitter — single post (under 280)

```
NexaRail mainnet is live.

Chain ID: nexarail-mainnet-1
Launched: 2026-06-18 18:00 UTC
Validators: 7 active (5 coord + 2 external)
Block time: ~5s

Live status: https://bookings-cpu.github.io/nexarail-status/
Validator onboarding: https://github.com/Bookings-cpu/nexarail
```

## 2. X / Twitter — thread (4 posts)

**Post 1**
```
NexaRail mainnet is live.

Chain nexarail-mainnet-1 went live yesterday at 18:00 UTC. 7-validator active set, ~5s blocks, slashing armed.

Status: https://bookings-cpu.github.io/nexarail-status/
```

**Post 2**
```
Genesis allocation locked at 1B NXRL across 13 buckets. Coordinator validators (alpha-echo) self-bonded 5M NXRL each at genesis with min_self_delegation 2.5M. External validators capped at 500 NXRL placeholder bonds during the open onboarding window.
```

**Post 3**
```
Slashing live from block 1:
- Downtime: >50% missed in last 10,000 blocks → 600s jail, 0.01% slash
- Double-sign: 5% slash + permanent tombstone

Don't run two nodes with the same priv_validator_key.json. Ever.
```

**Post 4**
```
Validator onboarding is open. Standard faucet grant: 500 NXRL self-bond + 100 NXRL gas.

Genesis, binary, onboarding runbook, peer endpoints — everything is in the repo:
https://github.com/Bookings-cpu/nexarail

DM to coordinate.
```

## 3. Discord — #announcements

```
NEXARAIL MAINNET IS LIVE

Chain:    nexarail-mainnet-1
Launched: 2026-06-18 18:00 UTC
Status:   https://bookings-cpu.github.io/nexarail-status/

Active set: 7 validators (5 coord + NodeSync + UTSA)
Block time: ~5 seconds
Slashing:   live from block 1 (downtime + double-sign)

Faucet grant for new validators: 500 NXRL self-bond + 100 NXRL gas.
Genesis SHA256: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

Onboarding runbook + peer endpoints:
https://github.com/Bookings-cpu/nexarail/blob/main/docs/mainnet/NEW_VALIDATOR_ONBOARDING.md

Drop your onboarding request in #validator-onboarding using the format pinned there.
```

## 4. Site / blog — short

```
NexaRail mainnet — nexarail-mainnet-1 — went live on 2026-06-18 at 18:00 UTC.

The network launched with a seven-validator active set: five coordinator validators run by the NexaRail core operations team, plus two community validators (NodeSync and UTSA) confirmed during the rehearsal window. Coordinator validators self-bonded 5,000,000 NXRL each at genesis with a 2,500,000 NXRL minimum self-delegation, ensuring tight alignment between operators and chain liveness during the initial onboarding window.

Slashing is live from block one. Downtime triggers a 600-second jail and 0.01% stake reduction after a 50% miss rate over 10,000 blocks. Double-signing triggers a 5% slash and permanent tombstone — operators must not run two nodes with the same consensus key.

Validator onboarding is open. The standard faucet grant is 600 NXRL (500 NXRL self-bond plus 100 NXRL gas), drawn from the 150M NXRL ecosystem grants bucket. The full onboarding runbook, genesis file, binary, and persistent peer endpoints are published in the public NexaRail repository. Live network status is available at https://bookings-cpu.github.io/nexarail-status/

For partnership or validator enquiries: Bradley Johnston, contact via the NexaRail repository discussions or Discord.
```

## 5. Email blast — short (~140 words)

```
Subject: NexaRail mainnet is live

NexaRail mainnet — chain ID nexarail-mainnet-1 — went live on 2026-06-18 at 18:00 UTC with a seven-validator active set and a 1B NXRL genesis supply. Block production is averaging 5.3 seconds, slashing is armed from block one, and the network is open to additional validators.

Live status: https://bookings-cpu.github.io/nexarail-status/
Onboarding runbook + genesis file: https://github.com/Bookings-cpu/nexarail
Genesis SHA256: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

If you operate validator infrastructure and want to join the active set during this window, reply to this email or open an issue in the repo. Standard onboarding grant is 600 NXRL (500 NXRL self-bond + 100 NXRL gas).

Bradley Johnston
NexaRail
```

---

## Distribution checklist (Bradley to action)

- [ ] X post / thread — paste from section 1 or 2
- [ ] Discord #announcements — paste from section 3
- [ ] Pin updated onboarding text in #validator-onboarding (already drafted at `mainnet-onboarding-pin-2026-06-19.md`)
- [ ] Site / blog post — section 4 (if site exists; otherwise skip)
- [ ] Direct email to partner shortlist — section 5

## Internal notes (do not publish)

- Router IPv4 port-forward (188.30.133.232:32656) is still pending — until that's open, validators must use the IPv6 peer endpoint. The pin already discloses this; no need to call it out in the announcement.
- Faucet wallet mnemonic was exposed in chat 2026-06-16; rotation/Ledger migration is post-launch hardening. Do not mention.
- Custody seed JSONs at ~/Documents/NEXARAIL-MAINNET-CRITICAL-DO-NOT-SHARE/ should be paper-backed and shredded before any wide announcement.
