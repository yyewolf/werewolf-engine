package engine

type PlayerID string

type RoleID string

const (
	RoleVillager     RoleID = "villager"
	RoleWerewolf     RoleID = "werewolf"
	RoleSeer         RoleID = "seer"
	RoleWitch        RoleID = "witch"
	RoleHunter       RoleID = "hunter"
	RoleLittleGirl   RoleID = "little_girl"
	RoleCupid        RoleID = "cupid"
	RoleThief        RoleID = "thief"
	RoleAncien       RoleID = "ancien"
	RoleScapegoat    RoleID = "scapegoat"
	RoleVillageIdiot RoleID = "village_idiot"
	RoleFlutePlayer  RoleID = "flute_player"
	RoleSavior       RoleID = "savior"
	RoleRaven        RoleID = "raven"
	RoleWhiteWolf    RoleID = "white_wolf"
	RolePyromaniac   RoleID = "pyromaniac"
)

type Team string

const (
	TeamVillagers  Team = "villagers"
	TeamWerewolves Team = "werewolves"
)

type OutcomeKind string

const (
	OutcomeNone        OutcomeKind = "none"
	OutcomeVillagers   OutcomeKind = "villagers"
	OutcomeWerewolves  OutcomeKind = "werewolves"
	OutcomeLovers      OutcomeKind = "lovers"
	OutcomeFlutePlayer OutcomeKind = "flute_player"
	OutcomeWhiteWolf   OutcomeKind = "white_wolf"
	OutcomePyromaniac  OutcomeKind = "pyromaniac"
)

type TitleKind string

const (
	TitleCaptain TitleKind = "captain"
)

type EliminationCause string

const (
	CauseDirect   EliminationCause = "direct"
	CauseAttack   EliminationCause = "attack"
	CauseVote     EliminationCause = "vote"
	CauseTie      EliminationCause = "tie"
	CauseIgnition EliminationCause = "ignition"
)

type Outcome struct {
	Ended          bool
	Kind           OutcomeKind
	Reason         string
	WinningPlayers []PlayerID
}

type Role interface {
	ID() RoleID
	Team() Team
}

type Player struct {
	ID    PlayerID
	Role  Role
	Alive bool
}

type LoverPair struct {
	A PlayerID
	B PlayerID
}

type GameState struct {
	Players       map[PlayerID]*Player
	Lovers        *LoverPair
	Votes         map[PlayerID]PlayerID
	Protected     map[PlayerID]bool
	Charmed       map[PlayerID]bool
	Doused        map[PlayerID]bool
	Revealed      map[PlayerID]bool
	VoteDisabled  map[PlayerID]bool
	VoteModifiers map[PlayerID]int
	Titles        map[TitleKind]PlayerID
	AncientSaved  map[PlayerID]bool
}
