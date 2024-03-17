
<a name="v4.3.1"></a>
## [v4.3.1](https://github.com/go-musicfox/go-musicfox/compare/v4.3.0...v4.3.1) (2024-02-24)

### Feat

* add vscode launch config
* foxful-cli ticker
* update scoop config

### Fix

* 每日推荐歌单
* use `AppName` in `internal.types`
* set mpris Identity an user-friendly name
* linux dbus `s.props` is nil
* cannot load framework in Mac Sonoma

### Optimize

* 默认关闭签到功能


<a name="v4.3.0"></a>
## [v4.3.0](https://github.com/go-musicfox/go-musicfox/compare/v4.2.1...v4.3.0) (2023-10-30)

### Feat

* update scoop config

### Fix

* missing import


<a name="v4.2.1"></a>
## [v4.2.1](https://github.com/go-musicfox/go-musicfox/compare/v4.2.0...v4.2.1) (2023-10-08)

### Feat

* add confirm menu when clear song cache
* update scoop config

### Fix

* [#199](https://github.com/go-musicfox/go-musicfox/issues/199)


<a name="v4.2.0"></a>
## [v4.2.0](https://github.com/go-musicfox/go-musicfox/compare/v4.1.7...v4.2.0) (2023-09-17)

### Feat

* 自定义下载文件名格式([#193](https://github.com/go-musicfox/go-musicfox/issues/193))
* 进度条配置优化 & 支持全局快捷键
* update scoop config

### Fix

* register global_key
* github action
* github action
* global hotkey build


<a name="v4.1.7"></a>
## [v4.1.7](https://github.com/go-musicfox/go-musicfox/compare/v4.1.6...v4.1.7) (2023-09-12)

### Feat

* update scoop config

### Fix

* err
* mac notification
* github action
* inject
* err
* lyrics view
* player view
* playlists
* w not quit
* [#183](https://github.com/go-musicfox/go-musicfox/issues/183)
* beep minimp3 err
* key

### Optimize

* lyrics
* loading
* login callback
* login callback

### Refactor

* pkg => internal


<a name="v4.1.6"></a>
## [v4.1.6](https://github.com/go-musicfox/go-musicfox/compare/v4.1.5...v4.1.6) (2023-08-21)

### Feat

* update scoop config

### Fix

* lyric


<a name="v4.1.5"></a>
## [v4.1.5](https://github.com/go-musicfox/go-musicfox/compare/v4.1.4...v4.1.5) (2023-08-20)

### Add

* Windows 检查更新提示

### Feat

* golang-lint
* golang-lint
* update scoop config
* 增加winget配置（待测试）
* update scoop config
* mpd volume日志
* 将歌曲添加至歌单
* 优化
* 优化ctrl
* 自动播放; fix: 修改一处小问题

### Fix

* [#167](https://github.com/go-musicfox/go-musicfox/issues/167)
* 修复播放列表切片越界的问题
* [#171](https://github.com/go-musicfox/go-musicfox/issues/171)
* error
* [#151](https://github.com/go-musicfox/go-musicfox/issues/151) [#152](https://github.com/go-musicfox/go-musicfox/issues/152)
* [#149](https://github.com/go-musicfox/go-musicfox/issues/149)
* 兼容[#147](https://github.com/go-musicfox/go-musicfox/issues/147)
* [#143](https://github.com/go-musicfox/go-musicfox/issues/143)
* panic
* bcef056后无法构建
* 在播放模式为单曲循环时手动切换歌曲
*  - 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件  - 将log输出至logfile
* mac图片错误时panic

### Optimize

* replace github.com/boltdb/bolt => go.etcd.io/bbolt


<a name="v4.1.4"></a>
## [v4.1.4](https://github.com/go-musicfox/go-musicfox/compare/v4.1.3...v4.1.4) (2023-07-16)

### Feat

* 模糊搜索本地歌单
* 添加了下载歌曲失败时的通知提醒
* add winget
* 增加winget配置（待测试）
* update scoop config
* 自动缓存音乐

### Fix

* actions
* actions
* actions
* actions
* actions
* actions
* actions
* actions
* actions
* actions
* actions
* winget
* winget
* [#158](https://github.com/go-musicfox/go-musicfox/issues/158) setting FLAC file will not halt; use tmp file
* [#149](https://github.com/go-musicfox/go-musicfox/issues/149) again
* error
* [#151](https://github.com/go-musicfox/go-musicfox/issues/151) [#152](https://github.com/go-musicfox/go-musicfox/issues/152)
* [#149](https://github.com/go-musicfox/go-musicfox/issues/149)
* 添加清除音乐缓存在帮助目录的条目
* 1. mpd在第一次下载缓存后panic 2. 删除多余的setSongTag调用 3. 修复cacheLimit语义
* 在缓存未完成时panic
* utils.ClearDir在目录不存在时不再创建目录


<a name="v4.1.3"></a>
## [v4.1.3](https://github.com/go-musicfox/go-musicfox/compare/v4.1.2...v4.1.3) (2023-07-11)

### Add

* Windows 检查更新提示

### Chore

* Move dir:`nix` to `deploy/nix`
* Move dir:`nix` to `deploy/nix`

### Enhancement

* Send music info to mpd

### Feat

* mpd volume日志
* 优化
* 优化ctrl
* 将歌曲添加至歌单
* update pkg
* [#128](https://github.com/go-musicfox/go-musicfox/issues/128)
* update dep
* 配置命令优化
* 从播放列表删除选中歌曲
* update scoop config
* Add flake support

### Fix

* 兼容[#147](https://github.com/go-musicfox/go-musicfox/issues/147)
* [#143](https://github.com/go-musicfox/go-musicfox/issues/143)
* panic
* bcef056后无法构建
* 在播放模式为单曲循环时手动切换歌曲
*  - 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件  - 将log输出至logfile
* mac图片错误时panic
* 歌单详情
* like song
* like selected song
* stack overflow
* 从播放列表删除选中歌曲


<a name="v4.1.2"></a>
## [v4.1.2](https://github.com/go-musicfox/go-musicfox/compare/v4.1.1...v4.1.2) (2023-06-06)

### Feat

* Add flake support
* update scoop config

### Fix

* 修复配置文件目录问题
* package name


<a name="v4.1.1"></a>
## [v4.1.1](https://github.com/go-musicfox/go-musicfox/compare/v4.1.0...v4.1.1) (2023-05-27)

### Feat

* update scoop config

### Fix

* 报错


<a name="v4.1.0"></a>
## [v4.1.0](https://github.com/go-musicfox/go-musicfox/compare/v4.0.6...v4.1.0) (2023-05-24)

### Add

* 播放结束上报网易云

### Feat

* support [#85](https://github.com/go-musicfox/go-musicfox/issues/85)
* update README
* support the base dir specifications of the platforms ([#107](https://github.com/go-musicfox/go-musicfox/issues/107))
* update scoop config

### Fix

* [#105](https://github.com/go-musicfox/go-musicfox/issues/105)
* 播放结束上报网易云

### Refactor

* polish README.md


<a name="v4.0.6"></a>
## [v4.0.6](https://github.com/go-musicfox/go-musicfox/compare/v4.0.5...v4.0.6) (2023-05-02)

### Feat

* 配置
* update Readme
* 自定义通知logo(linux/windows)
* update scoop config

### Fix

* 快退会将秒数退为负数


<a name="v4.0.5"></a>
## [v4.0.5](https://github.com/go-musicfox/go-musicfox/compare/v4.0.4...v4.0.5) (2023-04-22)

### Feat

* 更新issue template
* 更新issue template
* 更新issue template
* 更新issue template
* 更新issue template
* 更新go-mp3
* 更新go-mp3
* update scoop config


<a name="v4.0.4"></a>
## [v4.0.4](https://github.com/go-musicfox/go-musicfox/compare/v4.0.3...v4.0.4) (2023-04-20)

### Feat

* 更新libflac
* 优化libflac
* 增加等待次数
* 优化beep_player因网速慢切歌问题
* update scoop config

### Fix

* oto linux panic
* 渐变色


<a name="v4.0.3"></a>
## [v4.0.3](https://github.com/go-musicfox/go-musicfox/compare/v4.0.2...v4.0.3) (2023-04-14)

### Feat

* github action
* update scoop config
* 配置进度条
* 增加鼠标配置
* update scoop config

### Fix

* panic: slice bounds out of range [:33] with capacity 32


<a name="v4.0.2"></a>
## [v4.0.2](https://github.com/go-musicfox/go-musicfox/compare/v4.0.1...v4.0.2) (2023-04-12)

### Feat

* minimp3优化
* 优化
* 代码优化
* 增加构建
* update libflac
* 登陆后更新like list
* update scoop config

### Fix

* 重载后概率导致当前歌曲停止 && Position()偶尔panic


<a name="v4.0.1"></a>
## [v4.0.1](https://github.com/go-musicfox/go-musicfox/compare/v4.0.0...v4.0.1) (2023-04-10)

### Feat

* 增加linux arm64及windows arm64构建
* goreleaser调试
* goreleaser调试
* goreleaser调试
* goreleaser优化
* 优化
* goreleaser调试
* goreleaser调试
* rm c
* osx进度精确到毫秒
* update scoop config

### Fix

* goreleaser


<a name="v4.0.0"></a>
## [v4.0.0](https://github.com/go-musicfox/go-musicfox/compare/v3.7.7...v4.0.0) (2023-04-09)

### Add

* purego
* 鼠标事件

### Feat

* 优化
* purego
* 单曲循环以及歌单只有一首歌时不再请求网络
* purego替换
* 替换为purego
* purego替换
* purego替换
* 增加数字操作 & 优化代码
* 喜欢标识
* osx时间进度更换
* 添加最近播放功能 https://github.com/go-musicfox/go-musicfox/issues/63
* 显示是否为喜欢歌曲 https://github.com/go-musicfox/go-musicfox/issues/58
* 单列歌词显示优化
* 代码优化
* 代码优化
* 代码优化
* 滚动歌词
* 副标题滚动优化
* update scoop config
* 添加快进快退功能
* 添加快进快退功能

### Fix

* OSX内存问题
* 内存泄露
* 加载中标题截断问题
* 歌词到最后一行时拖动进度条无法正常更新
* 修复窗口大小变化导致副标题栏截断问题

### Fixme

* beep FLAC格式(其他未测)跳转卡顿

### Update

* Readme


<a name="v3.7.7"></a>
## [v3.7.7](https://github.com/go-musicfox/go-musicfox/compare/v3.7.6...v3.7.7) (2023-03-26)

### Feat

* LyricsX对接优化
* 移除延迟补偿
* 更新脚本
* 优化github action
* 优化github action
* scoop配置
* update scoop config


<a name="v3.7.6"></a>
## [v3.7.6](https://github.com/go-musicfox/go-musicfox/compare/v3.7.5...v3.7.6) (2023-03-25)

### Feat

* update scoop config
* 测试
* 优化
* 优化
* 优化
* 更新github action
* 代码优化
* 优化mac control
* 优化
* scoop调整

### Fix

* github action
* github action
* github action
* github action
* error
* error
* error
* 报错优化
* Windows下载的BUG

### Update

* macdriver
* README
* README


<a name="v3.7.5"></a>
## [v3.7.5](https://github.com/go-musicfox/go-musicfox/compare/v3.7.4...v3.7.5) (2023-03-19)

### Feat

* 进度优化
* 优化
* 优化
* 增加Scoop支持
* add build script
* update build

### Fix

* 修复近期因代码优化引入的mac端内存和带宽占用高的问题
* mac osx player memory leak


<a name="v3.7.4"></a>
## [v3.7.4](https://github.com/go-musicfox/go-musicfox/compare/v3.7.3...v3.7.4) (2023-03-12)

### Fix

* UNM


<a name="v3.7.3"></a>
## [v3.7.3](https://github.com/go-musicfox/go-musicfox/compare/v3.7.2...v3.7.3) (2023-03-09)

### Feat

* mpris使用高分辨率封面
* mpris封面显示

### Fix

* ci
* gitlab action
* gitlab-ci
* gitlab-ci
* gitlab-ci
* gitlab-ci
* timer leak

### Update

* go mod


<a name="v3.7.2"></a>
## [v3.7.2](https://github.com/go-musicfox/go-musicfox/compare/v3.7.1...v3.7.2) (2023-02-19)

### Fix

* 包升级
* 包升级


<a name="v3.7.1"></a>
## [v3.7.1](https://github.com/go-musicfox/go-musicfox/compare/v3.7.0...v3.7.1) (2023-02-14)

### Feat

* 优化

### Fix

* 歌词显示


<a name="v3.7.0"></a>
## [v3.7.0](https://github.com/go-musicfox/go-musicfox/compare/v3.6.1...v3.7.0) (2023-02-11)

### Feat

* 更新版本号
* 增加歌词翻译
* 代码结构优化 & 更新README及帮助菜单
* 优化
* 增加快捷键
* 电台分页优化
* 优化
* 优化
* 预览图更新

### Fix

* archlinux报错


<a name="v3.6.1"></a>
## [v3.6.1](https://github.com/go-musicfox/go-musicfox/compare/v3.6.0...v3.6.1) (2023-01-11)

### Feat

* 调试CGO内存问题
* 测试cgo内存泄露
* 优化
* 优化

### Fix

* beep音量调整


<a name="v3.6.0"></a>
## [v3.6.0](https://github.com/go-musicfox/go-musicfox/compare/v3.5.4...v3.6.0) (2022-12-24)

### Feat

* 增加音量🔊显示
* 下载增加tag信息
* 下载增加tag
* 优化
* 菜单更新
* 帮助菜单更新
* 菜单
* 增加了g和G的键位绑定支持
* 搜索分页
* 增加配置获取歌单下所有歌曲

### Fix

* 双行且奇数列表个数时G跳转位置错误


<a name="v3.5.4"></a>
## [v3.5.4](https://github.com/go-musicfox/go-musicfox/compare/v3.5.3...v3.5.4) (2022-12-22)

### Feat

* 优化
* 增加扫码登录
* 修复登录问题

### Fix

* 登录回车问题


<a name="v3.5.3"></a>
## [v3.5.3](https://github.com/go-musicfox/go-musicfox/compare/v3.5.2...v3.5.3) (2022-12-19)

### Feat

* 更新icon


<a name="v3.5.2"></a>
## [v3.5.2](https://github.com/go-musicfox/go-musicfox/compare/v3.5.1...v3.5.2) (2022-12-18)

### Feat

* 接口报460问题


<a name="v3.5.1"></a>
## [v3.5.1](https://github.com/go-musicfox/go-musicfox/compare/v3.5.0...v3.5.1) (2022-12-18)


<a name="v3.5.0"></a>
## [v3.5.0](https://github.com/go-musicfox/go-musicfox/compare/v3.4.2...v3.5.0) (2022-12-17)

### Feat

* 更新logo、增加beep mp3解码器（默认go-mp3，其他可选项minimp3）
* 更新icon
* 优化Mac退出
* 生成默认配置文件及通知优化
* 支持指定下载目录
* 增加config命令查看配置文件路径
* Mac通知优化，不再依赖外部
* 登录增加国家码支持，格式:+86 18299999999，默认国内
* 增加日志信息

### Stash

* 下载


<a name="v3.4.2"></a>
## [v3.4.2](https://github.com/go-musicfox/go-musicfox/compare/v3.4.1...v3.4.2) (2022-12-02)

### Fix

* 无权限自动切歌


<a name="v3.4.1"></a>
## [v3.4.1](https://github.com/go-musicfox/go-musicfox/compare/v3.4.0...v3.4.1) (2022-11-30)

### Feat

* 优化及修复bug


<a name="v3.4.0"></a>
## [v3.4.0](https://github.com/go-musicfox/go-musicfox/compare/v3.3.3...v3.4.0) (2022-11-28)

### Feat

* 下载优化
* 下载


<a name="v3.3.3"></a>
## [v3.3.3](https://github.com/go-musicfox/go-musicfox/compare/v3.3.2...v3.3.3) (2022-11-27)

### Fix

* 当前播放列表重复进入问题


<a name="v3.3.2"></a>
## [v3.3.2](https://github.com/go-musicfox/go-musicfox/compare/v3.3.1...v3.3.2) (2022-11-17)

### Feat

* 更新获取歌曲URL接口


<a name="v3.3.1"></a>
## [v3.3.1](https://github.com/go-musicfox/go-musicfox/compare/v3.3.0...v3.3.1) (2022-11-14)

### Feat

* 兼容

### Fix

* stop


<a name="v3.3.0"></a>
## [v3.3.0](https://github.com/go-musicfox/go-musicfox/compare/v3.2.2...v3.3.0) (2022-11-14)

### Feat

* linux接入mpris
* 优化linux通知

### Update

* Linux通知优化


<a name="v3.2.2"></a>
## [v3.2.2](https://github.com/go-musicfox/go-musicfox/compare/v3.2.1...v3.2.2) (2022-10-07)

### Update

* 优化搜索


<a name="v3.2.1"></a>
## [v3.2.1](https://github.com/go-musicfox/go-musicfox/compare/v3.2.0...v3.2.1) (2022-10-06)

### Update

* 帮助


<a name="v3.2.0"></a>
## [v3.2.0](https://github.com/go-musicfox/go-musicfox/compare/v3.1.1...v3.2.0) (2022-10-05)

### Feat

* 搜索
* 优化
* 使用libflac解析无损音质


<a name="v3.1.1"></a>
## [v3.1.1](https://github.com/go-musicfox/go-musicfox/compare/v3.1.0...v3.1.1) (2022-10-03)

### Fix

* 私人FM解析问题


<a name="v3.1.0"></a>
## [v3.1.0](https://github.com/go-musicfox/go-musicfox/compare/v3.0.2...v3.1.0) (2022-10-03)

### Feat

* 优化
* 「添加到喜欢」功能优化
* UNM优化
* 优化
* 播放列表


<a name="v3.0.2"></a>
## [v3.0.2](https://github.com/go-musicfox/go-musicfox/compare/v3.0.1...v3.0.2) (2022-09-27)

### Fix

* linux报错


<a name="v3.0.1"></a>
## [v3.0.1](https://github.com/go-musicfox/go-musicfox/compare/v2.2.1...v3.0.1) (2022-09-25)

### Feat

* 优化windows resize
* 优化windows resize
* 优化
* 优化
* 增加Lastfm
* mac增加睡眠回调
* 增加当前播放列表
* 记住音量🔊
* 优化seek
* 优化
* 增加artwork
* support mpd player
* vendor
* github ci
* github ci
* version reject

### Fix

* 内存泄露
* remote center
* 依赖为可选项

### Update

* 增加配置 notifySender
* 增加环境变量MUSICFOX_ROOT 及 通知logo
* 通知icon


<a name="v2.2.1"></a>
## [v2.2.1](https://github.com/go-musicfox/go-musicfox/compare/v2.2.0...v2.2.1) (2022-07-17)

### Fix

* 版本
* 修复登录问题


<a name="v2.2.0"></a>
## [v2.2.0](https://github.com/go-musicfox/go-musicfox/compare/v2.1.1...v2.2.0) (2022-05-29)


<a name="v2.1.1"></a>
## [v2.1.1](https://github.com/go-musicfox/go-musicfox/compare/v2.1.0...v2.1.1) (2022-05-29)

### Feat

* workflow
* update github-action
* goreleaser


<a name="v2.1.0"></a>
## [v2.1.0](https://github.com/go-musicfox/go-musicfox/compare/v2.0.1...v2.1.0) (2021-07-18)


<a name="v2.0.1"></a>
## [v2.0.1](https://github.com/go-musicfox/go-musicfox/compare/v2.0.0...v2.0.1) (2021-05-14)


<a name="v2.0.0"></a>
## v2.0.0 (2021-05-14)

