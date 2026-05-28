package quant_test

import (
	"testing"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	"github.com/stretchr/testify/assert"
)

func TestClampChromosome_EnforcesEMAShortLessThanLong(t *testing.T) {
	c := quant.DefaultSeedChromosome
	c.EMAShortBars = 80
	c.EMALongBars = 30
	quant.ClampChromosome(&c)
	assert.True(t, c.EMAShortBars < c.EMALongBars, "EMAShortBars must be < EMALongBars after clamp")
}

func TestClampChromosome_EnforcesDeadHoldMicroReserveConstraint(t *testing.T) {
	c := quant.DefaultSeedChromosome
	c.DeadHoldTarget = 0.6
	c.MicroReservePct = 0.5  // 0.6 + 0.5 = 1.1 > 0.95
	quant.ClampChromosome(&c)
	assert.LessOrEqual(t, c.DeadHoldTarget+c.MicroReservePct, 0.95, "sum must be <= 0.95")
}

func TestDefaultSeedChromosome_IsValid(t *testing.T) {
	c := quant.DefaultSeedChromosome
	assert.True(t, c.EMAShortBars < c.EMALongBars)
	assert.LessOrEqual(t, c.DeadHoldTarget+c.MicroReservePct, 0.95)
	assert.GreaterOrEqual(t, c.A, -3.0)
	assert.LessOrEqual(t, c.A, 3.0)
}
