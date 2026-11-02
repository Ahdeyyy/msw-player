package main

type densityMetrics struct{ unit, row, thumbnailW, thumbnailH, panel, gap float32 }

func (a *app) densityMetrics() densityMetrics {
	switch a.density {
	case "Compact":
		return densityMetrics{3, 64, 80, 45, 16, 12}
	case "Spacious":
		return densityMetrics{5, 100, 128, 72, 28, 24}
	default:
		return densityMetrics{4, 84, 104, 58.5, 24, 18}
	}
}

const scrollbarGutter float32 = 16
