package desktop_lyrics

type fontInk struct {
	Above, Below float64
}

func lyricLabelHeight(fontSize, lineSpacing, defaultSize float64, primary, fallback fontInk, needLargeHeight bool) float64 {
	if fontSize <= 0 {
		fontSize = defaultSize
	}
	inkHeight := max(primary.Above, fallback.Above) + max(primary.Below, fallback.Below)
	height := max(fontSize+lineSpacing, inkHeight) + lineSpacing
	if needLargeHeight {
		return 2 * height
	}
	return height
}
