package desktop_lyrics

// fontInk holds font line metrics above and below the baseline.
type fontInk struct {
	Above, Below float64
}

func lyricLabelHeight(fontSize, lineSpacing float64, primary, fallback fontInk, needLargeHeight bool) float64 {
	inkHeight := max(primary.Above, fallback.Above) + max(primary.Below, fallback.Below)
	height := max(fontSize+lineSpacing, inkHeight) + lineSpacing
	if needLargeHeight {
		return 2 * height
	}
	return height
}
