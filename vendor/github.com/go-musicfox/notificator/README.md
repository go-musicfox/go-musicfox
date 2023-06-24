notificator
===========================

Desktop notification with Golang for:

  * Windows with `growlnotify`;
  * Mac OS X with `terminal-notifier` (if installed) or `osascript` (native, 10.9 Mavericks or Up.);
  * Linux with `notify-send` for Gnome and `kdialog` for Kde.

Usage
------

```go
package main

import (
  "github.com/0xAX/notificator"
)

var notify *notificator.Notificator

func main() {

  notify = notificator.New(notificator.Options{
    DefaultIcon: "icon/default.png",
    AppName:     "My test App",
  })

  notify.Push(notificator.UrNormal, "title", "text", "/home/user/icon.png", "https://github.com/go-musicfox/go-musicfox")
}
```

TODO
-----

  * Add more options for different notificators.

Сontribution
------------

  * Fork;
  * Make changes;
  * Send pull request;
  * Thank you.
