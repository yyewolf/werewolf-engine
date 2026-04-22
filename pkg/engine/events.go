package engine

type EventKind string

const (
	EventPreTransition       EventKind = "pre_transition"
	EventPostTransition      EventKind = "post_transition"
	EventPlayerAdded         EventKind = "player_added"
	EventRoleQueued          EventKind = "role_queued"
	EventGameStarted         EventKind = "game_started"
	EventRoleAssigned        EventKind = "role_assigned"
	EventRoleActionRequested EventKind = "role_action_requested"
	EventPlayerInspected     EventKind = "player_inspected"
)

type Event struct {
	Kind        EventKind
	Transition  TransitionKind
	PlayerID    PlayerID
	RoleID      RoleID
	Phase       string
	Action      string
	TargetCount int // number of player targets the prompted action requires
}
