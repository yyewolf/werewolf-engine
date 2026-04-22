package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

func TestScenarioFullGameFromLobbyWithSystemComposition(t *testing.T) {
	e := mustNewEmptyEngine(t)
	players := []engine.PlayerID{"alice", "bob", "charlie", "diana", "edgar", "frank", "grace", "heidi", "ivan", "judy", "karl", "laura"}
	composition := []engine.RoleID{
		engine.RoleCupid,
		engine.RoleSavior,
		engine.RoleRaven,
		engine.RoleFlutePlayer,
		engine.RolePyromaniac,
		engine.RoleVillageIdiot,
		engine.RoleAncien,
		engine.RoleWerewolf,
		engine.RoleWerewolf,
		engine.RoleSeer,
		engine.RoleWitch,
		engine.RoleHunter,
	}

	for _, player := range players {
		mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: player})
	}
	for _, role := range composition {
		mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: role})
	}
	if e.Info().PlayerCount != len(players) || e.Info().QueuedRoleCount != len(composition) {
		t.Fatalf("unexpected lobby setup info: %+v", e.Info())
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	assignments := []struct {
		player engine.PlayerID
		role   engine.RoleID
	}{
		{player: "alice", role: engine.RoleCupid},
		{player: "bob", role: engine.RoleSavior},
		{player: "charlie", role: engine.RoleRaven},
		{player: "diana", role: engine.RoleFlutePlayer},
		{player: "edgar", role: engine.RolePyromaniac},
		{player: "frank", role: engine.RoleVillageIdiot},
		{player: "grace", role: engine.RoleAncien},
		{player: "heidi", role: engine.RoleWerewolf},
		{player: "ivan", role: engine.RoleWerewolf},
		{player: "judy", role: engine.RoleSeer},
		{player: "karl", role: engine.RoleWitch},
		{player: "laura", role: engine.RoleHunter},
	}
	for _, assignment := range assignments {
		mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: assignment.player, Role: assignment.role})
	}
	if e.Info().AssignedRoleCount != len(players) || e.Info().QueuedRoleCount != 0 {
		t.Fatalf("unexpected assignment info: %+v", e.Info())
	}

	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "alice",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"charlie", "heidi"},
	})
	if len(e.State().Relationships) != 1 {
		t.Fatalf("expected Cupid to create one lovers relationship, got %d", len(e.State().Relationships))
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "night"})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "bob",
		Action:  engine.RoleActionProtectPlayer,
		Targets: []engine.PlayerID{"grace"},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionMarkForCrows,
		Targets: []engine.PlayerID{"heidi"},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "diana",
		Action:  engine.RoleActionCharmPlayers,
		Targets: []engine.PlayerID{"judy", "karl"},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "edgar",
		Action:  engine.RoleActionDousePlayers,
		Targets: []engine.PlayerID{"heidi", "ivan"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "grace", Cause: engine.CauseAttack})
	if !findPlayerState(t, e, "grace").Alive {
		t.Fatalf("ancien should survive the first attack during the night")
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	votes := map[engine.PlayerID]engine.PlayerID{
		"alice":   "heidi",
		"bob":     "heidi",
		"charlie": "heidi",
		"diana":   "heidi",
		"edgar":   "heidi",
		"frank":   "heidi",
		"grace":   "heidi",
		"heidi":   "grace",
		"ivan":    "grace",
		"judy":    "heidi",
		"karl":    "heidi",
		"laura":   "heidi",
	}
	for actor, target := range votes {
		mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: actor, Target: target})
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	if findPlayerState(t, e, "heidi").Alive {
		t.Fatalf("expected the marked werewolf to be eliminated by vote")
	}
	if findPlayerState(t, e, "charlie").Alive {
		t.Fatalf("expected lover chain death after Heidi was eliminated")
	}
	if e.Position().Info.Outcome.Ended {
		t.Fatalf("expected one werewolf to remain so the game continues")
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "night"})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "bob",
		Action:  engine.RoleActionProtectPlayer,
		Targets: []engine.PlayerID{"judy"},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "diana",
		Action:  engine.RoleActionCharmPlayers,
		Targets: []engine.PlayerID{"frank", "laura"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "edgar", Action: engine.RoleActionIgnite})

	if findPlayerState(t, e, "ivan").Alive {
		t.Fatalf("expected the doused remaining werewolf to die on ignition")
	}
	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected full-game scenario to reach a terminal outcome")
	}
	if outcome.Kind != engine.OutcomeVillagers {
		t.Fatalf("expected villagers to win after the last werewolf dies, got %q", outcome.Kind)
	}
}
