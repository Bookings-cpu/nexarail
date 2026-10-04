package types

import "fmt"

type GenesisState struct {
	Params        Params            `json:"params"`
	Accounts      []TreasuryAccount `json:"accounts"`
	Budgets       []Budget          `json:"budgets"`
	Grants        []Grant           `json:"grants"`
	SpendRequests []SpendRequest    `json:"spend_requests"`
}

func DefaultGenesis() *GenesisState    { return &GenesisState{Params: DefaultParams()} }
func (gs *GenesisState) ProtoMessage() {}
func (gs *GenesisState) Reset()        { *gs = GenesisState{} }
func (gs *GenesisState) String() string {
	return fmt.Sprintf("Genesis{accts=%d,budgets=%d,grants=%d,spends=%d}", len(gs.Accounts), len(gs.Budgets), len(gs.Grants), len(gs.SpendRequests))
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for i, a := range gs.Accounts {
		if seen[a.AccountId] {
			return fmt.Errorf("dup account %s at %d", a.AccountId, i)
		}
		seen[a.AccountId] = true
		if err := a.ValidateWithParams(gs.Params); err != nil {
			return err
		}
	}
	seen = make(map[string]bool)
	for i, b := range gs.Budgets {
		if seen[b.BudgetId] {
			return fmt.Errorf("dup budget %s at %d", b.BudgetId, i)
		}
		seen[b.BudgetId] = true
		if err := b.ValidateWithParams(gs.Params); err != nil {
			return err
		}
	}
	budgetIDs := make(map[string]bool, len(gs.Budgets))
	for _, b := range gs.Budgets {
		budgetIDs[b.BudgetId] = true
	}
	seen = make(map[string]bool)
	for i, g := range gs.Grants {
		if seen[g.GrantId] {
			return fmt.Errorf("dup grant %s at %d", g.GrantId, i)
		}
		seen[g.GrantId] = true
		if err := g.ValidateWithParams(gs.Params); err != nil {
			return err
		}
		// Referential integrity: a grant's BudgetId must name a real budget in
		// this same genesis — without this, a grant can load pointing at a
		// nonexistent budget, which previously panicked the first time any
		// keeper function discarded GetBudget's "found" bool and dereferenced
		// a zero-value Budget with nil-Int coin fields.
		if g.BudgetId != "" && !budgetIDs[g.BudgetId] {
			return fmt.Errorf("grant %s references nonexistent budget %s", g.GrantId, g.BudgetId)
		}
	}
	grantIDs := make(map[string]bool, len(gs.Grants))
	for _, g := range gs.Grants {
		grantIDs[g.GrantId] = true
	}
	seen = make(map[string]bool)
	for i, s := range gs.SpendRequests {
		if seen[s.SpendId] {
			return fmt.Errorf("dup spend %s at %d", s.SpendId, i)
		}
		seen[s.SpendId] = true
		if err := s.ValidateWithParams(gs.Params); err != nil {
			return err
		}
		if s.BudgetId != "" && !budgetIDs[s.BudgetId] {
			return fmt.Errorf("spend %s references nonexistent budget %s", s.SpendId, s.BudgetId)
		}
		if s.GrantId != "" && !grantIDs[s.GrantId] {
			return fmt.Errorf("spend %s references nonexistent grant %s", s.SpendId, s.GrantId)
		}
	}
	return nil
}
