package model

type Loading struct {
	MenuTips
	displayNotOnlyOnMain bool
}

func NewLoading(m *Main, originMenu ...*MenuItem) *Loading {
	l := &Loading{
		MenuTips: MenuTips{
			main: m,
		},
	}
	if len(originMenu) > 0 {
		l.originMenu = originMenu[0]
	}
	return l
}

func (l *Loading) DisplayNotOnlyOnMain() {
	l.displayNotOnlyOnMain = true
}

func (l *Loading) Start() {
	if !l.displayNotOnlyOnMain {
		if _, ok := l.main.app.CurPage().(*Main); !ok {
			return
		}
	}
	l.DisplayTips(l.main.options.LoadingText)
}

func (l *Loading) Complete() {
	if !l.displayNotOnlyOnMain {
		if _, ok := l.main.app.CurPage().(*Main); !ok {
			return
		}
	}
	l.Recover()
}
