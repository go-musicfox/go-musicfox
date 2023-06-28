package notificator

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Options struct {
	DefaultIcon string
	AppName     string
	OSXSender   string
}

const (
	UrNormal   = "normal"
	UrCritical = "critical"
)

type notifier interface {
	push(title string, text string, iconPath string, redirectUrl string) *exec.Cmd
	pushCritical(title string, text string, iconPath string, redirectUrl string) *exec.Cmd
}

type Notificator struct {
	notifier    notifier
	defaultIcon string
}

func (n Notificator) Push(urgency string, title string, text string, iconPath string, redirectUrl string) error {
	icon := n.defaultIcon

	if iconPath != "" {
		icon = iconPath
	}

	if urgency == UrCritical {
		return n.notifier.pushCritical(title, text, icon, redirectUrl).Run()
	}

	return n.notifier.push(title, text, icon, redirectUrl).Run()

}

type osxNotificator struct {
	AppName string
	Sender  string
}

func (o osxNotificator) push(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {

	// Checks if terminal-notifier exists, and is accessible.

	// if terminal-notifier exists, use it.
	// else, fall back to osascript. (Mavericks and later.)
	if CheckTermNotif() {
		if redirectUrl != "" {
			return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-open", redirectUrl)
		}
		return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-sender", o.Sender)
	} else if CheckMacOSVersion() {
		title = strings.Replace(title, `"`, `\"`, -1)
		text = strings.Replace(text, `"`, `\"`, -1)

		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title)
		return exec.Command("osascript", "-e", notification)
	}

	// finally falls back to growlnotify.

	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title, "--url", redirectUrl)
}

// Causes the notification to stick around until clicked.
func (o osxNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {

	// same function as above...
	if CheckTermNotif() {
		// timeout set to 30 seconds, to show the importance of the notification
		if redirectUrl != "" {
			return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-timeout", "30", "-open", redirectUrl)
		}

		return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-timeout", "30", "-sender", o.Sender)
	} else if CheckMacOSVersion() {
		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title)
		return exec.Command("osascript", "-e", notification)
	}

	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title, "--url", redirectUrl)

}

type linuxNotificator struct {
	AppName string
}

func (l linuxNotificator) push(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {
	return exec.Command("notify-send", "-a", l.AppName, "-i", iconPath, title, text)
}

// Causes the notification to stick around until clicked.
func (l linuxNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {
	return exec.Command("notify-send", "-a", l.AppName, "-i", iconPath, title, text, "-u", "critical")
}

type windowsNotificator struct{}

func (w windowsNotificator) push(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {
	return exec.Command("growlnotify", "/i:", iconPath, "/t:", title, text)
}

// Causes the notification to stick around until clicked.
func (w windowsNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) *exec.Cmd {
	return exec.Command("growlnotify", "/i:", iconPath, "/t:", title, text, "/s", "true", "/p", "2")
}

func New(o Options) *Notificator {

	var Notifier notifier

	switch runtime.GOOS {

	case "darwin":
		Notifier = osxNotificator{AppName: o.AppName, Sender: o.OSXSender}
	case "linux":
		Notifier = linuxNotificator{AppName: o.AppName}
	case "windows":
		Notifier = windowsNotificator{}

	}

	return &Notificator{notifier: Notifier, defaultIcon: o.DefaultIcon}
}

// Helper function for macOS

func CheckTermNotif() bool {
	// Checks if terminal-notifier exists, and is accessible.
	if err := exec.Command("which", "terminal-notifier").Run(); err != nil {
		return false
	}
	// no error, so return true. (terminal-notifier exists)
	return true
}

func CheckMacOSVersion() bool {
	// Checks if the version of macOS is 10.9 or Higher (osascript support for notifications.)

	cmd := exec.Command("sw_vers", "-productVersion")
	check, _ := cmd.Output()

	version := strings.Split(strings.TrimSpace(string(check)), ".")

	// semantic versioning of macOS

	if len(version) < 2 {
		return false
	}

	major, _ := strconv.Atoi(version[0])
	minor, _ := strconv.Atoi(version[1])

	if major < 10 {
		return false
	} else if major == 10 && minor < 9 {
		return false
	} else {
		return true
	}
}
