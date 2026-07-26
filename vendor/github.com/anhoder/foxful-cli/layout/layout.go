// Package layout provides lipgloss layout primitives as a public API for
// downstream consumers. It re-exports core types (Layer, Compositor) and
// layout functions (JoinHorizontal, JoinVertical, Place, Overlay) so that
// callers do not need to import lipgloss directly for common compositing
// and positioning tasks.
//
// Usage:
//
//	// Simple overlay
//	result := layout.Overlay(background, popup, 10, 5)
//
//	// Multi-layer compositing
//	layers := []*layout.Layer{
//	    layout.NewLayer(background),
//	    layout.NewLayer(sidebar).X(0).Y(2),
//	    layout.NewLayer(modal).X(30).Y(10),
//	}
//	comp := layout.NewCompositor(layers...)
//	result := comp.Render()
//
//	// Centering content
//	centered := layout.CenterH(content, width)
//	fitted := layout.Center(content, width, height)
//
//	// Joining
//	row := layout.JoinHorizontal(layout.Top, col1, col2)
//	col := layout.JoinVertical(layout.Left, row1, row2)
package layout

import (
	"charm.land/lipgloss/v2"
)

// ---- re-export core lipgloss layout types ----

// Layer is a renderable layer with a position.
type Layer = lipgloss.Layer

// Compositor manages multiple layers and renders the final composite output.
type Compositor = lipgloss.Compositor

// Position controls horizontal or vertical alignment.
type Position = lipgloss.Position

// ---- re-export alignment constants ----

// Vertical and horizontal alignment positions.
const (
	Top    = lipgloss.Top
	Bottom = lipgloss.Bottom
	Left   = lipgloss.Left
	Right  = lipgloss.Right
	Center = lipgloss.Center
)

// ---- re-export factory functions ----

// NewLayer creates a new renderable layer from content.
var NewLayer = lipgloss.NewLayer

// NewCompositor creates a compositor that stacks layers from back to front.
var NewCompositor = lipgloss.NewCompositor

// ---- re-export layout functions ----

// JoinHorizontal places strings side-by-side with vertical alignment.
var JoinHorizontal = lipgloss.JoinHorizontal

// JoinVertical stacks strings vertically with horizontal alignment.
var JoinVertical = lipgloss.JoinVertical

// Place positions a string within a bounding box of the given dimensions.
var Place = lipgloss.Place

// Width returns the visual (grapheme) width of a string, ignoring ANSI escapes.
var Width = lipgloss.Width

// WithWhitespaceChars sets the fill characters used by Place.
var WithWhitespaceChars = lipgloss.WithWhitespaceChars

// WithWhitespaceStyle sets the style applied to fill characters used by Place.
var WithWhitespaceStyle = lipgloss.WithWhitespaceStyle

// ---- higher-level helpers ----

// Overlay places an overlay string on top of a background string at (x, y).
// This is a convenience shorthand for the most common compositing pattern.
func Overlay(background string, overlay string, x, y int) string {
	layers := []*Layer{
		NewLayer(background),
		NewLayer(overlay).X(x).Y(y),
	}
	return NewCompositor(layers...).Render()
}

// CenterH horizontally centers content within the given width.
// Equivalent to JoinHorizontal(Center, content) with full-width padding.
func CenterH(content string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(content)
}

// PlaceCenter centers content both horizontally and vertically within the
// given bounding box. Shorthand for Place(width, height, Center, Center, content).
func PlaceCenter(content string, width, height int) string {
	return Place(width, height, Center, Center, content)
}
