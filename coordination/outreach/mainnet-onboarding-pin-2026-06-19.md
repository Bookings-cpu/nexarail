# Discord pin — mainnet onboarding (paste 1:1)

Use this in the validator channel. Keep it pinned once mainnet status URL is published.

---

```
📍 MAINNET ONBOARDING — `nexarail-mainnet-1`

Chain:     nexarail-mainnet-1
Denom:     unxrl  (1 NXRL = 1,000,000 unxrl)
Min gas:   0.025unxrl
Binary:    https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64
Genesis:   https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json
SHA256:    f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

Persistent peers (use either or both — same node ID):

  IPv6 (recommended, direct, low latency):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656

  IPv4 (via bore.pub free TCP relay; works through CGNAT):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@bore.pub:32656

  Note: bore.pub is a community-run free service (single server, no SLA). It
  works today but a permanent v4 ingress (sentry node) is on the roadmap.

Slashing IS live:
  • Downtime: >50% missed in last 10,000 blocks → 600s jail, 0.01% slash
  • Double-sign: 5% slash + permanent tombstone
  • Don't run two nodes with the same priv_validator_key.json. Ever.

Standard faucet grant per new validator:
  500 NXRL self-bond + 100 NXRL gas = 600 NXRL total

To request a grant, post in this thread:
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

Status page:
  https://bookings-cpu.github.io/nexarail-status/

Questions: tag @Bradley.
```
