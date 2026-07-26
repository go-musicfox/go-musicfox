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
// 它会按顺序处理：认证检查 -> 加载状态 -> 核心逻辑。
//
// 当 showLoading 为 true 时，操作将通过 DeferWithLoading 异步执行，
// 使得加载提示能在核心逻辑运行前渲染至少一帧。此时 Execute 立即返回 nil，
// 实际的页面导航由核心逻辑通过副作用（如 EnterMenu）或返回值完成。
//
// 认证检查仍然同步执行：若需要登录，立即返回登录页；登录成功后回调会重新
// 进入 Execute，此时认证通过，showLoading 分支将异步执行核心逻辑。
func (op *Operation) Execute() model.Page {
	// 优先处理认证检查（同步），避免对未登录用户显示无意义的加载状态
	if op.needsAuth {
		if _struct.CheckUserInfo(op.n.user) == _struct.NeedLogin {
			page, _ := op.n.ToLoginPage(func() model.Page {
				return op.Execute()
			})
			return page
		}
	}

	// 若需要显示加载状态，委托给 DeferWithLoading 异步执行，
	// 使加载提示在核心逻辑运行前得以渲染。
	if op.showLoading {
		main := op.n.MustMain()
		scheduled := main.DeferWithLoading(func(m *model.Main) (bool, model.Page) {
			// 核心逻辑在 tick 处理器中执行，此时加载提示已渲染
			resultPage := op.coreFunc(op.n)
			if resultPage != nil {
				// 核心逻辑返回了新页面，通知框架结束加载并导航至该页面
				return false, resultPage
			}
			// 核心逻辑通过副作用完成导航（如 EnterMenu）或无需导航，
			// 返回 (true, nil) 正常结束加载并重渲染
			return true, nil
		})

		if !scheduled {
			// DeferWithLoading 因互斥检查失败（已有其他待执行操作），
			// 降级为同步执行以保证操作不丢失
			return op.coreFunc(op.n)
		}

		// 异步执行已调度，立即返回 nil。
		// 实际的页面导航将在 tick 处理器中通过 resultPage 或副作用完成。
		return nil
	}

	// 不需要加载状态，直接同步执行
	return op.coreFunc(op.n)
}
