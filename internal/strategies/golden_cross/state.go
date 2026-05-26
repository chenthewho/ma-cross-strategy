package golden_cross

import "github.com/chenthewho/ma-cross-strategy/internal/quant"

// RuntimeState is a re-export of quant.RuntimeState.
// The strategy kernel does not add extra fields beyond what the quant
// package provides.
type RuntimeState = quant.RuntimeState
