package cli

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yyewolf/werewolf-engine/pkg/engine"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Padding(0, 1)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	panelStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("84"))
	deadStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Bold(true)
	actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("154"))
	systemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true)
)

type prompt struct {
	actor            engine.PlayerID
	action           string
	targetCount      int
	collectedTargets []engine.PlayerID
	validTargets     []engine.PlayerID
}

type model struct {
	engine      engine.Engine
	input       textinput.Model
	logs        []string
	promptQueue []*prompt
	seenEvents  int
	width       int
	height      int
	errMsg      string
	infoMsg     string
}

func Run() error {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel() model {
	e, err := engine.NewPositionEngine(engine.Bootstrap{})
	if err != nil {
		panic(err)
	}
	input := textinput.New()
	input.Focus()
	input.Placeholder = "Type response or 'next' to continue"
	input.Prompt = "gm> "
	input.CharLimit = 256
	input.Width = 72
	return model{
		engine: e,
		input:  input,
		logs:   []string{"Welcome to Werewolf GM Terminal"},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.errMsg = ""
			m.infoMsg = ""
			m.input.SetValue("")
			return m, nil
		case "enter":
			input := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			if input == "" {
				return m, nil
			}
			m.logs = append(m.logs, "> "+input)
			m.errMsg = ""
			m.infoMsg = ""

			if len(m.promptQueue) > 0 {
				// We're waiting for a response to the current prompt
				if err := m.respondToPrompt(input); err != nil {
					m.errMsg = err.Error()
					m.logs = append(m.logs, "error: "+err.Error())
				}
			} else {
				// No pending prompt, handle general commands
				if err := m.handleCommand(input); err != nil {
					m.errMsg = err.Error()
					m.logs = append(m.logs, "error: "+err.Error())
				}
			}

			m.syncEvents()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) respondToPrompt(input string) error {
	if len(m.promptQueue) == 0 {
		return fmt.Errorf("no pending prompt")
	}
	current := m.promptQueue[0]

	// Validate the target player exists and is alive
	target := engine.PlayerID(input)
	pos := m.engine.Position()
	var targetPlayer *engine.PlayerState
	for i := range pos.State.Players {
		if pos.State.Players[i].ID == target {
			targetPlayer = &pos.State.Players[i]
			break
		}
	}
	if targetPlayer == nil {
		return fmt.Errorf("player %q does not exist", input)
	}
	if !targetPlayer.Alive {
		return fmt.Errorf("player %q is dead", input)
	}

	current.collectedTargets = append(current.collectedTargets, target)

	// Not enough targets yet — keep the prompt active
	if len(current.collectedTargets) < current.targetCount {
		m.infoMsg = fmt.Sprintf("target %d/%d selected", len(current.collectedTargets), current.targetCount)
		return nil
	}

	// All targets gathered — submit the action
	_, err := m.engine.Apply(engine.Transition{
		Kind:    engine.TransitionRoleAction,
		Actor:   current.actor,
		Action:  current.action,
		Targets: current.collectedTargets,
	})
	if err != nil {
		// Reset collected targets so the user can retry
		current.collectedTargets = nil
		return err
	}

	m.infoMsg = fmt.Sprintf("✓ %s completed %s", current.actor, current.action)
	m.promptQueue = m.promptQueue[1:]
	return nil
}

func (m *model) handleCommand(input string) error {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return nil
	}

	cmd := fields[0]
	pos := m.engine.Position()

	switch cmd {
	case "help":
		m.infoMsg = "Commands: players, roles, start, next, kill, resolve"
		return nil
	case "clear":
		m.logs = nil
		m.infoMsg = "Log cleared"
		return nil
	case "reset":
		e, err := engine.NewPositionEngine(engine.Bootstrap{})
		if err != nil {
			return err
		}
		m.engine = e
		m.seenEvents = 0
		m.promptQueue = nil
		m.logs = []string{"Game reset"}
		m.infoMsg = "Fresh game ready"
		return nil

	case "players":
		if len(fields) < 2 {
			return fmt.Errorf("usage: players <id> [id...]")
		}
		for _, name := range fields[1:] {
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionAddPlayer, Target: engine.PlayerID(name)}); err != nil {
				return err
			}
		}
		m.infoMsg = fmt.Sprintf("Added %d player(s)", len(fields)-1)
		return nil

	case "roles":
		if len(fields) < 2 {
			return fmt.Errorf("usage: roles <role> [role...]")
		}
		for _, role := range fields[1:] {
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionAddRole, Role: engine.RoleID(role)}); err != nil {
				return err
			}
		}
		m.infoMsg = fmt.Sprintf("Queued %d role(s)", len(fields)-1)
		return nil

	case "start":
		if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionStartGame}); err != nil {
			return err
		}

		// Auto-assign all roles
		pos := m.engine.Position()
		players := make([]engine.PlayerID, 0)
		for _, p := range pos.State.Players {
			players = append(players, p.ID)
		}

		// Collect queued roles
		events := m.engine.Events()
		roleQueue := map[engine.RoleID]int{}
		for _, e := range events {
			if e.Kind == engine.EventRoleQueued {
				roleQueue[e.RoleID]++
			}
		}

		roleList := make([]engine.RoleID, 0)
		for role, count := range roleQueue {
			for i := 0; i < count; i++ {
				roleList = append(roleList, role)
			}
		}

		// Shuffle
		rand.Shuffle(len(players), func(i, j int) { players[i], players[j] = players[j], players[i] })
		rand.Shuffle(len(roleList), func(i, j int) { roleList[i], roleList[j] = roleList[j], roleList[i] })

		// Assign
		for i, player := range players {
			if i >= len(roleList) {
				break
			}
			if _, err := m.engine.Apply(engine.Transition{
				Kind:   engine.TransitionAssignRole,
				Target: player,
				Role:   roleList[i],
			}); err != nil {
				return err
			}
		}

		m.infoMsg = "Game started, roles assigned"
		return nil

	case "next":
		if len(m.promptQueue) > 0 {
			m.infoMsg = fmt.Sprintf("waiting for %d prompt(s) to be answered", len(m.promptQueue))
			return nil
		}

		// Auto-advance turn if in role_assignment and no pending prompts
		if pos.Turn.Phase == "role_assignment" {
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn}); err != nil {
				return err
			}
			m.infoMsg = "Advanced to next phase"
			return nil
		}

		// Advance out of night when all prompts are answered
		if pos.Turn.Phase == "night" {
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn}); err != nil {
				return err
			}
			m.infoMsg = "Advanced to day vote"
			return nil
		}

		// If day_vote phase, resolve votes and advance
		if pos.Turn.Phase == "day_vote" {
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionResolveVotes}); err != nil {
				return err
			}
			if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionAdvanceTurn}); err != nil {
				return err
			}
			m.infoMsg = "Votes resolved, advancing"
			return nil
		}

		m.infoMsg = "Cannot advance now"
		return nil

	case "kill":
		if len(fields) < 2 {
			return fmt.Errorf("usage: kill <player>")
		}
		cause := engine.CauseDirect
		if len(fields) > 2 {
			cause = engine.EliminationCause(fields[2])
		}
		if _, err := m.engine.Apply(engine.Transition{
			Kind:   engine.TransitionKillPlayer,
			Target: engine.PlayerID(fields[1]),
			Cause:  cause,
		}); err != nil {
			return err
		}
		m.infoMsg = fmt.Sprintf("Killed %s", fields[1])
		return nil

	case "resolve":
		if _, err := m.engine.Apply(engine.Transition{Kind: engine.TransitionResolveVotes}); err != nil {
			return err
		}
		m.infoMsg = "Votes resolved"
		return nil

	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func (m *model) syncEvents() {
	events := m.engine.Events()
	for _, event := range events[m.seenEvents:] {
		if event.Kind == engine.EventRoleActionRequested {
			pos := m.engine.Position()
			var validTargets []engine.PlayerID
			for _, p := range pos.State.Players {
				if p.Alive && p.ID != event.PlayerID {
					validTargets = append(validTargets, p.ID)
				}
			}
			m.promptQueue = append(m.promptQueue, &prompt{
				actor:        event.PlayerID,
				action:       event.Action,
				targetCount:  event.TargetCount,
				validTargets: validTargets,
			})
		}
		m.logs = append(m.logs, formatEvent(event))
	}
	m.seenEvents = len(events)
}

func (m model) View() string {
	status := m.renderStatus()
	control := m.renderControls()
	footer := m.renderFooter()

	leftWidth := max(50, min(80, m.width/2))
	rightWidth := max(40, m.width-leftWidth-3)
	left := panelStyle.Width(leftWidth).Render(status)
	right := panelStyle.Width(rightWidth).Render(control)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Werewolf GM Terminal"),
		body,
		footer,
	)
}

func (m model) renderStatus() string {
	pos := m.engine.Position()
	var lines []string

	lines = append(lines, headerStyle.Render("Game State"))
	lines = append(lines, fmt.Sprintf("Phase: %s  Turn: %d", pos.Turn.Phase, pos.Turn.Index))
	lines = append(lines, fmt.Sprintf("Players: %d alive  (%d total)", pos.Info.AliveCount, pos.Info.PlayerCount))

	if pos.Info.Outcome.Ended {
		lines = append(lines, okStyle.Render(fmt.Sprintf("🎉 %s WIN", pos.Info.Outcome.Kind)))
		lines = append(lines, mutedStyle.Render(pos.Info.Outcome.Reason))
	}

	lines = append(lines, "")
	lines = append(lines, headerStyle.Render("Players"))
	for _, p := range pos.State.Players {
		status := "●"
		if !p.Alive {
			status = "○"
		}
		flags := playerFlags(p)
		flagStr := ""
		if flags != "" {
			flagStr = " [" + flags + "]"
		}
		line := fmt.Sprintf("%s %s  %s (%s)", status, p.ID, p.Role, p.Team)
		if flagStr != "" {
			line += flagStr
		}
		if !p.Alive {
			line = deadStyle.Render(line)
		}
		lines = append(lines, line)
	}

	if len(pos.State.Votes) > 0 {
		lines = append(lines, "")
		lines = append(lines, headerStyle.Render("Votes (day)"))
		for _, vote := range pos.State.Votes {
			lines = append(lines, fmt.Sprintf("%s → %s", vote.Voter, vote.Target))
		}
	}

	return strings.Join(lines, "\n")
}

func (m model) renderControls() string {
	pos := m.engine.Position()
	var lines []string

	if len(m.promptQueue) > 0 {
		current := m.promptQueue[0]
		queued := len(m.promptQueue) - 1
		header := "▶ PROMPT"
		if queued > 0 {
			header = fmt.Sprintf("▶ PROMPT  (+%d queued)", queued)
		}
		lines = append(lines, promptStyle.Render(header))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%s: %s", current.actor, current.action))
		if current.targetCount > 1 {
			lines = append(lines, fmt.Sprintf("select target %d/%d", len(current.collectedTargets)+1, current.targetCount))
		}
		lines = append(lines, "")
		lines = append(lines, headerStyle.Render("Valid targets:"))
		for _, t := range current.validTargets {
			lines = append(lines, actionStyle.Render("  → "+string(t)))
		}
	} else {
		lines = append(lines, headerStyle.Render("Available Actions"))
		lines = append(lines, "")

		switch pos.Turn.Phase {
		case "lobby":
			lines = append(lines, "Setup phase:")
			lines = append(lines, "  players <id> [id...]")
			lines = append(lines, "  roles <role> [role...]")
			lines = append(lines, "  start")
		case "role_assignment":
			lines = append(lines, "Roles assigned.")
			lines = append(lines, "  next  (to begin night)")
		case "night":
			lines = append(lines, "Waiting for role actions...")
			lines = append(lines, "  next  (advance phase)")
		case "day_vote":
			lines = append(lines, "Day voting phase.")
			lines = append(lines, "  next  (resolve votes)")
		default:
			lines = append(lines, "  next  (advance)")
		}

		lines = append(lines, "")
		lines = append(lines, mutedStyle.Render("reset | clear | help"))
	}

	lines = append(lines, "")
	lines = append(lines, headerStyle.Render("Event Log"))
	for _, log := range tail(m.logs, 8) {
		lines = append(lines, log)
	}

	return strings.Join(lines, "\n")
}

func (m model) renderFooter() string {
	status := mutedStyle.Render("↵ enter | esc clear | q quit")
	msg := ""
	if m.errMsg != "" {
		msg = errorStyle.Render(m.errMsg)
	} else if m.infoMsg != "" {
		msg = okStyle.Render(m.infoMsg)
	}
	return panelStyle.Width(max(80, m.width-2)).Render(
		status + "\n" + msg + "\n\n" + m.input.View(),
	)
}

func playerFlags(p engine.PlayerState) string {
	flags := []string{}
	if p.Protected {
		flags = append(flags, "🛡️")
	}
	if p.Charmed {
		flags = append(flags, "💕")
	}
	if p.Doused {
		flags = append(flags, "🔥")
	}
	if p.Revealed {
		flags = append(flags, "👁️")
	}
	if p.VoteDisabled {
		flags = append(flags, "🤐")
	}
	return strings.Join(flags, "")
}

func formatEvent(event engine.Event) string {
	switch event.Kind {
	case engine.EventPlayerAdded:
		return systemStyle.Render(fmt.Sprintf("✓ Player %s joined", event.PlayerID))
	case engine.EventRoleQueued:
		return systemStyle.Render(fmt.Sprintf("✓ Queued %s", event.RoleID))
	case engine.EventGameStarted:
		return systemStyle.Render("▶ Game started")
	case engine.EventRoleAssigned:
		return systemStyle.Render(fmt.Sprintf("✓ %s assigned %s", event.PlayerID, event.RoleID))
	case engine.EventRoleActionRequested:
		return promptStyle.Render(fmt.Sprintf("? %s: %s ?", event.PlayerID, event.Action))
	case engine.EventPlayerInspected:
		return okStyle.Render(fmt.Sprintf("🔍 seer: %s is %s", event.PlayerID, event.RoleID))
	case engine.EventPostTransition:
		if event.Phase != "" {
			return systemStyle.Render(fmt.Sprintf("↻ Entering %s", event.Phase))
		}
	}
	if event.Transition != "" {
		return fmt.Sprintf("→ %s", event.Transition)
	}
	return fmt.Sprintf("◇ %s", event.Kind)
}

func tail(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
