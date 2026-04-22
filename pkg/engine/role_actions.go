package engine

type DefaultRoleActionExecutor struct{}

func (DefaultRoleActionExecutor) Execute(game *Game, state GameState, actor PlayerID, action string, targets []PlayerID) error {
	actorPlayer, ok := state.Players[actor]
	if !ok || actorPlayer.Role == nil {
		return ErrPlayerUnknown
	}

	switch action {
	case RoleActionAttackPlayer:
		if (actorPlayer.Role.ID() != RoleWerewolf && actorPlayer.Role.ID() != RoleWhiteWolf) || len(targets) != 1 {
			return ErrRoleActionDenied
		}
		if _, ok := state.Players[targets[0]]; !ok {
			return ErrPlayerUnknown
		}
		return game.KillWithCause(targets[0], CauseAttack)
	case RoleActionInspectPlayer:
		if actorPlayer.Role.ID() != RoleSeer || len(targets) != 1 {
			return ErrRoleActionDenied
		}
		target, ok := state.Players[targets[0]]
		if !ok || !target.Alive {
			return ErrPlayerUnknown
		}
		return nil // result is emitted as EventPlayerInspected by the engine
	case RoleActionSetLovers:
		if actorPlayer.Role.ID() != RoleCupid {
			return ErrRoleActionDenied
		}
		if len(targets) != 2 {
			return ErrTransitionInvalid
		}
		return game.SetLovers(targets[0], targets[1])
	case RoleActionCharmPlayers:
		if actorPlayer.Role.ID() != RoleFlutePlayer || len(targets) != 2 {
			return ErrRoleActionDenied
		}
		for _, target := range targets {
			if target == actor {
				continue
			}
			if _, ok := state.Players[target]; !ok {
				return ErrPlayerUnknown
			}
			game.SetCharmed(target, true)
		}
		return nil
	case RoleActionDousePlayers:
		if actorPlayer.Role.ID() != RolePyromaniac || len(targets) == 0 {
			return ErrRoleActionDenied
		}
		for _, target := range targets {
			if target == actor {
				continue
			}
			if _, ok := state.Players[target]; !ok {
				return ErrPlayerUnknown
			}
			game.SetDoused(target, true)
		}
		return nil
	case RoleActionIgnite:
		if actorPlayer.Role.ID() != RolePyromaniac {
			return ErrRoleActionDenied
		}
		for target, doused := range state.Doused {
			if !doused || target == actor {
				continue
			}
			if err := game.KillWithCause(target, CauseIgnition); err != nil {
				return err
			}
		}
		return nil
	case RoleActionProtectPlayer:
		if actorPlayer.Role.ID() != RoleSavior || len(targets) != 1 {
			return ErrRoleActionDenied
		}
		if _, ok := state.Players[targets[0]]; !ok {
			return ErrPlayerUnknown
		}
		game.SetProtected(targets[0], true)
		return nil
	case RoleActionMarkForCrows:
		if actorPlayer.Role.ID() != RoleRaven || len(targets) != 1 {
			return ErrRoleActionDenied
		}
		if _, ok := state.Players[targets[0]]; !ok {
			return ErrPlayerUnknown
		}
		game.AddVoteModifier(targets[0], 2)
		return nil
	case RoleActionAssignCaptain:
		if len(targets) != 1 {
			return ErrTransitionInvalid
		}
		if _, ok := state.Players[targets[0]]; !ok {
			return ErrPlayerUnknown
		}
		game.AssignTitle(TitleCaptain, targets[0])
		return nil
	default:
		return ErrRoleActionUnknown
	}
}
