package engine

import "testing"

func TestLoversHelperBranches(t *testing.T) {
	if onlyLoversAlive(GameState{}) {
		t.Fatalf("expected onlyLoversAlive to be false when no lovers are configured")
	}

	checker := LoversWinChecker{}
	outcome, matched := checker.Check(GameState{
		Players: map[PlayerID]*Player{
			"alice": {ID: "alice", Alive: true},
			"bob":   {ID: "bob", Alive: true},
		},
		Lovers: &LoverPair{A: "alice", B: "bob"},
	})
	if matched || outcome.Ended {
		t.Fatalf("expected lovers check to fail when lover roles are missing, got %+v matched=%v", outcome, matched)
	}
}

func TestStateSortsCustomTitles(t *testing.T) {
	g := NewGame()
	if err := g.AddPlayer("alice"); err != nil {
		t.Fatalf("add alice: %v", err)
	}
	if err := g.AssignRole("alice", staticRole{id: RoleVillager, team: TeamVillagers}); err != nil {
		t.Fatalf("assign alice: %v", err)
	}
	g.AssignTitle(TitleKind("zeta"), "alice")
	g.AssignTitle(TitleKind("alpha"), "alice")

	e := &positionEngine{game: g, queuedRoles: map[RoleID]int{}, pendingActions: map[PlayerID]string{}}
	state := e.State()
	if len(state.Titles) != 2 {
		t.Fatalf("expected two custom titles, got %d", len(state.Titles))
	}
	if state.Titles[0].Kind != TitleKind("alpha") || state.Titles[1].Kind != TitleKind("zeta") {
		t.Fatalf("expected titles to be sorted alphabetically, got %+v", state.Titles)
	}
}
