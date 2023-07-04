package ui

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/utils"
)

type ClearCacheMenu struct {
	DefaultMenu
	ok bool
}

func NewClearCacheMenu() *ClearCacheMenu {
	return &ClearCacheMenu{}
}

func (m *ClearCacheMenu) GetMenuKey() string {
	return "clear_cache"
}

func (m *ClearCacheMenu) MenuViews() []MenuItem {
	if m.ok {
		return []MenuItem{
			{Title: "缓存清除成功"},
		}
	} else {
		return []MenuItem{
			{Title: "缓存清除失败", Subtitle: "请检查配置是否正确"},
		}
	}
}

func (m *ClearCacheMenu) SubMenu(_ *NeteaseModel, _ int) Menu {
	return nil
}

func (m *ClearCacheMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		var (
			cacheDir = utils.GetCacheDir()
			err      error
		)
		if utils.FileOrDirExists(cacheDir) {
			err = filepath.WalkDir(cacheDir, func(path string, d fs.DirEntry, err error) error {
				return os.Remove(path)
			})
		} else {
			err = os.MkdirAll(cacheDir, os.ModePerm)
		}
		if err != nil {
			m.ok = false
			return false
		} else {
			m.ok = true
			return true
		}
	}
}
