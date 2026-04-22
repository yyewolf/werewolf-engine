package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

func assertWinningPlayers(t *testing.T, outcome engine.Outcome, expected ...engine.PlayerID) {
	t.Helper()
	if len(outcome.WinningPlayers) != len(expected) {
		t.Fatalf("expected %d winning players, got %d (%v)", len(expected), len(outcome.WinningPlayers), outcome.WinningPlayers)
	}
	counts := map[engine.PlayerID]int{}
	for _, id := range outcome.WinningPlayers {
		counts[id]++
	}
	for _, id := range expected {
		counts[id]--
	}
	for id, count := range counts {
		if count != 0 {
			t.Fatalf("winning players mismatch for %q: delta=%d (actual=%v expected=%v)", id, count, outcome.WinningPlayers, expected)
		}
	}
}

func findPlayerState(t *testing.T, e engine.Engine, id engine.PlayerID) engine.PlayerState {
	t.Helper()
	for _, player := range e.State().Players {
		if player.ID == id {
			return player
		}
	}
	t.Fatalf("player %q not found in state", id)
	return engine.PlayerState{}
}

func assertTitleHolder(t *testing.T, e engine.Engine, kind engine.TitleKind, holder engine.PlayerID) {
	t.Helper()
	for _, title := range e.State().Titles {
		if title.Kind == kind {
			if title.Holder != holder {
				t.Fatalf("expected %s holder %q, got %q", kind, holder, title.Holder)
			}
			return
		}
	}
	t.Fatalf("title %s not found", kind)
}

func mustNewEngine(t *testing.T, players []engine.BootstrapPlayer) engine.Engine {
	t.Helper()
	e, err := engine.NewPositionEngine(engine.Bootstrap{
		Players: players,
		InitialTurn: engine.TurnInfo{
			Phase: "night",
		},
	})
	if err != nil {
		t.Fatalf("new position engine: %v", err)
	}
	return e
}

func mustNewEmptyEngine(t *testing.T) engine.Engine {
	t.Helper()
	e, err := engine.NewPositionEngine(engine.Bootstrap{})
	if err != nil {
		t.Fatalf("new empty engine: %v", err)
	}
	return e
}

func mustApply(t *testing.T, e engine.Engine, transition engine.Transition) {
	t.Helper()
	if _, err := e.Apply(transition); err != nil {
		t.Fatalf("apply transition %s: %v", transition.Kind, err)
	}
}

func classicPlayers() []engine.BootstrapPlayer {
	return []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Werewolf{}},
		{ID: "bob", Role: roles.Seer{}},
		{ID: "charlie", Role: roles.Villager{}},
		{ID: "diana", Role: roles.Witch{}},
	}
}
