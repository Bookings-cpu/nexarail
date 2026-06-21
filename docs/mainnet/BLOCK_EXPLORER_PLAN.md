# Block Explorer — Plan

**Status:** Scaffolded, deployment blocked on Oracle sentry (need public RPC + LCD).

## Why we need one

Right now:
- Validators can't show their delegators tx history
- Users can't look up tx hashes or addresses
- The status page only shows aggregate stats — no per-block, per-tx, per-address drill-down
- Outside observers have no way to evaluate chain health beyond "is the website up"

A block explorer is table-stakes for any L1 that wants to be taken seriously by validators, delegators, exchanges, or builders.

## Recommendation: fork `ping.pub`

Open-source cosmos-flavored explorer. Vue.js SPA, consumes chain RPC + REST APIs. Already supports the SDK v0.47 chains. Hostable on GitHub Pages alongside the existing status site.

Repo: https://github.com/ping-pub/explorer

**Pros**
- Cosmos-native, knows all the standard modules out of the box
- Static SPA, deploys to GitHub Pages free
- Used by 30+ cosmos chains in production (Juno, Stride, Osmosis testnet, etc.)
- Active maintenance

**Cons**
- Heavy on RPC calls (each page load = dozens of queries to public RPC)
- Vue.js — not the lightest framework
- Needs branding config per-chain

## Deployment shape

```
Bookings-cpu/nexarail-explorer  (new repo)
├── .github/workflows/deploy.yml  (auto-deploy on push to main)
├── chains/mainnet/
│   ├── nexarail.json   (chain config — RPC, LCD, denom, prefix, logo)
│   └── nexarail.png    (logo)
├── public/
│   ├── index.html
│   ├── assets/         (logos, fonts, colors)
└── src/                (forked ping.pub source, minimal changes)
```

Domain plan:
- Free tier: `https://bookings-cpu.github.io/nexarail-explorer/`
- Optional later: `https://explorer.nexarail.com` (needs domain + DNS)

## Chain config (drop-in for ping.pub `chains/` directory)

```json
{
  "chain_name": "nexarail",
  "registry_name": "nexarail",
  "logo": "/logos/nexarail.png",
  "api": ["https://api.nexarail.com"],
  "rpc": ["https://rpc.nexarail.com"],
  "sdk_version": "0.47.17",
  "coin_type": "118",
  "min_tx_fee": "25000",
  "addr_prefix": "nxr",
  "logo_XL": "/logos/nexarail-xl.png",
  "assets": [
    {
      "base": "unxrl",
      "symbol": "NXRL",
      "exponent": "6",
      "coingecko_id": "",
      "logo": "/logos/nexarail.png"
    }
  ],
  "themeColor": "#FFD700",
  "features": [
    "ibc-transfer",
    "ibc-go",
    "no-legacy-stdTx",
    "cosmwasm"
  ]
}
```

(Note: `api.nexarail.com` and `rpc.nexarail.com` don't yet exist. These will resolve to the Oracle sentry once it lands. Until then, we can hardcode the bore.pub IPv4 endpoint or skip explorer deploy.)

## What's blocking deployment today

1. **No public RPC** — Alpha's RPC is loopback-only. Bore.pub exposes p2p (32656) not RPC (32657).
2. **No public LCD/REST** — Same as above; api.address is tcp://127.0.0.1:1417.
3. **No DNS** — `explorer.nexarail.com` and `rpc.nexarail.com` need to exist.

The Oracle sentry, once provisioned, can serve as the public RPC + LCD layer:
- Sentry runs `nexaraild start` with the validator key empty (sentry mode)
- Sentry's RPC + REST exposed on its public IP (firewall opens 26657 + 1317 to 0.0.0.0)
- DNS A records for `rpc.nexarail.com` and `api.nexarail.com` → sentry public IP
- Cloudflare in front for caching + DDoS protection (cheap or free)
- Explorer queries cloudflare-fronted sentry

## What I can do right now (without sentry)

1. **Create the explorer repo** (`nexarail-explorer`) with ping.pub fork
2. **Scaffold the chain config + logo + branding**
3. **Wire the GitHub Actions deploy workflow**
4. **Deploy a build that points at `localhost:32657`** — not useful for users, but proves the build pipeline works

That's a half-day of work and de-risks the actual go-live to a 5-minute DNS swap once sentry IPs are known.

## What I'll do once sentry is live

1. Configure `nexaraild` on sentry: validator key empty, persistent_peers point at alpha + IPv6 endpoint, RPC + LCD exposed on 0.0.0.0
2. Bradley sets up DNS (`rpc.nexarail.com` + `api.nexarail.com` → sentry IP), optional Cloudflare
3. Swap explorer chain config to public endpoints
4. Add link to explorer from status page

## Open questions for Bradley

1. **Domain ownership**: do you own `nexarail.com`? Or another suitable domain? Needed for `rpc.`, `api.`, `explorer.` subdomains.
2. **Cloudflare in front**: free tier is fine — want me to add the proxy + DDoS protection to the plan?
3. **Theme**: ping.pub explorers are configurable. Gold accent (matching the existing mainnet status page)? Or something else?
4. **Scaffolded repo now**: want me to create `nexarail-explorer` and push the scaffold today, or wait until sentry lands?

## Next move

Awaiting Bradley's answers to 4 open questions above before I create the repo + push the scaffold. Otherwise on standby.
