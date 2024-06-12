
4.5.0
=============
2024-06-12

* fix: goreleaser config (0c095bef)
* fix: download failed(#274) (f3e2c534)
* fix: `UrlMusic` (9fea879a)
* fix: panic nil ptr (8ebe8180)
* feat(mpd): add config item `MpdAutoStart` to contrl whether mpd should be started (b667fbf4)
* fix: refresh login (048092f3)
* chore: update cache (e7803c52)
* chore: fix homebrew (9bfd4572)
* Update go-musicfox.ini (10211c28)
* chore: add linux desktop file (c64747ba)
* docs: 更新了ubuntu编译问题 (7ee81e1f)
* docs: 更新了debian系发行版安装方法 (8a26c13e)
* chore: update deps (3228d555)
* 新建了desktop (0b6f4946)
* fix: unm request (67777c6b)
* Update go-musicfox.ini (fb606585)
* fix: UNM `migu`, `kuwo` (e59fff1c)
* chore: optimize log (b33cf3cc)
* chore: fix ci (69b19b15)
* chore:add ci when pr (97ab6a9b)
* chore: mod tidy & vendor (3fc84ec2)
* feat: support more level & refactor `utils` package (b9580588)
* feat: add more song level (1fab47bb)
* 更新了ubuntu编译问题 (7aa5eed3)
* 更新了debian系发行版安装方法 (690c0250)
* fix: make install error (7939738e)
* docs: update README (cb367118)
* ci: update release.yaml (0be2e771)
* ci: update release.yaml (52e6f51b)
* ci: fix branch match (96f4c1f1)
* update ci (#269) (f28c965f)
* ci: rm unsupported option (#268) (20ba1bea)
* ci: upload more artifacts (7d3d4e5c)
* ci: add deb, rpm... (95b4a30e)
* Update README.md (e7912bb0)
* Update README.md (b1f69a4e)
* optimize: catch panic in beep player (328f4e26)
* doc: add termux install (5e7a4702)
* chore: update README & go.mod (98ea6dce)
* feat: update scoop config (96d27b77)

4.4.1
=============
2024-05-13

* chore: rm useless log (d233096e)
* feat: update version.sh & fix windows control (3e262fbd)
* fix: subscribe_list unresponsive (6211e0cd)
* ci: support env `RELEASE_TAG`,`GIT_REVISION` (9e007943)
* feat: update scoop config (37263a0f)

4.4.0
=============
2024-05-12

* chore: rename playing info `AsText` (ff135fa2)
* feat: append all songs of playlist to now playing list (#232) (5837b0dc)
* feat: impl xesam:asText (#202) (0edfcf8b)
* fix: notificator panic (c71521e4)
* refactor: project layout & fix some bugs (54f8cb44)
* chore: format (f0892131)
* fix(remote_control): linux mpris (72bd99b5)
* fix: smtc with beep (63a229d8)
* feat: media player & smtc (87d21367)
* chore: optimize for ambiguity (20e63a88)
* update: path.Join to filepath.Join (71acf30e)
* update: change to utils (e095c776)
* fix: opening the QR code image fails when the MUSICFOX_ROOT environment variable is set in Windows (abf1a06a)
* optimize: media player command manager (17b5ffc4)
* feat: add win media player #246 (c5939128)
* fix: upgrade deps (fcf7686f)
* feat: update win media player (6e45899e)
* mod: 函数名和注释 (b9a2c5c1)
* feat: #196 (894b18eb)
* feat(keys): #248 (87db3fac)
* feat: perfect windows media player (ca23f453)
* feat: update win player (3ba594b8)
* Update README.md (0a207066)
* update readme for linux building and troubleshooting (cde265be)
* fix: defer ignore (3a72db2e)
* chore(player): comment uncompleted code (508b580f)
* chore(player): comment uncompleted code (bd03d1de)
* ci(build): update build sh (f3728b5d)
* ci: add lint (e97e0f12)
* feat(player): add windows media player (da4d48b8)
* feat(version): add build tags to version (f9afd6c4)
* Update README.md (ebf90e81)
* Update README.md (19396a78)
* fix(beep): panic when player (0e7cfbf5)
* fix(beep): fix deadlock (bfaa5cf2)
* feat(homebrew): add --head install (35337988)
* upgrade: upgrade bubbletea to latest (312413ca)
* chore(log): add debug log (78bbf470)
* chore(mod): rm useless dep (ac850283)
* chore(mod): tidy go mod (a250ead9)
* refactor(beep): update beep to latest (with `oto`) (0024f5e6)
* docs(keymap): update keymap doc (4042f095)
* feat: update scoop config (45fa4e17)

4.3.3
=============
2024-03-31

* chore: update CHANGELOG (e6d65829)
* fix(macdriver): upgrade `purego` to fix stuck in macOS (7a8ccbc3)
* chore: mod tidy (c64c4853)
* docs(CHANGELOG): update CHANGELOG (d0190a11)
* feat: update scoop config (2d1c2406)

4.3.2
=============
2024-03-30

* feat: add goreleaser changelog (b49d3535)
* feat: add CHANGELOG generator (1aa26469)
* update goreleaser (bd705ea2)
* fix: fix faulty position (8d017e01)
* feat: update scoop config (a92a0f41)
* fix: 每日推荐歌单 (eac333fc)
* fix: use `AppName` in `internal.types` (9ec40ea8)
* fix: set mpris Identity an user-friendly name (5541b1f4)
* optimize: 默认关闭签到功能 (75ea70d7)
* build(deps): bump golang.org/x/crypto from 0.10.0 to 0.17.0 (64a8f797)
* fix: linux dbus `s.props` is nil (ff462888)
* fix: cannot load framework in Mac Sonoma (b6f63984)
* feat: add vscode launch config (6654436a)
* fix(utils): music cache priority (4130389e)
* fix(utils): music cache priority (da33abbb)
* feat: foxful-cli ticker (d5852352)
* feat: update scoop config (fcb13414)
* optimize: only `W` for logout(#215) (605d9c81)
* fix: panic when parse lrc(#240) (65bb5950)
* feat: page up and page down(#241) (ff95dd3d)
* feat: add goreleaser changelog (957e3346)
* feat: add CHANGELOG generator (e8e723e8)
* update goreleaser (a31ceea7)
* fix: fix faulty position (fb7e4865)
* feat: update scoop config (c0a7513e)

4.3.1
=============
2024-02-24

* fix: 每日推荐歌单 (608f2d67)
* fix: use `AppName` in `internal.types` (6d4fe1a3)
* fix: set mpris Identity an user-friendly name (a44d0c2e)
* optimize: 默认关闭签到功能 (3a49acc4)
* build(deps): bump golang.org/x/crypto from 0.10.0 to 0.17.0 (1e05b6a9)
* fix: linux dbus `s.props` is nil (af433e02)
* Update event_handler.go (cd2f4b9b)
* fix: cannot load framework in Mac Sonoma (3892ffa3)
* feat: add vscode launch config (31d0122b)
* fix(utils): music cache priority (041c9f06)
* fix(utils): music cache priority (4ceb7d93)
* feat: foxful-cli ticker (7368a25b)
* feat: update scoop config (9810f207)

4.3.0
=============
2023-10-30

* fix: missing import (9bf9b11d)
* improvement(build): omit `hostname` error message in `version.sh` (fd1c6190)
* fix(lyric): fix panic when no lyric (4692ee47)
* feat(player): provide xseam:asText for mpris (4ba8d3fa)
* fix(state_handler): update playing position every 200ms (b69ae5dc)
* fix iss203 (f1158cd2)
* feat: update scoop config (6d18b6ec)

4.2.1
=============
2023-10-08

* fix: #199 (ab0862e8)
* feat: add confirm menu when clear song cache (55807f5b)
* Update go-musicfox.ini (a6dea574)
* Update README.md (65a41ffc)
* feat: update scoop config (a24af5d4)

4.2.0
=============
2023-09-17

* fix: register global_key (9b1db07a)
* fix: github action (12607463)
* fix: github action (7c2a102f)
* fix: global hotkey build (d9dfe6ed)
* feat: 自定义下载文件名格式(#193) (9632ffc9)
* feat: 进度条配置优化 & 支持全局快捷键 (f7cb76f6)
* feat: update scoop config (e1f735c4)

4.1.7
=============
2023-09-12

* fix: err (8c5d65b9)
* fix: mac notification (9a99ba17)
* Update README.md (977d5cb5)
* Update README.md (1c7f344b)
* Update README.md (94cc971e)
* Update README.md (47622d9f)
* fix: github action (e10495a9)
* fix: inject (c0059834)
* fix: err (3eaf9d3f)
* update foxful-cli (9161738e)
* refactor (bd131ed1)
* fix: lyrics view (ac0c7e77)
* refactor: pkg => internal (046d7315)
* optimize: lyrics (3b6aaef8)
* fix: player view (b16ebae8)
* optimize (fa4a9842)
* fix: playlists (d65862fb)
* optimize: loading (c80fb990)
* optimize: login callback (8765c96d)
* optimize: login callback (b959aa49)
* fix: w not quit (a84e25cf)
* fix: #183 (1c85e032)
* Update Makefile (7940c652)
* Update release.yaml (4ed85894)
* fix: beep minimp3 err (520c0313)
* fix: key (43f517f6)
* update depend (80e7d6fa)
* refactor by foxful-cli (b134daea)
* update (06645142)
* add patch (0840bf58)
* gitignore (786fe827)
* fix(utils): 当无法正常下载音乐且开启缓存时panic (a6af3bb6)
* feat: update scoop config (cf50bfe1)

4.1.6
=============
2023-08-21

* fix: lyric (6fdef846)
* feat: update scoop config (267dc710)

4.1.5
=============
2023-08-20

* fix: #167 (ae6eb437)
* fix(player): 修复了当mpd repeat为true时不能自动切歌的问题 (a6cbb3f6)
* feat: golang-lint (8373239a)
* feat: golang-lint (755a8f37)
* optimize: replace github.com/boltdb/bolt => go.etcd.io/bbolt (4706f989)
* fix: 修复播放列表切片越界的问题 (4db9962b)
* fix: #171 (025fa87b)
* docs(README): 新增快捷键项目，对齐MarkDown表格 (e02eb309)
* feat: update scoop config (44101318)

4.1.4
=============
2023-07-16

* fix: actions (57eab0e4)
* fix: actions (1a6280ae)
* fix: actions (94fe01d3)
* fix: actions (7abe38ac)
* fix: actions (7981c4a6)
* fix: actions (a486aa95)
* fix: actions (ecb634c7)
* fix: actions (6161d8d7)
* fix: actions (3dcf0534)
* fix: actions (e5faa4ef)
* fix: actions (64026757)
* fix: winget (db934eaa)
* fix: winget (f1bf252e)
* feat: 模糊搜索本地歌单 (932046da)
* fix: #158 setting FLAC file will not halt; use tmp file (cd3fe90c)
* feat: 添加了下载歌曲失败时的通知提醒 (e5ab571f)
* fix(utils.go): 修复了下载未缓存文件时文件名错误的问题 (7ae9c9a5)
* fix: #149 again (0b9c850c)
* feat: add winget (6a02f9ef)
* fix: error (7923b2da)
* fix: #151 #152 (ae9f57d8)
* feat: 增加winget配置（待测试） (610d8de4)
* fix: #149 (b56e8838)
* fix: 添加清除音乐缓存在帮助目录的条目 (6f831c4f)
* feat: update scoop config (f9444422)

4.1.3
=============
2023-07-11

* fix: 兼容#147 (f062a07b)
* feat: mpd volume日志 (2f32ac5b)
* fix: #143 (6d5f3c10)
* fix: panic (5e07e445)
* update (1fd8bfb6)
* feat: 优化 (61392e40)
* fix: bcef056后无法构建 (6ae05a54)
* feat: 优化ctrl (b4e431ef)
* add: Windows 检查更新提示 (b864a819)
* fix: 在播放模式为单曲循环时手动切换歌曲 (7ce4c651)
* fix: (a7ec978a)
* feat: 将歌曲添加至歌单 (8ae13370)
* fix: mac图片错误时panic (d4fb7435)
* feat: update pkg (73d68982)
* fix: 歌单详情 (a0b80e23)
* optimized code (e9fc158a)
* optimized code (d1e07ba0)
* refactor (581171d2)
* feat: #128 (09c399c4)
* Enhancement: Send music info to mpd (a19d6723)
* fix: like song (a5098fe0)
* fix: like selected song (089afb56)
* feat: update dep (2f4395da)
* update (3c86fd2a)
* feat: 配置命令优化 (c1605dc7)
* chore: Move dir:`nix` to `deploy/nix` (cc8c4813)
* Add flake support (b987dae7)
* Add flake support (275ce3cc)
* fix: stack overflow (341d5936)
* fix: 从播放列表删除选中歌曲 (b559ba7e)
* feat: 从播放列表删除选中歌曲 (eadf7500)
* feat: update scoop config (524a011c)

4.1.2
=============
2023-06-06

* fix 逻辑错误，修复后续Access is denied错误 (b35c884d)
* fix: 修复配置文件目录问题 (ccf4885c)
* Update README.md (0e7d86ce)
* fix: package name (634a2854)
* feat: Add flake support (7bfab6b5)
* feat: update scoop config (f8f091ad)

4.1.1
=============
2023-05-27

* fix: 报错 (6bb739e1)
* feat: update scoop config (c2dff1f8)

4.1.0
=============
2023-05-24

* feat: support #85 (df5a6053)
* feat: update README (0c8f5d1b)
* refactor: polish README.md (398796e3)
* feat: support the base dir specifications of the platforms (#107) (e97019be)
* fix: #105 (d63897e3)
* Fix typo in UI (eacd3284)
* Update player.go (fe9e5bca)
* fix: 播放结束上报网易云 (a254a2ed)
* add: 播放结束上报网易云 (b11b8d3b)
* feat: update scoop config (eb24d0af)

4.0.6
=============
2023-05-02

* fix: 快退会将秒数退为负数 (db429bd9)
* feat: 配置 (7552594a)
* feat: update Readme (5a4e66ad)
* feat: 自定义通知logo(linux/windows) (d53cb82b)
* Update README.md (a294bde6)
* feat: update scoop config (e4ee2f8a)

4.0.5
=============
2023-04-21

* feat: 更新issue template (db136399)
* feat: 更新issue template (f30f1164)
* feat: 更新issue template (42873ce2)
* feat: 更新issue template (f3d6a4c5)
* feat: 更新issue template (30aadf89)
* Update issue templates (5a52eb90)
* feat: 更新go-mp3 (7721c8e7)
* feat: 更新go-mp3 (15616a2c)
* feat: update scoop config (7ae8df35)

4.0.4
=============
2023-04-20

* fix: oto linux panic (4e89aa41)
* feat: 更新libflac (f2b22b75)
* feat: 优化libflac (f0fa2b03)
* feat: 增加等待次数 (3bdc9ced)
* feat: 优化beep_player因网速慢切歌问题 (5cdbbdce)
* fix: 渐变色 (68664d2d)
* feat: update scoop config (be73b159)

4.0.3
=============
2023-04-14

* feat: github action (6f621c9d)
* feat: update scoop config (dc9d4360)
* Update README.md (bad968ee)
* feat: 配置进度条 (a62fac1e)
* fix: panic: slice bounds out of range [:33] with capacity 32 (8dde7614)
* feat: 增加鼠标配置 (d920d28d)
* feat: update scoop config (c78f3a77)

4.0.2
=============
2023-04-12

* feat: minimp3优化 (ff81a020)
* feat: 优化 (938cb857)
* feat: 代码优化 (90352f20)
* fix: 重载后概率导致当前歌曲停止 && Position()偶尔panic (9fc95e05)
* feat: 增加构建 (8dedbdc3)
* feat: update libflac (eff98519)
* feat: 登陆后更新like list (b45ce39a)
* feat: update scoop config (ca56c329)

4.0.1
=============
2023-04-10

* feat: 增加linux arm64及windows arm64构建 (36bb8bc0)
* fix: goreleaser (86596485)
* feat: goreleaser调试 (e909d21a)
* feat: goreleaser调试 (012a8a6a)
* feat: goreleaser调试 (73034d57)
* feat: goreleaser优化 (15e3d25c)
* feat: 优化 (5adfce23)
* feat: goreleaser调试 (03d53a17)
* feat: goreleaser调试 (89e2a819)
* feat: rm c (e993081a)
* feat: osx进度精确到毫秒 (3c685211)
* feat: update scoop config (103d2adf)

4.0.0
=============
2023-04-09

* feat: 优化 (4c5a7476)
* fix: OSX内存问题 (baa29575)
* feat: purego (910e7271)
* fix: 内存泄露 (f5a9b8a6)
* feat: 单曲循环以及歌单只有一首歌时不再请求网络 (488573ff)
* feat: purego替换 (dd4ca178)
* feat: 替换为purego (b3168dbc)
* feat: purego替换 (7a3ac397)
* feat: purego替换 (d3aff7f8)
* feat: 增加数字操作 & 优化代码 (9b7fbd1b)
* feat: 喜欢标识 (b082145c)
* add: purego (95505334)
* feat: osx时间进度更换 (426ff034)
* feat: 添加最近播放功能 https://github.com/go-musicfox/go-musicfox/issues/63 (26dfe05f)
* feat: 显示是否为喜欢歌曲 https://github.com/go-musicfox/go-musicfox/issues/58 (f5fd60f4)
* fix: 加载中标题截断问题 (9f9d45f4)
* fix: 歌词到最后一行时拖动进度条无法正常更新 (a73af781)
* fix: 修复窗口大小变化导致副标题栏截断问题 (54cfd030)
* feat: 单列歌词显示优化 (0a04861a)
* feat: 代码优化 (a5b658e2)
* feat: 代码优化 (ed20f51c)
* feat: 代码优化 (f7fa96c2)
* update: Readme (9b4d42f4)
* add: 鼠标事件 (9a485856)
* fixme: beep FLAC格式(其他未测)跳转卡顿 (8aad39fd)
* feat: 滚动歌词 (6784362d)
* feat: 副标题滚动优化 (15119836)
* update bubbletea (ba0d7e37)
* update README (535f9009)
* feat: update scoop config (33267d35)

3.7.7
=============
2023-03-26

* update README (edf883df)
* feat: LyricsX对接优化 (4b1df028)
* feat: 移除延迟补偿 (547c7ccd)
* feat: 更新脚本 (fb6ee77a)
* feat: 优化github action (b166c8a1)
* feat: 优化github action (e9f746b5)
* feat: scoop配置 (3ad6ae67)
* feat: update scoop config (dd36ce75)

3.7.6
=============
2023-03-25

* fix: github action (c6f040d4)
* fix: github action (34f0b8b4)
* fix: github action (17232228)
* Update go-musicfox.json (5770a3aa)
* update (f5e7dafb)
* fix: github action (a96daea6)
* feat: update scoop config (6587359a)
* feat: 测试 (f2e00afe)
* feat: 优化 (5be610c0)
* feat: 优化 (922cd36e)
* feat: 优化 (0f7ba737)
* fix: error (ad996366)
* fix: error (0d2dc64f)
* fix: error (180bfe25)
* feat: 更新github action (ba603fdd)
* Update go-musicfox.json (5faf267f)
* feat: 代码优化 (5617c83b)
* fix: 报错优化 (3af02c8d)
* update: macdriver (3e8eb432)
* feat: 优化mac control (f4918861)
* feat: 优化 (d0184c5f)
* fix: Windows下载的BUG (b1406501)
* add 歌曲列表菜单副标题内容超出显示宽度时滚动显示 (17d558c5)
* Update README.md (f067d76d)
* update: README (e9d05fb0)
* update: README (18281219)
* feat: scoop调整 (26308eab)
* Scoop update for go-musicfox version v3.7.5 (fe982b85)

3.7.5
=============
2023-03-19

* feat: 进度优化 (4a2b6fb1)
* fix: 修复近期因代码优化引入的mac端内存和带宽占用高的问题 (9944d5d3)
* feat: 优化 (35e40821)
* feat: 优化 (50d2b5d3)
* fix: mac osx player memory leak (deffa16b)
* feat: 增加Scoop支持 (891ead87)
* feat: add build script (e9e14e02)
* feat: update build (19169a39)
* 修改NixOS安装方式 (4ba4a0f2)

3.7.4
=============
2023-03-12

* fix: UNM (f04fcebc)

3.7.3
=============
2023-03-09

* fix: ci (0ea6da26)
* fix: gitlab action (41a91741)
* fix: gitlab-ci (a707d6f6)
* fix: gitlab-ci (c1cacb70)
* fix: gitlab-ci (2c289054)
* fix: gitlab-ci (ce78b506)
* update: go mod (f6fc0341)
* build(deps): bump github.com/gookit/goutil from 0.5.2 to 0.6.0 (d858d406)
* fix: timer leak (9b95b0f1)
* Update README.md (81600cce)
* 添加了 Void Linux 的安装方式 (1230a08d)
* feat: mpris使用高分辨率封面 (ebf39962)
* feat: mpris封面显示 (4bdb26d9)
* Update README.md (4505ca35)

3.7.2
=============
2023-02-19

* fix: 包升级 (7ed5207e)
* fix: 包升级 (0c79ec0b)

3.7.1
=============
2023-02-14

* feat: 优化 (44a6fc97)
* fix: 歌词显示 (a5339f53)

3.7.0
=============
2023-02-11

* feat: 更新版本号 (ef71d38d)
* feat: 增加歌词翻译 (393af877)
* fix: archlinux报错 (bfb7854c)
* feat: 代码结构优化 & 更新README及帮助菜单 (03e6f67a)
* feat: 优化 (17e8dccc)
* feat: 增加快捷键 (49e9fd2d)
* feat: 电台分页优化 (32e26739)
* feat: 优化 (f9eb14e8)
* Update README.md (f1a47ab3)
* feat: 优化 (7002706f)
* 添加NUR安装方式 (dfd874f6)
* feat: 预览图更新 (870b9e1f)

3.6.1
=============
2023-01-11

* fix: beep音量调整 (df236764)
* feat: 调试CGO内存问题 (27822c11)
* feat: 测试cgo内存泄露 (179557b4)
* Update README.md (e85509e6)
* Update README.md (591d1185)
* Update README.md (7770cb36)
* feat: 优化 (9e4581cc)
* feat: 优化 (6c651060)

3.6.0
=============
2022-12-24

* feat: 增加音量🔊显示 (427806f9)
* feat: 下载增加tag信息 (59cf9255)
* feat: 下载增加tag (620633d1)
* feat: 优化 (82def2b3)
* fix: 双行且奇数列表个数时G跳转位置错误 (b5c74e58)
* feat: 菜单更新 (2b2eb5e1)
* feat: 帮助菜单更新 (d76b10a1)
* feat: 菜单 (b9cd3389)
* feat: 增加了g和G的键位绑定支持 (19de53c4)
* feat: 搜索分页 (fd4d3dfd)
* feat: 增加配置获取歌单下所有歌曲 (f361aa6e)

3.5.4
=============
2022-12-22

* feat: 优化 (6d66a58a)
* fix: 登录回车问题 (de7e324e)
* feat: 增加扫码登录 (d3b3610d)
* feat: 修复登录问题 (53ce2bab)

3.5.3
=============
2022-12-19

* update logo (088ee249)
* update logo (1c5a69cd)
* feat: 更新icon (767b6737)

3.5.2
=============
2022-12-18

* feat: 接口报460问题 (f6814d0a)

3.5.1
=============
2022-12-17

* update logo (0f15ef3b)

3.5.0
=============
2022-12-17

* update README (0a290bda)
* feat: 更新logo、增加beep mp3解码器（默认go-mp3，其他可选项minimp3） (6943d678)
* feat: 更新icon (e69b14e1)
* Update README.md (651cb91c)
* Fix libFLAC.so.8 dependency in AUR (a1d523da)
* stash: 下载 (d97440c3)
* feat: 优化Mac退出 (ff57f3b5)
* 优化 (e8f0d483)
* feat: 生成默认配置文件及通知优化 (46d4598b)
* feat: 支持指定下载目录 (d10498fc)
* feat: 增加config命令查看配置文件路径 (c9978211)
* feat: Mac通知优化，不再依赖外部 (85e78ea1)
* feat: 登录增加国家码支持，格式:+86 18299999999，默认国内 (1fc3bec9)
* feat: 增加日志信息 (35493d8c)

3.4.2
=============
2022-12-02

* fix: 无权限自动切歌 (d8c87c27)
* Update README.md (5ca509ea)
* Update README.md (ac2b1ad4)

3.4.1
=============
2022-11-30

* feat: 优化及修复bug (4746e3a1)

3.4.0
=============
2022-11-28

* feat: 下载优化 (43d64111)
* Update README.md (8047928b)
* feat: 下载 (d9bc884c)

3.3.3
=============
2022-11-27

* fix: 当前播放列表重复进入问题 (42ab2289)

3.3.2
=============
2022-11-17

* Update README.md (dd1dd7de)
* feat: 更新获取歌曲URL接口 (ffc8a4d1)

3.3.1
=============
2022-11-14

* feat: 兼容 (78f21af9)
* fix: stop (0be4cec0)

3.3.0
=============
2022-11-14

* feat: linux接入mpris (4d6504d6)
* feat: 优化linux通知 (59691d63)
* update: Linux通知优化 (a7ae4eb2)
* Update README.md (31b2753c)
* Update README.md (96151a1f)
* 修改Arch和Gentoo的安装方式 (a8f99229)

3.2.2
=============
2022-10-07

* update: 优化搜索 (bb8efeac)

3.2.1
=============
2022-10-06

* update: 帮助 (e679a657)

3.2.0
=============
2022-10-05

* feat: 搜索 (7d64fdc0)
* update (c96d4073)
* update github action (8e2a1ccd)
* update github action (c1d15cff)
* update github action (c9362982)
* update github action (4aa0bbf4)
* update github action (cb1b24a7)
* update github action (c0e57f83)
* update github action (0c6d0ceb)
* update flac (e5ca7bdd)
* update flac (a030698a)
* fix github action (8e20def6)
* Update .goreleaser.yaml (3fcc1eef)
* Update .goreleaser.yaml (db0398a3)
* Update .goreleaser.yaml (e599de82)
* Update .goreleaser.yaml (08634ff1)
* Update .goreleaser.yaml (6a8c3834)
* Update .goreleaser.yaml (66e0c4bb)
* Update .goreleaser.yaml (1f5c6cc3)
* Update .goreleaser.yaml (6821a120)
* Update .goreleaser.yaml (49636fca)
* Update .goreleaser.yaml (a6f1b69a)
* fix github action (76fedcc7)
* feat: 优化 (23f7f1ba)
* feat: 使用libflac解析无损音质 (1d3ff9ac)

3.1.1
=============
2022-10-03

* fix: 私人FM解析问题 (16a61e17)

3.1.0
=============
2022-10-03

* feat: 优化 (220ba865)
* feat: 「添加到喜欢」功能优化 (bd351a76)
* Update README.md (21a3a955)
* Update README.md (369b4ebc)
* Update README.md (bbf8cd05)
* feat: UNM优化 (0b6fc1bb)
* update (353e310b)
* feat: 优化 (50ff212c)
* feat: 播放列表 (c407ea2e)
* Update README.md (709f2bf1)
* Update README.md (947084ef)
* Update README.md (1b2a13ff)

3.0.2
=============
2022-09-27

* fix: linux报错 (b0388799)
* update mac优化 (00e88023)
* update runewidth (e2420b10)
* Update README.md (3355ef35)
* update (8bfc2373)
* Update README.md (9f4b3521)
* Update README.md (9e574134)
* Update README.md (ec05da90)
* Update README.md (27923ed2)

3.0.1
=============
2022-09-25

* fix gitlab cd (0fa74c4b)
* fix gitlab cd (072fe53e)
* fix gitlab cd (4746e0c3)
* fix gitlab cd (253265f9)
* fix (70fbb946)
* update gitlab-ci (fe70de36)
* Update README.md (4bba56b7)
* update README (fcf8a94d)
* feat: 优化windows resize (16a95cfa)
* feat: 优化windows resize (34bfbbb4)
* feat: 优化 (6b893d0a)
* feat: 优化 (f1e9cf85)
* feat: 增加Lastfm (cf4e94cd)
* feat: mac增加睡眠回调 (e2f9416d)
* update (35196707)
* feat: 增加当前播放列表 (04816c26)
* fix warning (5e635394)
* fix: 内存泄露 (42bd4aa8)
* fix (0555332d)
* update (0cfd27f5)
* feat: 记住音量🔊 (849e6aec)
* update (620da236)
* fix bug (29c14018)
* add. AVPlayer (09add3fd)
* feat: 优化seek (3072a983)
* fix (36e3a543)
* feat: 优化 (fce6af62)
* feat: 增加artwork (19a719d6)
* feat: support mpd player (6f7581a0)
* update. minimp3 (c19c9fcf)
* fix player listen (39141a64)
* fix: remote center (ce681155)
* fix timer error when next song (709d24b8)
* fix some bugs (6bb44b4d)
* update go version (1a1970a8)
* fix gc panic and data race (495b2f4e)
* feat: vendor (c86a6b7a)
* update (4c89701e)
* refactor (01680c72)
* feat: github ci (afbb71bd)
* feat: github ci (b3d1893e)
* test (cccdfcf7)
* update (d0301776)
* update (63f29597)
* update (bafca22b)
* update (3a90d8e1)
* stash (82de75ea)
* update (79468b65)
* update: 增加配置 notifySender (13fc9604)
* update: 增加环境变量MUSICFOX_ROOT 及 通知logo (bab4dd88)
* update: 通知icon (5c6a7f70)
* fix: 依赖为可选项 (f688d769)
* update (31691864)
* update (3f9b0a77)
* update. github ci (de29c1f3)
* feat: version reject (f31178f7)
* refactor (31d840a8)

2.2.1
=============
2022-07-17

* fix: 版本 (7d74f6e7)
* fix: 修复登录问题 (8246036d)
* Update README.md (b6099d40)
* update. README (e66c4c13)

2.2.0
=============
2022-05-29

* fix. homebrew install (71e01613)

2.1.1
=============
2022-05-29

* feat. update ignore (c6a50a67)
* feat. github workflow (3dcf0fd3)
* feat. github workflow (aa0ceb4a)
* feat. github workflow (bab2a705)
* feat. github workflow (1fdef76b)
* feat: workflow (9288c38f)
* feat. github workflow (c7986024)
* feat. github workflow (c2001718)
* feat: update github-action (a6fd45db)
* feat. github workflow (b68aa014)
* feat. goreleaser (7894ef70)
* feat: goreleaser (af631bb1)
* Update build.yml (23fada7c)
* Update build.yml (6f1db592)
* Update build.yml (a3b1be3f)
* Update build.yml (383aa190)
* Create build.yml (7f7969ed)
* feat. 更新版本 (9ea08db0)
* feat. update README (afd418d5)
* feat. 配置变更 (fcab6963)
* feat. 音乐缓存到文件 (359fdb97)
* update. 支持更多符号 (d88eed73)
* update. 允许全角符号 resolve #9 (a2ee9db1)
* Update README.md (c857c30d)
* update (10dd1375)
* update. (ed658d3e)
* Update README.md (e97021c3)
* update. 检查更新 (d08c947c)

2.1.0
=============
2021-07-18

* update. (c39ff782)
* update. (37f77b29)
* update. (a37fe73e)
* update. (e68e7a60)
* update. (a04930ea)
* update. 增加通知 (71ea3072)
* update. 增加歌词显示、签到配置 (87afd8d9)
* update. 新增主题色设置 (aa040a0c)
* update. 新增欢迎语配置 (05818093)
* update. 新增配置 (cd2c717d)
* add. linux (4d361222)
* Update README.md (9f446a53)

2.0.1
=============
2021-05-14

* fix. 若干Bug (33b8743c)
* fix. 若干Bug (7809d33a)
* fix. 私人FM (7efe8526)
* Update README.md (b0dcf420)
* Update README.md (1abdb447)
* Update README.md (05ccf84d)
* Update README.md (8b0073bb)
* Update README.md (7a65b17e)
* Update README.md (e7670503)

2.0.0
=============
2021-05-14

* update. (f7049608)
* add. README (94fef1e4)
* add. 版本更新 (f20493d3)
* add. 主播电台 (a638338f)
* add. 帮助 (2ecf6525)
* add. 电台、精选歌单 (186f777a)
* complete. 搜索 (a241b1a8)
* add. so many (1de84fe1)
* add. 专辑菜单; fix. 播放 (2cf8d6a5)
* fix. 若干Bug (dbf5d9d1)
* fix. 播放列表显示Bug (76b9057b)
* add. 播放模式、同步到本地 (46c6c852)
* add. 心动模式 (57e97f3f)
* fix. 登录UI (d75e07ef)
* fix. ui (ac5073bd)
* update. ui (f7cd2801)
* update. ui (162bc54a)
* update. (b9e5c3d0)
* update. (f27aec19)
* update. (7377cba5)
* update. 歌词更新 (2f294b34)
* add. 歌词 (668b1162)
* update. fm下一曲 (5791f08e)
* update. (e3dbf3c3)
* update. 菜单 (5ed6ebec)
* update. flac (fb7276e4)
* add. 用户登录信息、用户歌单 (170f4e0c)
* add. 每日推荐歌单 (432ef986)
* update. 上下曲 (8c02d6d6)
* fix. 菜单未对齐 (4fe70924)
* fix. 菜单未对齐 (d99e0ebe)
* update. (ea1c325e)
* add. (60db5fc6)
* fix. 协程泄漏 (9a0db97c)
* update. (0764b316)
* add. player (bda5547c)
* update. (9e40ce64)
* add. 每日推荐 (53957dd8)
* add. 登录校验 (9998e4b9)
* update. 菜单 (8026bbf1)
* update. 依赖 (4d7b25ee)
* update (ca37b30b)
* update. 优化 (edb9090c)
* update. 优化菜单UI (006b9579)
* fix. 清屏 (df12338a)
* add. 菜单返回 (7f4dd0e7)
* update. menu (ff71c936)
* add. enter menu (af42f3fe)
* update. 优化UI (2feb5e29)
* update. (21e40b99)
* update. menu (75cb6d42)
* fix. 菜单翻页 (78109d2f)
* update. (f9045565)
* add. loading (b777678e)
* update. menu (f149dea2)
* update. menu (7dc4c52a)
* update. tab => space (94aa1475)
* update (2d38b5ad)
* update (cfc4d2f2)
* add. title (4b13fd80)
* update. (2d4604c2)
* update. (5bbd513d)
* add. menu interface (9cc6421e)
* update. progress width (899f713a)
* reformat (fa46a64e)
* add. main ui entry (937f64cf)
* complete. startup ui (eee62ff6)
* update. start up (32c2e5f2)
* init (6454db4c)
* init (241a85e8)


