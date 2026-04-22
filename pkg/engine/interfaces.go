package engine

type Engine interface {
	Position() Position
	State() StateView
	Turn() TurnInfo
	Info() GameInfo
	Events() []Event
	Subscribe() <-chan Event
	Apply(transition Transition) (Position, error)
}

type RoleActionExecutor interface {
	Execute(game *Game, state GameState, actor PlayerID, action string, targets []PlayerID) error
}

type RelationshipStore interface {
	SetLovers(pair *LoverPair)
	Lovers() *LoverPair
}

type DeathSideEffectChecker interface {
	OnDeath(state GameState, dead PlayerID) ([]PlayerID, error)
}

type WinConditionChecker interface {
	Check(state GameState) (Outcome, bool)
}
