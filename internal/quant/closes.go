package quant

// ExtractCloses extracts close prices from a Bar slice.
// This is the ACL degradation layer: spot-equity strategies (IsSpot=true)
// must degrade Bar to []float64 here; strategy kernel must never reference Bar directly.
func ExtractCloses(bars []Bar) []float64 {
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	return closes
}

// ExtractTimestamps extracts OpenTime timestamps from a Bar slice.
func ExtractTimestamps(bars []Bar) []int64 {
	ts := make([]int64, len(bars))
	for i, b := range bars {
		ts[i] = b.OpenTime
	}
	return ts
}
