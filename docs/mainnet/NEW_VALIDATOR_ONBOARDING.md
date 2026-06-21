# Mainnet new-validator onboarding — `nexarail-mainnet-1`

_Written 2026-06-19, post-rescue. Send this to any operator who wants to join._

## TL;DR for the operator

You can join `nexarail-mainnet-1` right now via the standard cosmos-sdk
`tx staking create-validator` flow. Block 1 already shipped — we're past
genesis. You sync from height 1, then submit a tx to register.

You need:
1. **NXRL** to self-bond + pay gas. Request a grant from Bradley (see
   *Faucet request* below) or bring your own.
2. **A box** that can run `nexaraild start` 24/7.
3. **Routable inbound TCP** on whichever port you choose for p2p (default
   26656). IPv6 is fine; IPv4 with a port-forward also fine.

## Network parameters

| Field | Value |
| --- | --- |
| Chain ID | `nexarail-mainnet-1` |
| Native denom | `unxrl`  (1 NXRL = 1,000,000 unxrl) |
| Genesis time | `2026-06-18T18:00:00Z` |
| Genesis URL | `https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json` |
| Genesis SHA256 | `f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32` |
| Binary | `https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64` |
| Min gas price | `0.025unxrl` |
| App version | `0.37.16` |
| Status page | `https://github.com/Bookings-cpu/nexarail-status` (rendered from `public/mainnet-status.json`) |

## Peer endpoints (use both if you can)

The coordinator alpha node is the public entrypoint. IPv6 is open; the
IPv4 router port-forward is in progress.

```
# IPv6 (working now):
96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656

# IPv4 (via bore.pub community TCP relay; permanent Oracle sentry in flight):
96e659f9a87723304dcd614e3ca89d9b6daf26cc@bore.pub:32656
```

Set as `persistent_peers` in `~/.nexarail/config/config.toml`.

> Note: Bradley's home IP can rotate. If you can't connect after a while,
> ping him and he'll repost the live string. The node_id is permanent;
> the host part can change.

## Slashing in force

- Downtime: >50% missed in last 10,000 blocks → 600s jail, 0.01% slash
- Double-sign: 5% slash + permanent tombstone

Do not run two nodes with the same consensus key. Ever.

## Step-by-step

### 1. Install

```bash
curl -L -o nexaraild https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64
chmod +x nexaraild
sudo mv nexaraild /usr/local/bin/
nexaraild version
```

### 2. Init

```bash
MONIKER="<your-moniker>"
nexaraild init "$MONIKER" --chain-id nexarail-mainnet-1
```

This creates `~/.nexarail/`. Validator consensus key sits at
`~/.nexarail/config/priv_validator_key.json` — **back this up encrypted,
treat it as sensitive as a wallet seed**.

### 3. Drop the genesis

```bash
curl -L -o ~/.nexarail/config/genesis.json \
  https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json

# verify
shasum -a 256 ~/.nexarail/config/genesis.json
# expect: f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32
```

If the SHA doesn't match, stop and ping Bradley.

### 4. Configure

`~/.nexarail/config/config.toml`:

```toml
[p2p]
persistent_peers = "96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656"
allow_duplicate_ip = true
addr_book_strict = false
```

`~/.nexarail/config/app.toml`:

```toml
minimum-gas-prices = "0.025unxrl"
```

### 5. Start syncing

```bash
nexaraild start
```

Watch height climb. With current chain height in the low thousands, a
fresh node syncs in well under an hour.

### 6. Create your wallet + request faucet

```bash
nexaraild keys add validator --keyring-backend file
# write the mnemonic on PAPER, store offline.

# print your operator address (this is what Bradley sends NXRL to):
nexaraild keys show validator -a --keyring-backend file
# nxr1...

# print your consensus pubkey (Bradley needs this too):
nexaraild tendermint show-validator
# nxrvalconspub1...
```

### 7. Faucet request — what to send Bradley

Paste this block in DM:

```
Mainnet validator onboarding request
Moniker:          <your moniker>
Operator address: nxr1...
Consensus pubkey: nxrvalconspub1...
Self-bond target: <X> NXRL          (default: 500 NXRL)
Gas budget:       100 NXRL
Total grant:      <X + 100> NXRL
Contact:          <email/Discord/Signal>
```

Bradley sends the grant once approved.

### 8. Submit `create-validator`

After funds arrive:

> ⚠ **STOP — DOUBLE-CHECK YOUR `--amount` BEFORE SIGNING.**
> The amount is in **unxrl** (micro-NXRL), not NXRL.
> `1 NXRL = 1,000,000 unxrl`, so a 500 NXRL self-bond is **`500000000unxrl`** — that is **EIGHT** zeros after the `5`, not six.
> Several validators have accidentally typed `5000000unxrl` (= 5 NXRL) and ended up with ~1/100 their intended voting power. Count the zeros twice.

```bash
nexaraild tx staking create-validator \
  --amount=500000000unxrl \                  # 500,000,000 unxrl = 500 NXRL  (count the zeros)
  --pubkey=$(nexaraild tendermint show-validator) \
  --moniker="<your moniker>" \
  --identity="<keybase 16-char id, optional>" \
  --website="<optional>" \
  --details="<optional>" \
  --chain-id=nexarail-mainnet-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --gas-adjustment=1.4 \
  --gas-prices="0.025unxrl" \
  --from=validator \
  --keyring-backend=file \
  --yes
```

The `--amount` is your self-bond in **unxrl**: `500000000unxrl` = 500 NXRL.

#### If you got the amount wrong

If your create-validator landed but with a typoed self-bond (for example 5 NXRL
instead of 500), don't redo create-validator. Just top up with a `delegate` tx
from the same operator wallet:

```bash
nexaraild tx staking delegate \
  <your-valoper-address> \
  495000000unxrl \                           # 495 NXRL — adjust to (target - current_self_bond) NXRL × 1,000,000
  --from=<your-operator-key-name> \
  --chain-id=nexarail-mainnet-1 \
  --gas=auto --gas-adjustment=1.4 --gas-prices=0.025unxrl \
  --yes
```

Your voting power updates in the next block. No re-registration needed.

### 9. Verify you're in the active set

```bash
nexaraild query staking validators --output json --limit 100 \
  | python3 -c "import sys,json; [print(v['description']['moniker'], v['tokens']) for v in json.load(sys.stdin)['validators']]"
```

You should see your moniker. Then check the status page.

## Faucet send-side (for Bradley)

Mainnet faucet source (locked 2026-06-19): **`ecosystem_grants`**
(`nxr14mnya6lj7ay43zyy95lhe8qppgr3nys9xte6gh`, balance 150,000,000 NXRL).

Standard grant per new mainnet validator (locked 2026-06-19):

| Item | Amount |
| --- | --- |
| Self-bond | 500 NXRL |
| Gas budget | 100 NXRL |
| **Total** | **600 NXRL  (600,000,000 unxrl)** |

Send command (run on Bradley's Mac mini #1):

```bash
BIN=~/workspace/nexarail/build/nexaraild
RPC=http://127.0.0.1:32657
HOME_=~/.nexarail-mainnet-keys
TO=nxr1...                                    # operator's address from the request
AMOUNT=600000000unxrl                         # 600 NXRL

$BIN tx send ecosystem_grants "$TO" "$AMOUNT" \
  --node "$RPC" \
  --chain-id nexarail-mainnet-1 \
  --keyring-backend os \
  --home "$HOME_" \
  --gas auto --gas-adjustment 1.4 --gas-prices 0.025unxrl \
  --yes
```

(Reminder: this is the cosmos-sdk fork — it's `tx send`, NOT `tx bank send`.)

Log every grant in `coordination/validators/mainnet-faucet-tracker.csv`.

## Known constraints

- IPv4 inbound is currently relayed via `bore.pub:32656` (free
  community-run TCP relay — single instance, no SLA). A permanent
  Oracle Always Free sentry is in deployment; once live, the published
  IPv4 endpoint moves to that sentry. IPv6 path direct to alpha is
  live and stable.
- Genesis is fixed. Late validators do not get vesting allocations —
  they buy/earn NXRL.
- Slashing is live from block 1. Do not run a hot wallet on the
  validator host.

## Changelog

- 2026-06-19 — Doc created post-launch rescue.
- 2026-06-20 — Replaced stale ephemeral IPv6 (`a872:d882:8cd5:dd0c`)
  with stable autoconf address (`ca7:b2ed:6262:4005`). Replaced
  obsolete router-forward IPv4 reference (`188.30.133.232`) with
  `bore.pub` relay endpoint pending Oracle sentry cutover.
