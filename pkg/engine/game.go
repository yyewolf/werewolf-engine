package engine

// Game is an in-memory deterministic implementation of Engine.
type Game struct {
	players            map[PlayerID]*Player
	votes              map[PlayerID]PlayerID
	relationships      RelationshipStore
	protected          map[PlayerID]bool
	charmed            map[PlayerID]bool
	doused             map[PlayerID]bool
	revealed           map[PlayerID]bool
	voteDisabled       map[PlayerID]bool
	voteModifiers      map[PlayerID]int
	titles             map[TitleKind]PlayerID
	ancientSaved       map[PlayerID]bool
	deathSideEffects   []DeathSideEffectChecker
	winConditionChecks []WinConditionChecker
}

type GameOption func(*Game)

func WithDeathSideEffects(checkers ...DeathSideEffectChecker) GameOption {
	return func(g *Game) {
		g.deathSideEffects = checkers
	}
}

func WithWinConditionCheckers(checkers ...WinConditionChecker) GameOption {
	return func(g *Game) {
		g.winConditionChecks = checkers
	}
}

func NewGame(opts ...GameOption) *Game {
	g := &Game{
		players:       make(map[PlayerID]*Player),
		votes:         make(map[PlayerID]PlayerID),
		relationships: NewMemoryRelationshipStore(),
		protected:     make(map[PlayerID]bool),
		charmed:       make(map[PlayerID]bool),
		doused:        make(map[PlayerID]bool),
		revealed:      make(map[PlayerID]bool),
		voteDisabled:  make(map[PlayerID]bool),
		voteModifiers: make(map[PlayerID]int),
		titles:        make(map[TitleKind]PlayerID),
		ancientSaved:  make(map[PlayerID]bool),
		deathSideEffects: []DeathSideEffectChecker{
			LoverChainDeathChecker{},
		},
		winConditionChecks: []WinConditionChecker{
			FlutePlayerWinChecker{},
			PyromaniacWinChecker{},
			WhiteWolfWinChecker{},
			LoversWinChecker{},
			VillagersWinChecker{},
			WerewolvesWinChecker{},
		},
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Game) AddPlayer(id PlayerID) error {
	if _, exists := g.players[id]; exists {
		return ErrPlayerExists
	}
	g.players[id] = &Player{ID: id, Alive: true}
	return nil
}

func (g *Game) AssignRole(playerID PlayerID, role Role) error {
	player, ok := g.players[playerID]
	if !ok {
		return ErrPlayerUnknown
	}
	if role == nil {
		return ErrRoleMissing
	}
	player.Role = role
	return nil
}

func (g *Game) SetLovers(a PlayerID, b PlayerID) error {
	if a == b {
		return ErrInvalidLovers
	}
	if _, ok := g.players[a]; !ok {
		return ErrInvalidLovers
	}
	if _, ok := g.players[b]; !ok {
		return ErrInvalidLovers
	}
	g.relationships.SetLovers(&LoverPair{A: a, B: b})
	return nil
}

func (g *Game) CastVote(voter PlayerID, target PlayerID) error {
	voterPlayer, ok := g.players[voter]
	if !ok {
		return ErrPlayerUnknown
	}
	if !voterPlayer.Alive {
		return ErrPlayerDead
	}
	if g.voteDisabled[voter] {
		return ErrPhaseInvalid
	}
	targetPlayer, ok := g.players[target]
	if !ok {
		return ErrPlayerUnknown
	}
	if !targetPlayer.Alive {
		return ErrPlayerDead
	}
	g.votes[voter] = target
	return nil
}

func (g *Game) ResolveVotes() error {
	counts := map[PlayerID]int{}
	for voter, target := range g.votes {
		voterPlayer := g.players[voter]
		targetPlayer := g.players[target]
		if voterPlayer == nil || targetPlayer == nil || !voterPlayer.Alive || !targetPlayer.Alive {
			continue
		}
		weight := 1
		if g.titles[TitleCaptain] == voter {
			weight++
		}
		counts[target] += weight
	}
	for target, delta := range g.voteModifiers {
		counts[target] += delta
	}

	var selected PlayerID
	maxVotes := 0
	tied := false
	for target, count := range counts {
		if count > maxVotes {
			selected = target
			maxVotes = count
			tied = false
			continue
		}
		if count == maxVotes {
			tied = true
		}
	}

	clear(g.votes)
	clear(g.voteModifiers)
	if maxVotes == 0 || tied {
		if tied {
			if scapegoat := g.findAliveRole(RoleScapegoat); scapegoat != "" {
				return g.KillWithCause(scapegoat, CauseTie)
			}
		}
		return nil
	}
	if player := g.players[selected]; player != nil && player.Role != nil && player.Role.ID() == RoleVillageIdiot {
		g.revealed[selected] = true
		g.voteDisabled[selected] = true
		return nil
	}
	return g.KillWithCause(selected, CauseVote)
}

func (g *Game) OnTurnAdvanced() {
	clear(g.protected)
}

func (g *Game) Kill(playerID PlayerID) error {
	return g.KillWithCause(playerID, CauseDirect)
}

func (g *Game) KillWithCause(playerID PlayerID, cause EliminationCause) error {
	if _, ok := g.players[playerID]; !ok {
		return ErrPlayerUnknown
	}
	if cause == CauseAttack {
		if g.protected[playerID] {
			return nil
		}
		if player := g.players[playerID]; player != nil && player.Role != nil && player.Role.ID() == RoleAncien && !g.ancientSaved[playerID] {
			g.ancientSaved[playerID] = true
			return nil
		}
	}

	state := g.state()
	queue := []PlayerID{playerID}

	for len(queue) > 0 {
		dead := queue[0]
		queue = queue[1:]

		player := g.players[dead]
		if player == nil || !player.Alive {
			continue
		}
		player.Alive = false

		for _, checker := range g.deathSideEffects {
			nextVictims, err := checker.OnDeath(state, dead)
			if err != nil {
				return err
			}
			for _, victim := range nextVictims {
				if _, ok := g.players[victim]; ok {
					queue = append(queue, victim)
				}
			}
		}
	}

	return nil
}

func (g *Game) Outcome() Outcome {
	state := g.state()
	for _, checker := range g.winConditionChecks {
		if outcome, matched := checker.Check(state); matched {
			return outcome
		}
	}
	return Outcome{Ended: false, Kind: OutcomeNone, Reason: "game is still in progress"}
}

func (g *Game) state() GameState {
	return GameState{
		Players:       g.players,
		Lovers:        g.relationships.Lovers(),
		Votes:         g.votes,
		Protected:     g.protected,
		Charmed:       g.charmed,
		Doused:        g.doused,
		Revealed:      g.revealed,
		VoteDisabled:  g.voteDisabled,
		VoteModifiers: g.voteModifiers,
		Titles:        g.titles,
		AncientSaved:  g.ancientSaved,
	}
}

func (g *Game) findAliveRole(roleID RoleID) PlayerID {
	for id, player := range g.players {
		if player.Alive && player.Role != nil && player.Role.ID() == roleID {
			return id
		}
	}
	return ""
}

func (g *Game) SetProtected(playerID PlayerID, value bool) {
	g.protected[playerID] = value
}

func (g *Game) SetCharmed(playerID PlayerID, value bool) {
	g.charmed[playerID] = value
}

func (g *Game) SetDoused(playerID PlayerID, value bool) {
	g.doused[playerID] = value
}

func (g *Game) AddVoteModifier(playerID PlayerID, delta int) {
	g.voteModifiers[playerID] += delta
}

func (g *Game) AssignTitle(kind TitleKind, playerID PlayerID) {
	g.titles[kind] = playerID
}
