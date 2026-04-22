package tests

import (
	"errors"
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

type chainingDeathChecker struct {
	called bool
	dead   []engine.PlayerID
}

func (c *chainingDeathChecker) OnDeath(state engine.GameState, dead engine.PlayerID) ([]engine.PlayerID, error) {
	c.called = true
	c.dead = append(c.dead, dead)
	if dead == "alice" {
		return []engine.PlayerID{"bob"}, nil
	}
	return nil, nil
}

type fixedWinChecker struct {
	called  bool
	outcome engine.Outcome
	match   bool
}

type errorDeathChecker struct {
	err error
}

func (c *fixedWinChecker) Check(state engine.GameState) (engine.Outcome, bool) {
	c.called = true
	return c.outcome, c.match
}

func (c errorDeathChecker) OnDeath(state engine.GameState, dead engine.PlayerID) ([]engine.PlayerID, error) {
	return nil, c.err
}

func TestGameOptionsAndKillUseCustomCheckers(t *testing.T) {
	deathChecker := &chainingDeathChecker{}
	winChecker := &fixedWinChecker{
		outcome: engine.Outcome{Ended: true, Kind: engine.OutcomeVillagers, Reason: "custom", WinningPlayers: []engine.PlayerID{"charlie"}},
		match:   true,
	}
	g := engine.NewGame(engine.WithDeathSideEffects(deathChecker), engine.WithWinConditionCheckers(winChecker))
	if err := g.AddPlayer("alice"); err != nil {
		t.Fatalf("add alice: %v", err)
	}
	if err := g.AddPlayer("bob"); err != nil {
		t.Fatalf("add bob: %v", err)
	}
	if err := g.AddPlayer("charlie"); err != nil {
		t.Fatalf("add charlie: %v", err)
	}
	if err := g.AssignRole("alice", roles.Villager{}); err != nil {
		t.Fatalf("assign alice: %v", err)
	}
	if err := g.AssignRole("bob", roles.Villager{}); err != nil {
		t.Fatalf("assign bob: %v", err)
	}
	if err := g.AssignRole("charlie", roles.Villager{}); err != nil {
		t.Fatalf("assign charlie: %v", err)
	}

	if err := g.Kill("alice"); err != nil {
		t.Fatalf("kill alice: %v", err)
	}
	if !deathChecker.called {
		t.Fatalf("expected custom death checker to be called")
	}
	if len(deathChecker.dead) != 2 || deathChecker.dead[0] != "alice" || deathChecker.dead[1] != "bob" {
		t.Fatalf("expected chained deaths for alice then bob, got %v", deathChecker.dead)
	}
	outcome := g.Outcome()
	if !winChecker.called {
		t.Fatalf("expected custom win checker to be called")
	}
	if outcome.Kind != engine.OutcomeVillagers || len(outcome.WinningPlayers) != 1 || outcome.WinningPlayers[0] != "charlie" {
		t.Fatalf("unexpected custom outcome: %+v", outcome)
	}
}

func TestGameValidationErrors(t *testing.T) {
	g := engine.NewGame()
	inProgress := engine.NewGame()
	if err := inProgress.AddPlayer("villager"); err != nil {
		t.Fatalf("add villager: %v", err)
	}
	if err := inProgress.AddPlayer("villager_two"); err != nil {
		t.Fatalf("add villager_two: %v", err)
	}
	if err := inProgress.AddPlayer("wolf"); err != nil {
		t.Fatalf("add wolf: %v", err)
	}
	if err := inProgress.AssignRole("villager", roles.Villager{}); err != nil {
		t.Fatalf("assign villager: %v", err)
	}
	if err := inProgress.AssignRole("villager_two", roles.Villager{}); err != nil {
		t.Fatalf("assign villager_two: %v", err)
	}
	if err := inProgress.AssignRole("wolf", roles.Werewolf{}); err != nil {
		t.Fatalf("assign wolf: %v", err)
	}
	if outcome := inProgress.Outcome(); outcome.Ended || outcome.Kind != engine.OutcomeNone {
		t.Fatalf("expected in-progress outcome, got %+v", outcome)
	}

	if err := g.AddPlayer("alice"); err != nil {
		t.Fatalf("add alice: %v", err)
	}
	if err := g.AddPlayer("alice"); err != engine.ErrPlayerExists {
		t.Fatalf("expected ErrPlayerExists, got %v", err)
	}
	if err := g.AssignRole("missing", roles.Villager{}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing player, got %v", err)
	}
	if err := g.AssignRole("alice", nil); err != engine.ErrRoleMissing {
		t.Fatalf("expected ErrRoleMissing, got %v", err)
	}
	if err := g.AssignRole("alice", roles.Villager{}); err != nil {
		t.Fatalf("assign alice: %v", err)
	}
	if err := g.SetLovers("alice", "alice"); err != engine.ErrInvalidLovers {
		t.Fatalf("expected ErrInvalidLovers for duplicate lover, got %v", err)
	}
	if err := g.SetLovers("alice", "missing"); err != engine.ErrInvalidLovers {
		t.Fatalf("expected ErrInvalidLovers for missing lover, got %v", err)
	}
	if err := g.CastVote("missing", "alice"); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing voter, got %v", err)
	}
	if err := g.AddPlayer("bob"); err != nil {
		t.Fatalf("add bob: %v", err)
	}
	if err := g.AssignRole("bob", roles.Villager{}); err != nil {
		t.Fatalf("assign bob: %v", err)
	}
	if err := g.Kill("bob"); err != nil {
		t.Fatalf("kill bob: %v", err)
	}
	if err := g.CastVote("alice", "missing"); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing target, got %v", err)
	}
	if err := g.CastVote("alice", "bob"); err != engine.ErrPlayerDead {
		t.Fatalf("expected ErrPlayerDead for dead target, got %v", err)
	}
	if err := g.CastVote("bob", "alice"); err != engine.ErrPlayerDead {
		t.Fatalf("expected ErrPlayerDead for dead voter, got %v", err)
	}
	if err := g.Kill("missing"); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing kill target, got %v", err)
	}
}

func TestPositionEngineValidationBranches(t *testing.T) {
	e := mustNewEmptyEngine(t)
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionKind("unknown")}); err != engine.ErrTransitionUnknown {
		t.Fatalf("expected ErrTransitionUnknown, got %v", err)
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid when adding players after start, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleWerewolf}); err != engine.ErrRoleUnavailable {
		t.Fatalf("expected ErrRoleUnavailable, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionRoleAction}); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for empty role action, got %v", err)
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn}); err != nil {
		t.Fatalf("expected blank phase advance to succeed, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionResolveVotes}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for vote resolution outside day_vote, got %v", err)
	}

	e2 := mustNewEmptyEngine(t)
	mustApply(t, e2, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e2, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleID("mystery")})
	mustApply(t, e2, engine.Transition{Kind: engine.TransitionStartGame})
	if _, err := e2.Apply(engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleID("mystery")}); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for unknown role, got %v", err)
	}

	e3 := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Villager{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
		{ID: "diana", Role: roles.Villager{}},
	})
	if _, err := e3.Apply(engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: "mystery_action"}); err != engine.ErrRoleActionUnknown {
		t.Fatalf("expected ErrRoleActionUnknown, got %v", err)
	}
	if _, err := e3.Apply(engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice"}); err != nil {
		t.Fatalf("kill existing player: %v", err)
	}
	if _, err := e3.Apply(engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "charlie"}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid outside day_vote, got %v", err)
	}
}

func TestNewPositionEngineDerivesDefaultPhase(t *testing.T) {
	lobby, err := engine.NewPositionEngine(engine.Bootstrap{})
	if err != nil {
		t.Fatalf("new lobby engine: %v", err)
	}
	if lobby.Turn().Phase != "lobby" {
		t.Fatalf("expected lobby phase, got %q", lobby.Turn().Phase)
	}

	assignment, err := engine.NewPositionEngine(engine.Bootstrap{Players: []engine.BootstrapPlayer{{ID: "alice"}}})
	if err != nil {
		t.Fatalf("new assignment engine: %v", err)
	}
	if assignment.Turn().Phase != "role_assignment" {
		t.Fatalf("expected role_assignment phase, got %q", assignment.Turn().Phase)
	}

	night, err := engine.NewPositionEngine(engine.Bootstrap{Players: []engine.BootstrapPlayer{{ID: "alice", Role: roles.Villager{}}}})
	if err != nil {
		t.Fatalf("new night engine: %v", err)
	}
	if night.Turn().Phase != "night" {
		t.Fatalf("expected night phase, got %q", night.Turn().Phase)
	}
}

func TestBuiltInRolesExposeIDsAndTeams(t *testing.T) {
	tests := []struct {
		name string
		role engine.Role
		id   engine.RoleID
		team engine.Team
	}{
		{name: "villager", role: roles.Villager{}, id: engine.RoleVillager, team: engine.TeamVillagers},
		{name: "werewolf", role: roles.Werewolf{}, id: engine.RoleWerewolf, team: engine.TeamWerewolves},
		{name: "seer", role: roles.Seer{}, id: engine.RoleSeer, team: engine.TeamVillagers},
		{name: "witch", role: roles.Witch{}, id: engine.RoleWitch, team: engine.TeamVillagers},
		{name: "hunter", role: roles.Hunter{}, id: engine.RoleHunter, team: engine.TeamVillagers},
		{name: "little girl", role: roles.LittleGirl{}, id: engine.RoleLittleGirl, team: engine.TeamVillagers},
		{name: "cupid", role: roles.Cupid{}, id: engine.RoleCupid, team: engine.TeamVillagers},
		{name: "thief", role: roles.Thief{}, id: engine.RoleThief, team: engine.TeamVillagers},
		{name: "ancien", role: roles.Ancien{}, id: engine.RoleAncien, team: engine.TeamVillagers},
		{name: "scapegoat", role: roles.Scapegoat{}, id: engine.RoleScapegoat, team: engine.TeamVillagers},
		{name: "village idiot", role: roles.VillageIdiot{}, id: engine.RoleVillageIdiot, team: engine.TeamVillagers},
		{name: "flute player", role: roles.FlutePlayer{}, id: engine.RoleFlutePlayer, team: engine.TeamVillagers},
		{name: "savior", role: roles.Savior{}, id: engine.RoleSavior, team: engine.TeamVillagers},
		{name: "raven", role: roles.Raven{}, id: engine.RoleRaven, team: engine.TeamVillagers},
		{name: "white wolf", role: roles.WhiteWolf{}, id: engine.RoleWhiteWolf, team: engine.TeamWerewolves},
		{name: "pyromaniac", role: roles.Pyromaniac{}, id: engine.RolePyromaniac, team: engine.TeamVillagers},
	}

	for _, test := range tests {
		if test.role.ID() != test.id {
			t.Fatalf("%s ID mismatch: got %q want %q", test.name, test.role.ID(), test.id)
		}
		if test.role.Team() != test.team {
			t.Fatalf("%s team mismatch: got %q want %q", test.name, test.role.Team(), test.team)
		}
	}
}

func TestRoleActionExecutorValidationBranches(t *testing.T) {
	executor := engine.DefaultRoleActionExecutor{}
	g := engine.NewGame()
	for _, player := range []struct {
		id   engine.PlayerID
		role engine.Role
	}{
		{id: "cupid", role: roles.Cupid{}},
		{id: "flute", role: roles.FlutePlayer{}},
		{id: "pyro", role: roles.Pyromaniac{}},
		{id: "savior", role: roles.Savior{}},
		{id: "raven", role: roles.Raven{}},
		{id: "target", role: roles.Villager{}},
		{id: "other", role: roles.Werewolf{}},
	} {
		if err := g.AddPlayer(player.id); err != nil {
			t.Fatalf("add %s: %v", player.id, err)
		}
		if err := g.AssignRole(player.id, player.role); err != nil {
			t.Fatalf("assign %s: %v", player.id, err)
		}
	}

	players := map[engine.PlayerID]*engine.Player{
		"cupid":  {ID: "cupid", Role: roles.Cupid{}, Alive: true},
		"flute":  {ID: "flute", Role: roles.FlutePlayer{}, Alive: true},
		"pyro":   {ID: "pyro", Role: roles.Pyromaniac{}, Alive: true},
		"savior": {ID: "savior", Role: roles.Savior{}, Alive: true},
		"raven":  {ID: "raven", Role: roles.Raven{}, Alive: true},
		"target": {ID: "target", Role: roles.Villager{}, Alive: true},
		"other":  {ID: "other", Role: roles.Werewolf{}, Alive: true},
	}
	baseState := engine.GameState{
		Players:       players,
		Protected:     map[engine.PlayerID]bool{},
		Charmed:       map[engine.PlayerID]bool{},
		Doused:        map[engine.PlayerID]bool{},
		Revealed:      map[engine.PlayerID]bool{},
		VoteDisabled:  map[engine.PlayerID]bool{},
		VoteModifiers: map[engine.PlayerID]int{},
		Titles:        map[engine.TitleKind]engine.PlayerID{},
		AncientSaved:  map[engine.PlayerID]bool{},
	}

	if err := executor.Execute(g, engine.GameState{Players: map[engine.PlayerID]*engine.Player{}}, "ghost", engine.RoleActionSetLovers, []engine.PlayerID{"target", "other"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing actor, got %v", err)
	}
	if err := executor.Execute(g, baseState, "cupid", engine.RoleActionSetLovers, []engine.PlayerID{"target"}); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for wrong lover target count, got %v", err)
	}
	if err := executor.Execute(g, baseState, "flute", engine.RoleActionCharmPlayers, []engine.PlayerID{"flute", "missing"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing charmed target, got %v", err)
	}
	if err := executor.Execute(g, baseState, "pyro", engine.RoleActionDousePlayers, []engine.PlayerID{"pyro", "missing"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing douse target, got %v", err)
	}
	igniteState := baseState
	igniteState.Doused = map[engine.PlayerID]bool{"pyro": true, "target": true, "other": false}
	if err := executor.Execute(g, igniteState, "pyro", engine.RoleActionIgnite, nil); err != nil {
		t.Fatalf("ignite should succeed, got %v", err)
	}
	if err := executor.Execute(g, baseState, "target", engine.RoleActionIgnite, nil); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-pyro ignite, got %v", err)
	}
	if err := executor.Execute(g, baseState, "savior", engine.RoleActionProtectPlayer, []engine.PlayerID{"missing"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing protect target, got %v", err)
	}
	if err := executor.Execute(g, baseState, "raven", engine.RoleActionMarkForCrows, []engine.PlayerID{"missing"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing crow target, got %v", err)
	}
	if err := executor.Execute(g, baseState, "target", engine.RoleActionAssignCaptain, nil); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for missing captain target, got %v", err)
	}
	if err := executor.Execute(g, baseState, "target", engine.RoleActionAssignCaptain, []engine.PlayerID{"missing"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing captain target, got %v", err)
	}
}

func TestGameEdgeBranches(t *testing.T) {
	g := engine.NewGame()
	if err := g.AddPlayer("missing_a"); err != nil {
		t.Fatalf("add missing_a: %v", err)
	}
	if err := g.AddPlayer("existing_b"); err != nil {
		t.Fatalf("add existing_b: %v", err)
	}
	if err := g.AssignRole("missing_a", roles.Villager{}); err != nil {
		t.Fatalf("assign missing_a: %v", err)
	}
	if err := g.AssignRole("existing_b", roles.Villager{}); err != nil {
		t.Fatalf("assign existing_b: %v", err)
	}
	if err := g.SetLovers("ghost", "existing_b"); err != engine.ErrInvalidLovers {
		t.Fatalf("expected ErrInvalidLovers for missing first lover, got %v", err)
	}

	deathErr := errors.New("death checker failed")
	g2 := engine.NewGame(engine.WithDeathSideEffects(errorDeathChecker{err: deathErr}))
	if err := g2.AddPlayer("alice"); err != nil {
		t.Fatalf("add alice: %v", err)
	}
	if err := g2.AssignRole("alice", roles.Villager{}); err != nil {
		t.Fatalf("assign alice: %v", err)
	}
	if err := g2.KillWithCause("alice", engine.CauseDirect); !errors.Is(err, deathErr) {
		t.Fatalf("expected propagated death checker error, got %v", err)
	}
	if err := g2.KillWithCause("alice", engine.CauseDirect); err != nil {
		t.Fatalf("killing already dead player should be a no-op, got %v", err)
	}
}

func TestPositionEngineVoteAndBootstrapEdgeBranches(t *testing.T) {
	if _, err := engine.NewPositionEngine(engine.Bootstrap{Players: []engine.BootstrapPlayer{{ID: "alice"}, {ID: "alice"}}}); err != engine.ErrPlayerExists {
		t.Fatalf("expected ErrPlayerExists for duplicate bootstrap players, got %v", err)
	}

	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.VillageIdiot{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "charlie", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice", Target: "bob"}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for disabled voter, got %v", err)
	}

	e2 := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Villager{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e2, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e2, engine.Transition{Kind: engine.TransitionResolveVotes})
	if len(e2.State().Votes) != 0 || !findPlayerState(t, e2, "alice").Alive || !findPlayerState(t, e2, "bob").Alive || !findPlayerState(t, e2, "charlie").Alive {
		t.Fatalf("expected no-vote resolution to leave players alive and votes empty")
	}
}

func TestSameTeamLoversFallBackToTeamWin(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Cupid{}},
		{ID: "bob", Role: roles.Seer{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionSetLovers, Targets: []engine.PlayerID{"alice", "bob"}})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "charlie"})
	if e.Position().Info.Outcome.Kind != engine.OutcomeVillagers {
		t.Fatalf("expected same-team lovers scenario to fall back to villagers win, got %q", e.Position().Info.Outcome.Kind)
	}
}

func TestGameResolveVotesIgnoresDeadVotes(t *testing.T) {
	g := engine.NewGame()
	for _, player := range []struct {
		id   engine.PlayerID
		role engine.Role
	}{
		{id: "alice", role: roles.Villager{}},
		{id: "bob", role: roles.Werewolf{}},
		{id: "charlie", role: roles.Villager{}},
		{id: "diana", role: roles.Villager{}},
	} {
		if err := g.AddPlayer(player.id); err != nil {
			t.Fatalf("add %s: %v", player.id, err)
		}
		if err := g.AssignRole(player.id, player.role); err != nil {
			t.Fatalf("assign %s: %v", player.id, err)
		}
	}
	if err := g.CastVote("alice", "bob"); err != nil {
		t.Fatalf("cast vote: %v", err)
	}
	if err := g.Kill("alice"); err != nil {
		t.Fatalf("kill voter: %v", err)
	}
	if err := g.ResolveVotes(); err != nil {
		t.Fatalf("resolve votes: %v", err)
	}
	if outcome := g.Outcome(); outcome.Ended || outcome.Kind != engine.OutcomeNone {
		t.Fatalf("expected dead-voter vote to be ignored, got %+v", outcome)
	}
}

func TestAdditionalPositionEngineValidationBranches(t *testing.T) {
	e := mustNewEmptyEngine(t)
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAddRole}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for empty queued role, got %v", err)
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "diana"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "eve"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "frank"})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"}); err != engine.ErrPlayerExists {
		t.Fatalf("expected ErrPlayerExists for duplicate add player, got %v", err)
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleWerewolf})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleWerewolf}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for add role outside lobby, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionStartGame}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for start_game outside lobby, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAssignRole, Role: engine.RoleVillager}); err != engine.ErrPhaseInvalid {
		t.Fatalf("expected ErrPhaseInvalid for missing assign target, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAssignRole, Target: "ghost", Role: engine.RoleVillager}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing assignee, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionKillPlayer}); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for empty kill target, got %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionKillPlayer, Target: "ghost"}); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown for missing kill target, got %v", err)
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "bob", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "charlie", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "diana", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "eve", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "frank", Role: engine.RoleWerewolf})
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"}); err != nil {
		t.Fatalf("advance to day_vote: %v", err)
	}
	if _, err := e.Apply(engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice"}); err != engine.ErrTransitionInvalid {
		t.Fatalf("expected ErrTransitionInvalid for missing vote target, got %v", err)
	}

	erroringEngine, err := engine.NewPositionEngine(engine.Bootstrap{
		Players: []engine.BootstrapPlayer{
			{ID: "alice", Role: roles.Villager{}},
			{ID: "bob", Role: roles.Villager{}},
			{ID: "carol", Role: roles.Werewolf{}},
		},
		InitialTurn: engine.TurnInfo{Phase: "day_vote"},
	}, engine.WithDeathSideEffects(errorDeathChecker{err: errors.New("resolve failed")}))
	if err != nil {
		t.Fatalf("new erroring engine: %v", err)
	}
	mustApply(t, erroringEngine, engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice", Target: "bob"})
	if _, err := erroringEngine.Apply(engine.Transition{Kind: engine.TransitionResolveVotes}); err == nil {
		t.Fatalf("expected resolve votes to propagate kill error")
	}

	unassigned := mustNewEmptyEngine(t)
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionStartGame})
	mustApply(t, unassigned, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})
	if _, err := unassigned.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "night"}); err != engine.ErrGameNotReady {
		t.Fatalf("expected ErrGameNotReady with unassigned players, got %v", err)
	}
}

func TestAdditionalRoleActionDeniedBranches(t *testing.T) {
	executor := engine.DefaultRoleActionExecutor{}
	g := engine.NewGame()
	players := map[engine.PlayerID]*engine.Player{
		"villager": {ID: "villager", Role: roles.Villager{}, Alive: true},
		"flute":    {ID: "flute", Role: roles.FlutePlayer{}, Alive: true},
		"pyro":     {ID: "pyro", Role: roles.Pyromaniac{}, Alive: true},
		"savior":   {ID: "savior", Role: roles.Savior{}, Alive: true},
		"raven":    {ID: "raven", Role: roles.Raven{}, Alive: true},
	}
	state := engine.GameState{
		Players:       players,
		Protected:     map[engine.PlayerID]bool{},
		Charmed:       map[engine.PlayerID]bool{},
		Doused:        map[engine.PlayerID]bool{"missing": true},
		Revealed:      map[engine.PlayerID]bool{},
		VoteDisabled:  map[engine.PlayerID]bool{},
		VoteModifiers: map[engine.PlayerID]int{},
		Titles:        map[engine.TitleKind]engine.PlayerID{},
		AncientSaved:  map[engine.PlayerID]bool{},
	}
	if err := executor.Execute(g, state, "villager", engine.RoleActionSetLovers, []engine.PlayerID{"villager", "flute"}); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-Cupid lovers action, got %v", err)
	}
	if err := executor.Execute(g, state, "villager", engine.RoleActionCharmPlayers, []engine.PlayerID{"villager", "flute"}); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-Flute charm action, got %v", err)
	}
	if err := executor.Execute(g, state, "villager", engine.RoleActionDousePlayers, []engine.PlayerID{"flute"}); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-Pyro douse action, got %v", err)
	}
	if err := executor.Execute(g, state, "villager", engine.RoleActionProtectPlayer, []engine.PlayerID{"flute"}); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-Savior protect action, got %v", err)
	}
	if err := executor.Execute(g, state, "villager", engine.RoleActionMarkForCrows, []engine.PlayerID{"flute"}); err != engine.ErrRoleActionDenied {
		t.Fatalf("expected ErrRoleActionDenied for non-Raven crow action, got %v", err)
	}
	if err := executor.Execute(g, state, "pyro", engine.RoleActionIgnite, nil); err != engine.ErrPlayerUnknown {
		t.Fatalf("expected ErrPlayerUnknown when ignition reaches missing player, got %v", err)
	}
}
