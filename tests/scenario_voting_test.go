package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

func TestScenarioVoteEliminatesUniqueTopTarget(t *testing.T) {
	e := mustNewEngine(t, classicPlayers())
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "charlie", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "diana", Target: "charlie"})

	stateBefore := e.State()
	if len(stateBefore.Votes) != 3 {
		t.Fatalf("expected 3 votes before resolution, got %d", len(stateBefore.Votes))
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})

	stateAfter := e.State()
	if len(stateAfter.Votes) != 0 {
		t.Fatalf("expected votes to clear after resolution, got %d", len(stateAfter.Votes))
	}
	outcome := e.Position().Info.Outcome
	if !outcome.Ended {
		t.Fatalf("expected game to end after eliminating the last werewolf")
	}
	if outcome.Kind != engine.OutcomeVillagers {
		t.Fatalf("expected villagers to win, got %q", outcome.Kind)
	}
	assertWinningPlayers(t, outcome, "bob", "charlie", "diana")
}

func TestScenarioVoteTieEliminatesNobody(t *testing.T) {
	e := mustNewEngine(t, classicPlayers())
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice", Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})

	state := e.State()
	if len(state.Votes) != 0 {
		t.Fatalf("expected votes to clear after tie resolution, got %d", len(state.Votes))
	}
	if e.Info().AliveCount != 4 {
		t.Fatalf("expected no elimination on tie, got alive count %d", e.Info().AliveCount)
	}
	if e.Position().Info.Outcome.Ended {
		t.Fatalf("expected game to continue after tie")
	}
}

func TestScenarioVoteRequiresDayVotePhase(t *testing.T) {
	e := mustNewEngine(t, classicPlayers())
	_, err := e.Apply(engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	if err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid, got %v", err)
	}
}
