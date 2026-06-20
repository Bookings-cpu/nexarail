# #validator-setup — Pinned messages

Two messages designed to be pinned together. Each fits inside Discord's 2000-character per-message limit. Copy-paste verbatim. Repin a new version whenever the binary, genesis SHA, or peer list changes.

---

## Pin #1 — Run a NexaRail validator (facts & downloads)

```
🛠 Run a NexaRail Testnet Validator

Chain ID: nexarail-testnet-1 · denom: unxrl (1 NXRL = 1,000,000 unxrl) · Standard faucet grant: 2,000,000 unxrl (1M self-delegation + 1M gas buffer)

📥 Binary (linux-amd64) — ONLY use this build:
https://github.com/Bookings-cpu/nexarail/releases/download/v0.1.0-rc1-validator-recovery-hotfix/nexaraild-linux-amd64
SHA256: cdb03d84e2d998e3581f368cc3440fce179c34010398c7343298e94a3d82112c

📥 Genesis:
https://github.com/Bookings-cpu/nexarail/releases/download/testnet-genesis-nexarail-testnet-1/genesis.json
SHA256: c9877720485c598da72579be98614059954fbe051f3fde29eea0a1a2a1057fe3

🌐 Persistent peers (paste into config.toml):
2bb62d82b4dbf820fdafd843816f1e72a84ffa8f@nexarail-testnet-peer.nodesync.top:26656,7d64eb6003fc12b8a174e6e9720e45a6412c4195@65.109.104.118:60656

🖥 Min host: Linux, 4 vCPU, 8 GB RAM, 200 GB SSD, static IP, TCP 26656 open inbound.

⚠ CLI quirks (this cosmos-sdk fork):
• Send command is `nexaraild tx send [from] [to] [amount]` — NOT `tx bank send`.
• If `tx staking` errors with "unknown command", you are on the old binary. Upgrade to the hotfix above. Same for `tx slashing unjail`.

🚦 Setup order:
1. Download binary, verify sha256, chmod +x.
2. ./nexaraild init <moniker> --chain-id nexarail-testnet-1
3. Replace ~/.nexarail/config/genesis.json. Verify SHA256.
4. In ~/.nexarail/config/config.toml: set persistent_peers to the list above.
5. In ~/.nexarail/config/app.toml: set minimum-gas-prices = "0.025unxrl".
6. ./nexaraild keys add <your-key-name> — capture mnemonic OFFLINE.
7. Start the node, wait until catching_up=false and peer count > 0.
8. Post a faucet request in #faucet-requests (template in pin #2).
9. Once your 2M unxrl grant lands, broadcast tx staking create-validator (command in pin #2).

🚫 Never paste your mnemonic anywhere. We only send the 2M grant after your node is fully synced.

❓ Stuck? Ask here. Coordinator usually responds within a few hours on UK daytime.
```

---

## Pin #2 — Templates & commands

````
📋 NexaRail Validator — Templates & Commands

💧 Faucet request — paste in #faucet-requests once your node is synced:

```
Faucet request — nexarail-testnet-1
Moniker:          <your moniker>
Operator address: <nxr1…>
Discord handle:   @you
Server country:   <country>
Synced (catching_up=false)? YES
Local block height:       <your node's height>
Coordinator height seen:  <same height seen by a peer at that moment>

I have read the faucet policy.
I have NOT shared my mnemonic or private key.
```

🛠 Create validator — after your 2M unxrl grant lands:

```
./nexaraild tx staking create-validator \
  --amount=1000000unxrl \
  --pubkey=$(./nexaraild tendermint show-validator) \
  --moniker="<your moniker>" \
  --chain-id=nexarail-testnet-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas=auto --gas-adjustment=1.4 --gas-prices=0.025unxrl \
  --from=<your-operator-key-name>
```

The 1M unxrl bonded is your self-delegation. The other 1M from the grant stays in your operator account as gas buffer.

🔓 Unjail — if you get downtime-jailed:

```
./nexaraild tx slashing unjail \
  --from=<your-operator-key-name> \
  --chain-id=nexarail-testnet-1 \
  --gas=auto --gas-adjustment=1.3 --gas-prices=0.025unxrl \
  -y
```

Wait until the jail timer (10 min from jailing) has expired before broadcasting, or it will fail.

🔒 Slashing IS ON:
• Downtime: jail + 1% slash after missing >50% of the last 100 blocks. 10-minute jail; you must `tx slashing unjail` to rejoin.
• Double-sign: 5% slash + tombstone (permanent, no recovery on this key). Be very careful with state restores, key copies, and parallel signing.

📤 After your create-validator tx lands, drop the tx hash in this channel and we will confirm you are in the active set.
````

---

## Maintenance notes (for the moderator pinning this — do not paste into Discord)

- Repin Pin #1 whenever the binary, genesis SHA, or persistent peer list changes.
- Repin Pin #2 whenever the standard grant amount, slashing parameters, or operator command surface changes.
- Source of truth:
  - Binary: latest `validator-recovery-hotfix`-style release on the `Bookings-cpu/nexarail` GitHub repo.
  - Genesis: `releases/testnet-genesis/nexarail-testnet-1-final-delivery/genesis.json` locally; corresponds to the `testnet-genesis-nexarail-testnet-1` GitHub release.
  - Peers: `coordination/validators/launch-endpoint-inventory.csv` (NodeSync + UTSA post-launch entries).
  - Slashing params: `app_state.slashing.params` in the live genesis (`signed_blocks_window=100`, `min_signed_per_window=0.5`, `downtime_jail_duration=600s`, `slash_fraction_double_sign=0.05`, `slash_fraction_downtime=0.01`).
- The earlier draft of this pin said "Slashing module is currently OFF" — that has been wrong since launch. Genesis ships slashing on. Do not reintroduce that line.
