package model

type Loading struct {
	MenuTips
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

func (l *Loading) Start() {
	l.DisplayTips(l.main.options.LoadingText)
}

func (l *Loading) Complete() {
	l.Recover()
}
