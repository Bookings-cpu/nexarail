# Validator broadcast — mainnet open for onboarding

Paste-ready single message for the validator community (Discord #announcements, Telegram, validator-channel email blast). Authoritative as of 2026-06-19 evening — v4 ingress is now live via bore.pub.

---

```
NEXARAIL MAINNET IS LIVE AND OPEN FOR VALIDATOR ONBOARDING

Chain:        nexarail-mainnet-1
Native denom: unxrl  (1 NXRL = 1,000,000 unxrl)
Genesis time: 2026-06-18 18:00:00 UTC
Min gas:      0.025unxrl
Block time:   ~5 seconds
Active set:   7 validators (5 coord + NodeSync + UTSA)
Status page:  https://bookings-cpu.github.io/nexarail-status/

Binary (linux/amd64):
  https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64

Genesis file:
  https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json
  SHA256: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

Persistent peers (use either or both — same node ID):

  IPv6 (recommended, direct, low latency):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656

  IPv4 (via bore.pub free TCP relay; works through CGNAT):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@bore.pub:32656

SLASHING IS LIVE FROM BLOCK 1
  • Downtime: >50% missed in last 10,000 blocks → 600s jail, 0.01% slash
  • Double-sign: 5% slash + permanent tombstone
  • Do NOT run two nodes with the same priv_validator_key.json. Ever.

Standard faucet grant for new validators (drawn from ecosystem_grants):
  500 NXRL self-bond + 100 NXRL gas = 600 NXRL total

To request a grant, reply in this channel with:
  Mainnet validator onboarding request
  Moniker:          <your moniker>
  Operator address: nxr1...
  Consensus pubkey: nxrvalconspub1...
  Self-bond target: <X> NXRL          (default: 500)
  Gas budget:       100 NXRL
  Total grant:      <X + 100> NXRL
  Contact:          <how to reach you>

Full step-by-step (install → init → genesis → sync → create-validator):
  https://github.com/Bookings-cpu/nexarail/blob/main/docs/mainnet/NEW_VALIDATOR_ONBOARDING.md

Questions: tag @Bradley.
```

---

## Distribution

- Discord #announcements + #validator-onboarding (also pin this in #validator-onboarding)
- Telegram channels for validator communities
- Direct DM to anyone who's expressed interest but not onboarded yet
- Email blast to the partner shortlist

## Notes

- v4 via bore.pub is a free TCP relay (`github.com/ekzhang/bore`) — community-run, no SLA. Permanent v4 sentry is on the roadmap.
- The IPv6 endpoint uses the Mac's stable autoconf address — survives reboots and ISP prefix renewals as long as the /64 stays the same.
- Status page auto-refreshes every 30s; backend pushes every 5 min via the `com.nexarail.status-publish` LaunchAgent.
