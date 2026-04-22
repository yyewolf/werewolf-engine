package engine

import "errors"

var (
	ErrPlayerExists      = errors.New("player already exists")
	ErrPlayerUnknown     = errors.New("player does not exist")
	ErrRoleMissing       = errors.New("role is required")
	ErrPlayerDead        = errors.New("player is dead")
	ErrInvalidLovers     = errors.New("lovers must reference two different existing players")
	ErrPhaseInvalid      = errors.New("transition is invalid for the current phase")
	ErrGameNotReady      = errors.New("game is not ready to start")
	ErrRoleUnavailable   = errors.New("role is not available for assignment")
	ErrTransitionUnknown = errors.New("unknown transition kind")
	ErrTransitionInvalid = errors.New("invalid transition payload")
	ErrRoleActionUnknown = errors.New("unknown role action")
	ErrRoleActionDenied  = errors.New("role action denied")
	ErrGameEnded         = errors.New("game has already ended")
)
