# Sentry Node — Install Runbook

**Status:** Ready-to-paste. Triggers when the Oracle Cloud Always Free Ampere VM finally provisions. Until then, the `com.nexarail.sentry-retry` LaunchAgent on Mac mini #1 retries VM creation every 15 minutes.

## Trigger condition

You'll know the VM has provisioned when the file `/tmp/nexarail-sentry-ready` exists on Mac mini #1. It will contain the instance OCID and public IP. macOS will also pop a notification.

```bash
cat /tmp/nexarail-sentry-ready
# expected:
#   2026-MM-DDTHH:MM:SSZ
#   INSTANCE_ID=ocid1.instance...
#   PUBLIC_IP=129.151.xx.xx
#   SSH=ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@129.151.xx.xx
```

## Step 1 — SSH in (from Mac mini #1)

```bash
PUBLIC_IP="$(grep '^PUBLIC_IP=' /tmp/nexarail-sentry-ready | cut -d= -f2)"
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP"
```

First connection asks to accept the host key. Accept.

## Step 2 — Server-side prep (paste on the sentry)

```bash
# OS bring-up
sudo apt-get update -y
sudo apt-get install -y curl jq ufw

# Ubuntu firewall — open just the ports the sentry needs
sudo ufw allow OpenSSH
sudo ufw allow 32656/tcp comment 'tendermint p2p'
sudo ufw allow 26657/tcp comment 'tendermint rpc'
sudo ufw allow 1317/tcp  comment 'cosmos-sdk lcd'
sudo ufw --force enable
sudo ufw status verbose
```

(Oracle security list was opened for 22 + 32656 in the bootstrap script. The 26657 and 1317 openings need adding too — see Step 4.)

## Step 3 — Drop the nexaraild binary (from Mac mini #1, separate terminal)

```bash
PUBLIC_IP="$(grep '^PUBLIC_IP=' /tmp/nexarail-sentry-ready | cut -d= -f2)"
scp -i ~/.ssh/nexarail-sentry-oracle \
  ~/workspace/nexarail/build/nexaraild-linux-arm64 \
  ubuntu@"$PUBLIC_IP":/tmp/nexaraild

ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  sudo install -o root -g root -m 755 /tmp/nexaraild /usr/local/bin/nexaraild
  nexaraild version
'
```

Expected output: `0.1.0-dev` (or the binary's stamped version).

Also drop the mainnet genesis file:

```bash
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  set -e
  nexaraild init nexarail-sentry-1 --chain-id nexarail-mainnet-1 --home ~/.nexarail
  curl -L -o ~/.nexarail/config/genesis.json \
    https://github.com/Bookings-cpu/nexarail/releases/download/mainnet-genesis-nexarail-mainnet-1/genesis.json
  # Verify SHA
  echo "f84f5f03d4d54945153c3f68e20e9864fc03c7f35dbeec2b40274f18d152db32  $HOME/.nexarail/config/genesis.json" | sha256sum -c -
'
```

## Step 4 — Open Oracle security list for RPC + LCD

These weren't in the bootstrap script (only 22 + 32656 were). Run on Mac mini #1:

```bash
export OCI_CLI_PROFILE=nexarail-apikey SUPPRESS_LABEL_WARNING=True
TENANCY=ocid1.tenancy.oc1..aaaaaaaagkrbfwvsjlyrxf5euze3ste2biyvlqhuddb44mvyvsnannzfzmkq
VCN_ID="$(oci network vcn list --compartment-id $TENANCY --display-name nexarail-vcn --query 'data[0].id' --raw-output)"
DSL_ID="$(oci network vcn get --vcn-id $VCN_ID --query 'data."default-security-list-id"' --raw-output)"
oci network security-list update --security-list-id "$DSL_ID" --force \
  --ingress-security-rules '[
    {"source":"0.0.0.0/0","protocol":"6","isStateless":false,"tcpOptions":{"destinationPortRange":{"min":22,"max":22}}},
    {"source":"0.0.0.0/0","protocol":"6","isStateless":false,"tcpOptions":{"destinationPortRange":{"min":32656,"max":32656}}},
    {"source":"0.0.0.0/0","protocol":"6","isStateless":false,"tcpOptions":{"destinationPortRange":{"min":26657,"max":26657}}},
    {"source":"0.0.0.0/0","protocol":"6","isStateless":false,"tcpOptions":{"destinationPortRange":{"min":1317,"max":1317}}}
  ]' --wait-for-state AVAILABLE
```

## Step 5 — Configure sentry mode

Edit `~/.nexarail/config/config.toml` on the sentry. Key changes from defaults:

```toml
[p2p]
laddr = "tcp://0.0.0.0:32656"
persistent_peers = "96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656"   # alpha IPv6
addr_book_strict = false
allow_duplicate_ip = true
pex = true

[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]   # required for explorer frontend

[mempool]
broadcast = true
```

Edit `~/.nexarail/config/app.toml`:

```toml
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
enabled-unsafe-cors = true

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
```

**Critical — wipe the validator key.** A sentry should never sign blocks:

```bash
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  # Replace with a zero-value placeholder so the node still starts but cannot sign.
  cat > ~/.nexarail/config/priv_validator_key.json <<JSON
{"address":"0000000000000000000000000000000000000000","pub_key":{"type":"tendermint/PubKeyEd25519","value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}
JSON
  chmod 600 ~/.nexarail/config/priv_validator_key.json
'
```

## Step 6 — systemd unit

```bash
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  sudo tee /etc/systemd/system/nexaraild.service > /dev/null <<EOF
[Unit]
Description=NexaRail mainnet sentry
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ubuntu
ExecStart=/usr/local/bin/nexaraild start --home /home/ubuntu/.nexarail --minimum-gas-prices 0.025unxrl
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
  sudo systemctl daemon-reload
  sudo systemctl enable --now nexaraild
  sleep 3
  sudo systemctl status nexaraild --no-pager
'
```

## Step 7 — Verify sync from genesis

```bash
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  curl -s http://localhost:26657/status | jq ".result.sync_info"
'
```

Expected: `catching_up: true` initially, progressing through blocks. Full sync from genesis (3-7 days of state) takes ~30-60 minutes on the Ampere ARM 4-CPU shape.

When `catching_up: false`, sentry is caught up and serving consensus state at the current height.

## Step 8 — Register sentry as a persistent_peer on alpha

Edit `~/.nexarail-mainnet-alpha/config/config.toml` on Mac mini #1:

```toml
persistent_peers = "<existing-coord-peers>,<sentry_node_id>@<sentry_public_ip>:32656"
```

Get the sentry node ID:

```bash
ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@"$PUBLIC_IP" '
  nexaraild tendermint show-node-id --home ~/.nexarail
'
```

Restart alpha to pick up the new peer:

```bash
launchctl kickstart -k gui/$(id -u)/com.nexarail.mainnet.alpha
```

## Step 9 — Update onboarding runbook + status page

Add the sentry to the public peers list in `docs/mainnet/NEW_VALIDATOR_ONBOARDING.md`:

```
Persistent peers (use any):
  IPv6 (alpha direct):  96e659f9a87723304dcd614e3ca89d9b6daf26cc@[2a04:4a43:867f:f226:ca7:b2ed:6262:4005]:32656
  IPv4 (sentry-1):      <sentry_node_id>@<sentry_public_ip>:32656
```

Also add to the status page collector's metadata so the public status shows "Sentry: live".

## Step 10 — DNS (optional but recommended for the block explorer)

Once a domain is registered (see `docs/mainnet/BLOCK_EXPLORER_PLAN.md`), point:

- `rpc.<domain>` → sentry public IP (A record)
- `api.<domain>` → sentry public IP (A record)
- `seed.<domain>` → sentry public IP (A record)

Cloudflare in front of these is recommended for caching + DDoS protection (free tier sufficient).

## Verification checklist

- [ ] SSH works from Mac mini #1 to sentry public IP
- [ ] `nexaraild version` runs on the sentry
- [ ] Genesis SHA matches the published value
- [ ] `priv_validator_key.json` has been zeroed (sentry cannot sign)
- [ ] systemd service `nexaraild` running, status active
- [ ] `curl http://<sentry_public_ip>:26657/status` returns chain info
- [ ] `curl http://<sentry_public_ip>:1317/cosmos/staking/v1beta1/validators` returns validator list
- [ ] Sentry peered with alpha (verify via `curl localhost:26657/net_info | jq '.result.n_peers'`)
- [ ] Sync caught up (`catching_up: false`)
- [ ] Sentry registered in alpha's persistent_peers and alpha restarted
- [ ] Onboarding runbook updated with sentry IPv4 peer
