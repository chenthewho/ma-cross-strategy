package golden_cross

import (
	"encoding/json"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// Params bundles the evolvable chromosome with the epoch-frozen spawn point.
type Params struct {
	Chromosome quant.Chromosome `json:"chromosome"`
	SpawnPoint quant.SpawnPoint `json:"spawn_point"`
}

// ParseParams unmarshals a JSON parameter pack.  On failure it returns the
// default seed chromosome and an empty spawn point together with the error.
func ParseParams(paramPackJSON string) (Params, error) {
	var p Params
	if err := json.Unmarshal([]byte(paramPackJSON), &p); err != nil {
		return Params{
			Chromosome: quant.DefaultSeedChromosome,
			SpawnPoint: quant.SpawnPoint{},
		}, err
	}
	return p, nil
}
