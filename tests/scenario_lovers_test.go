package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

func TestScenarioLoversWinWhenOnlyPairSurvives(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Werewolf{}},
		{ID: "bob", Role: roles.Cupid{}},
		{ID: "charlie", Role: roles.Hunter{}},
		{ID: "diana", Role: roles.Hunter{}},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "bob",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "diana"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeLovers {
		t.Fatalf("expected lovers to win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "alice", "bob")
}

func TestScenarioLoversBothVillagersWinWithVillage(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Cupid{}},
		{ID: "bob", Role: roles.Villager{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "alice",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "charlie"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeVillagers {
		t.Fatalf("expected villagers to win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "alice", "bob")
}

func TestScenarioLoversBothWerewolvesWinWithWerewolves(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Cupid{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Werewolf{}},
		{ID: "diana", Role: roles.Villager{}},
		{ID: "eve", Role: roles.Villager{}},
		{ID: "frank", Role: roles.Villager{}},
		{ID: "grace", Role: roles.Villager{}},
		{ID: "henry", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "alice",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"bob", "charlie"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "diana"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "eve"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "frank"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "grace"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "henry"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeWerewolves {
		t.Fatalf("expected werewolves to win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "bob", "charlie")
}

func TestScenarioLoverChainDeath(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Cupid{}},
		{ID: "bob", Role: roles.Villager{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "alice",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeWerewolves {
		t.Fatalf("expected werewolves to win after lover chain death, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "charlie")
}
