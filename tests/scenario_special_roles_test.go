package tests

import (
	"testing"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
	"github.com/yyewolf/werewolf-engine/pkg/engine/roles"
)

func TestScenarioAncienSurvivesFirstAttackButNotSecond(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Ancien{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice", Cause: engine.CauseAttack})
	if !findPlayerState(t, e, "alice").Alive {
		t.Fatalf("ancien should survive first attack")
	}
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "alice", Cause: engine.CauseAttack})
	if findPlayerState(t, e, "alice").Alive {
		t.Fatalf("ancien should die on second attack")
	}
}

func TestScenarioScapegoatDiesOnVoteTie(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Scapegoat{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "charlie", Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	if findPlayerState(t, e, "alice").Alive {
		t.Fatalf("scapegoat should die on tie")
	}
}

func TestScenarioVillageIdiotSurvivesVoteAndLosesVoteRight(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.VillageIdiot{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "charlie", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	player := findPlayerState(t, e, "alice")
	if !player.Alive || !player.Revealed || !player.VoteDisabled {
		t.Fatalf("village idiot should survive, be revealed, and lose voting right")
	}
}

func TestScenarioSaviorProtectsAgainstAttack(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Savior{}},
		{ID: "bob", Role: roles.Villager{}},
		{ID: "charlie", Role: roles.Werewolf{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionProtectPlayer, Targets: []engine.PlayerID{"bob"}})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionKillPlayer, Target: "bob", Cause: engine.CauseAttack})
	if !findPlayerState(t, e, "bob").Alive {
		t.Fatalf("protected player should survive attack")
	}
}

func TestScenarioRavenAddsTwoVotesToTarget(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Raven{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionMarkForCrows, Targets: []engine.PlayerID{"bob"}})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "charlie", Target: "charlie"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	if findPlayerState(t, e, "bob").Alive {
		t.Fatalf("raven-marked target should be eliminated from +2 vote modifier")
	}
}

func TestScenarioCaptainVoteBreaksTie(t *testing.T) {
	e := mustNewEngine(t, []engine.BootstrapPlayer{
		{ID: "alice", Role: roles.Villager{}},
		{ID: "bob", Role: roles.Werewolf{}},
		{ID: "charlie", Role: roles.Villager{}},
	})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionRoleAction, Actor: "alice", Action: engine.RoleActionAssignCaptain, Targets: []engine.PlayerID{"alice"}})
	assertTitleHolder(t, e, engine.TitleCaptain, "alice")
	mustApply(t, e, engine.Transition{Kind: engine.TransitionAdvanceTurn, NextPhase: "day_vote"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "alice", Target: "bob"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionCastVote, Actor: "bob", Target: "alice"})
	mustApply(t, e, engine.Transition{Kind: engine.TransitionResolveVotes})
	if findPlayerState(t, e, "bob").Alive {
		t.Fatalf("captain double vote should break the tie against bob")
	}
}
