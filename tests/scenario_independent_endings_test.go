package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

func TestScenarioFlutePlayerWinsWhenAllOthersAreCharmed(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.FlutePlayer{}},
		{ID: "bob", Role: roles.Villager{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionCharmPlayers, Targets: []engine.PlayerID{"bob", "charlie"}})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeFlutePlayer {
		t.Fatalf("expected flute player win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "alice")
}

func TestScenarioWhiteWolfWinsAsLastSurvivor(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.WhiteWolf{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "charlie"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeWhiteWolf {
		t.Fatalf("expected white wolf win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "alice")
}

func TestScenarioPyromaniacWinsAfterIgnition(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Pyromaniac{}},
		{ID: "bob", Role: roles.Villager{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionDousePlayers, Targets: []engine.PlayerID{"bob", "charlie"}})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionIgnite})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomePyromaniac {
		t.Fatalf("expected pyromaniac win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "alice")
}
