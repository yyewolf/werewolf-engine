package engine

type FlutePlayerWinChecker struct{}

func (FlutePlayerWinChecker) Check(state GameState) (Outcome, bool) {
	for id, player := range state.Players {
		if !player.Alive || player.Role == nil || player.Role.ID() != RoleFlutePlayer {
			continue
		}
		for otherID, other := range state.Players {
			if otherID == id || !other.Alive {
				continue
			}
			if !state.Charmed[otherID] {
				return Outcome{}, false
			}
		}
		return Outcome{Ended: true, Kind: OutcomeFlutePlayer, Reason: "all other alive players are charmed", WinningPlayers: []PlayerID{id}}, true
	}
	return Outcome{}, false
}

type WhiteWolfWinChecker struct{}

func (WhiteWolfWinChecker) Check(state GameState) (Outcome, bool) {
	whiteWolf := PlayerID("")
	aliveCount := 0
	for id, player := range state.Players {
		if !player.Alive {
			continue
		}
		aliveCount++
		if player.Role != nil && player.Role.ID() == RoleWhiteWolf {
			whiteWolf = id
		}
	}
	if whiteWolf != "" && aliveCount == 1 {
		return Outcome{Ended: true, Kind: OutcomeWhiteWolf, Reason: "white wolf is the sole survivor", WinningPlayers: []PlayerID{whiteWolf}}, true
	}
	return Outcome{}, false
}

type PyromaniacWinChecker struct{}

func (PyromaniacWinChecker) Check(state GameState) (Outcome, bool) {
	pyro := PlayerID("")
	aliveCount := 0
	for id, player := range state.Players {
		if !player.Alive {
			continue
		}
		aliveCount++
		if player.Role != nil && player.Role.ID() == RolePyromaniac {
			pyro = id
		}
	}
	if pyro != "" && aliveCount == 1 {
		return Outcome{Ended: true, Kind: OutcomePyromaniac, Reason: "pyromaniac is the sole survivor", WinningPlayers: []PlayerID{pyro}}, true
	}
	return Outcome{}, false
}

// LoverChainDeathChecker enforces the deterministic lover death side-effect.
type LoverChainDeathChecker struct{}

func (LoverChainDeathChecker) OnDeath(state GameState, dead PlayerID) ([]PlayerID, error) {
	if state.Lovers == nil {
		return nil, nil
	}
	switch dead {
	case state.Lovers.A:
		return []PlayerID{state.Lovers.B}, nil
	case state.Lovers.B:
		return []PlayerID{state.Lovers.A}, nil
	default:
		return nil, nil
	}
}

// LoversWinChecker has priority over team win checks.
type LoversWinChecker struct{}

func (LoversWinChecker) Check(state GameState) (Outcome, bool) {
	a, b, ok := aliveLovers(state)
	if !ok {
		return Outcome{}, false
	}
	if !onlyLoversAlive(state) {
		return Outcome{}, false
	}
	if a.Role == nil || b.Role == nil {
		return Outcome{}, false
	}
	if a.Role.Team() == b.Role.Team() {
		return Outcome{}, false
	}
	return Outcome{
		Ended:          true,
		Kind:           OutcomeLovers,
		Reason:         "only mixed-team lovers are alive",
		WinningPlayers: []PlayerID{state.Lovers.A, state.Lovers.B},
	}, true
}

type VillagersWinChecker struct{}

func (VillagersWinChecker) Check(state GameState) (Outcome, bool) {
	aliveWerewolves := 0
	alivePlayers := 0
	winners := []PlayerID{}
	for _, player := range state.Players {
		if !player.Alive {
			continue
		}
		alivePlayers++
		if player.Role != nil && player.Role.Team() == TeamWerewolves {
			aliveWerewolves++
			continue
		}
		winners = append(winners, player.ID)
	}
	if alivePlayers > 0 && aliveWerewolves == 0 {
		return Outcome{Ended: true, Kind: OutcomeVillagers, Reason: "no werewolves remain", WinningPlayers: winners}, true
	}
	return Outcome{}, false
}

type WerewolvesWinChecker struct{}

func (WerewolvesWinChecker) Check(state GameState) (Outcome, bool) {
	aliveWerewolves := 0
	aliveNonWerewolves := 0
	winners := []PlayerID{}
	for _, player := range state.Players {
		if !player.Alive {
			continue
		}
		if player.Role != nil && player.Role.Team() == TeamWerewolves {
			aliveWerewolves++
			winners = append(winners, player.ID)
			continue
		}
		aliveNonWerewolves++
	}
	if aliveWerewolves > 0 && aliveWerewolves >= aliveNonWerewolves {
		return Outcome{Ended: true, Kind: OutcomeWerewolves, Reason: "werewolves reached parity", WinningPlayers: winners}, true
	}
	return Outcome{}, false
}

func aliveLovers(state GameState) (*Player, *Player, bool) {
	if state.Lovers == nil {
		return nil, nil, false
	}
	a, okA := state.Players[state.Lovers.A]
	b, okB := state.Players[state.Lovers.B]
	if !okA || !okB || !a.Alive || !b.Alive {
		return nil, nil, false
	}
	return a, b, true
}

func onlyLoversAlive(state GameState) bool {
	if state.Lovers == nil {
		return false
	}
	for id, player := range state.Players {
		if id == state.Lovers.A || id == state.Lovers.B {
			continue
		}
		if player.Alive {
			return false
		}
	}
	return true
}
