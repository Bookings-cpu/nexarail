# Mainnet operations runbook

_Living document. Owner: Bradley. Last touched 2026-06-20 by Clove._

This is the consolidated procedure book for running the live `nexarail-mainnet-1` chain on Mac mini #1. It assumes the post-launch state captured in `MEMORY.md` and complements the public [`NEW_VALIDATOR_ONBOARDING.md`](./NEW_VALIDATOR_ONBOARDING.md). External operators do not need this file — it documents the coordinator side.

---

## 1. Where everything lives

```
Source / repo:          ~/workspace/nexarail
Build binary (live):    ~/workspace/nexarail/build/nexaraild
                        (used by the 5 coord LaunchAgents)
Coord home dirs:        ~/.nexarail-mainnet-{alpha,bravo,charlie,delta,echo}
Operator keyring (OS):  ~/.nexarail-mainnet-keys           (macOS Keychain backend)
LaunchAgents (user):    ~/Library/LaunchAgents/
LaunchAgent logs:       ~/Library/Logs/nexarail-*
Backups (chmod 600/700) ~/Library/Application Support/nexarail/
  ├── backups/keys/<date>/<coord>/   (priv_validator_key + node_key + state)
  ├── backups/state/<coord>/<hour>.json   (hourly priv_validator_state)
  ├── backups/logs/<rotated logs>
  └── snapshots/<YYYY-MM-DD>.txt          (daily nexarail-mainnet-snap output)
```

## 2. LaunchAgents — quick reference

| Label | Role | Trigger |
|---|---|---|
| `ai.nexarail.mainnet.{alpha,bravo,charlie,delta,echo}` | The 5 coord validators | KeepAlive=true |
| `com.nexarail.bore-mainnet-p2p` | IPv4 ingress relay via bore.pub | KeepAlive=true (until sentry) |
| `com.nexarail.mainnet-status-collector` | Pulls chain state for the public page | StartInterval (5 min) |
| `com.nexarail.status-publish` | Pushes the JSON to GitHub Pages | StartInterval |
| `com.nexarail.mainnet-validator-watchdog` | Per-minute health tick, alerts | StartInterval 60 s |
| `com.nexarail.watchdog-notifier` | Tails alert log, fires macOS notifications | KeepAlive=true |
| `com.nexarail.mainnet-state-snapshot` | Hourly `priv_validator_state.json` backup | StartInterval 3600 s |
| `com.nexarail.log-rotate` | Daily log rotation (copytruncate, 14-day retention) | CalendarInterval 00:30 UTC |
| `com.nexarail.mainnet-daily-snapshot` | Daily snap to file (90-day retention) | CalendarInterval 00:05 UTC |

Health one-liner:

```bash
~/bin/nexarail-mainnet-snap
```

## 3. Coord validator lifecycle

### Restart a single coord (clean, normal)

```bash
launchctl kickstart -k gui/$(id -u)/ai.nexarail.mainnet.alpha
```

Replace `alpha` with the target coord. The `-k` flag kills first, then starts. KeepAlive=true ensures it stays up after.

### Stop a coord (rare — emergency only)

```bash
launchctl bootout gui/$(id -u)/ai.nexarail.mainnet.alpha
```

To bring it back:

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/ai.nexarail.mainnet.alpha.plist
```

> ⚠️ Stopping a coord while the chain is at 5/7 voting power risk: with 5 coords at 5M NXRL each = 25M NXRL of voting power and 1,000 NXRL between the two externals, losing 1 coord drops the network to 4/5 of coord power, still well above 2/3 quorum. **Do not stop 2 coords simultaneously.**

### Hot config edit (e.g. add a peer)

1. Edit `~/.nexarail-mainnet-<v>/config/config.toml`
2. `cp file.toml file.toml.bak-$(date -u +%FT%H%M%SZ)` first if it's a meaningful change
3. `launchctl kickstart -k gui/$(id -u)/ai.nexarail.mainnet.<v>` to reload

Common edits: `persistent_peers`, `allow_duplicate_ip`, `addr_book_strict`, prometheus.

### Recover a corrupted coord

Worst case: a coord's data dir is destroyed (disk failure, bad shutdown).

```bash
# 1. Stop the coord
launchctl bootout gui/$(id -u)/ai.nexarail.mainnet.alpha

# 2. Move the broken home aside, do NOT delete (forensics)
mv ~/.nexarail-mainnet-alpha ~/.nexarail-mainnet-alpha.broken-$(date -u +%F-%H%M)

# 3. Re-init from scratch
nexaraild init "nxrl-controlled-alpha" --chain-id nexarail-mainnet-1 --home ~/.nexarail-mainnet-alpha

# 4. Drop the mainnet genesis (verify SHA)
curl -L -o ~/.nexarail-mainnet-alpha/config/genesis.json \
  https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json
shasum -a 256 ~/.nexarail-mainnet-alpha/config/genesis.json
# expect: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32

# 5. Restore the BACKED-UP priv_validator_key.json + node_key.json
cp "$HOME/Library/Application Support/nexarail/backups/keys/<latest-date>/alpha/priv_validator_key.json" \
   ~/.nexarail-mainnet-alpha/config/priv_validator_key.json
cp "$HOME/Library/Application Support/nexarail/backups/keys/<latest-date>/alpha/node_key.json" \
   ~/.nexarail-mainnet-alpha/config/node_key.json
chmod 600 ~/.nexarail-mainnet-alpha/config/priv_validator_key.json ~/.nexarail-mainnet-alpha/config/node_key.json

# 6. Re-apply config tweaks (mesh flags etc.)
# See: ~/.nexarail-mainnet-alpha.broken-*/config/config.toml for what was there.

# 7. Start, watch sync
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/ai.nexarail.mainnet.alpha.plist
journalctl … # macOS: tail -f ~/Library/Logs/ai.nexarail.mainnet.alpha.stderr.log
```

> ❗ **Never** restore a stale `priv_validator_state.json` from the hourly snapshot onto a live validator that has already signed past that height. The state snapshots are FORENSIC, not for replay. Let the rebuilt node sign fresh from the height it joins.

### If multiple coords die at once

Chain halts at the height where voting power < 2/3. To recover:

1. Verify each coord's `priv_validator_state.json` last-signed height. The coord with the highest height is the canonical state.
2. Rebuild lower-height coords from genesis + the backed-up `priv_validator_key.json`; let them sync up.
3. Do NOT manually edit `priv_validator_state.json` to bump heights — that risks double-sign.

## 4. External validator operations

### Grant a new validator (post-onboarding-request)

```bash
~/workspace/nexarail/scripts/mainnet/grant-validator.sh "Moniker" nxr1...
# defaults to 600 NXRL standard grant
```

The helper:
- validates the bech32 address format
- prompts `Send? [y/N]` (pipe `y` if non-interactive)
- broadcasts the tx, captures the hash
- logs the result to `coordination/validators/mainnet-faucet-tracker.csv`

Verify on chain after a few seconds:

```bash
nexaraild query bank balances <operator-addr> --node http://127.0.0.1:32657
```

Then DM the operator with the tx hash and a pointer to `NEW_VALIDATOR_ONBOARDING.md`. They run `tx staking create-validator` to enter the active set.

### Confirm a validator is signing

```bash
nexaraild query staking validators --node http://127.0.0.1:32657 \
  --home ~/.nexarail-mainnet-alpha -o json \
  | python3 -c "import json,sys; vs=json.load(sys.stdin)['validators']; \
                 [print(v['description']['moniker'], v['jailed'], v['status']) for v in vs]"
```

Watch alpha's `/net_info`:

```bash
curl -s http://127.0.0.1:32657/net_info | jq '.result.peers[].node_info | {moniker, id}'
```

### An external validator was auto-jailed

This is normal. They missed >50% over 10,000 blocks. Slash = 0.01% (cosmetic, ~0.05 NXRL on 500 bond). Jail period = 600 s.

There's nothing for us to do. Operator must:

1. Bring their node back online and let it sync.
2. After `jailed_until` time has passed, send:
   ```bash
   nexaraild tx slashing unjail <operator-addr> \
     --from <their operator key> \
     --chain-id nexarail-mainnet-1 \
     --node http://localhost:26657 \
     --gas auto --gas-adjustment 1.3 --fees 5000unxrl \
     --keyring-backend <theirs>
   ```

The watchdog will catch the unjail event and stop alerting.

### One of our coords got jailed

⚠️ This means a coord missed too many blocks. Investigate why first (load? RPC unresponsive? wrong config?). Then:

1. Fix the underlying issue.
2. Verify the coord is signing again (check `priv_validator_state.json` height advancing).
3. Send unjail from the coord's operator key:
   ```bash
   nexaraild tx slashing unjail nxrvaloper1<coord-operator> \
     --from <coord-operator-key-name> \
     --keyring-backend os --home ~/.nexarail-mainnet-keys \
     --chain-id nexarail-mainnet-1 \
     --node http://localhost:32657 \
     --gas auto --gas-adjustment 1.3 --fees 5000unxrl --yes
   ```

## 5. Faucet (`ecosystem_grants`)

- Source address: `nxr14mnya6lj7ay43zyy95lhe8qppgr3nys9xte6gh`
- Original balance: 150,000,000 NXRL
- Standard grant: 600 NXRL (500 self-bond + 100 gas)
- Helper: `~/workspace/nexarail/scripts/mainnet/grant-validator.sh`
- Tracker: `~/workspace/nexarail/coordination/validators/mainnet-faucet-tracker.csv`
- Send variant: this fork uses `tx send <from> <to> <amount>` (NOT `tx bank send`).

Check balance:

```bash
nexaraild query bank balances nxr14mnya6lj7ay43zyy95lhe8qppgr3nys9xte6gh \
  --node http://127.0.0.1:32657 -o json | jq
```

## 6. Watchdog + notifier

Live tail:

```bash
tail -f ~/Library/Logs/nexarail-mainnet-validator-watchdog.log
tail -f ~/Library/Logs/nexarail-mainnet-validator-watchdog.alerts.log
```

Force a re-tick (after editing the script):

```bash
launchctl kickstart -k gui/$(id -u)/com.nexarail.mainnet-validator-watchdog
```

Restart the notifier (if alerts stop firing):

```bash
launchctl kickstart -k gui/$(id -u)/com.nexarail.watchdog-notifier
```

## 7. Backups

### One-time key backup (do after any key rotation)

```bash
~/bin/nexarail-mainnet-key-backup.sh
# writes ~/Library/Application Support/nexarail/backups/keys/<today>/<coord>/
```

### Hourly state snapshots

Already running via LaunchAgent. Files at:

```
~/Library/Application Support/nexarail/backups/state/<coord>/<YYYY-MM-DDTHH>.json
```

14-day rolling retention. Forensics only — DO NOT replay.

### Recommended external backups (Bradley, manual)

The local backups protect against most disk-corruption scenarios. For full disaster recovery:

- Copy `~/Library/Application Support/nexarail/backups/keys/<latest>/` to an encrypted USB or 1Password Document. Refresh whenever any key rotates.
- The macOS Keychain backup (Time Machine) covers the operator keyring as long as Time Machine is on.
- Custody seed mnemonics for the operator wallets are paper-backed and shredded per the 2026-06-19 ceremony.

## 8. Public ingress

Today: `bore.pub:32656` via `com.nexarail.bore-mainnet-p2p` LaunchAgent. Free community-run TCP relay, single instance, no SLA.

Roadmap: replace with Oracle Always Free ARM64 sentry per [`SENTRY_NODE_ORACLE_RUNBOOK_2026-06-20.md`](./SENTRY_NODE_ORACLE_RUNBOOK_2026-06-20.md). Pre-built ARM64 binary at `~/workspace/nexarail/build/nexaraild-linux-arm64-v0.1.0-rc1-validator-recovery-hotfix`.

To check bore tunnel state:

```bash
pgrep -fl 'bore local 32656'
launchctl list | grep bore
```

If bore.pub goes down, IPv6 is unaffected (direct to alpha). Externals using v6 keep peering. Externals on v4-only networks lose peer until bore comes back (or sentry replaces it).

## 9. Status page

- JSON: `https://bookings-cpu.github.io/nexarail-status/mainnet-status.json`
- HTML: `https://bookings-cpu.github.io/nexarail-status/`
- Refresh: every 5 minutes via `com.nexarail.mainnet-status-collector` + `com.nexarail.status-publish`
- Backing repo: `Bookings-cpu/nexarail-status` (GitHub Pages)

If status page goes stale (>15 min):

```bash
launchctl kickstart -k gui/$(id -u)/com.nexarail.mainnet-status-collector
launchctl kickstart -k gui/$(id -u)/com.nexarail.status-publish
tail -20 ~/Library/Logs/nexarail-mainnet-status-collector.stdout.log
tail -20 ~/Library/Logs/nexarail-status-publish.stdout.log
```

## 10. Common issues

### "alpha RPC unreachable"

Usually transient under high load. The watchdog tolerates one failed tick. If it persists more than ~3 ticks:

1. `ps -p <alpha-pid>` — confirm process alive
2. `tail -100 ~/Library/Logs/ai.nexarail.mainnet.alpha.stderr.log`
3. If process hung, `launchctl kickstart -k gui/$(id -u)/ai.nexarail.mainnet.alpha`

### Block production slowed

Check load:

```bash
~/bin/nexarail-mainnet-snap | tail -20
uptime
```

Likely culprit: another heavy process on this Mac (a former incident: a testnet daemon was eating CPU + RAM, see `memory/2026-06-20.md`). Use `top -o cpu -l 1 -n 12` to identify and decide whether to throttle.

### Disk space concern

```bash
df -h /
du -sh ~/.nexarail-mainnet-*
du -sh ~/workspace/nexarail/rehearsals
```

Retired testnet data at `~/workspace/nexarail/rehearsals/controlled-testnet/` is ~47 GB and can be archived/deleted if needed (research material per MEMORY.md, not live state — coordinate with Bradley before deleting).

### A new operator says peer fails

Send them the IPv6 endpoint:

```
96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656
```

If they're v4-only, point at `bore.pub:32656` (or the sentry once deployed). Also confirm their `allow_duplicate_ip = true` and `addr_book_strict = false` since alpha-side flags are permissive.

## 11. Reference

- Public onboarding doc: [`docs/mainnet/NEW_VALIDATOR_ONBOARDING.md`](./NEW_VALIDATOR_ONBOARDING.md)
- Sentry runbook: [`docs/mainnet/SENTRY_NODE_ORACLE_RUNBOOK_2026-06-20.md`](./SENTRY_NODE_ORACLE_RUNBOOK_2026-06-20.md)
- Genesis tokenomics: [`docs/mainnet/MAINNET_TOKENOMICS_APPROVED.json`](./MAINNET_TOKENOMICS_APPROVED.json) (local-only; not in public repo)
- Status page repo: `Bookings-cpu/nexarail-status`
- Mainnet binary release: `v0.1.0-rc1-validator-recovery-hotfix`
- Mainnet genesis release: `mainnet-genesis-nexarail-mainnet-1`
- Faucet send policy: locked 2026-06-19 (600 NXRL standard)
- Local daily logs / longer-term memory: `~/.openclaw/workspace/memory/`

## Changelog

- 2026-06-20 — Written by Clove after the post-launch ops session that retired testnet, deployed the watchdog/notifier/state-snapshot stack, processed the first external grant (OneNov), and shipped the curated outreach copy pack publicly.
