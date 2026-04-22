package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

func TestEngineHandlesLobbyStartAndRoleAssignment(t *testing.T) {
	e := mustNewEmptyEngine(t)
	if e.Turn().Phase != "lobby" {
		t.Fatalf("expected initial phase lobby, got %q", e.Turn().Phase)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleWerewolf})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleSeer})

	info := e.Info()
	if info.PlayerCount != 2 || info.QueuedRoleCount != 2 || info.AssignedRoleCount != 0 {
		t.Fatalf("unexpected lobby info: %+v", info)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	if e.Turn().Phase != "role_assignment" {
		t.Fatalf("expected phase role_assignment, got %q", e.Turn().Phase)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleWerewolf})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "bob", Role: engine.RoleSeer})

	info = e.Info()
	if info.QueuedRoleCount != 0 || info.AssignedRoleCount != 2 {
		t.Fatalf("unexpected assignment info: %+v", info)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "night"})
	if e.Turn().Phase != "night" {
		t.Fatalf("expected phase night, got %q", e.Turn().Phase)
	}
}

func TestEngineEmitsRoleAssignedEvent(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})

	found := false
	for _, event := range e.Events() {
		if event.Kind == engine.EventRoleAssigned && event.PlayerID == "alice" && event.RoleID == engine.RoleVillager {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected role_assigned event in event history")
	}
}

func TestEngineSubscribeReceivesPregameEvents(t *testing.T) {
	e := mustNewEmptyEngine(t)
	sub := e.Subscribe()
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})

	seenPre := false
	seenAdded := false
	seenPost := false
	for i := 0; i < 3; i++ {
		event := <-sub
		switch event.Kind {
		case engine.EventPreTransition:
			seenPre = true
		case engine.EventPlayerAdded:
			seenAdded = true
		case engine.EventPostTransition:
			seenPost = true
		}
	}
	if !seenPre || !seenAdded || !seenPost {
		t.Fatalf("expected pre/player_added/post events, got pre=%v added=%v post=%v", seenPre, seenAdded, seenPost)
	}
}

func TestEngineEmitsThiefActionRequestedDuringAssignment(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleThief})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleThief})

	found := false
	for _, event := range e.Events() {
		if event.Kind == engine.EventRoleActionRequested && event.PlayerID == "alice" && event.Action == engine.RoleActionChooseRole {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected thief role_action_requested event")
	}
}

func TestEngineE2ECupidActionRequiresPrompt(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleCupid})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "bob", Role: engine.RoleVillager})

	_, err := e.Apply(engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})
	if err != engine.ErrRoleActionDenied {
		t.Fatalf("expected Cupid action before prompt to fail with ErrRoleActionDenied, got %v", err)
	}

	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "charlie", Role: engine.RoleCupid})

	foundPrompt := false
	for _, event := range e.Events() {
		if event.Kind == engine.EventRoleActionRequested && event.PlayerID == "charlie" && event.RoleID == engine.RoleCupid && event.Action == engine.RoleActionSetLovers {
			foundPrompt = true
			break
		}
	}
	if !foundPrompt {
		t.Fatalf("expected cupid role_action_requested event")
	}

	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})

	relationships := e.State().Relationships
	if len(relationships) != 1 {
		t.Fatalf("expected one relationship after cupid action, got %d", len(relationships))
	}
	relationship := relationships[0]
	if relationship.Kind != engine.RelationshipLovers || relationship.A != "alice" || relationship.B != "bob" {
		t.Fatalf("unexpected lovers relationship: %+v", relationship)
	}

	_, err = e.Apply(engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})
	if err != engine.ErrRoleActionDenied {
		t.Fatalf("expected Cupid action after prompt consumption to fail with ErrRoleActionDenied, got %v", err)
	}
}

func TestEngineE2ECannotVoteDuringNight(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleWerewolf})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleSeer})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleWerewolf})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "bob", Role: engine.RoleSeer})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "night"})

	_, err := e.Apply(engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice", Target: "bob"})
	if err != engine.ErrPhaseInvalid {
		t.Fatalf("expected night vote to fail with ErrPhaseInvalid, got %v", err)
	}
	if len(e.State().Votes) != 0 {
		t.Fatalf("expected no votes to be recorded during night, got %d", len(e.State().Votes))
	}
}

func TestEngineE2ECupidCannotUseOtherRoleAction(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleCupid})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionStartGame})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "alice", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "bob", Role: engine.RoleVillager})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAssignRole, Target: "charlie", Role: engine.RoleCupid})

	_, err := e.Apply(engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionProtectPlayer,
		Targets: []engine.PlayerID{"alice"},
	})
	if err != engine.ErrRoleActionDenied {
		t.Fatalf("expected Cupid using another role action to fail with ErrRoleActionDenied, got %v", err)
	}
	if len(e.State().Relationships) != 0 {
		t.Fatalf("expected no lovers relationship after denied Cupid action, got %d", len(e.State().Relationships))
	}

	mustApply(t, e, engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   "charlie",
		Action:  engine.RoleActionSetLovers,
		Targets: []engine.PlayerID{"alice", "bob"},
	})

	relationships := e.State().Relationships
	if len(relationships) != 1 {
		t.Fatalf("expected Cupid prompt to remain usable after denied wrong action, got %d relationships", len(relationships))
	}
}

func TestEngineCannotStartWithoutEnoughQueuedRoles(t *testing.T) {
	e := mustNewEmptyEngine(t)
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAddPlayer, Target: "alice"})
	_, err := e.Apply(engine.Transition{Kind: engine.TransitionStartGame})
	if err != engine.ErrGameNotReady {
		t.Fatalf("expected ErrGameNotReady, got %v", err)
	}
}
