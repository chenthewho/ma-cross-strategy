package quant

// LotManager provides position lot aggregation and dead-hold release logic.
// All methods are pure functions operating on lot slices — no I/O, no DB access.

// TotalDeadHold returns the sum of all non-cold-sealed DEAD_STACK lots.
func TotalDeadHold(lots []SpotLot) float64 {
	sum := 0.0
	for _, l := range lots {
		if l.LotType == LotDeadStack && !l.IsColdSealed {
			sum += l.Amount
		}
	}
	return sum
}

// TotalFloatHold returns the sum of all FLOATING lots.
func TotalFloatHold(lots []SpotLot) float64 {
	sum := 0.0
	for _, l := range lots {
		if l.LotType == LotFloating {
			sum += l.Amount
		}
	}
	return sum
}

// TotalColdSealedHold returns the sum of all cold-sealed lots.
func TotalColdSealedHold(lots []SpotLot) float64 {
	sum := 0.0
	for _, l := range lots {
		if l.IsColdSealed {
			sum += l.Amount
		}
	}
	return sum
}

// SoftRelease performs a soft release of aged DEAD_STACK lots to FLOATING.
//
// Parameters:
//   - lots: current position lots (modified in place)
//   - currentTime: current bar timestamp (ms)
//   - price: current asset price
//   - totalEquity: total portfolio equity
//   - softReleaseMonths: age threshold in months (from chromosome)
//   - maxSoftReleasePct: max fraction of DeadHold to release (from chromosome)
//   - deadHoldTarget: target dead hold fraction (from chromosome)
//
// Returns: release amount (shares) and a modified lots slice.
// Only lots aged > softReleaseMonths and not cold-sealed are eligible.
// Only releases when DeadHold exceeds the target threshold.
func SoftRelease(
	lots []SpotLot,
	currentTime int64,
	price float64,
	totalEquity float64,
	softReleaseMonths int,
	maxSoftReleasePct float64,
	deadHoldTarget float64,
) ([]SpotLot, float64) {
	if softReleaseMonths < 1 {
		return lots, 0
	}
	if totalEquity <= 0 || price <= 0 {
		return lots, 0
	}

	totalDead := TotalDeadHold(lots)
	targetAmount := deadHoldTarget * totalEquity / price

	// Only release if dead hold exceeds target
	excess := totalDead - targetAmount
	if excess <= 0 {
		return lots, 0
	}

	// Calculate aged dead hold
	thresholdMs := currentTime - int64(softReleaseMonths)*30*24*3600*1000
	agedAmount := 0.0
	for _, l := range lots {
		if l.LotType == LotDeadStack && !l.IsColdSealed && l.CreatedAt < thresholdMs {
			agedAmount += l.Amount
		}
	}
	if agedAmount <= 0 {
		return lots, 0
	}

	// Cap the release amount
	releaseAmount := min(excess, agedAmount, totalDead*maxSoftReleasePct)
	if releaseAmount <= 0 {
		return lots, 0
	}

	// Release aged lots proportionally
	remaining := releaseAmount
	for i := range lots {
		if remaining <= 0 {
			break
		}
		l := &lots[i]
		if l.LotType != LotDeadStack || l.IsColdSealed || l.CreatedAt >= thresholdMs {
			continue
		}
		// Release from this lot
		release := min(l.Amount, remaining)
		l.Amount -= release
		remaining -= release
	}

	// Create a new FLOATING lot for the released shares
	newLot := SpotLot{
		LotType:      LotFloating,
		Amount:       releaseAmount,
		CostPrice:    price, // mark to current price on release
		CreatedAt:    currentTime,
		IsColdSealed: false,
	}

	// Remove zero-amount dead lots
	filtered := lots[:0]
	for _, l := range lots {
		if l.Amount > 0 {
			filtered = append(filtered, l)
		}
	}
	filtered = append(filtered, newLot)

	return filtered, releaseAmount
}

// HardRelease performs a hard release when the micro engine wants to sell
// but FloatHold is insufficient.
//
// This draws from non-cold-sealed DEAD_STACK lots to fill the gap.
// ColdSealed lots are NEVER released.
//
// Returns: modified lots and the amount released (shares).
func HardRelease(lots []SpotLot, neededAmount float64) ([]SpotLot, float64) {
	if neededAmount <= 0 {
		return lots, 0
	}

	// Calculate total releasable dead hold (non-cold-sealed)
	releasable := 0.0
	for _, l := range lots {
		if l.LotType == LotDeadStack && !l.IsColdSealed {
			releasable += l.Amount
		}
	}
	releaseAmount := min(neededAmount, releasable)
	if releaseAmount <= 0 {
		return lots, 0
	}

	// Draw from dead lots
	remaining := releaseAmount
	for i := range lots {
		if remaining <= 0 {
			break
		}
		l := &lots[i]
		if l.LotType != LotDeadStack || l.IsColdSealed {
			continue
		}
		release := min(l.Amount, remaining)
		l.Amount -= release
		remaining -= release
	}

	// Create FLOATING lot for released shares
	newLot := SpotLot{
		LotType:      LotFloating,
		Amount:       releaseAmount,
		CostPrice:    lots[0].CostPrice, // preserve original cost
		CreatedAt:    lots[0].CreatedAt,
		IsColdSealed: false,
	}

	filtered := lots[:0]
	for _, l := range lots {
		if l.Amount > 0 {
			filtered = append(filtered, l)
		}
	}
	filtered = append(filtered, newLot)

	return filtered, releaseAmount
}

// ColdSealedDeadHold returns the sum of cold-sealed DEAD_STACK lots.
func ColdSealedDeadHold(lots []SpotLot) float64 {
	sum := 0.0
	for _, l := range lots {
		if l.LotType == LotDeadStack && l.IsColdSealed {
			sum += l.Amount
		}
	}
	return sum
}
