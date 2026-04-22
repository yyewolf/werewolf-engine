package engine

import "sort"

type positionEngine struct {
	game           *Game
	turn           TurnInfo
	actionRunner   RoleActionExecutor
	queuedRoles    map[RoleID]int
	pendingActions map[PlayerID]string
	events         []Event
	subscribers    []chan Event
}

func NewPositionEngine(cfg Bootstrap, opts ...GameOption) (Engine, error) {
	g := NewGame(opts...)
	queuedRoles := map[RoleID]int{}
	for _, p := range cfg.Players {
		if err := g.AddPlayer(p.ID); err != nil {
			return nil, err
		}
		if p.Role != nil {
			g.players[p.ID].Role = p.Role
		}
	}
	turn := cfg.InitialTurn
	if turn.Phase == "" {
		if len(cfg.Players) == 0 {
			turn.Phase = "lobby"
		} else if hasUnassignedPlayers(g) {
			turn.Phase = "role_assignment"
		} else {
			turn.Phase = "night"
		}
	}
	return &positionEngine{
		game:           g,
		turn:           turn,
		actionRunner:   DefaultRoleActionExecutor{},
		queuedRoles:    queuedRoles,
		pendingActions: map[PlayerID]string{},
	}, nil
}

func (e *positionEngine) Apply(transition Transition) (Position, error) {
	// Prevent transitions after game has ended (except during setup phases)
	if e.turn.Phase != "lobby" && e.turn.Phase != "role_assignment" && e.game.Outcome().Ended {
		return Position{}, ErrGameEnded
	}

	e.emit(Event{Kind: EventPreTransition, Transition: transition.Kind, Phase: e.turn.Phase, PlayerID: transition.Target, RoleID: transition.Role, Action: transition.Action})
	switch transition.Kind {
	case TransitionAddPlayer:
		if e.turn.Phase != "lobby" || transition.Target == "" {
			return Position{}, ErrPhaseInvalid
		}
		if err := e.game.AddPlayer(transition.Target); err != nil {
			return Position{}, err
		}
		e.emit(Event{Kind: EventPlayerAdded, PlayerID: transition.Target, Phase: e.turn.Phase})
	case TransitionAddRole:
		if e.turn.Phase != "lobby" || transition.Role == "" {
			return Position{}, ErrPhaseInvalid
		}
		e.queuedRoles[transition.Role]++
		e.emit(Event{Kind: EventRoleQueued, RoleID: transition.Role, Phase: e.turn.Phase})
	case TransitionStartGame:
		if e.turn.Phase != "lobby" {
			return Position{}, ErrPhaseInvalid
		}
		if len(e.game.players) == 0 || e.totalQueuedRoles() < len(e.game.players) {
			return Position{}, ErrGameNotReady
		}
		e.turn.Phase = "role_assignment"
		e.emit(Event{Kind: EventGameStarted, Phase: e.turn.Phase})
	case TransitionAssignRole:
		if e.turn.Phase != "role_assignment" || transition.Target == "" || transition.Role == "" {
			return Position{}, ErrPhaseInvalid
		}
		if e.queuedRoles[transition.Role] == 0 {
			return Position{}, ErrRoleUnavailable
		}
		role, err := newRoleByID(transition.Role)
		if err != nil {
			return Position{}, err
		}
		if err := e.game.AssignRole(transition.Target, role); err != nil {
			return Position{}, err
		}
		e.queuedRoles[transition.Role]--
		e.emit(Event{Kind: EventRoleAssigned, PlayerID: transition.Target, RoleID: transition.Role, Phase: e.turn.Phase})
		if transition.Role == RoleCupid {
			e.requestRoleAction(transition.Target, transition.Role, RoleActionSetLovers)
		}
		if transition.Role == RoleThief {
			e.requestRoleAction(transition.Target, transition.Role, RoleActionChooseRole)
		}
	case TransitionKillPlayer:
		if transition.Target == "" {
			return Position{}, ErrTransitionInvalid
		}
		cause := transition.Cause
		if cause == "" {
			cause = CauseDirect
		}
		if err := e.game.KillWithCause(transition.Target, cause); err != nil {
			return Position{}, err
		}
	case TransitionCastVote:
		if e.turn.Phase != "day_vote" {
			return Position{}, ErrPhaseInvalid
		}
		if transition.Actor == "" || transition.Target == "" {
			return Position{}, ErrTransitionInvalid
		}
		if err := e.game.CastVote(transition.Actor, transition.Target); err != nil {
			return Position{}, err
		}
	case TransitionResolveVotes:
		if e.turn.Phase != "day_vote" {
			return Position{}, ErrPhaseInvalid
		}
		if err := e.game.ResolveVotes(); err != nil {
			return Position{}, err
		}
	case TransitionRoleAction:
		if transition.Actor == "" || transition.Action == "" {
			return Position{}, ErrTransitionInvalid
		}
		if e.turn.Phase == "role_assignment" {
			if err := e.consumePendingRoleAction(transition.Actor, transition.Action); err != nil {
				return Position{}, err
			}
		}
		if err := e.actionRunner.Execute(e.game, e.game.state(), transition.Actor, transition.Action, transition.Targets); err != nil {
			return Position{}, err
		}
		if transition.Action == RoleActionInspectPlayer && len(transition.Targets) == 1 {
			if p, ok := e.game.players[transition.Targets[0]]; ok && p.Role != nil {
				e.emit(Event{Kind: EventPlayerInspected, PlayerID: transition.Targets[0], RoleID: p.Role.ID(), Phase: e.turn.Phase})
			}
		}
	case TransitionAdvanceTurn:
		if e.turn.Phase == "role_assignment" && hasUnassignedPlayers(e.game) {
			return Position{}, ErrGameNotReady
		}
		e.game.OnTurnAdvanced()
		e.turn.Index++
		if transition.NextPhase != "" {
			e.turn.Phase = transition.NextPhase
		} else {
			e.turn.Phase = defaultNextPhase(e.turn.Phase)
		}
		if e.turn.Phase == "night" {
			e.emitNightPrompts()
		}
	default:
		return Position{}, ErrTransitionUnknown
	}
	position := e.Position()
	e.emit(Event{Kind: EventPostTransition, Transition: transition.Kind, Phase: e.turn.Phase, PlayerID: transition.Target, RoleID: transition.Role, Action: transition.Action})
	return position, nil
}

func (e *positionEngine) Position() Position {
	state := e.State()
	return Position{
		State: state,
		Turn:  e.turn,
		Info:  e.buildInfo(state),
	}
}

func (e *positionEngine) State() StateView {
	state := e.game.state()
	players := make([]PlayerState, 0, len(state.Players))
	for _, p := range state.Players {
		var roleID RoleID
		var team Team
		if p.Role != nil {
			roleID = p.Role.ID()
			team = p.Role.Team()
		}
		players = append(players, PlayerState{
			ID:           p.ID,
			Role:         roleID,
			Team:         team,
			Alive:        p.Alive,
			Protected:    state.Protected[p.ID],
			Charmed:      state.Charmed[p.ID],
			Doused:       state.Doused[p.ID],
			Revealed:     state.Revealed[p.ID],
			VoteDisabled: state.VoteDisabled[p.ID],
		})
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })

	relationships := []Relationship{}
	if state.Lovers != nil {
		relationships = append(relationships, Relationship{
			Kind: RelationshipLovers,
			A:    state.Lovers.A,
			B:    state.Lovers.B,
		})
	}
	votes := make([]Vote, 0, len(state.Votes))
	for voter, target := range state.Votes {
		votes = append(votes, Vote{Voter: voter, Target: target})
	}
	sort.Slice(votes, func(i, j int) bool { return votes[i].Voter < votes[j].Voter })
	titles := make([]TitleAssignment, 0, len(state.Titles))
	for kind, holder := range state.Titles {
		titles = append(titles, TitleAssignment{Kind: kind, Holder: holder})
	}
	sort.Slice(titles, func(i, j int) bool { return titles[i].Kind < titles[j].Kind })
	return StateView{Players: players, Relationships: relationships, Titles: titles, Votes: votes}
}

func (e *positionEngine) Turn() TurnInfo {
	return e.turn
}

func (e *positionEngine) Info() GameInfo {
	return e.buildInfo(e.State())
}

func (e *positionEngine) Events() []Event {
	return append([]Event(nil), e.events...)
}

func (e *positionEngine) Subscribe() <-chan Event {
	ch := make(chan Event, 128)
	e.subscribers = append(e.subscribers, ch)
	return ch
}

func (e *positionEngine) requestRoleAction(playerID PlayerID, roleID RoleID, action string) {
	e.pendingActions[playerID] = action
	e.emitRoleActionRequest(playerID, roleID, action)
}

func (e *positionEngine) emitRoleActionRequest(playerID PlayerID, roleID RoleID, action string) {
	e.emit(Event{Kind: EventRoleActionRequested, PlayerID: playerID, RoleID: roleID, Phase: e.turn.Phase, Action: action, TargetCount: roleActionTargetCount(action)})
}

func (e *positionEngine) emitNightPrompts() {
	for _, player := range e.game.players {
		if !player.Alive || player.Role == nil {
			continue
		}
		action := nightActionForRole(player.Role.ID())
		if action == "" {
			continue
		}
		e.emitRoleActionRequest(player.ID, player.Role.ID(), action)
	}
}

func nightActionForRole(roleID RoleID) string {
	switch roleID {
	case RoleWerewolf, RoleWhiteWolf:
		return RoleActionAttackPlayer
	case RoleSeer:
		return RoleActionInspectPlayer
	case RoleSavior:
		return RoleActionProtectPlayer
	case RoleRaven:
		return RoleActionMarkForCrows
	case RoleFlutePlayer:
		return RoleActionCharmPlayers
	case RolePyromaniac:
		return RoleActionDousePlayers
	default:
		return ""
	}
}

func roleActionTargetCount(action string) int {
	switch action {
	case RoleActionSetLovers:
		return 2
	default:
		return 1
	}
}

func (e *positionEngine) consumePendingRoleAction(actor PlayerID, action string) error {
	pendingAction, ok := e.pendingActions[actor]
	if !ok || pendingAction != action {
		return ErrRoleActionDenied
	}
	delete(e.pendingActions, actor)
	return nil
}

func (e *positionEngine) buildInfo(state StateView) GameInfo {
	alive := 0
	for _, p := range state.Players {
		if p.Alive {
			alive++
		}
	}
	return GameInfo{
		PlayerCount:       len(state.Players),
		AliveCount:        alive,
		QueuedRoleCount:   e.totalQueuedRoles(),
		AssignedRoleCount: e.assignedRoleCount(),
		Outcome:           e.game.Outcome(),
	}
}

func (e *positionEngine) totalQueuedRoles() int {
	total := 0
	for _, count := range e.queuedRoles {
		total += count
	}
	return total
}

func (e *positionEngine) assignedRoleCount() int {
	count := 0
	for _, player := range e.game.players {
		if player.Role != nil {
			count++
		}
	}
	return count
}

func (e *positionEngine) emit(event Event) {
	e.events = append(e.events, event)
	for _, subscriber := range e.subscribers {
		subscriber <- event
	}
}

// defaultNextPhase returns the standard next phase for a given phase.
func defaultNextPhase(phase string) string {
	switch phase {
	case "role_assignment":
		return "night"
	case "night":
		return "day_vote"
	case "day_vote":
		return "night"
	default:
		return phase
	}
}

func hasUnassignedPlayers(g *Game) bool {
	for _, player := range g.players {
		if player.Role == nil {
			return true
		}
	}
	return false
}
