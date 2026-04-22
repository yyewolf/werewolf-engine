package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

func TestScenarioVillagersWinWhenNoWerewolvesRemain(t *testing.T) {
	e := mustNewEngine(t, classicPlayers())
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice"})

	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end")
	}
	if outcome.Kind != engine.OutcomeVillagers {
		t.Fatalf("expected villagers to win, got %q", outcome.Kind)
	}
}
