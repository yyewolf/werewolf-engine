package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

func TestEngineExposesStateTurnAndInfo(t *testing.T) {
	e := mustNewEngine(t, classicPlayers())
	state := e.State()
	turn := e.Turn()
	info := e.Info()

	if len(state.Players) != 4 {
		t.Fatalf("expected 4 players, got %d", len(state.Players))
	}
	if turn.Phase != "night" {
		t.Fatalf("expected initial phase night, got %q", turn.Phase)
	}
	if info.PlayerCount != 4 || info.AliveCount != 4 {
		t.Fatalf("expected info counts 4/4, got %d/%d", info.PlayerCount, info.AliveCount)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	if e.Turn().Phase != "day_vote" {
		t.Fatalf("expected phase to advance to day_vote")
	}
}
