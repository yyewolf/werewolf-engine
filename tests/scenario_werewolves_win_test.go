package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

func TestScenarioWerewolvesWinOnParity(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Werewolf{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
		{ID: "diana", Role: roles.Seer{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "diana"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeWerewolves {
		t.Fatalf("expected werewolves to win, got %q", outcome.Kind)
	}
}
