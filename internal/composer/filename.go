package composer

import (
	"fmt"
	"regexp"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

const (
	tplDownloadSong  = "downloadSong"
	tplDownloadLyric = "downloadLyric"
)

var defaultDownloadTemplates = map[string]string{
	tplDownloadSong:  "{{.SongName}}-{{.ArtistName}}.{{.FileExt}}",
	tplDownloadLyric: "{{.SongName}}-{{.ArtistName}}.{{.FileExt}}",
}

type FileNameProps struct {
	ItemProps
	FileExt  string
	SongType string // 兼容旧配置
}

// FileNameGenerator 文件名构建器
type FileNameGenerator struct {
	tplManager *TemplateManager
	reg        *regexp.Regexp // 文件名修正
}

// NewFileNameGenerator
func NewFileNameGenerator() *FileNameGenerator {
	tplManager := NewTemplateManager()
	if err := tplManager.LoadFromMap(defaultDownloadTemplates); err != nil {
		panic(fmt.Sprintf("加载默认模板失败: %v", err))
	}
	return &FileNameGenerator{
		tplManager: tplManager,
		reg:        regexp.MustCompile("[<>:\"/\\|?*\000]"),
	}
}

// RegisterSongTemplate 自定义歌曲下载模板
func (m *FileNameGenerator) RegisterSongTemplate(template string) error {
	return m.tplManager.Register(tplDownloadSong, template)
}

// RegisterLyricTemplate 自定义歌词下载模板
func (m *FileNameGenerator) RegisterLyricTemplate(template string) error {
	return m.tplManager.Register(tplDownloadLyric, template)
}

func (m *FileNameGenerator) Song(song structs.Song, fileExt string) (string, error) {
	return m.generator(tplDownloadSong, song, fileExt)
}

func (m *FileNameGenerator) Lyric(song structs.Song, fileExt string) (string, error) {
	return m.generator(tplDownloadLyric, song, fileExt)
}

func (m *FileNameGenerator) generator(name string, song structs.Song, fileExt string) (string, error) {
	props := FileNameProps{
		ItemProps: NewPropsBuilder().WithSong(song).Build(),
		FileExt:   fileExt,
		SongType:  fileExt,
	}
	str, err := m.tplManager.Execute(name, props)
	if err == nil {
		// Windows Linux 均不允许文件名中出现 / \ 替换为 _
		return m.reg.ReplaceAllString(str, "_"), nil
	}
	return "", err
}
