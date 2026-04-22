# RULES.md

## Scope
This document defines the baseline rules implemented by the engine.

Current baseline:
- **Classic Werewolf (Loup-Garou de Thiercelieux)** role set only
- No expansion roles in v1 baseline
- Additional supported roles/titles currently implemented in engine tests: Ancien, Bouc Emissaire, Idiot du Village, Joueur de Flute, Salvateur, Corbeau, Loup Blanc, Pyromane, Capitaine

When behavior changes, update this document in the same change set.

## Game Setup
- Minimum players: 5
- Recommended players: 8 to 18
- Each player is assigned exactly one hidden role.
- Roles are distributed according to a role configuration chosen at game creation.
- The game starts in `Night` phase.

## Baseline Roles
Baseline role IDs and teams:
- `villager`: Villager team
- `werewolf`: Werewolf team
- `seer`: Villager team
- `witch`: Villager team
- `hunter`: Villager team
- `little_girl`: Villager team (optional in config)
- `cupid`: Villager team
- `thief`: Villager team
- `ancien`: Villager team
- `scapegoat`: Villager team
- `village_idiot`: Villager team
- `flute_player`: Villager-aligned role with independent win condition
- `savior`: Villager team
- `raven`: Villager team
- `white_wolf`: Werewolf-aligned role with independent win condition
- `pyromaniac`: Villager-aligned role with independent win condition

Supported title IDs:
- `captain`: title, not a role

Notes:
- Optional roles can be disabled by configuration.
- Unknown or disabled roles must be rejected during game setup.

## Phase Model
The baseline loop is:
1. `Night`
2. `DayDiscussion`
3. `DayVote`
4. `Resolution`
5. Repeat until a win condition is met, then `Ended`

Engine constraints:
- Only actions legal for the current phase are accepted.
- Phase progression is explicit through engine commands.
- Timers are not managed by the engine; caller decides when to advance.

Pre-night setup actions:
- `thief` replacement choice occurs before the first `Night` resolution.
- `cupid` lover selection occurs on first night only.

## Night Actions
Action resolution order must be deterministic:
1. First-night setup effects (cupid selection) if not already resolved.
2. Werewolves choose one victim collectively.
3. Seer inspects one player.
4. Witch may use heal poison powers according to remaining charges.
5. Deferred death effects (for example hunter shot) are resolved by role rules.

Conflict rules (baseline):
- If witch heals the werewolf target, that target survives the night kill.
- If multiple effects target the same player, deterministic priority decides final outcome.
- A dead player cannot perform new actions unless role rules explicitly allow it.
- Lover link: when one lover dies, the linked lover dies immediately as a chained effect.

## Day Actions
- During `DayDiscussion`, no lethal vote is resolved.
- During `DayVote`, each alive player can cast one vote.
- If a strict majority or configured vote rule selects a target, that player is eliminated.
- Tie behavior must follow configured policy (default: no elimination on tie).

## Role Behavior (Baseline)

### Villager
- No active night power.
- Participates in day vote.

### Werewolf
- Participates in werewolf night kill decision.
- Counts toward werewolf win condition.

### Seer
- Once per night may inspect one alive player.
- Inspection reveals team alignment in baseline rules.

### Witch
- Starts with one heal and one poison charge.
- Heal and poison can each be used at most once per game.
- Poison targets one alive player at night.

### Hunter
- On death, may execute one final shot action if enabled by config.
- Final shot timing and legality must be deterministic and tested.

### Little Girl (Optional)
- Included as a configurable classic role.
- Baseline behavior is passive unless explicit action rules are enabled by config.

### Cupid
- Acts only once at the beginning of the game (first night setup).
- Chooses two alive players as lovers.
- If one lover dies, the other dies immediately (chain death).

### Thief
- At setup, may swap to one of two extra configured roles if available.
- If no swap is performed, keeps current assigned role according to config policy.
- Swap rules and visibility must be deterministic and covered by tests.

### Ancien
- Survives the first attack-type elimination.
- Dies normally on subsequent attack-type eliminations.

### Bouc Emissaire
- If the village vote ends in a tie, the scapegoat is eliminated.

### Idiot du Village
- If chosen by vote, survives the vote elimination.
- Is revealed and loses the right to vote afterwards.

### Joueur de Flute
- Can charm other players via role action.
- Wins independently when all other alive players are charmed.

### Salvateur
- Can protect one player against attack-type elimination for the current turn window.

### Corbeau
- Can mark a player, adding two virtual votes against that player in the next vote resolution.

### Loup Blanc
- Wins independently if it is the only surviving player.

### Pyromane
- Can douse players and later ignite them.
- Wins independently if it becomes the only surviving player.

### Capitaine
- Title, not a role.
- Captain's vote counts double.

## Win Conditions
The game ends immediately when one condition is met:
- Villagers win: no alive werewolves remain.
- Werewolves win: alive werewolves are greater than or equal to alive non-werewolves.
- Lovers special win: if surviving players are exactly the lover pair, they win together.
- Flute player win: a living flute player has charmed all other alive players.
- White wolf win: the white wolf is the only survivor.
- Pyromaniac win: the pyromaniac is the only survivor.

Additional constraints:
- Win checks occur after each lethal resolution step.
- If multiple end states are theoretically reached in one resolution cycle, deterministic priority decides the final outcome.

Default end-state priority:
1. Flute player independent win
2. Pyromaniac independent win
3. White wolf independent win
4. Lovers special win (when exact mixed-team lover pair survives)
5. Villager or werewolf standard team wins

## Determinism Requirements
- Random events (role assignment, tie breaks if enabled) must use injected seeded RNG.
- Same seed + same command sequence must produce the same events and final state.
- Engine behavior must not depend on wall-clock timing.

## Validation Rules
- Invalid actions return typed validation errors.
- Out-of-phase actions are rejected.
- Duplicate commands with same idempotency key (if enabled) must not duplicate effects.
- Commands for unknown game/player IDs must return not-found errors.

## Testing Requirements For Rules
Each rule in this file must have test coverage via:
- Unit tests for role/action legality and transitions
- Scenario tests for full game flow
- Determinism tests comparing repeated seeded runs
- Tie and edge-case tests for vote/action conflicts

Coverage target:
- 100% in core engine/rules packages for implemented baseline behavior

## Versioning
- This file represents implemented behavior, not aspirational behavior.
- Any added/changed rule must update this file and corresponding tests in the same pull request.
