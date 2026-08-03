# 许可证审计报告 — GPL 污染风险分析

> 审计日期：2026-07-29 | 原目标许可证：MIT | 审计范围：go.mod 全部依赖（vendor/ 内 LICENSE 文件）
> **后续行动**：审计确认两个 GPL-3.0 依赖（tag、UnblockNeteaseMusic）污染了 MIT 项目。\
> **2026-07-29 已将项目许可证从 MIT 切换为 GPL-3.0，与所有依赖兼容。**

---

## 总体概览

| 风险等级 | 数量 | 说明 |
|:---|---:|:---|
| 🔴 **HIGH** — GPL 家族（copyleft） | **3** | 直接 / 间接依赖使用了 GPL-3.0 / LGPL v3 |
| 🟡 **MEDIUM** — 缺少 LICENSE 文件 | **2** | 无许可证声明，默认 All Rights Reserved |
| 🟢 **SAFE** — MIT 兼容 | **106** | MIT / BSD / Apache-2.0 / ISC / Unlicense |

**许可证类型分布**：

| 许可证 | 数量 |
|:---|---:|
| MIT | ~68 |
| BSD-3-Clause | ~18 |
| BSD-2-Clause | ~6 |
| Apache-2.0 | ~12 |
| Unlicense (Public Domain) | 1 |
| GPL-3.0 | 2 |
| LGPL v3 (with linking exception) | 1 |
| 缺失 | 2 |

---

## 🔴 HIGH RISK — GPL 家族许可证

**GPL 是强 copyleft 许可证，与 MIT 不兼容。** 静态链接 GPL 代码会强制整个分发作品以 GPL 许可证发布。LGPL 约束较弱但仍是 GPL 家族。

### 1. `github.com/frolovo22/tag` — GPL-3.0

```
go.mod:     github.com/frolovo22/tag v0.0.2
replace:    github.com/frolovo22/tag v0.0.2 => github.com/go-musicfox/tag v1.0.2
vendor:     vendor/github.com/frolovo22/tag/LICENSE → 完整 GPL-3.0 文本
```

| 属性 | 值 |
|---|---|
| **风险等级** | 🔴 **HIGH** |
| **许可证** | GPL-3.0 |
| **类型** | **直接依赖** |
| **使用位置** | `internal/track/tagger.go:16` — `import songtag "github.com/frolovo22/tag"` |
| **功能** | 通用音频标签读写（FLAC/OGG/MP4 等格式的 Metadata）|
| **fork 版本** | `github.com/go-musicfox/tag v1.0.2`，但 fork **仍使用 GPL-3.0** |
| **影响范围** | 编译进主二进制，标签编辑核心逻辑 |

**结论**：`go-musicfox/tag` 是原 `frolovo22/tag` 的增强 fork，但未更换许可证。该库与项目核心二进制**静态链接**，GPL-3.0 的 §5(c) 要求"整个作品必须以本许可证授权"。这使整个 go-musicfox 面临被 GPL-3.0 "污染"的风险。

**建议**：
- **最佳**：联系 fork 作者（项目自身）更换为 MIT/Apache-2.0，或使用 MIT 兼容的替代库
- **备选**：将音频标签功能抽成独立可执行文件，通过 IPC 调用（进程边界隔离）
- **不可行**：保留 GPL-3.0 但声称项目整体是 MIT — 这在法律上无效

---

### 2. `github.com/cnsilvan/UnblockNeteaseMusic` — GPL-3.0

```
go.mod:     github.com/cnsilvan/UnblockNeteaseMusic v0.0.0-... (indirect)
replace:    github.com/cnsilvan/UnblockNeteaseMusic => github.com/go-musicfox/UnblockNeteaseMusic v0.1.6
vendor:     vendor/github.com/cnsilvan/UnblockNeteaseMusic/LICENSE → 完整 GPL-3.0 文本
```

| 属性 | 值 |
|---|---|
| **风险等级** | 🔴 **HIGH** |
| **许可证** | GPL-3.0 |
| **类型** | **间接依赖**（通过 `go-musicfox/netease-music` SDK）|
| **使用位置** | `go-musicfox/netease-music` → UNM 模块，用于"解锁灰色歌曲"功能 |
| **功能** | 多源歌曲匹配（酷我/酷狗/咪咕/QQ 音乐），替代网易云无版权歌曲 |
| **fork 版本** | `github.com/go-musicfox/UnblockNeteaseMusic v0.1.6`，fork **仍使用 GPL-3.0** |

**结论**：UNM 通过 netease-music SDK **间接静态链接**进主二进制。GPL 的 copyleft 效力通过间接依赖同样生效。Go 的静态编译使整个二进制都是"covered work"。

**建议**：
- **最佳**：更换 fork 许可证为 MIT/Apache-2.0（需取得原 `cnsilvan/UnblockNeteaseMusic` 作者同意，因为 fork 基于 GPL 代码）
- **备选**：将 UNM 功能作为独立 HTTP 代理进程运行（如原项目设计意图），进程边界隔离 GPL 代码

---

### 3. `gopkg.in/retry.v1` — LGPL v3（with static-linking exception）

```
go.mod:     gopkg.in/retry.v1 v1.0.3 (indirect)
vendor:     vendor/gopkg.in/retry.v1/LICENSE → LGPL v3 + 静态链接例外
```

| 属性 | 值 |
|---|---|
| **风险等级** | 🔴 **HIGH**（存在争议空间） |
| **许可证** | LGPL v3 + static-linking exception |
| **类型** | **间接依赖** |
| **使用位置** | `vendor/github.com/juju/persistent-cookiejar/serialize.go:16` → cookie 持久化的重试逻辑 |
| **功能** | 文件写入时的指数退避重试策略 |

LGPL v3 的静态链接例外声明：

> As a special exception to the GNU Lesser General Public License version 3 ("LGPL3"), the copyright holders of this Library give you permission to convey to a third party a Combined Work that links statically or dynamically to this Library without providing any Minimal Corresponding Source...

**结论**：LGPL v3 比 GPL-3.0 宽松很多，且 Canonical Ltd. 添加了静态链接例外。许多项目在 MIT 下使用 LGPL 依赖（含例外条款）被认为是可接受的。但这仍是一个灰色地带 — 严格来说它仍是 GPL 家族许可证。

**建议**：
- **可接受**：LGPL v3 + linking exception 在实践中广泛被接受
- **严格合规**：替换为 MIT 许可证的重试库（如 `github.com/cenkalti/backoff/v4`）

---

## 🟡 MEDIUM RISK — 缺少 LICENSE 文件

### 4. `github.com/anhoder/foxful-cli` — 无 LICENSE

```
go.mod:     github.com/anhoder/foxful-cli v1.0.5 (direct)
vendor:     vendor/github.com/anhoder/foxful-cli/ — 无 LICENSE/COPYING 文件
```

| 属性 | 值 |
|---|---|
| **风险等级** | 🟡 **MEDIUM** |
| **类型** | **直接依赖**（项目自有） |
| **功能** | 核心 UI 框架（bubbletea 的 foxful 定制版） |

**结论**：无许可证意味着默认 "All Rights Reserved"，他人无权复制/分发/修改。但这是**同一作者（anhoder）**的项目，可以在上游仓库添加 MIT LICENSE 文件即可解决。

**建议**：在 `github.com/anhoder/foxful-cli` 仓库添加 MIT LICENSE 文件，然后执行 `make vendor` 同步。

---

### 5. `github.com/mewkiz/pkg` — 无 LICENSE

```
go.mod:     github.com/mewkiz/pkg v0.0.0-20230226050401-4010bf0fec14 (indirect)
vendor:     vendor/github.com/mewkiz/pkg/ — 无 LICENSE/COPYING 文件
```

| 属性 | 值 |
|---|---|
| **风险等级** | 🟡 **MEDIUM** |
| **类型** | **间接依赖** |
| **使用位置** | `mewkiz/flac` → `mewkiz/pkg`（FLAC 解码的工具包）|
| **关联信息** | 同作者 `mewkiz/flac` 使用 **Unlicense**（公共领域） |

**结论**：作者 `mewkiz` 的其他库使用 Unlicense，`mewkiz/pkg` 极可能是相同意图。但缺少正式声明无法确认。

**建议**：向 `github.com/mewkiz/pkg` 提 issue 请求添加 LICENSE 文件，或评估其使用量是否值得替换。

---

## 🟢 其他值得关注的点

### Fork 许可证一致性

| 原始依赖 | Fork | Fork 许可证 |
|---|---|---|
| `frolovo22/tag` | `go-musicfox/tag` v1.0.2 | GPL-3.0 ⚠️ |
| `cnsilvan/UnblockNeteaseMusic` | `go-musicfox/UnblockNeteaseMusic` v0.1.6 | GPL-3.0 ⚠️ |
| `charm.land/bubbletea/v2` | `go-musicfox/bubbletea/v2` | MIT ✅ |
| `gopxl/beep` | `go-musicfox/beep` v1.4.1 | MIT ✅ |
| `hajimehoshi/go-mp3` | `go-musicfox/go-mp3` v0.3.3 | Apache-2.0 ✅ |
| `shkh/lastfm-go` | `go-musicfox/lastfm-go` v0.0.2 | MIT ✅ |
| `saltosystems/winrt-go` | `go-musicfox/winrt-go` v0.1.4 | MIT ✅ |
| `cocoonlife/goflac` | `go-musicfox/goflac` v0.1.5 | BSD-3 ✅ |
| `gookit/gcli/v2` | `anhoder/gcli/v2` v2.3.5 | MIT ✅ |

### 嵌入字体许可证

`github.com/alecthomas/chroma/v2` 嵌入 **Liberation Mono** 字体（base64 编码），该字体使用 **SIL Open Font License 1.1**。SIL OFL 是字体专用许可证，与代码的 MIT 兼容，但分发时需要保留字体版权声明。

### 完整安全清单（部分）

以下列出所有已验证为 MIT 兼容的依赖（代表性样本）：

| # | 依赖 | 许可证 |
|---:|---|:---:|
| | `charm.land/bubbles/v2`, `bubbletea/v2`, `lipgloss/v2` | MIT |
| | `charm.land/glamour/v2`, `charmbracelet/*` (harmonica/ultraviolet/x/*/colorprofile) | MIT |
| | `github.com/BurntSushi/toml`, `adrg/xdg`, `bogem/id3v2/v2`, `buger/jsonparser` | MIT |
| | `github.com/go-musicfox/netease-music`, `go-musicfox/requests` | MIT / Apache-2.0 |
| | `github.com/knadh/koanf/*`, `lucasb-eyer/go-colorful`, `mattn/go-runewidth` | MIT |
| | `github.com/pelletier/go-toml/*`, `rivo/uniseg`, `robotn/gohook` | MIT |
| | `github.com/skip2/go-qrcode`, `skratchdot/open-golang`, `tosone/minimp3` | MIT |
| | `go.etcd.io/bbolt`, `github.com/muesli/cancelreader`, `tidwall/*` | MIT |
| | `github.com/atotto/clipboard`, `go-musicfox/notificator`, `go-ole/go-ole` | BSD-3 |
| | `github.com/gen2brain/beeep`, `godbus/dbus/v5`, `pkg/errors` | BSD-2 |
| | `github.com/ebitengine/purego`, `ebitengine/oto/v3` | Apache-2.0 |
| | `github.com/forgoer/openssl`, `go-flac/flacpicture`, `go-flac/go-flac` | Apache-2.0 |
| | `github.com/mewkiz/flac` | Unlicense |
| | `golang.org/x/*` (sync/sys/net/crypto/term/text) | BSD-3 |
| | `github.com/fsnotify/fsnotify`, `microcosm-cc/bluemonday`, `gorilla/css` | BSD-3 |

---

## 行动建议优先级

| 优先级 | 行动 | 影响 |
|:---:|---|:---|
| **P0** | `go-musicfox/tag` — 更换许可证或替换为 MIT 兼容库 | 阻断级：核心二进制被 GPL-3.0 污染 |
| **P0** | `go-musicfox/UnblockNeteaseMusic` — 更换许可证或进程隔离 | 阻断级：间接依赖同样污染整个二进制 |
| **P1** | `anhoder/foxful-cli` — 在上游仓库添加 MIT LICENSE | 自有项目，一劳永逸 |
| **P2** | `gopkg.in/retry.v1` — 评估替换或接受 LGPL v3 + exception | 存在争议空间，多数场景可接受 |
| **P3** | `mewkiz/pkg` — 联系作者确认许可证 | 间接依赖，影响有限 |

---

*审计方法：逐一读取 vendor/ 下每个依赖的 LICENSE 文件，交叉验证 go.mod replace 指令。所有结论基于实际 vendor 内容，非元数据推断。*
