# NexaRail Discord — Moderation Rules

**Effective:** when the server goes live.
**Owner:** Coordinator + moderators.

## Short version (pinned in every channel)

```
1. Be useful. No spam, no shilling, no off-topic chat in technical channels.
2. No scams, impersonation, or fake faucet/airdrop posts. Instant ban.
3. Never paste mnemonics, private keys, or seed phrases anywhere.
4. Coordinator will NEVER DM you about funds. Anyone who does is impersonating.
5. Keep validator support in #validator-support, not DMs.
6. English in technical channels for first pass; other languages welcome alongside.
7. Respect operators of different backgrounds, time zones, and skill levels.
```

## Full rules

### R-1. No financial promises or solicitation
NexaRail Testnet 1 is a testnet. Tokens on this chain are not investments. Do not post price talk, "next 1000x", airdrop promises, IDO claims, or any solicitation framed around the testnet token. Coordinator will delete and warn; repeat = ban.

### R-2. No scams or impersonation
Impersonating coordinator, validators, sponsors, or any other community member = instant ban. Posting fake faucet sites, fake genesis files, fake binaries = instant ban + report to platform.

### R-3. Never share secrets
Never paste a mnemonic, private key, keyring password, validator consensus key, or `priv_validator_key.json` content. We will redact and warn; repeat = ban. Treat any compromised secret as lost — generate a new key and request rotation through the proper process.

### R-4. DMs about funds = scam
The coordinator will NEVER DM you to request a faucet, mnemonic, password, or wallet send. Any DM claiming to be coordinator about funds is impersonation. Screenshot and report in `#node-status`.

### R-5. Stay on-topic
- Validator setup → `#validator-setup`
- Operator-to-operator debugging → `#validator-support`
- Faucet → `#faucet-requests`
- Health / restarts / peer issues → `#node-status`
- Governance discussion → `#governance-discussion`
- Anything dev-related from the team → `#dev-updates`
- General chit-chat → keep it minimal and in info channels only

Cross-posting the same question in 4 channels = warning.

### R-6. No NDAs in public channels
If you have a confidential issue (security disclosure, key compromise, paid engagement), reach the coordinator privately via the security contact listed in `NEXARAIL_TESTNET_1_STATUS.md`. Do not post it in any public channel.

### R-7. Respect
No harassment, racism, sexism, threats, doxxing, or targeted abuse. Instant ban.

### R-8. Bot policy
Personal bots welcome IF they don't touch funds, don't DM members, and don't post automatically more than once per hour. Coordinator-approved automation only handles read-only chain data.

### R-9. Security disclosure
If you discover a security issue, do NOT post it publicly. Reach the coordinator privately. We will acknowledge within 48 hours and coordinate a fix + disclosure window.

### R-10. Moderation appeals
If you think a moderation action was wrong, DM the coordinator (NOT a moderator) within 7 days. We will review.

## Enforcement

| Severity | Action |
|---|---|
| Minor (off-topic, low-effort spam) | Warning + message removed |
| Repeat minor | Temporary mute (24h) |
| Major (scam, impersonation, secret-pasting after warning, harassment) | Permanent ban |
| Security threat (active exploit, attempted social engineering) | Permanent ban + report to platform |

## Logging

Coordinator + moderators keep a moderation log in `coordination/discord/moderation-log.md` (private to mod team). Logged actions: warnings, mutes, bans, with reason and link to the offending message where possible.
