# go-musicfox v2

基于微内核插件架构的 go-musicfox 重构版本。

## 项目结构

```
v2/
├── cmd/                    # 应用程序入口
├── pkg/                    # 公共库代码
│   ├── kernel/            # 微内核核心
│   ├── plugin/            # 插件系统
│   ├── model/             # 数据模型
│   ├── config/            # 配置管理
│   ├── event/             # 事件总线
│   ├── registry/          # 服务注册表
│   └── security/          # 安全管理
├── internal/              # 内部实现代码
├── test/                  # 测试文件
├── go.mod                 # Go模块定义
├── .golangci.yml         # 代码质量检查配置
├── Makefile              # 构建脚本
└── README.md             # 项目说明
```

## 技术栈

- **Go**: 1.21+
- **依赖注入**: go.uber.org/dig
- **配置管理**: github.com/knadh/koanf/v2
- **日志**: log/slog (Go标准库)
- **测试**: github.com/stretchr/testify

## 开发命令

```bash
# 下载依赖
make deps

# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行测试
make test

# 构建项目
make build

# 安装开发工具
make tools

# 清理构建产物
make clean
```

## 架构设计

本项目采用微内核插件架构，支持多种插件实现方式：

- **动态链接库插件**: 用于音频处理等性能敏感组件
- **RPC插件**: 用于音乐源等独立服务组件
- **WebAssembly插件**: 用于第三方插件的安全执行
- **热加载插件**: 用于UI扩展等需要快速迭代的组件

## 开发指南

1. 所有新功能开发都应该在 `v2/` 目录下进行
2. 遵循Go标准项目布局和编码规范
3. 提交代码前请运行 `make lint` 和 `make test`
4. 插件开发请参考 `pkg/plugin/` 下的接口定义

## 许可证

本项目继承原项目的许可证。