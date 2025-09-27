# 插件配置和部署指南

## 概述

本指南详细介绍 go-musicfox v2 插件的配置管理、打包和部署流程，帮助开发者正确配置和部署插件。

## 插件配置文件

### 配置文件格式

插件配置使用 JSON 格式，文件名为 `plugin.json`：

```json
{
  "id": "my-audio-plugin",
  "name": "My Audio Plugin",
  "version": "1.0.0",
  "description": "A sample audio processing plugin",
  "author": "Your Name",
  "license": "MIT",
  "homepage": "https://github.com/yourname/my-audio-plugin",
  "type": "dynamic_library",
  "category": "audio-effect",
  "tags": ["audio", "effect", "filter"],
  "api_version": "v2.0.0",
  "min_kernel_version": "2.0.0",
  "entry_point": "plugin.so",
  "dependencies": [
    "audio-codec",
    "network-client"
  ],
  "permissions": [
    {
      "id": "audio-processing",
      "description": "Process audio data",
      "required": true
    },
    {
      "id": "file-access",
      "description": "Access audio files",
      "required": false
    }
  ],
  "config_schema": {
    "type": "object",
    "properties": {
      "sample_rate": {
        "type": "integer",
        "minimum": 8000,
        "maximum": 192000,
        "default": 44100
      },
      "buffer_size": {
        "type": "integer",
        "minimum": 64,
        "maximum": 8192,
        "default": 1024
      },
      "enabled": {
        "type": "boolean",
        "default": true
      }
    }
  },
  "resources": {
    "max_memory": "100MB",
    "max_cpu": 0.5,
    "timeout": "30s"
  },
  "security": {
    "sandbox": true,
    "allowed_paths": [
      "./data",
      "./temp"
    ],
    "allowed_networks": [
      "api.example.com"
    ]
  }
}
```

### 配置字段说明

#### 基础信息

- **id**: 插件唯一标识符，必须全局唯一
- **name**: 插件显示名称
- **version**: 插件版本，遵循语义化版本规范
- **description**: 插件描述
- **author**: 插件作者
- **license**: 许可证类型
- **homepage**: 插件主页或仓库地址

#### 插件类型

- **type**: 插件类型，可选值：
  - `dynamic_library`: 动态链接库插件
  - `rpc`: RPC 插件
  - `webassembly`: WebAssembly 插件
  - `hot_reload`: 热加载插件

- **category**: 插件分类，如 `audio-effect`、`music-source`、`ui-extension`
- **tags**: 插件标签，用于搜索和分类

#### 版本兼容性

- **api_version**: 支持的 API 版本
- **min_kernel_version**: 最低内核版本要求

#### 入口点和依赖

- **entry_point**: 插件入口文件
- **dependencies**: 依赖的其他插件列表

#### 权限配置

```json
"permissions": [
  {
    "id": "permission-id",
    "description": "Permission description",
    "required": true
  }
]
```

常见权限类型：
- `audio-processing`: 音频处理权限
- `file-access`: 文件访问权限
- `network-access`: 网络访问权限
- `system-info`: 系统信息访问权限

#### 配置模式

使用 JSON Schema 定义插件配置参数：

```json
"config_schema": {
  "type": "object",
  "properties": {
    "parameter_name": {
      "type": "string|integer|boolean|number",
      "minimum": 0,
      "maximum": 100,
      "default": "default_value",
      "description": "Parameter description"
    }
  },
  "required": ["required_parameter"]
}
```

#### 资源限制

```json
"resources": {
  "max_memory": "100MB",     // 最大内存使用
  "max_cpu": 0.5,           // 最大 CPU 使用率 (0-1)
  "max_disk_io": "10MB/s",  // 最大磁盘 I/O
  "max_network_io": "1MB/s", // 最大网络 I/O
  "timeout": "30s",         // 操作超时时间
  "max_goroutines": 100     // 最大协程数
}
```

#### 安全配置

```json
"security": {
  "sandbox": true,              // 是否启用沙箱
  "allowed_paths": [            // 允许访问的路径
    "./data",
    "./config"
  ],
  "allowed_networks": [         // 允许访问的网络
    "api.example.com",
    "*.trusted-domain.com"
  ],
  "allowed_syscalls": [         // 允许的系统调用
    "read",
    "write",
    "open"
  ]
}
```

## 插件类型特定配置

### 动态链接库插件

```json
{
  "type": "dynamic_library",
  "entry_point": "plugin.so",
  "build": {
    "go_version": ">=1.21",
    "build_mode": "plugin",
    "cgo_enabled": false,
    "ldflags": "-s -w"
  }
}
```

### RPC 插件

```json
{
  "type": "rpc",
  "rpc": {
    "protocol": "grpc",
    "address": "localhost:50051",
    "timeout": "30s",
    "retry_policy": {
      "max_attempts": 3,
      "initial_backoff": "1s",
      "max_backoff": "10s"
    },
    "health_check": {
      "enabled": true,
      "interval": "10s",
      "timeout": "5s"
    }
  }
}
```

### WebAssembly 插件

```json
{
  "type": "webassembly",
  "entry_point": "plugin.wasm",
  "wasm": {
    "runtime": "wasmtime",
    "memory_limit": "50MB",
    "timeout": "30s",
    "host_functions": [
      "host_log",
      "host_get_time",
      "host_http_request"
    ]
  }
}
```

### 热加载插件

```json
{
  "type": "hot_reload",
  "hot_reload": {
    "enabled": true,
    "watch_paths": [
      "./src",
      "./templates"
    ],
    "ignore_patterns": [
      "*.tmp",
      "*.log"
    ],
    "debounce_time": "500ms",
    "auto_reload": true
  }
}
```

## 环境配置

### 开发环境配置

创建 `config/development.yaml`：

```yaml
kernel:
  log_level: debug
  debug_mode: true

plugins:
  auto_load: true
  load_paths:
    - "./plugins"
    - "./examples"
  
  # 插件特定配置
  my-audio-plugin:
    enabled: true
    config:
      sample_rate: 44100
      buffer_size: 1024
      debug: true

security:
  strict_mode: false
  allow_unsigned_plugins: true

resources:
  default_limits:
    max_memory: "200MB"
    max_cpu: 0.8
    timeout: "60s"
```

### 生产环境配置

创建 `config/production.yaml`：

```yaml
kernel:
  log_level: info
  debug_mode: false

plugins:
  auto_load: false
  load_paths:
    - "/opt/musicfox/plugins"
  
  # 只加载必要的插件
  enabled_plugins:
    - "audio-processor"
    - "netease-music"
    - "local-music"

security:
  strict_mode: true
  allow_unsigned_plugins: false
  require_signature: true

resources:
  default_limits:
    max_memory: "100MB"
    max_cpu: 0.5
    timeout: "30s"

monitoring:
  enabled: true
  metrics_endpoint: ":9090"
  health_check_endpoint: ":8080/health"
```

## 插件打包

### 目录结构

```
my-plugin/
├── plugin.json          # 插件配置文件
├── plugin.so           # 编译后的插件文件
├── README.md           # 插件说明
├── LICENSE             # 许可证文件
├── CHANGELOG.md        # 变更日志
├── docs/               # 文档目录
│   ├── api.md         # API 文档
│   └── examples.md    # 使用示例
├── examples/           # 示例代码
│   └── basic.go
└── tests/              # 测试文件
    └── plugin_test.go
```

### 打包脚本

创建 `scripts/package.sh`：

```bash
#!/bin/bash

set -e

PLUGIN_NAME="my-plugin"
VERSION="1.0.0"
DIST_DIR="dist"

echo "Packaging plugin ${PLUGIN_NAME} v${VERSION}..."

# 创建发布目录
mkdir -p "$DIST_DIR"

# 构建插件
echo "Building plugin..."
go build -buildmode=plugin -ldflags="-s -w" -o plugin.so

# 验证插件配置
echo "Validating plugin configuration..."
musicfox-cli plugin validate plugin.json

# 创建插件包
echo "Creating plugin package..."
tar -czf "$DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz" \
    plugin.json \
    plugin.so \
    README.md \
    LICENSE \
    docs/ \
    examples/

# 生成校验和
echo "Generating checksums..."
cd "$DIST_DIR"
sha256sum "${PLUGIN_NAME}-${VERSION}.tar.gz" > "${PLUGIN_NAME}-${VERSION}.sha256"

echo "Package created: $DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz"
```

### 签名和验证

```bash
# 生成密钥对
musicfox-cli plugin keygen --name="Your Name" --email="your@email.com"

# 签名插件包
musicfox-cli plugin sign \
    --key=private.key \
    --package=my-plugin-1.0.0.tar.gz \
    --output=my-plugin-1.0.0.sig

# 验证插件包
musicfox-cli plugin verify \
    --key=public.key \
    --package=my-plugin-1.0.0.tar.gz \
    --signature=my-plugin-1.0.0.sig
```

## 插件部署

### 本地部署

```bash
# 安装插件
musicfox-cli plugin install my-plugin-1.0.0.tar.gz

# 启用插件
musicfox-cli plugin enable my-plugin

# 配置插件
musicfox-cli plugin config my-plugin \
    --set sample_rate=48000 \
    --set buffer_size=2048

# 重启服务
sudo systemctl restart musicfox
```

### 容器化部署

创建 `Dockerfile`：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -buildmode=plugin -o plugin.so

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /plugins
COPY --from=builder /app/plugin.so .
COPY --from=builder /app/plugin.json .

VOLUME ["/plugins"]
```

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  musicfox:
    image: go-musicfox:v2
    ports:
      - "8080:8080"
    volumes:
      - ./plugins:/app/plugins
      - ./config:/app/config
    environment:
      - MUSICFOX_CONFIG=/app/config/production.yaml
      - MUSICFOX_PLUGINS_DIR=/app/plugins
    depends_on:
      - redis
  
  my-plugin:
    build: .
    volumes:
      - plugin-data:/plugins
    
  redis:
    image: redis:alpine
    volumes:
      - redis-data:/data

volumes:
  plugin-data:
  redis-data:
```

### Kubernetes 部署

创建 `k8s/plugin-configmap.yaml`：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-plugin-config
data:
  plugin.json: |
    {
      "id": "my-plugin",
      "name": "My Plugin",
      "version": "1.0.0",
      "config": {
        "sample_rate": 44100,
        "buffer_size": 1024
      }
    }
```

创建 `k8s/deployment.yaml`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: musicfox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: musicfox
  template:
    metadata:
      labels:
        app: musicfox
    spec:
      containers:
      - name: musicfox
        image: go-musicfox:v2
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: plugin-config
          mountPath: /app/plugins/my-plugin
        - name: plugin-binary
          mountPath: /app/plugins/my-plugin/plugin.so
          subPath: plugin.so
        env:
        - name: MUSICFOX_PLUGINS_DIR
          value: "/app/plugins"
      volumes:
      - name: plugin-config
        configMap:
          name: my-plugin-config
      - name: plugin-binary
        secret:
          secretName: my-plugin-binary
```

## 配置管理最佳实践

### 1. 配置分层

```
配置优先级（从高到低）：
1. 命令行参数
2. 环境变量
3. 配置文件
4. 插件默认配置
```

### 2. 敏感信息处理

```yaml
# 使用环境变量存储敏感信息
api_key: ${MUSIC_API_KEY}
database_password: ${DB_PASSWORD}

# 或使用外部密钥管理
secrets:
  provider: "vault"
  endpoint: "https://vault.example.com"
  path: "secret/musicfox/plugins"
```

### 3. 配置验证

```go
func (p *MyPlugin) ValidateConfig(config map[string]interface{}) error {
    // 验证必需参数
    if _, ok := config["api_key"]; !ok {
        return fmt.Errorf("missing required config: api_key")
    }
    
    // 验证参数范围
    if sampleRate, ok := config["sample_rate"].(float64); ok {
        if sampleRate < 8000 || sampleRate > 192000 {
            return fmt.Errorf("invalid sample_rate: %v", sampleRate)
        }
    }
    
    return nil
}
```

### 4. 配置热更新

```go
func (p *MyPlugin) UpdateConfig(config map[string]interface{}) error {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    // 验证新配置
    if err := p.ValidateConfig(config); err != nil {
        return err
    }
    
    // 应用新配置
    p.config = config
    
    // 通知配置变更
    p.eventBus.Publish("config.updated", &ConfigEvent{
        PluginID: p.GetInfo().ID,
        Config:   config,
    })
    
    return nil
}
```

## 监控和诊断

### 健康检查

```go
func (p *MyPlugin) HealthCheck() error {
    // 检查关键组件状态
    if !p.isConnected {
        return fmt.Errorf("plugin not connected")
    }
    
    // 检查资源使用情况
    if p.getMemoryUsage() > p.maxMemory {
        return fmt.Errorf("memory usage exceeded limit")
    }
    
    return nil
}
```

### 指标收集

```go
func (p *MyPlugin) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "requests_total":     p.requestsTotal,
        "requests_failed":    p.requestsFailed,
        "response_time_avg":  p.avgResponseTime,
        "memory_usage":       p.getMemoryUsage(),
        "cpu_usage":          p.getCPUUsage(),
        "uptime":             time.Since(p.startTime).Seconds(),
    }
}
```

### 日志配置

```yaml
logging:
  level: info
  format: json
  output: stdout
  
  # 插件特定日志配置
  plugins:
    my-plugin:
      level: debug
      output: /var/log/musicfox/my-plugin.log
      max_size: 100MB
      max_backups: 5
      max_age: 30
```

## 故障排除

### 常见问题

1. **插件加载失败**
   ```bash
   # 检查插件配置
   musicfox-cli plugin validate plugin.json
   
   # 检查依赖
   musicfox-cli plugin deps my-plugin
   
   # 查看详细错误
   musicfox --log-level=debug --load-plugin=plugin.so
   ```

2. **配置错误**
   ```bash
   # 验证配置格式
   musicfox-cli config validate config.yaml
   
   # 检查配置合并结果
   musicfox-cli config show --plugin=my-plugin
   ```

3. **权限问题**
   ```bash
   # 检查插件权限
   musicfox-cli plugin permissions my-plugin
   
   # 授予权限
   musicfox-cli plugin grant my-plugin audio-processing
   ```

### 调试工具

```bash
# 插件状态检查
musicfox-cli plugin status

# 插件性能分析
musicfox-cli plugin profile my-plugin --duration=30s

# 插件日志查看
musicfox-cli plugin logs my-plugin --tail=100

# 插件配置导出
musicfox-cli plugin export my-plugin > my-plugin-config.json
```

---

通过遵循本指南，您可以正确配置和部署 go-musicfox v2 插件，确保插件在各种环境中稳定运行。