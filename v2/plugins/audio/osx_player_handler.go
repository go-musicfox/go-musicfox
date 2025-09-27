//go:build darwin

package audio

import (
	"github.com/ebitengine/purego/objc"
)

// playerHandler 处理AVPlayer通知事件
type playerHandler struct {
	ID objc.ID
	p  *OSXPlayer
}

// sel_handleFinish 播放完成选择器
var sel_handleFinish = objc.RegisterName("handleFinish:")

// sel_handleFailed 播放失败选择器
var sel_handleFailed = objc.RegisterName("handleFailed:")

// newPlayerHandler 创建新的播放器处理器
func newPlayerHandler(p *OSXPlayer) *playerHandler {
	h := &playerHandler{p: p}

	// 简化实现，直接创建一个空的ID
	// 在实际使用中，这里应该创建一个真正的Objective-C对象
	h.ID = objc.ID(0)

	return h
}

// handleFinish 处理播放完成通知
func (h *playerHandler) handleFinish(id objc.ID, cmd objc.SEL, notification objc.ID) {
	if h.p != nil {
		h.p.setState(Stopped)
		if h.p.timer != nil {
			h.p.timer.Stop()
		}
	}
}

// handleFailed 处理播放失败通知
func (h *playerHandler) handleFailed(id objc.ID, cmd objc.SEL, notification objc.ID) {
	if h.p != nil {
		h.p.setState(Stopped)
		if h.p.timer != nil {
			h.p.timer.Stop()
		}
	}
}

// release 释放处理器资源
func (h *playerHandler) release() {
	if h.ID != 0 {
		// 简化实现，直接重置ID
		// 在实际使用中，这里应该移除通知监听器并释放对象
		h.ID = 0
	}
}
