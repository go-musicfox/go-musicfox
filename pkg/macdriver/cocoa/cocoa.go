//go:build darwin

package cocoa

type CGFloat = float64

type CGSize struct {
	Width, Height CGFloat
}

type CGPoint struct {
	X, Y float64
}

type CGRect struct {
	Origin CGPoint
	Size   CGSize
}

type NSPoint = CGPoint
type NSRect = CGRect
type NSSize = CGSize
