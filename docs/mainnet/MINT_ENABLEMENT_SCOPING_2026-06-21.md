# Mint Module Enablement — Scoping (2026-06-21)

## Discovery

The `mint` module is **already wired into the binary** (per litepaper §5,
"All standard Cosmos SDK v0.47.17 modules are wired"). What I previously
read as "no mint module" was a missing CLI subcommand only — the module
state exists in genesis with all inflation parameters **hardcoded to
zero**:

```
mint.params:
  inflation_max:         0.000
  inflation_min:         0.000
  inflation_rate_change: 0.000
  goal_bonded:           0.670
  blocks_per_year:       6,311,520
  mint_denom:            unxrl
```

This means there is **no chain upgrade required** to enable inflation. A
single `MsgUpdateParams` governance proposal can change these values to
positive numbers and inflation begins at the next block.

## Why this matters

Validators currently earn nothing from securing the chain. Once the
faucet/grant onboarding pipeline normalises (or runs out), there is no
economic reason for external operators to keep their nodes up. Without
yield, operator attrition is a tail risk that compounds — every week
without rewards is a week operators are running for goodwill alone.

## The decision: mint vs reserve drip

Two ways to deliver validator yield. Bradley needs to pick one before the
proposal goes to a vote.

### Option A — Enable mint (recommended)

Standard cosmos pattern. Inflation creates new NXRL each block, routed
into the `distribution` module, paid pro-rata to bonded validators (and
their delegators).

**Recommended params** (conservative, sub-Cosmos-Hub):

| Param | Proposed | Cosmos Hub | Reasoning |
|---|---|---|---|
| `inflation_min` | 0.05 (5%) | 0.07 | Floor — low at 67%+ bonded |
| `inflation_max` | 0.12 (12%) | 0.20 | Cap — high when bond ratio low |
| `inflation_rate_change` | 0.10 | 0.13 | Slow drift between min/max |
| `goal_bonded` | 0.67 | 0.67 | Standard 67% target |
| `blocks_per_year` | unchanged | — | Already correct for 5s blocks |

**Pros**
- Universally understood by validators — exactly the cosmos staking
  primitive they know
- Yield is automatic and continuous
- Reflects "real" staking economics; supports a price narrative
- No manual operator distribution work
- Compounds for stakers (compounding APY)

**Cons**
- Dilutes the fixed-supply narrative. 1B supply isn't truly fixed if
  mint expands it. At 12% max on 1B = +120M NXRL/year worst case.
- Most reward goes to coord nodes today (99.97% bonded power = 99.97%
  of new mint). Externals see effectively zero direct yield until power
  rebalances.

### Option B — Reserve drip from `validator_rewards_reserve`

Use the pre-allocated 150M NXRL `validator_rewards_reserve` bucket.
Periodic governance-authorised sends from that bucket to
`distribution`'s community pool, then funded as community rewards via
`FundCommunityPool` or direct transfers to validators.

**Pros**
- No new tokens minted — preserves the "1B fixed supply" narrative
- Pre-allocated bucket already exists for exactly this purpose
- Can target external validators specifically (not pro-rata) by
  manually distributing

**Cons**
- Requires manual or periodic governance proposals — operationally
  heavier than mint
- Yield isn't smooth/continuous — comes in bursts
- Validators expect mint-based yield; reserve-drip is a non-standard
  pattern (operators may discount it in their decision to stay)
- 150M finite cap → eventually runs out unless tokenomics restructured

### Recommendation

**Option A (enable mint), with the coord-share problem solved via the
decentralisation ceremony**:

1. Enable mint now with the conservative params above.
2. Coord nodes accept that early staking rewards accrue to them by
   default (because they hold 99.97% of bonded power).
3. **Commit publicly** that all coord-side staking rewards will be
   pooled and redistributed during the decentralisation ceremony at
   100 validators (or earlier if external count grows fast). This
   converts "coord scoops all the rewards" into "coord is custodying
   yield on behalf of the ecosystem until the milestone".
4. Preserve `validator_rewards_reserve` for the ceremony bonus payout
   and ongoing external grants — don't burn it for yield drip.

This gives validators the yield pattern they expect, gives Bradley a
crisp narrative ("we are not scooping the rewards — we are custodying
them for the milestone"), and preserves the 150M reserve for a
high-impact milestone moment rather than slow operational dripping.

## Coord-side acknowledgment

With 5 coord at 5M NXRL each and only 8,398 NXRL of external bonded:

- Coord share: 99.9664%
- External share: 0.0336%

At 12% max inflation on 25M bonded:
- Total new mint: ~3M NXRL/year
- Coord receives: ~2.999M NXRL/year
- External receives: ~1,000 NXRL/year (~£X spread across 16 operators)

That external yield is **not material** today. The reward narrative
only matters after coord rebalance. Today's mint enablement is more
about (a) creating the yield mechanism and (b) starting the coord
reward accrual that funds the future decentralisation ceremony.

## Proposed governance proposal (JSON sketch)

```json
{
  "messages": [
    {
      "@type": "/cosmos.mint.v1beta1.MsgUpdateParams",
      "authority": "<gov module account>",
      "params": {
        "mint_denom": "unxrl",
        "inflation_rate_change": "0.100000000000000000",
        "inflation_max":         "0.120000000000000000",
        "inflation_min":         "0.050000000000000000",
        "goal_bonded":           "0.670000000000000000",
        "blocks_per_year":       "6311520"
      }
    }
  ],
  "metadata": "Enable NXRL staking yield via mint module. See docs/mainnet/MINT_ENABLEMENT_SCOPING_2026-06-21.md.",
  "deposit": "1000000000unxrl",
  "title":   "Enable NXRL staking yield",
  "summary": "Change mint module parameters from zero-inflation to 5-12% bonded-ratio-targeted inflation. Validators and delegators begin earning staking rewards from the next block after this proposal passes. Coord-side rewards are committed to the decentralisation ceremony at 100 validators."
}
```

(Authority address is the gov module account — pulled at proposal-build
time from `nexaraild query auth module-account gov`.)

## Open questions for Bradley before the vote

1. **Confirm the params.** 5%/12% conservative — happy or want to push
   to standard Cosmos Hub 7%/20% (more aggressive)?
2. **Coord-rewards commitment.** Are we OK publicly committing that all
   coord staking rewards will go to the decentralisation ceremony?
   That promise needs to be on-chain or public before the vote.
3. **Voting period.** Default is 5 days. Want to shorten with a
   `--expedited` proposal? Coord can pass alone but the 5-day window
   demonstrates governance respect.
4. **Timing.** Submit now (low validator count, smooth rollout) or
   wait until ~30 external validators (more eyes on the vote, more
   credibility)?

## Status & next move

- Scoping done; this doc.
- Params chosen; ready to build the actual proposal.
- Awaiting Bradley's call on the four open questions above before
  building + submitting.
