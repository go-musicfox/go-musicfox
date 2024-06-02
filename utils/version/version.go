package version

import (
	"io"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

// CheckUpdate 检查更新
func CheckUpdate() (bool, string) {
	response, err := http.Get(types.AppCheckUpdateUrl)
	if err != nil {
		return false, ""
	}
	defer response.Body.Close()

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, ""
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false, ""
	}

	return CompareVersion(tag, types.AppVersion, false), tag
}

func CompareVersion(v1, v2 string, equal bool) bool {
	var (
		v1IsDev = strings.HasSuffix(v1, "-dev")
		v2IsDev = strings.HasSuffix(v2, "-dev")
	)
	if v1IsDev && !v2IsDev {
		return true
	}
	if !v1IsDev && v2IsDev {
		return false
	}

	v1 = strings.Trim(v1, "v")
	v2 = strings.Trim(v2, "v")
	if equal && v1 == v2 {
		return true
	}
	if v1 != "" && v2 == "" {
		return true
	}

	v1Arr := strings.Split(v1, ".")
	v2Arr := strings.Split(v2, ".")
	if len(v1Arr) >= 1 && len(v2Arr) >= 1 {
		if v1Arr[0] > v2Arr[0] {
			return true
		}
		if v1Arr[0] < v2Arr[0] {
			return false
		}
	}

	if len(v1Arr) >= 2 && len(v2Arr) >= 2 {
		if v1Arr[1] > v2Arr[1] {
			return true
		}
		if v1Arr[1] < v2Arr[1] {
			return false
		}
	}

	if len(v1Arr) >= 3 && len(v2Arr) >= 3 {
		if v1Arr[2] > v2Arr[2] {
			return true
		}
		if v1Arr[2] < v2Arr[2] {
			return false
		}
	}
	return false
}
