//go:build darwin

package commands

import (
	"fmt"
	"time"

	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/utils/notify"
)

var notifyOpts struct {
	title string
	text  string
}

func NewNotifyCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "notify",
		UseFor: "发送测试通知 (仅 macOS)",
		Examples: "{$binName} {$cmd}\n" +
			"  {$binName} {$cmd} -t \"标题\" -m \"内容\"\n",
		Config: func(c *gcli.Command) {
			c.Flags.StrOpt(&notifyOpts.title, "title", "t", "测试通知", "通知标题")
			c.Flags.StrOpt(&notifyOpts.text, "message", "m", "这是一条测试通知", "通知内容")
		},
		Func: runNotify,
	}
	return cmd
}

func runNotify(_ *gcli.Command, _ []string) error {
	notify.Notify(notify.NotifyContent{
		Title: notifyOpts.title,
		Text:  notifyOpts.text,
	})
	fmt.Println("通知已发送")

	time.Sleep(time.Second * 2)

	return nil
}
