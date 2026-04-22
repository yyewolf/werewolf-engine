package engine

type TurnInfo struct {
	Index int
	Phase string
}

type TitleAssignment struct {
	Kind   TitleKind
	Holder PlayerID
}

type RelationshipKind string

const (
	RelationshipLovers RelationshipKind = "lovers"
)

type PlayerState struct {
	ID           PlayerID
	Role         RoleID
	Team         Team
	Alive        bool
	Protected    bool
	Charmed      bool
	Doused       bool
	Revealed     bool
	VoteDisabled bool
}

type Relationship struct {
	Kind RelationshipKind
	A    PlayerID
	B    PlayerID
}

type Vote struct {
	Voter  PlayerID
	Target PlayerID
}

type StateView struct {
	Players       []PlayerState
	Relationships []Relationship
	Titles        []TitleAssignment
	Votes         []Vote
}

type GameInfo struct {
	PlayerCount       int
	AliveCount        int
	QueuedRoleCount   int
	AssignedRoleCount int
	Outcome           Outcome
}

type Position struct {
	State StateView
	Turn  TurnInfo
	Info  GameInfo
}

type TransitionKind string

const (
	TransitionAddPlayer    TransitionKind = "add_player"
	TransitionAddRole      TransitionKind = "add_role"
	TransitionStartGame    TransitionKind = "start_game"
	TransitionAssignRole   TransitionKind = "assign_role"
	TransitionKillPlayer   TransitionKind = "kill_player"
	TransitionCastVote     TransitionKind = "cast_vote"
	TransitionResolveVotes TransitionKind = "resolve_votes"
	TransitionRoleAction   TransitionKind = "role_action"
	TransitionAdvanceTurn  TransitionKind = "advance_turn"
)

const (
	RoleActionChooseRole    = "choose_role"
	RoleActionSetLovers     = "set_lovers"
	RoleActionAttackPlayer  = "attack_player"
	RoleActionInspectPlayer = "inspect_player"
	RoleActionCharmPlayers  = "charm_players"
	RoleActionDousePlayers  = "douse_players"
	RoleActionIgnite        = "ignite"
	RoleActionProtectPlayer = "protect_player"
	RoleActionMarkForCrows  = "mark_for_crows"
	RoleActionAssignCaptain = "assign_captain"
)

type Transition struct {
	Kind      TransitionKind
	Target    PlayerID
	Actor     PlayerID
	Role      RoleID
	Action    string
	Targets   []PlayerID
	NextPhase string
	Cause     EliminationCause
}

type BootstrapPlayer struct {
	ID   PlayerID
	Role Role
}

type Bootstrap struct {
	Players     []BootstrapPlayer
	InitialTurn TurnInfo
}
