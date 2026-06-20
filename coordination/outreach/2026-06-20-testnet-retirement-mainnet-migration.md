# Testnet retirement + mainnet migration — broadcast

_Drafted 2026-06-20 by Clove. Post to wherever testnet operators live: testnet Discord channel, validator Telegram, and direct DM to NodeSync + UTSA (who are the only two non-coord testnet validators; they also receive the urgent activation DM separately)._

**Context for Bradley:** as of today, the 5 local-coord testnet validators are permanently disabled at the LaunchDaemon level (the system-wide plists are still on disk for emergency restore, but `launchctl disable` prevents them from auto-starting). Mainnet is the only live NexaRail chain from this point on. The testnet chain (`nexarail-testnet-1`) has stopped producing blocks since the coord quorum is down — any operator still running a testnet node will see height freeze and `catching_up=false` with no new commits coming in. This message tells them what happened and where to go next.

---

## Single broadcast message (Discord / Telegram / channel post)

```
NexaRail testnet (nexarail-testnet-1) is retired as of 2026-06-20.

Mainnet (nexarail-mainnet-1) has been live since 2026-06-18 18:00 UTC and is now the only NexaRail chain operating in production. Coord validators on testnet have been shut down — if your testnet node is still up, you'll see block production stopped on its end.

What this means:
• Testnet had served its purpose: validator readiness, slashing rehearsal, peer-mesh proving, faucet pipeline calibration. All gates green.
• Going forward, all validator operations, faucet grants, slashing, and onboarding happen on mainnet only.
• Your testnet priv_validator_key.json has no role on mainnet — different chain, different consensus keys. Archive it; do not import it anywhere.

Joining mainnet:

Chain:        nexarail-mainnet-1
Genesis:      2026-06-18 18:00 UTC, currently at block 9,200+
Block time:   ~5 seconds
Slashing:     live from block 1 (downtime + double-sign)
Status page:  https://bookings-cpu.github.io/nexarail-status/

Binary (linux/amd64):
  https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64

Genesis file:
  https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json
  SHA256: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

Persistent peers (use either or both — same node ID):

  IPv6 (recommended):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656

  IPv4 (via bore.pub community relay during the sentry rollout):
    96e659f9a87723304dcd614e3ca89d9b6daf26cc@bore.pub:32656

Standard mainnet faucet grant for new validators (drawn from ecosystem_grants):
  500 NXRL self-bond + 100 NXRL gas = 600 NXRL total

To request a grant, reply in #validator-onboarding with:
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

Important: NEW operators only need to do the full onboarding flow above. If you already submitted a mainnet gentx during the genesis window (NodeSync, UTSA) you're already in the active set — you've received a separate DM with the activation steps.

Questions: tag @Bradley.
```

---

## Distribution

1. **Testnet Discord channel** (whatever you've been using for testnet ops). Pin it.
2. **Validator Telegram** (if you've been using one).
3. **DM to NodeSync** — say "see also the urgent activation DM I sent — that one is your specific path".
4. **DM to UTSA** — same as above.
5. Once posted, link this message in the existing `mainnet-onboarding-pin-2026-06-19.md` Discord pin so newcomers find both.

## After this goes out

- Delete or archive any testnet faucet bots that were still running.
- Keep `~/workspace/nexarail/rehearsals/...` data on disk for now — it's research material, not live state.
- Treat any future "can I join testnet?" question as "you mean mainnet" by default. Policy locked in MEMORY.md.
