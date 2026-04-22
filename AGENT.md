# AGENT.md

## Mission
Build a production-grade **Werewolf (Loup-Garou de Thiercelieux)** game engine in Go that can be used in two ways:
- As a **Go library** (embedded in another app)
- As a **gRPC service** (remote orchestration)

The core game logic must be deterministic, modular, and fully validated by tests.

Delivery priority for v1:
- Build and stabilize the **library-first engine** first.
- Keep gRPC support as a later adapter once core behavior is locked.

## Product Requirements
- Language: **Go** (stable toolchain)
- Interfaces:
  - Library API (`pkg/`): direct in-process usage (v1 focus)
  - gRPC API (`api/proto`, `pkg/server/grpc`): remote usage (deferred until core is stable)
- Architecture:
  - Baseline gameplay scope: **classic role set** only
  - Classic baseline includes at minimum: villager, werewolf, seer, witch, hunter, little_girl, cupid, thief
  - Modular role system (easy to add/compose custom roles)
  - Clean separation of core domain logic and transport layers
  - Storage abstraction with pluggable backends; in-memory backend as default
- Quality:
  - 100% test coverage target for core engine package(s)
  - High-confidence scenario and property tests for full game flow

## Core Engineering Principles
1. **Core-first**: The domain engine is transport-agnostic and contains no gRPC-specific code.
2. **Deterministic execution**: Inject RNG/time and avoid hidden global state. Random behavior must be reproducible from explicit seeds.
3. **Event-driven state transitions**: State changes are explicit and traceable through events.
4. **Strong typing over stringly logic**: Use enums/typed IDs for phases, roles, actions, and outcomes.
5. **Extensibility without rewrites**: New roles or rules should be plug-in style.
6. **Reproducible tests**: Every scenario can be replayed with a fixed seed.

## Suggested Repository Layout

```
/api/proto/                # protobuf contracts
/cmd/werewolfd/            # gRPC server binary entrypoint
/internal/                 # internal-only adapters and wiring
/pkg/engine/               # core domain engine (public library)
/pkg/engine/roles/         # built-in role modules and role registry
/pkg/engine/rules/         # win conditions, phase rules, validators
/pkg/storage/              # storage interfaces and backend implementations
/pkg/storage/memory/       # default in-memory storage backend
/pkg/server/grpc/          # gRPC handlers mapping proto <-> engine
/pkg/testkit/              # test helpers, scenario builders, fakes
/RULES.md                  # authoritative implemented game rules
```

Notes:
- Keep domain model in `pkg/engine` usable with zero network stack.
- gRPC package must call engine APIs; never duplicate game logic there.
- Timers and deadlines are caller-managed; core engine validates phase/action legality only.

## Domain Model Guidelines
- `GameID`, `PlayerID`, `RoleID`, `ActionID` as typed values.
- Explicit phases, e.g.:
  - Lobby
  - Night
  - DayDiscussion
  - DayVote
  - Resolution
  - Ended
- Immutable or copy-on-write style state updates preferred where practical.
- Every command returns:
  - Updated state (or event stream to apply)
  - Domain events
  - Validation error (if invalid)

## Role System (Modular by Design)
Implement roles using a contract-based module pattern.

Example conceptual interface:

```go
type Role interface {
    ID() RoleID
    Team() Team
    Priority() int // resolves action order

    OnPhaseStart(ctx Context, state *State) ([]Event, error)
    ValidateAction(ctx Context, state *State, action Action) error
    ResolveAction(ctx Context, state *State, action Action) ([]Event, error)
    WinCondition(ctx Context, state *State) (won bool, decisive bool)
}
```

Guidelines:
- v1 role baseline is classic game roles only; add expansions behind explicit modules later.
- Roles are registered in a registry/factory and instantiated from config.
- Avoid switch-heavy monoliths; prefer per-role module files.
- Role interactions should be resolved by deterministic ordering and conflict rules.

## Storage Abstraction
- Define storage contracts behind interfaces (game repository, event log, snapshots if used).
- Provide a default in-memory implementation for local usage and tests.
- Engine should depend on interfaces, not concrete storage types.
- Storage must not break deterministic replay for identical seed + command sequence.

## Library API Expectations
Expose a minimal stable API for embedders:
- `NewGame(config)`
- `AddPlayer(...)`, `AssignRoles(...)`, `StartGame(...)`
- `SubmitAction(...)`, `AdvancePhase(...)`
- `State()`, `Events(...)`, `Outcome()`

API design rules:
- Return typed domain errors.
- No panics for user-input errors.
- Context-aware methods where cancellation matters.

## gRPC Service Expectations
- Protobuf-first contract under `api/proto`.
- gRPC methods should map 1:1 to use-cases (create game, join, submit action, advance, query state).
- Keep protobuf messages transport-friendly; map to typed domain models in adapter layer.
- Use gRPC status codes consistently:
  - `InvalidArgument` for validation
  - `NotFound` for unknown IDs
  - `FailedPrecondition` for invalid phase/action timing

Implementation order:
- v1 may ship with no gRPC server implementation.
- v2 adds gRPC as a thin adapter over stable library APIs.

## Testing Mandate (100%)
The engine must be fully tested for unit, branch, and scenario behavior.

Required test categories:
1. **Unit tests**
   - Every exported function and critical internal branch
   - Role-specific behavior and edge cases
2. **Scenario tests**
   - End-to-end game flows by seeded simulations
   - Multiple player counts and role compositions
3. **Table-driven tests**
   - Validation matrices for action legality per phase/role/state
4. **Property/fuzz tests**
   - Invariants: no duplicate players, no impossible phase transitions, winner consistency
5. **Contract tests**
   - gRPC handler tests: request/response mapping and error code translation

Coverage policy:
- `go test ./... -coverprofile=coverage.out`
- `go tool cover -func=coverage.out` must show:
  - 100% for `pkg/engine` and role/rules packages
  - No untested critical branch in state transitions

## CI Quality Gates
A CI pipeline should fail unless all pass:
- `go fmt ./...` (or `gofmt -w` validation)
- `go vet ./...`
- `go test ./... -race -coverprofile=coverage.out`
- Coverage threshold checks (hard fail below policy)
- Optional: `staticcheck ./...`

## Error Handling and Observability
- Domain errors: typed and comparable (`errors.Is`/`errors.As` compatible)
- Include structured logs at adapter boundaries, not deep in pure domain logic.
- Emit domain events for auditability and replay in tests.

## Performance and Concurrency
- Prefer single-writer game loop model per game instance.
- If concurrent access is required, guard mutable state carefully.
- No data races (validated by `-race`).

## Documentation Expectations
At minimum:
- `README.md`: quickstart for library and gRPC modes
- `RULES.md`: precise list of currently implemented rules and role behavior
- Role creation guide: how to add a custom role module
- API docs for exported engine types and methods
- Example integrations (small, runnable)

## Non-Goals (unless explicitly requested)
- UI/frontend
- Persistent storage implementation details (unless needed for examples)
- Matchmaking platform features beyond engine scope

## Delivery Checklist
- [ ] Core engine compiles and runs via public library API
- [ ] Classic role baseline implemented and documented in `RULES.md`
- [ ] Storage abstraction implemented with default in-memory backend
- [ ] Role modules are pluggable and independently testable
- [ ] Full scenario matrix implemented and reproducible with seeds
- [ ] Coverage targets reached and enforced in CI
- [ ] gRPC adapter planned or implemented without duplicated business logic
- [ ] Documentation updated with runnable examples

## Confirmed Scope Decisions
1. Baseline roles: classic game only.
2. Timers/deadlines: caller-managed, not engine-owned.
3. Determinism: randomness must be seed-driven and replayable.
4. Storage: abstract interface with in-memory implementation by default.
5. Delivery order: do not focus on gRPC in initial implementation phase.
