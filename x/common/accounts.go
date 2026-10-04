package common

// Canonical module account names for NexaRail's live fund infrastructure.
// Previously redeclared independently as identical string literals in
// app.go, x/treasury/types, x/payout/types, and x/settlement/keeper — any one
// of those drifting out of sync with app.go's maccPerms registration would
// silently misdirect or panic on a live fund transfer. Single source of
// truth here; every other declaration now aliases these.
const (
	TreasuryModuleAccount = "nexarail_treasury"
	BurnerModuleAccount   = "nexarail_burner"
)
