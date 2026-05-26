package golden_cross

import (
	"math"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeDeadRelease is a thin wrapper around quant.SoftRelease and
// quant.HardRelease.  It orchestrates the two-phase dead-hold release
// logic and returns a single ReleaseIntent (or nil).
//
// Phase 1 — SoftRelease: aged dead-hold lots that exceed the target
// threshold are converted to floating.
//
// Phase 2 — HardRelease: if the micro engine has produced a SELL that
// exceeds available FloatHold, additional dead-hold lots are released
// to cover the gap.
//
// The function copies input lots to avoid mutating the caller's data.
func ComputeDeadRelease(
	lots []quant.SpotLot,
	currentTime int64,
	price float64,
	totalEquity float64,
	floatHold float64,
	microIntent *quant.MicroIntent,
	chromo quant.Chromosome,
) *quant.ReleaseIntent {
	// Copy to preserve caller immutability.
	lotsCopy := make([]quant.SpotLot, len(lots))
	copy(lotsCopy, lots)

	// ── Phase 1: soft release ──
	newLots, softAmount := quant.SoftRelease(
		lotsCopy, currentTime, price, totalEquity,
		chromo.SoftReleaseMonths,
		chromo.MaxSoftReleasePct,
		chromo.DeadHoldTarget,
	)
	if softAmount > 0 {
		return &quant.ReleaseIntent{
			ReleaseType:   "SOFT",
			ReleaseAmount: softAmount,
			Reason:        "aged dead hold soft release",
		}
	}

	// ── Phase 2: hard release ──
	if microIntent != nil && microIntent.Action == "SELL" {
		// Convert sell CNY to required shares.
		neededShares := math.Abs(microIntent.AmountCNY) / price
		gap := neededShares - floatHold
		if gap > 0 {
			_, hardAmount := quant.HardRelease(newLots, gap)
			if hardAmount > 0 {
				return &quant.ReleaseIntent{
					ReleaseType:   "HARD",
					ReleaseAmount: hardAmount,
					Reason:        "hard release to cover micro sell shortfall",
				}
			}
		}
	}

	return nil
}
