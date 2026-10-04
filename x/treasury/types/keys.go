package types

import "github.com/nexarail/chain/x/common"

const (
	ModuleName = "treasury"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	// Aliases x/common's canonical constant — see x/common/accounts.go.
	TreasuryModuleAccount = common.TreasuryModuleAccount
)

var (
	ParamsKey        = []byte{0x01}
	AccountKeyPrefix = []byte{0x02}
	BudgetKeyPrefix  = []byte{0x03}
	GrantKeyPrefix   = []byte{0x04}
	SpendKeyPrefix   = []byte{0x05}
)

func AccountKey(id string) []byte { return append(AccountKeyPrefix, []byte(id)...) }
func BudgetKey(id string) []byte  { return append(BudgetKeyPrefix, []byte(id)...) }
func GrantKey(id string) []byte   { return append(GrantKeyPrefix, []byte(id)...) }
func SpendKey(id string) []byte   { return append(SpendKeyPrefix, []byte(id)...) }

func BudgetByAccountKey(accountID, budgetID string) []byte { return idx(0x11, accountID, budgetID) }
func GrantByBudgetKey(budgetID, grantID string) []byte     { return idx(0x12, budgetID, grantID) }
func GrantByRecipientKey(recipient, grantID string) []byte { return idx(0x13, recipient, grantID) }
func SpendByAccountKey(accountID, spendID string) []byte   { return idx(0x14, accountID, spendID) }
func SpendByBudgetKey(budgetID, spendID string) []byte     { return idx(0x15, budgetID, spendID) }
func SpendByGrantKey(grantID, spendID string) []byte       { return idx(0x16, grantID, spendID) }
func SpendByRequesterKey(requester, spendID string) []byte { return idx(0x17, requester, spendID) }
func SpendByRecipientKey(recipient, spendID string) []byte { return idx(0x18, recipient, spendID) }

func idx(prefix byte, a, b string) []byte {
	return append(append([]byte{prefix}, []byte(a)...), []byte(b)...)
}
