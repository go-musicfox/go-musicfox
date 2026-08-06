# PR 标题

fix(login): 注入随机 NMTID 并统一登录身份为 pc，降低登录风控触发率

# PR 描述

## 背景

go-musicfox 登录功能被网易云服务端风控（-462 人机验证）拦截。排查后发现 SDK 层存在两处反风控缺陷：

1. **NMTID 为写死的固定字符串**：`util.ApplyRequestStrategy` 注入的 NMTID 固定为 `"some_random_id_from_strategy"`。服务端极易通过设备指纹识别固定串；已有公开实验证实 NMTID 缺失或形状固定是触发 verifyType=50 硬拦截的已知原因。
2. **邮箱登录身份不一致**：邮箱登录显式携带 `os=ios`、`appver=8.7.01` 的过时客户端身份，与手机号/二维码登录的 `os=pc` 身份互相矛盾，多条登录路径指纹不一致。

## 改动

| 文件 | 改动 |
|---|---|
| `util/common.go` | `ApplyRequestStrategy` 注入的 NMTID 由固定字符串改为运行时随机生成（复用已导出的 `util.RandStringRunes(32)`） |
| `service/login_email_service.go` | 移除显式 `os=ios`/`appver=8.7.01` cookie，邮箱登录与其它登录路径统一为 `os=pc` 身份（由 `ApplyRequestStrategy` 注入全局 cookie jar） |

`ApplyRequestStrategy` 是所有登录相关请求（含二维码 `CheckQR` 轮询、登录刷新）的共同注入点，在源头随机化可保证轮询/刷新期间 jar 中的 NMTID 始终是随机值，不会被固定值覆盖。

## 验证

- 改动仅引用已导出的 `util.RandStringRunes`，未新增依赖与导出符号；编译正确性由 CI `go build` 最终确认
- go-musicfox 已在 vendor 中同步本改动，随 go-musicfox PR #617 一并验证登录回归

## 备注

- 新增注释遵循 go-musicfox 项目的英文注释规范（本仓库现有注释为中文，如需统一可反馈调整）
- 建议发布 v1.6.1，供下游项目 `make vendor` 同步
