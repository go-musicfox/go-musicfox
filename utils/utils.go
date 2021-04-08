package utils

import (
    "bytes"
    "errors"
    "os"
    "os/exec"
    "os/user"
    "runtime"
    "strings"
)

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

type ResCode uint8
const (
    Success ResCode = iota
    UnknownError
    NeedLogin
    PasswordError
)

// CheckCodeFromResponse check response code
func CheckCodeFromResponse(response map[string]interface{}) ResCode {
    code, ok := response["code"].(float64);
    if !ok {
        return UnknownError
    }

    switch code {
    case 301:
        fallthrough
    case 302:
        return NeedLogin
    case 200:
        return Success
    }

    return PasswordError
}