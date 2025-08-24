### 全面支持 XDG 目录规范并提供自动迁移

自 [#453](https://github.com/go-musicfox/go-musicfox/pull/453) 起，引入了 [https://github.com/adrg/xdg](https://github.com/adrg/xdg) 以提供完整的 XDG 规范支持。

#### 1. 路径结构变更

为了实现标准化，文件路径已从原有的 `MUSICFOX_ROOT` (默认为 `~/.config/musicfox`) 迁移至 XDG 标准目录。

**路径对照表如下：**

| 原路径 (`MUSICFOX_ROOT`) | 新路径 (XDG 标准) |
| :--- | :--- |
| `cookie` | `XDG_DATA_HOME/go-musicfox/cookie` |
| `db` | `XDG_DATA_HOME/go-musicfox/db` |
| `logo.png` | `XDG_DATA_HOME/go-musicfox/logo.png` |
| `musicfox.log` | `XDG_STATE_HOME/go-musicfox/musicfox.log` |
| `music_cache` | `XDG_RUNTIME_DIR/beep_playing` |
| `qrcode.png` | `XDG_RUNTIME_DIR/qrcode.png` |
| `qrcode_lastfm.png` | `XDG_RUNTIME_DIR/qrcode_lastfm.png` |
| `download` | `XDG_DOWNLOAD_DIR/go-musicfox` |
| `XDG_CACHE_HOME/go-musicfox` | `XDG_CACHE_HOME/go-musicfox/music_cache` |

#### 2. 无缝自动迁移

考虑到这是一次重大的结构变更，引入了自动迁移功能，以确保用户可以平滑过渡。

*   **触发条件**：当程序启动时，会自动检测旧路径下是否存在 `db` 及 `cookie` 文件，并且新路径下不存在。
*   **迁移行为**：如果满足条件，程序会自动将所有相关文件迁移至新的 XDG 路径，并根据运行结果向用户展示所有路径的变更情况。
*   **已知限制**：迁移是采用文件重命名（`os.Rename`）的方式实现的，因此**不支持跨磁盘设备或文件系统的迁移**。如果您的旧配置文件位于与其他 XDG 目录不同的磁盘分区，需要用户手动处理。

#### 3. 保留 `MUSICFOX_ROOT` 兼容性

为了照顾习惯使用 `MUSICFOX_ROOT` 环境变量自定义存储位置的用户，保留了对该变量的特别支持。

*   当检测到 `MUSICFOX_ROOT` 环境变量时，程序的所有文件（包括数据、缓存、日志等）都将统一存放在该变量指定的目录下，以保持行为的一致性。
*   同时，恢复了此前移除的在 `MUSICFOX_ROOT` 变量下缓存至 `cache` 子目录的能力。

<details>
<summary><b><code>MUSICFOX_ROOT</code> 模式下的目录结构示例</b> (点击展开)</summary>

```tree
/tmp/musocfox_root
├── cache
│   └── music_cache
│       └── 2668056312-2.mp3
├── data
│   ├── cookie
│   ├── db
│   │   └── musicfox.db
│   └── logo.png
├── download
│   ├── 天天天国地獄国 (feat. ななひら & P丸様。)-Aiobahn +81,ななひら,P丸様。.lrc
│   └── 天天天国地獄国 (feat. ななひら & P丸様。)-Aiobahn +81,ななひら,P丸様。.mp3
├── go-musicfox.ini
└── log
    └── musicfox.log
```
</details>

#### 4. 配置优化

*   移除了默认配置文件中预置的下载路径（包括歌曲和歌词），使其默认为空。

---

### 运行预览

相关功能只在 linux 系统下进行过充分测试，尚未在其他系统进行测试

<details>
<summary><b>预览：常规启动 (XDG 路径)</b> (点击展开)</summary>
<img width="1593" height="783" alt="image" src="https://github.com/user-attachments/assets/b47efea9-91ed-4f9e-bdd2-6804aa952173" />
</details>

<details>
<summary><b>预览：在 <code>MUSICFOX_ROOT</code> 变量下启动</b> (点击展开)</summary>
<img width="908" height="614" alt="image" src="https://github.com/user-attachments/assets/d8f4adb8-14f6-48d7-84d2-246e4cef9dd9" />
</details>
