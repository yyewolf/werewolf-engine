# github.com/yyewolf/werewolf-engine
An open-source engine to run werewolf games.

## Terminal CLI

The repository includes a terminal-based game master CLI built with Charmbracelet's TUI libraries.

Run it with:

```bash
go run ./cmd/werewolf-cli
```

The terminal UI assumes the operator is the game master. It prompts for system composition, role assignment, phase changes, votes, eliminations, and supported role actions on behalf of the actual players.
