package migration

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// NeedsMigration 检查是否有任何迁移任务需要执行。
func NeedsMigration() (bool, error) {
	items := getMigrationItems()
	for _, item := range items {
		needs, err := checkItemNeedsMigration(item)
		if err != nil {
			return false, fmt.Errorf("failed to check item '%s': %w", item.description, err)
		}
		if needs {
			return true, nil
		}
	}
	return false, nil
}

// RunAndReport 执行所有必要的迁移，并通过 TUI 向用户报告结果。
// 这个函数是阻塞的，并且在它返回后，主程序应该总是退出。
func RunAndReport() error {
	results := performMigrations()

	if len(results) == 0 {
		return nil
	}

	m := newReportModel(results)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

type migrationStatus int

const (
	StatusSuccess migrationStatus = iota // 迁移成功
	StatusFailure                        // 迁移失败
	StatusSkipped                        // 无需迁移（已是最新）
)

// migrationResult 单项迁移任务的结果
type migrationResult struct {
	Item        migrationItem
	Description string
	Status      migrationStatus
	Error       error
}

// performMigrations 执行所有已定义的迁移任务并返回结果。
func performMigrations() []migrationResult {
	var results []migrationResult
	items := getMigrationItems()

	for _, item := range items {
		needs, checkErr := checkItemNeedsMigration(item)
		if checkErr != nil {
			results = append(results, migrationResult{
				Item:        item,
				Description: item.description,
				Status:      StatusFailure,
				Error:       checkErr,
			})
			continue
		}

		if !needs {
			results = append(results, migrationResult{
				Item:        item,
				Description: item.description,
				Status:      StatusSkipped,
			})
			continue
		}

		oldPath := item.oldPathFn()
		newPath := item.newPathFn()

		var err error
		if item.actionFn != nil {
			err = item.actionFn(oldPath, newPath)
		} else {
			err = movePath(oldPath, newPath)
		}

		if err == nil {
			results = append(results, migrationResult{
				Item:        item,
				Description: item.description,
				Status:      StatusSuccess,
			})
		} else {
			results = append(results, migrationResult{
				Item:        item,
				Description: item.description,
				Status:      StatusFailure,
				Error:       err,
			})
		}
	}
	return results
}

func checkItemNeedsMigration(item migrationItem) (bool, error) {
	if item.displayOnly {
		return false, nil
	}

	oldPath := item.oldPathFn()
	newPath := item.newPathFn()

	if oldPath == newPath {
		return false, nil
	}

	oldExists, err := pathExists(oldPath)
	if err != nil {
		return false, fmt.Errorf("failed to check old path '%s': %w", oldPath, err)
	}
	if !oldExists {
		return false, nil
	}

	newExists, err := pathExists(newPath)
	if err != nil {
		return false, fmt.Errorf("failed to check new path '%s': %w", newPath, err)
	}

	return !newExists, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func movePath(oldPath, newPath string) error {
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory '%s': %w", newDir, err)
	}
	return os.Rename(oldPath, newPath)
}
