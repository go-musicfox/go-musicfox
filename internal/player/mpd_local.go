package player

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const mpdStagingDir = "musicfox-staging"

// mpdConfigMusicDirectory reads "music_directory" from the MPD configuration
// file. It is only used for TCP connections because MPD hides local paths from
// non-local clients.
func mpdConfigMusicDirectory(configFile string) string {
	if configFile == "" {
		return ""
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return ""
	}

	var dir string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "music_directory" {
			continue
		}
		value := parseMpdConfigValue(strings.TrimSpace(strings.TrimPrefix(line, fields[0])))
		if value == "" {
			continue
		}
		dir = expandHomeDir(value)
	}
	return dir
}

// parseMpdConfigValue extracts the first value from an MPD config line,
// honoring double or single quotes and stopping at an inline comment.
func parseMpdConfigValue(s string) string {
	if s == "" {
		return ""
	}
	switch s[0] {
	case '"', '\'':
		quote := s[0]
		if end := strings.IndexByte(s[1:], quote); end >= 0 {
			return s[1 : end+1]
		}
		return ""
	}
	if end := strings.IndexAny(s, " \t#"); end >= 0 {
		return s[:end]
	}
	return s
}

func expandHomeDir(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// relativeLocalURI returns path relative to base using MPD's slash separator
// when path resides inside base.
func relativeLocalURI(base, path string) (string, bool) {
	baseReal, err := filepath.EvalSymlinks(base)
	if err != nil {
		baseReal = base
	}
	pathReal, err := filepath.EvalSymlinks(path)
	if err != nil {
		pathReal = path
	}
	rel, err := filepath.Rel(baseReal, pathReal)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// stageLocalFile makes path reachable from MPD's music directory and returns
// the URI to add. Files already inside musicDir are used in place; files
// elsewhere are hard-linked (or copied) into a staging directory.
func stageLocalFile(musicDir, path string) (uri, staged string, err error) {
	if uri, ok := relativeLocalURI(musicDir, path); ok {
		return uri, "", nil
	}
	if !filepath.IsAbs(musicDir) {
		return "", "", fmt.Errorf("cannot stage %q: mpd music_directory %q is not local", path, musicDir)
	}
	stageDir := filepath.Join(musicDir, mpdStagingDir)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", "", err
	}
	target := filepath.Join(stageDir, filepath.Base(path))
	_ = os.Remove(target)
	if err := os.Link(path, target); err != nil {
		if err := copyLocalFile(path, target); err != nil {
			return "", "", err
		}
	}
	uri, ok := relativeLocalURI(musicDir, target)
	if !ok {
		_ = os.Remove(target)
		return "", "", fmt.Errorf("staged file %q is outside mpd music directory", target)
	}
	return uri, target, nil
}

func copyLocalFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(dst)
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
