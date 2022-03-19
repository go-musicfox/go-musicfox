# go-musicfox

**给个star✨吧**

go-musicfox是 [musicfox](https://github.com/anhoder/musicfox) 的重写版，为了解决某些问题，提升体验，因此采用go进行重写。*go初学者，顺便熟悉go，哈哈RwR*

> UI基于 [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) ，做了一些定制

![GitHub repo size](https://img.shields.io/github/repo-size/anhoder/go-musicfox) ![GitHub](https://img.shields.io/github/license/anhoder/go-musicfox) ![Last Tag](https://badgen.net/github/tag/anhoder/go-musicfox) ![GitHub last commit](https://badgen.net/github/last-commit/anhoder/go-musicfox) ![GitHub All Releases](https://img.shields.io/github/downloads/anhoder/go-musicfox/total)

![GitHub stars](https://img.shields.io/github/stars/anhoder/go-musicfox?style=social) ![GitHub forks](https://img.shields.io/github/forks/anhoder/go-musicfox?style=social)

## 最近更新（v2.1.0）

### 增加配置文件

配置文件路径为用户目录下的.go-musicfox/go-musicfox.ini，相关配置有：

```ini
# 启动页配置
[startup]
# 是否显示启动页
show=true
# 启动页进度条是否有回弹效果
progressOutBounce=true
# 启动页时长
loadingSeconds=2
# 启动页欢迎语，支持字母、数字、部分英文字符
# welcome=welcome!
welcome=musicfox
# 启动时自动签到
signin=true

# 进度条配置
[progress]
# 进度条已加载字符
fullChar=#
# 进度条未加载字符
emptyChar=

# 主页面配置
[main]
# 是否显示标题
showTitle=true
# 加载中提示
loadingText=[加载中...]
# 歌曲码率，128000, 320000...，视网络情况而定
songBr=999000
# 主题颜色，0-256或#233445
primaryColor=random
# 是否显示歌词
showLyric=true
# 是否显示通知信息
showNotify=true
```

### 增加通知

目前触发通知的操作有：
* 播放歌曲
* 自动签到成功
* 添加歌曲到喜欢
* 从喜欢移除
* 标记为不喜欢
* 登出


## 预览

![欢迎界面](preview/img.png)
![主界面1](preview/img1.png)
![主界面2](preview/img2.png)

## 安装

### Mac

提供两种方式安装：

#### 使用brew安装

```sh
brew tap anhoder/go-musicfox && brew install go-musicfox
```

如果你之前安装过musicfox，需要使用下列命令重新链接:

```sh
brew unlink musicfox && brew link --overwrite go-musicfox
```

#### 直接下载

下载[Mac可执行文件](https://github.com/anhoder/go-musicfox/releases/latest/download/musicfox.mac)，在iTerm或Terminal中打开

### Linux

下载Linux可执行文件: [amd64版](https://github.com/anhoder/go-musicfox/releases/download/v2.0.1/musicfox.linux_amd64), [arm64版](https://github.com/anhoder/go-musicfox/releases/download/v2.0.1/musicfox.linux_arm64)，在命令行中运行即可。

ArchLinux可使用AUR安装：
```sh
yay -S go-musicfox-bin
```

### Windows

下载[Windows可执行文件](https://github.com/anhoder/go-musicfox/releases/latest/download/musicfox.exe)，在命令行中运行即可。

> **推荐使用Windows Terminal，UI及体验好很多**

## 使用

```sh
$ musicfox
```

| 按键 | 作用 | 备注 |
| --- | --- | --- |
| h/H/LEFT | 左 |  |
| l/L/RIGHT | 右 |  |
| k/K/UP | 上 |  |
| j/J/DOWN | 下 | |
| q/Q | 退出 | |
| space | 暂停/播放 | |
| [ | 上一曲 | |
| ] | 下一曲 | |
| - | 减小音量 | |
| = | 加大音量 | |
| n/N/ENTER | 进入选中的菜单 | |
| b/B/ESC | 返回上级菜单 | |
| w/W | 退出并退出登录 | |
| p | 切换播放方式 | |
| P | 心动模式(仅在歌单中时有效) | |
| r/R | 重新渲染UI | Windows调整窗口大小后，没有事件触发，可以使用该方法手动重新渲染 |
| , | 喜欢当前播放歌曲 | |
| < | 喜欢当前选中歌曲 | |
| . | 当前播放歌曲移除出喜欢 | |
| > | 当前选中歌曲移除出喜欢 | |
| / | 标记当前播放歌曲为不喜欢 | |
| ? | 标记当前选中歌曲为不喜欢 | |

## TODO

* [x] 我的歌单
* [x] 每日推荐歌曲
* [x] 每日推荐歌单
* [x] 私人FM
* [x] 歌词显示
* [x] 欢迎界面
* [x] 搜索
    * [x] 按歌曲
    * [x] 按歌手
    * [x] 按歌词
    * [x] 按歌单
    * [x] 按专辑
    * [x] 按用户
* [x] 排行榜
* [x] 精选歌单
* [x] 最新专辑
* [x] 热门歌手
* [x] 云盘
* [x] 播放方式切换
* [x] 喜欢/取消喜欢
* [x] 心动模式/智能模式
* [x] 音乐电台
* [x] 配置文件
* [x] 通知功能
    
## 伴生项目

1. [anhoder/bubbletea](https://github.com/anhoder/bubbletea): 基于 [bubbletea](https://github.com/charmbracelet/bubbletea) 进行部分定制 
2. [anhoder/bubbles](https://github.com/anhoder/bubbles): 基于 [bubbles](https://github.com/charmbracelet/bubbles) 进行部分定制
3. [anhoder/netease-music](https://github.com/anhoder/netease-music): fork自 [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo) ，在原项目的基础上去除API功能，只保留service、util作为一个独立的包，方便在其他go项目中调用

## 感谢

感谢以下项目及其贡献者们（不限于）：

* [bubbletea](https://github.com/charmbracelet/bubbletea)
* [beep](https://github.com/faiface/beep)
* [musicbox](https://github.com/darknessomi/musicbox)
* [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)
* [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo)
* [gcli](https://github.com/gookit/gcli)
* ...
