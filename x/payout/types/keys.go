package types

import "github.com/nexarail/chain/x/common"

const (
	ModuleName = "payout"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	// TreasuryModuleAccount is the module account that funds live payouts in v1.
	// Aliases x/common's canonical constant (see x/common/accounts.go) rather
	// than redeclaring the literal, to avoid a payout -> treasury module
	// dependency while still staying in sync with app.go's registration.
	TreasuryModuleAccount = common.TreasuryModuleAccount
)

var (
	ParamsKey            = []byte{0x01}
	PayoutKeyPrefix      = []byte{0x02}
	BatchPayoutKeyPrefix = []byte{0x03}
)

func PayoutKey(id string) []byte      { return append(PayoutKeyPrefix, []byte(id)...) }
func BatchPayoutKey(id string) []byte { return append(BatchPayoutKeyPrefix, []byte(id)...) }
func PayoutByMerchantKey(m, id string) []byte {
	return append(append([]byte{0x11}, []byte(m)...), []byte(id)...)
}
func PayoutByRecipientKey(r, id string) []byte {
	return append(append([]byte{0x12}, []byte(r)...), []byte(id)...)
}
func PayoutByInitiatorKey(i, id string) []byte {
	return append(append([]byte{0x13}, []byte(i)...), []byte(id)...)
}
