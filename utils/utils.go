package utils

import (
	"bytes"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"go-musicfox/pkg/constants"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/buger/jsonparser"
	"go-musicfox/configs"
)

//go:embed embed
var embedDir embed.FS

// GetLocalDataDir 获取本地数据存储目录
func GetLocalDataDir() string {
	var projectDir string
	if root := os.Getenv("MUSICFOX_ROOT"); root != "" {
		projectDir = root
	} else {
		// Home目录
		homeDir, err := Home()
		if nil != err {
			panic("未获取到用户Home目录: " + err.Error())
		}
		projectDir = homeDir + "/" + constants.AppLocalDataDir
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		_ = os.Mkdir(projectDir, os.ModePerm)
	}
	return projectDir
}

// Home 获取当前用户的Home目录
func Home() (string, error) {
	curUser, err := user.Current()
	if nil == err {
		return curUser.HomeDir, nil
	}

	// cross compile support
	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

// IDToBin convert autoincrement ID to []byte
func IDToBin(ID uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ID)
	return b
}

// BinToID convert []byte to autoincrement ID
func BinToID(bin []byte) uint64 {
	ID := binary.BigEndian.Uint64(bin)

	return ID
}

// OpenUrl 打开链接
func OpenUrl(url string) error {
	commands := map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	run, ok := commands[runtime.GOOS]
	if !ok {
		return errors.New(fmt.Sprintf("don't know how to open things on %s platform", runtime.GOOS))
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", run, url)
	} else {
		cmd = exec.Command(run, url)
	}
	return cmd.Start()
}

// LoadIniConfig 加载ini配置信息
func LoadIniConfig() {
	projectDir := GetLocalDataDir()
	configs.ConfigRegistry = configs.NewRegistryFromIniFile(projectDir + "/" + constants.AppIniFile)
}

// CheckUpdate 检查更新
func CheckUpdate() bool {
	response, err := http.Get(constants.AppCheckUpdateUrl)
	if err != nil {
		return false
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false
	}

	curTagArr := strings.Split(strings.Trim(constants.AppVersion, "v"), ".")
	tagArr := strings.Split(strings.Trim(tag, "v"), ".")
	if len(tagArr) >= 1 && len(curTagArr) >= 1 {
		if tagArr[0] > curTagArr[0] {
			return true
		}

		if tagArr[0] < curTagArr[0] {
			return false
		}
	}

	if len(tagArr) >= 2 && len(curTagArr) >= 2 {
		if tagArr[1] > curTagArr[1] {
			return true
		}

		if tagArr[1] < curTagArr[1] {
			return false
		}
	}

	if len(tagArr) >= 3 && len(curTagArr) >= 3 {
		if tagArr[2] > curTagArr[2] {
			return true
		}

		if tagArr[2] < curTagArr[2] {
			return false
		}
	}

	return false
}
