//go:build darwin
// +build darwin

package cocoa

type NSBundle struct{ gen_NSBundle }

func NSBundle_Main() NSBundle {
	return NSBundle_mainBundle()
}
