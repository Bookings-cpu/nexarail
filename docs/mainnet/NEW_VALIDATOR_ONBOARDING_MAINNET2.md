# Mainnet relaunch new-validator onboarding — `nexarail-mainnet-2`

_Written 2026-09-15, post-relaunch. Send this to anyone who wants to join. Every step below has been tested end-to-end against the live network before publishing this doc._

## TL;DR for the operator

`nexarail-mainnet-2` is a fresh chain — new genesis, new validator set, new
tokenomics baked in properly this time. It is **not** a continuation of
`nexarail-mainnet-1` (which had a fatal bug and was retired). If you were a
validator on mainnet-1, you need to onboard fresh here — old balances and
validator registrations do not carry over.

You can join right now via the standard cosmos-sdk `tx staking
create-validator` flow. You sync from height 1, then submit a tx to register.

You need:
1. **NXRL** to self-bond + pay gas. Request a grant from Bradley (see
   *Faucet request* below).
2. **A box** that can run `nexaraild start` 24/7.
3. **Routable inbound TCP** on port 26656 for p2p.

## Network parameters

| Field | Value |
| --- | --- |
| Chain ID | `nexarail-mainnet-2` |
| Native denom | `unxrl` (1 NXRL = 1,000,000 unxrl) |
| Genesis time | `2026-09-15T13:19:13Z` |
| Genesis URL | `https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-2/genesis.json` |
| Genesis SHA256 | `d2d6933fbdf2fed1727c9906dfb41024871c3c20636a6b366ae8414d5af54d62` |
| Binary | `https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-2/nexaraild-linux-amd64` |
| Min gas price | `0.001unxrl` |
| App version | `0.37.16` |
| Status page | `https://bookings-cpu.github.io/nexarail-status/mainnet.html` |
| Public RPC | `http://5.161.72.48:26657` (for syncing checks / queries only — do not persistent-peer against this) |

## Peer endpoints (persistent_peers — use all 5)

```
1af9139d59677cc42ff2924c89f9cf9e0c980246@5.161.85.160:26656,d91372d5918d870c7861a2eb3e99b90d7c5882ca@5.161.103.203:26656,8776e496483fefbbc4dc17749f5ec80ab121fe2c@5.161.99.76:26656,45668d1ae375f39fce47dfa96e76db7946da7188@5.161.94.47:26656,bbd2b07264da053541128a0677540bc95bf415c6@5.161.72.48:26656
```

Set as `persistent_peers` in `~/.nexarail/config/config.toml`.

## Slashing in force

- Downtime: >50% missed in last 10,000 blocks → 600s jail, 0.01% slash
- Double-sign: 5% slash + permanent tombstone

Do not run two nodes with the same consensus key. Ever.

## Step-by-step

### 1. Install

```bash
curl -L -o nexaraild https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-2/nexaraild-linux-amd64
chmod +x nexaraild
sudo mv nexaraild /usr/local/bin/
nexaraild version
```

### 2. Init

```bash
MONIKER="<your-moniker>"
nexaraild init "$MONIKER" --chain-id nexarail-mainnet-2
```

This creates `~/.nexarail/`. Validator consensus key sits at
`~/.nexarail/config/priv_validator_key.json` — **back this up encrypted,
treat it as sensitive as a wallet seed**.

### 3. Drop the genesis

```bash
curl -L -o ~/.nexarail/config/genesis.json \
  https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-2/genesis.json

shasum -a 256 ~/.nexarail/config/genesis.json
# expect: d2d6933fbdf2fed1727c9906dfb41024871c3c20636a6b366ae8414d5af54d62
```

If the SHA doesn't match, stop and ping Bradley.

### 4. Configure

`~/.nexarail/config/config.toml`:

```toml
[p2p]
persistent_peers = "1af9139d59677cc42ff2924c89f9cf9e0c980246@5.161.85.160:26656,d91372d5918d870c7861a2eb3e99b90d7c5882ca@5.161.103.203:26656,8776e496483fefbbc4dc17749f5ec80ab121fe2c@5.161.99.76:26656,45668d1ae375f39fce47dfa96e76db7946da7188@5.161.94.47:26656,bbd2b07264da053541128a0677540bc95bf415c6@5.161.72.48:26656"
```

`~/.nexarail/config/app.toml`:

```toml
minimum-gas-prices = "0.001unxrl"
```

### 5. Start syncing

```bash
nexaraild start
```

Watch height climb. Chain is only minutes old at launch — you'll be caught up in seconds to minutes depending on when you join.

### 6. Create your wallet + request faucet

```bash
nexaraild keys add validator --keyring-backend file
# write the mnemonic on PAPER, store offline. Never paste it anywhere, including Discord DMs.

nexaraild keys show validator -a --keyring-backend file
# nxr1...  — this is what Bradley sends NXRL to

nexaraild tendermint show-validator
# {"@type":"/cosmos.crypto.ed25519.PubKey","key":"..."}  — Bradley needs this too, for reference only
```

### 7. Faucet request — what to send Bradley

Post this in the validator Discord channel or DM:

```
Mainnet validator onboarding request
Moniker:          <your moniker>
Operator address: nxr1...
Self-bond target: 500 NXRL (default)
Gas budget:       100 NXRL
Total grant:      600 NXRL
Contact:          <email/Discord/Signal>
```

### 8. Submit `create-validator`

After funds arrive:

> ⚠ **STOP — DOUBLE-CHECK YOUR `--amount` BEFORE SIGNING.**
> The amount is in **unxrl** (micro-NXRL), not NXRL.
> `1 NXRL = 1,000,000 unxrl`, so a 500 NXRL self-bond is **`500000000unxrl`** — **EIGHT** zeros after the `5`, not six. Count them twice.

```bash
nexaraild tx staking create-validator \
  --amount=500000000unxrl \
  --pubkey=$(nexaraild tendermint show-validator) \
  --moniker="<your moniker>" \
  --identity="<keybase 16-char id, optional>" \
  --website="<optional>" \
  --details="<optional>" \
  --chain-id=nexarail-mainnet-2 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --gas-adjustment=1.4 \
  --gas-prices="0.001unxrl" \
  --from=validator \
  --keyring-backend=file \
  --yes
```

#### If you got the amount wrong

Don't redo create-validator — top up with a `delegate` tx from the same wallet:

```bash
nexaraild tx staking delegate \
  <your-valoper-address> \
  <difference>unxrl \
  --from=validator \
  --chain-id=nexarail-mainnet-2 \
  --gas=auto --gas-adjustment=1.4 --gas-prices=0.001unxrl \
  --yes
```

### 9. Verify you're in the active set

```bash
nexaraild query staking validators --node http://5.161.72.48:26657 --output json --limit 100 \
  | python3 -c "import sys,json; [print(v['description']['moniker'], v['tokens']) for v in json.load(sys.stdin)['validators']]"
```

Or check the status page: https://bookings-cpu.github.io/nexarail-status/mainnet.html

## Faucet send-side (for Bradley)

Mainnet faucet source: **`ecosystem_grants`** (`nxr1rd3pqfnazunaym0ezcjw4vrlpjy6dle5zuhaqf`, ~164,285,714 NXRL).

Standard grant per new validator: **600 NXRL** (500 self-bond + 100 gas).

SSH into alpha (5.161.85.160), then:

```bash
./grant-validator.sh <operator-address> 600000000
```

(Script lives at `/root/grant-validator.sh` on alpha — sends from the on-host `ecosystem_grants` test-keyring key, never leaves that host.)

## Known constraints

- Public RPC (`5.161.72.48:26657`) is temporarily hosted on a validator node (`echo`) due to a Hetzner account server-cap — a dedicated non-validator sentry node is planned once the cap is raised. Use it for queries/sync checks; don't hammer it.
- Genesis is fixed. Late validators do not get vesting allocations — same policy as before.
- Slashing is live from block 1. Do not run a hot wallet on the validator host.

## Verified working (2026-09-15)

This entire flow — genesis download, hash verify, wallet creation, faucet grant, create-validator, active-set confirmation — was tested end-to-end against the live network before this doc was published. Test validator `onboarding-test` is visible on the status page as proof.
