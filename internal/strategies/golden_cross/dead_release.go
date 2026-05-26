package golden_cross

import (
	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeDeadRelease determines whether any soft or hard release should occur.
//
// Soft release: when DeadHold exceeds the target threshold AND some lots have
// aged past SoftReleaseMonths, release a capped fraction of aged dead hold to
// FLOATING.
//
// Hard release: when the micro engine wants to sell but FloatHold is
// insufficient, draw from non-cold-sealed DEAD_STACK lots to fill the gap.
func ComputeDeadRelease(
	lots []quant.SpotLot,
	currentTime int64,
	price float64,
	totalEquity float64,
	floatHold float64,
	microIntent *quant.MicroIntent,
	chromo quant.Chromosome,
) *quant.ReleaseIntent {
	if totalEquity <= 0 || price <= 0 {
		return nil
	}

	// ── Soft release check ──
	totalDead := quant.TotalDeadHold(lots)
	targetAmount := chromo.DeadHoldTarget * totalEquity / price

	excess := totalDead - targetAmount
	if excess > 0 && chromo.SoftReleaseMonths > 0 {
		thresholdMs := currentTime - int64(chromo.SoftReleaseMonths)*30*24*3600*1000

		agedAmount := 0.0
		for _, l := range lots {
			if l.LotType == quant.LotDeadStack && !l.IsColdSealed && l.CreatedAt < thresholdMs {
				agedAmount += l.Amount
			}
		}

		if agedAmount > 0 {
			releaseAmount := min(excess, agedAmount, totalDead*chromo.MaxSoftReleasePct)
			if releaseAmount > 0 {
				return &quant.ReleaseIntent{
					ReleaseType:   "SOFT",
					ReleaseAmount: releaseAmount,
					Reason:        "dead hold exceeds target, aged lots available",
				}
			}
		}
	}

	// ── Hard release check ──
	if microIntent != nil && microIntent.Action == "SELL" {
		neededShares := microIntent.AmountCNY / price
		if neededShares > floatHold {
			gap := neededShares - floatHold
			releasable := 0.0
			for _, l := range lots {
				if l.LotType == quant.LotDeadStack && !l.IsColdSealed {
					releasable += l.Amount
				}
			}
			hardRelease := min(gap, releasable)
			if hardRelease > 0 {
				return &quant.ReleaseIntent{
					ReleaseType:   "HARD",
					ReleaseAmount: hardRelease,
					Reason:        "float hold insufficient for sell, drawing from dead hold",
				}
			}
		}
	}

	return nil
}
