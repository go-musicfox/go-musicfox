# go-musicfox | **另一个Spotify版 [Spotifox](https://github.com/go-musicfox/spotifox)**

go-musicfox 是用 Go 写的又一款网易云音乐命令行客户端，支持各种音质级别、UnblockNeteaseMusic、Last.fm、MPRIS 和 macOS 交互响应（睡眠暂停、蓝牙耳机连接断开响应和菜单栏控制等）等功能特性。

> UI 基于 [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) 进行了部分定制

![GitHub repo size](https://img.shields.io/github/repo-size/go-musicfox/go-musicfox) ![GitHub](https://img.shields.io/github/license/go-musicfox/go-musicfox) ![Last Tag](https://badgen.net/github/tag/go-musicfox/go-musicfox) ![GitHub last commit](https://badgen.net/github/last-commit/go-musicfox/go-musicfox) ![GitHub All Releases](https://img.shields.io/github/downloads/go-musicfox/go-musicfox/total) ![GitHub stars](https://img.shields.io/github/stars/go-musicfox/go-musicfox?style=social) ![GitHub forks](https://img.shields.io/github/forks/go-musicfox/go-musicfox?style=social)

<p><img src="previews/logo.png" alt="logo" width="256"/></p>

([The icon](https://github.com/go-musicfox/go-musicfox-icon) is based on [kitty-icon](https://github.com/DinkDonk/kitty-icon))

------------------------------
<details>
<summary>

## 预览
</summary>

#### 1. 启动

![启动界面](previews/boot.png)

#### 2. 主界面

![主界面](previews/main.png)

#### 3. 通知

![通知](previews/notify.png)

#### 4. 登录

![登录界面](previews/login.png)

#### 5. 搜索

![搜索界面](previews/search.png)

#### 6. Last.fm 授权

![lastfm](previews/lastfm.png)

#### 7. macOS NowPlaying

![NowPlaying](previews/nowplaying.png)

#### 8. UnblockNeteaseMusic

![UNM](previews/unm.png)

#### 9. macOS 歌词显示

![LyricsX](previews/lyricsX.gif)

> [!IMPORTANT]
> 需要满足以下条件：
> 1. go-musicfox >= v3.7.7
> 2. 下载和安装 [LyricsX 的 go-musicfox 的 fork 版本](https://github.com/go-musicfox/LyricsX/releases/latest)
> 3. 在 LyricsX 设置中，打开`使用系统正在播放的应用`

</details>
<details>
<summary>

## 安装
</summary>

<details>
<summary>

### macOS
</summary>

#### 1. 通过 Homebrew 安装

```sh
$ brew install anhoder/go-musicfox/go-musicfox  // 指定 --head 使用master代码编译安装
```

如果你之前安装过 musicfox，需要使用下列命令重新链接:

```sh
$ brew unlink musicfox && brew link --overwrite go-musicfox
```

#### 2. 直接下载

在 [Release](https://github.com/go-musicfox/go-musicfox/releases/latest) 下载 macOS 的可执行文件。

</details>

<details>
<summary>
  
### Linux
</summary>

#### 1. 使用发行版软件包（推荐）

<details>
<summary>
  
##### Arch Linux
</summary>

###### 从 [AUR](https://aur.archlinux.org/) 安装

```sh
$ paru -S go-musicfox # 下载源代码编译安装
$ paru -S go-musicfox-bin # 下载安装预编译好的二进制
```

###### 从 `archlinuxcn` 安装

首先[添加 archlinuxcn 仓库到系统](https://www.archlinuxcn.org/archlinux-cn-repo-and-mirror/)。

```sh
# pacman -S go-musicfox
```
</details>

<details>
<summary>

##### Fedora Linux
</summary>

###### 从 [Copr](https://copr.fedorainfracloud.org/coprs/poesty/go-musicfox/) 安装。

```sh
$ sudo dnf copr enable poesty/go-musicfox
$ sudo dnf install go-musicfox
```

</details>

<details>
<summary>

##### Debian系发行版（Ubuntu、Deepin、UOS等）
</summary>

###### 从 [星火商店](https://spark-app.store/) 安装。

```sh
$ sudo aptss install go-musicfox  //二进制包部署，同步较慢
$ sudo aptss install go-musicfox-git  //从源码编译，请保持网络通畅
```

</details>

<details>
<summary>

##### Gentoo Linux
</summary>

###### 从 [gentoo-zh Overlay](https://github.com/microcai/gentoo-zh) 安装

```sh
$ eselect repository enable gentoo-zh
$ emerge --sync
$ emerge -a media-sound/go-musicfox
```

</details>

<details>
<summary>

##### NixOS
</summary>

<details>
<summary>
1. flake support
</summary>
下面是一个在nixos配置中使用它的例子

```nix
{
  description = "My configuration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    go-musicfox.url = "github:go-musicfox/go-musicfox";
  };

  outputs = { nixpkgs, go-musicfox, ... }:
    {
      nixosConfigurations = {
        hostname = nixpkgs.lib.nixosSystem
          {
            system = "x86_64-linux";
            modules = [
              {
                nixpkgs.overlays = [ go-musicfox.overlays.default ];
                environment.systemPackages = with pkgs;[
                  go-musicfox
                ];
              }
            ];
          };
      };
    };
}
```

临时运行:

```sh
$ nix run github:go-musicfox/go-musicfox
```

</details>
<details>
<summary>
2. 配置 configuration.nix 或使用 Home Manager（推荐）
</summary>

```nix
# configuration.nix
environment.systemPackages = [
  pkgs.go-musicfox
];

# home manager
home.packages = [
  pkgs.go-musicfox
];
```

</details>
<details>
<summary>
3. 从 <a href="https://search.nixos.org/packages?channel=unstable&show=go-musicfox&from=0&size=50&sort=relevance&type=packages&query=go-musicfox">Nixpkgs </a>安装
</summary>
安装到本地 profile：

```sh
$ nix-env -iA nixos.go-musicfox
```

临时安装：

```sh
$ nix-shell -p go-musicfox
```
</details>

</details>

<details>
<summary>

##### Void Linux
</summary>

从 [void-packages-zh](https://github.com/voidlinux-zh-association/void-packages-zh#readme) 安装。

</details>

<details>
<summary>
  
##### Termux(Android)
</summary>

```sh
$ apt install go-musicfox
```
> 如果遇到卡顿，请切换到mpd播放引擎
</details>
  
#### 2. 通过 Homebrew 安装

```sh
$ brew install anhoder/go-musicfox/go-musicfox  // 指定 --head 使用master代码编译安装
```

如果你之前安装过 musicfox，需要使用下列命令重新链接:

```sh
$ brew unlink musicfox && brew link --overwrite go-musicfox
```

#### 3. 直接下载

在 [Release](https://github.com/go-musicfox/go-musicfox/releases/latest) 下载 Linux 的可执行文件。

</details>

<details>
<summary>

### Windows
</summary>

#### 1. 通过 scoop 安装

```sh
scoop bucket add go-musicfox https://github.com/go-musicfox/go-musicfox.git

scoop install go-musicfox
```

#### 2. 直接下载

在 [Release](https://github.com/go-musicfox/go-musicfox/releases/latest) 下载 Windows 的可执行文件。

</details>

<details>
<summary>
  
### 手动编译
</summary>

注：需要 Go v1.22 及以上版本

前往 [下载 Go ](https://go.dev/dl/)页面选择适合你的 Go 安装包体。

#### 在 Linux 上编译

Linux 需要 `libFLAC-dev` 开发套件

请根据你的发行版，选择适合你的安装命令：

* APT (Debian, Ubuntu)

```sh
$ sudo apt install software-properties-common build-essential
$ sudo add-apt-repository ppa:longsleep/golang-backports //ubuntu默认go语言版本为1.18，需要更新到1.21
$ sudo apt install libflac-dev libasound2-dev golang-go
```

* pacman (Arch)

```sh
$ sudo pacman -S flac
```

* DNF (Fedora)

```sh
$ sudo dnf install flac-devel
```

其他发行版请根据相应文档寻找 `libflac-dev` 开发套件安装说明。

#### 开始编译

```sh
$ git clone https://github.com/go-musicfox/go-musicfox
$ go mod download
$ make # 编译到 bin 目录下
$ make install # 安装到 $GOPATH/bin下
```

</details>
</details>
<details>
<summary>

## 使用
</summary>

```sh
$ musicfox
```

<details>
<summary>
  
### 注意事项
</summary>

- **请务必使用等宽字体，或将配置项 `doubleColumn` 设为 `false`，否则双列显示排版可能会混乱**
- **如果在使用时出现莫名奇妙的光标移动、切歌或暂停等现象，请将配置项 `enableMouseEvent` 设置为 `false`** 
- **本应用不对 macOS 原生终端和 Windows 的命令提示符（CMD）做兼容处理（[#99](https://github.com/go-musicfox/go-musicfox/issues/99)）**   
  > macOS 用户推荐使用 [iTerm2](https://iterm2.com/) 或 [Kitty](https://sw.kovidgoyal.net/kitty/) 
  >
  > Linux 用户推荐使用 [Kitty](https://sw.kovidgoyal.net/kitty/)
  >
  > Windows 用户推荐使用 [Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)，使用体验更佳
- 如果在执行文件时遇到以下错误，说明你的操作系统内不包含 `libFLAC.so.8`：
  ```
  ./musicfox: error while loading shared libraries: libFLAC.so.8: cannot open shared object file: No such file or directory
  ```
  例如 Ubuntu 23.10 及它的衍生版系列，`libFLAC.so.12` 已经将 `libFLAC.so.8` 替换。
  
  遇到这种问题，你可以：
  * 找到已安装的新版 `libFLAC.so`，将其软链为`libFLAC.so.8`: `ln -s /xxx/libFLAC.so /xxx/libFLAC.so.8` （**推荐**）
  * 自行安装 `libflac8` （不推荐）
  * 参照[手动编译](#手动编译)一节自行编译。

  > 这里之所以使用 FLAC8，主要是为了兼容大部分系统，因为FLAC是向前兼容的（也就是说 `≥ 8` 的FLAC都可以使用）

</details>
<details>
<summary>
  
### 快捷键
</summary>

#### 应用内快捷键

|          按键           |       作用       |              备注               |
|:---------------------:|:--------------:|:-----------------------------:|
|   `h`/`H`/`← (左方向)`   |       左        |                               |
|   `l`/`L`/`→ (右方向)`   |       右        |                               |
|   `k`/`K`/`↑ (上方向)`   |       上        |                               |
|   `j`/`J`/`↓ (下方向)`   |       下        |                               |
|          `g`          |     上移到顶部      |                               |
|          `G`          |     下移到底部      |                               |
|        `q`/`Q`        |       退出       |                               |
|     `Space (空格)`      |     暂停/播放      |                               |
|          `[`          |      上一曲       |                               |
|          `]`          |      下一曲       |                               |
|        `-/滚轮下`        |      减小音量      |                               |
|        `=/滚轮上`        |      加大音量      |                               |
| `n`/`N`/`Enter (回车)`  |    进入选中的菜单     |                               |
| `b`/`B`/`Escape (退出)` |     返回上级菜单     |                               |
|        `w`/`W`        |    退出并退出登录     |                               |
|          `p`          |     切换播放方式     |                               |
|          `P`          | 心动模式（仅在歌单中时有效） |                               |
|        `r`/`R`        |    重新渲染 UI     | 如果 UI 界面因为某种原因出现错乱，可以使用这个重新渲染 |
|        `c`/`C`        |     当前播放列表     |                               |
|        `v`/`V`        | 快进 5 s / 10 s  |                               |
|        `x`/`X`        |  快退 1 s / 5 s  |                               |
|          `,`          |    喜欢当前播放歌曲    |                               |
|          `<`          |    喜欢当前选中歌曲    |                               |
|          `.`          |  当前播放歌曲移除出喜欢   |                               |
|          `>`          |  当前选中歌曲移除出喜欢   |                               |
|        `` ` ``        |   当前播放歌曲加入歌单   |                               |
|          `~`          |   当前播放歌曲移出歌单   |                               |
|         `Tab`         |   当前选中歌曲加入歌单   |                               |
|      `Shift+Tab`      |   当前选中歌曲移出歌单   |                               |
|          `>`          |  当前选中歌曲移除出喜欢   |                               |
|          `>`          |  当前选中歌曲移除出喜欢   |                               |
|          `t`          |  标记当前播放歌曲为不喜欢  |                               |
|          `T`          |  标记当前选中歌曲为不喜欢  |                               |
|          `d`          |    下载当前播放歌曲    |                               |
|          `D`          |    下载当前选中歌曲    |                               |
|        `ctrl`+`l`     |    下载当前播放歌曲歌词    |                               |
|          `/`          |     搜索当前列表     |                               |
|          `?`          |      帮助信息      |                               |
|          `a`          |   播放中歌曲的所属专辑   |                               |
|          `A`          |   选中歌曲的所属专辑    |                               |
|          `s`          |   播放中歌曲的所属歌手   |                               |
|          `S`          |   选中歌曲的所属歌手    |                               |
|          `o`          |   网页打开播放中歌曲    |                               |
|          `O`          | 网页打开选中歌曲/专辑... |                               |
|          `e`          |    添加为下一曲播放    |                               |
|          `E`          |   添加到播放列表末尾    |                               |
|          `\`          |  从播放列表删除选中歌曲   |         仅在当前播放列表界面有效          |
|        `;`/`:`        |     收藏选中歌单     |                               |
|        `'`/`"`        |    取消收藏选中歌单    |                               |
|        `u`/`U`        |     清除音乐缓存     |                               |
|        `ctrl`+`u`        |     上一页     |                               |
|        `ctrl`+`d`        |     下一页     |                               |


#### 全局快捷键

默认不设置任何全局快捷键，如果需要请在配置文件中的`global_hotkey`下进行配置，例如：

```ini
[global_hotkey]
# 格式：键=功能 (https://github.com/go-musicfox/go-musicfox/blob/master/internal/ui/event_handler.go#L15)
ctrl+shift+space=toggle
```

> 因为Linux下开启全局快捷键需要安装比较多的依赖，可能你并不需要这个功能，所以Releases中的Linux二进制文件是不支持全局快捷键的
> 
> 如果需要开启，请安装[依赖](https://github.com/go-vgo/robotgo#requirements)后手动进行编译: 
> 
> ```shell
> BUILD_TAGS=enable_global_hotkey make build
> ```

</details>
</details>
<details>
<summary>
  
## 配置文件
</summary>

配置文件路径为用户配置目录下的 `go-musicfox.ini` 文件，详细可参见[配置示例](./utils/filex/embed/go-musicfox.ini)。

> 用户配置目录路径：
> 
> macOS：`$HOME/Library/Application Support/go-musicfox`
>
> Linux：`$XDG_CONFIG_HOME/go-musicfox` 或 `$HOME/.config/go-musicfox`
> 
> Windows：`%AppData%\go-musicfox`
> 
> 你可以通过设置 `MUSICFOX_ROOT` 环境变量来自定义用户配置的存储位置
> 
> 旧版本的 go-musicfox 的默认用户配置目录为 `$HOME/.go-musicfox`（*nix）或 `%USERPROFILE%\.go-musicfox`（Windows），升级到新版本时将自动迁移到上述的新路径

</details>
<details>
<summary>

## CHANGELOG
</summary>

See [CHANGELOG.md](./CHANGELOG.md)

</details>
<details>
<summary>

## 相关项目
</summary>

1. [go-musicfox/bubbletea](https://github.com/go-musicfox/bubbletea)：基于 [bubbletea](https://github.com/charmbracelet/bubbletea) 进行部分定制
2. [go-musicfox/netease-music](https://github.com/go-musicfox/netease-music)：fork 自 [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo) ，在原项目的基础上去除 API 功能，只保留 service 和 util 作为一个独立的包，方便在其他 Go 项目中调用

</details>
<details>
<summary>

## 感谢
</summary>

感谢以下项目及其贡献者们（但不限于）：

* [bubbletea](https://github.com/charmbracelet/bubbletea)
* [beep](https://github.com/faiface/beep)
* [musicbox](https://github.com/darknessomi/musicbox)
* [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)
* [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo)
* [gcli](https://github.com/gookit/gcli)
* ...

感谢 [JetBrains Open Source](https://www.jetbrains.com/zh-cn/opensource/?from=archery) 为项目提供免费的 IDE 授权    
[<img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" width="200"/>](https://www.jetbrains.com/opensource/)

</details>

[![Star History Chart](https://api.star-history.com/svg?repos=go-musicfox/go-musicfox&type=Date)](https://star-history.com/#go-musicfox/go-musicfox&Date)
