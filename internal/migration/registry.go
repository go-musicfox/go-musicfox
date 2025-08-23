package migration

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// migrationItem 定义了一个需要迁移的单元。
type migrationItem struct {
	description string
	oldPathFn   func() string
	newPathFn   func() string
	actionFn    func(oldPath, newPath string) error
	displayOnly bool
}

// getMigrationItems 集中定义所有需要迁移的项目。
func getMigrationItems() []migrationItem {
	items := []migrationItem{
		{
			description: "迁移数据库目录",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, "db") },
			newPathFn:   app.DBDir,
		},
		{
			description: "迁移配置文件",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, types.AppIniFile) },
			newPathFn:   app.ConfigFilePath,
		},
		{
			description: "迁移 cookie",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, "cookie") },
			newPathFn:   func() string { return filepath.Join(app.DataDir(), "cookie") },
		},
		{
			description: "迁移 logo.png",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, "logo.png") },
			newPathFn:   func() string { return filepath.Join(app.DataDir(), "logo.png") },
		},
		{
			description: "音乐缓存路径已变更，请手动处理或清空缓存",
			oldPathFn:   func() string { return filepath.Join(app.CacheDir(), "xxx.mp3") },
			newPathFn:   func() string { return filepath.Join(app.MusicCacheDir(), "xxx.mp3") },
			displayOnly: true,
		},

		{
			description: "音乐下载路径已变更，请手动处理",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, "download") },
			newPathFn:   app.DownloadDir,
			displayOnly: true,
		},
	}

	if runtime.GOOS == "darwin" {
		macOnlyItem := migrationItem{
			description: "[仅macOS] 移除旧路径的通知程序",
			oldPathFn:   func() string { return filepath.Join(oldRootDir, "musicfox-notifier.app") },
			newPathFn:   func() string { return filepath.Join(app.DataDir(), "musicfox-notifier.app") },
			actionFn:    func(oldPath, newPath string) error { return os.RemoveAll(oldPath) },
		}
		items = append(items, macOnlyItem)
	}

	return items
}

var oldRootDir = oldDataRootDir()

func oldDataRootDir() string {
	var projectDir string
	if root := os.Getenv("MUSICFOX_ROOT"); root != "" {
		projectDir = root
	} else {
		configDir, err := os.UserConfigDir()
		if nil != err {
			panic("未获取到本地数据目录：" + err.Error())
		}
		projectDir = filepath.Join(configDir, types.AppLocalDataDir)
	}
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		_ = os.MkdirAll(projectDir, os.ModePerm)
	}
	return projectDir
}
