package common

import "errors"

var TransitionCancelErr = errors.New("state transition was cancelled")
