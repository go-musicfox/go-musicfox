package ui

import (
	"github.com/anhoder/foxful-cli/model"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// CoreFunc 是操作的核心逻辑函数，它接收 Netease 实例作为参数。
type CoreFunc func(m *Netease) model.Page

// Operation 代表一个可执行的操作单元。
type Operation struct {
	n           *Netease
	coreFunc    CoreFunc
	needsAuth   bool
	showLoading bool
}

// NewOperation 创建一个新操作。
// 参数 m 是 Netease 主模型，coreFunc 是要执行的核心业务逻辑。
func NewOperation(n *Netease, coreFunc CoreFunc) *Operation {
	return &Operation{
		n:        n,
		coreFunc: coreFunc,
	}
}

// NeedsAuth 将操作标记为需要用户登录。
// 如果不调用此方法，则默认为不需要登录。
// 返回 Operation 指指针以支持链式调用。
func (op *Operation) NeedsAuth() *Operation {
	op.needsAuth = true
	return op
}

// ShowLoading 将操作标记为在执行期间应显示加载状态。
// 如果不调用此方法，则默认为不显示加载状态。
// 返回 Operation 指针以支持链式调用。
func (op *Operation) ShowLoading() *Operation {
	op.showLoading = true
	return op
}

// Execute 按照配置执行操作。
// 它会按顺序处理：加载状态 -> 认证检查 -> 核心逻辑。
func (op *Operation) Execute() model.Page {
	if op.showLoading {
		loading := model.NewLoading(op.n.MustMain())
		loading.Start()
		defer loading.Complete()
	}

	if op.needsAuth {
		if _struct.CheckUserInfo(op.n.user) == _struct.NeedLogin {
			page, _ := op.n.ToLoginPage(func() model.Page {
				return op.Execute()
			})
			return page
		}
	}

	return op.coreFunc(op.n)
}
