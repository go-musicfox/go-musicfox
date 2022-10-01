package utils

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/cookiestxt"
	"golang.org/x/text/unicode/norm"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/width"
)

func UnGzipV2(gzipData io.Reader) (io.Reader, error) {
	r, err := gzip.NewReader(gzipData)
	if err != nil {
		log.Println("UnGzipV2 error:", err)
		return gzipData, err
	}
	//defer r.Close()
	return r, nil
}
func UnGzip(gzipData []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(gzipData))
	if err != nil {
		log.Println("UnGzip error:", err)
		return gzipData, err
	}
	defer r.Close()
	var decryptECBBytes = gzipData
	decryptECBBytes, err = ioutil.ReadAll(r)
	if err != nil {
		log.Println("UnGzip")
		return gzipData, err
	}
	return decryptECBBytes, nil
}
func LogInterface(i interface{}) string {
	return fmt.Sprintf("%+v", i)
}
func ReplaceAll(str string, expr string, replaceStr string) string {
	reg := regexp.MustCompile(expr)
	str = reg.ReplaceAllString(str, replaceStr)
	return str
}
func ParseJson(data []byte) map[string]interface{} {
	var result map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	d.Decode(&result)
	return result
}
func ParseJsonV2(reader io.Reader) map[string]interface{} {
	var result map[string]interface{}
	d := json.NewDecoder(reader)
	d.UseNumber()
	d.Decode(&result)
	return result
}
func ParseJsonV3(data []byte, dest interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(dest)
}
func ParseJsonV4(reader io.Reader, dest interface{}) error {
	d := json.NewDecoder(reader)
	d.UseNumber()
	return d.Decode(dest)
}
func PanicWrapper(f func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recover panic : ", r)
		}
	}()
	f()
}

func ToJson(object interface{}) string {
	result := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(result)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(object)
	if err != nil {
		log.Println("ToJson Error：", err)
		return "{}"
	}
	return result.String()
}
func Exists(keys []string, h map[string]interface{}) bool {
	for _, key := range keys {
		if !Exist(key, h) {
			return false
		}
	}
	return true
}
func Exist(key string, h map[string]interface{}) bool {
	_, ok := h[key]
	return ok
}
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return path[0 : i+1], nil
}
func MD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
func GenRandomBytes(size int) (blk []byte, err error) {
	blk = make([]byte, size)
	_, err = rand.Read(blk)
	return
}

func CalMatchScoresV2(beMatchedData string, beSplitedData string, matchType string) float32 {
	var score float32 = 0.0
	beMatchedData = width.Narrow.String(strings.ToUpper(strings.TrimSpace(norm.NFC.String(beMatchedData))))
	beSplitedData = width.Narrow.String(strings.ToUpper(strings.TrimSpace(norm.NFC.String(beSplitedData))))
	orginData := beMatchedData
	if len(beMatchedData) < len(beSplitedData) {
		orginData = beSplitedData
		beSplitedData = beMatchedData
		beMatchedData = orginData

	}
	//log.Printf("1:orginData:%s,beMatchedData:%s,beSplitedData:%s\n",orginData,beMatchedData,beSplitedData)
	var keyword []string
	if matchType == "songName" {
		keyword = ParseSongNameKeyWord(beSplitedData)
	} else if matchType == "singerName" {
		keyword = ParseSingerKeyWord(beSplitedData)
	}

	for _, key := range keyword {
		beMatchedData = strings.Replace(beMatchedData, key, "", 1)
	}
	//log.Printf("2:orginData:%s,beMatchedData:%s,beSplitedData:%s\n",orginData,beMatchedData,beSplitedData)
	if beMatchedData == orginData {
		return 0.0
	}
	//beMatchedData = ReplaceAll(beMatchedData, "[`~!@#$%^&*()_\\-+=|{}':;',\\[\\]\\\\.<>/?~！@#￥%……&*（）——+|{}【】‘；：”“’。，、？]", "")
	//beMatchedData = strings.ReplaceAll(beMatchedData, " ", "")
	score = 1 - (float32(len(beMatchedData)) / (float32(len(orginData))))
	return score
}

func CalMatchScores(beMatchedData string, keyword []string) float32 {
	var score float32 = 0.0
	beMatchedData = width.Narrow.String(strings.ToUpper(beMatchedData))
	orginData := beMatchedData
	for _, key := range keyword {
		beMatchedData = strings.Replace(beMatchedData, key, "", 1)
	}
	if beMatchedData == orginData {
		return 0.0
	}
	//beMatchedData = ReplaceAll(beMatchedData, "[`~!@#$%^&*()_\\-+=|{}':;',\\[\\]\\\\.<>/?~！@#￥%……&*（）——+|{}【】‘；：”“’。，、？]", "")
	//beMatchedData = strings.ReplaceAll(beMatchedData, " ", "")
	score = 1 - (float32(len(beMatchedData)) / (float32(len(orginData))))
	return score
}

var leftPairedSymbols = width.Narrow.String("(（《<[{「【『\"'")
var rightPairedSymbols = width.Narrow.String(")）》>]}」】』\"'")

func parsePairedSymbols(data string, sub string, substr []string, keyword map[string]int) string {
	data = strings.TrimSpace(data)
	leftIndex := strings.Index(leftPairedSymbols, sub)
	rightIndex := strings.Index(rightPairedSymbols, sub)
	subIndex := 0
	for index, key := range substr {
		if key == sub {
			subIndex = index
			break
		}
	}
	index := -1
	if leftIndex != -1 {
		index = leftIndex
	} else if rightIndex != -1 {
		index = rightIndex
	}
	if index != -1 {
		leftSymbol := leftPairedSymbols[index : index+len(sub)]
		rightSymbol := rightPairedSymbols[index : index+len(sub)]
		leftCount := strings.Count(data, leftSymbol)
		rightCount := strings.Count(data, rightSymbol)
		if leftCount == rightCount && leftCount > 0 {
			for i := 0; i < leftCount; i++ {
				lastLeftIndex := strings.LastIndex(data, leftSymbol)
				matchedRightIndex := strings.Index(data[lastLeftIndex:], rightSymbol)
				if matchedRightIndex == -1 {
					continue
				}
				key := strings.TrimSpace(data[lastLeftIndex+len(leftSymbol) : lastLeftIndex+matchedRightIndex])
				data = data[:lastLeftIndex] + " " + data[lastLeftIndex+matchedRightIndex+len(rightSymbol):]
				substr2 := substr[subIndex+1:]
				parseKeyWord(key, substr2, keyword)

			}
		}
	}
	return data
}
func parseKeyWord(data string, substr []string, keyword map[string]int) {
	if len(data) == 0 {
		return
	}
	data = strings.TrimSpace(data)
	for _, sub := range substr {
		if strings.Contains(data, sub) {
			if strings.Contains(leftPairedSymbols, sub) || strings.Contains(rightPairedSymbols, sub) {
				data = parsePairedSymbols(data, sub, substr, keyword)
			} else {
				splitData := strings.Split(data, sub)
				for _, key := range splitData {
					newKey := strings.TrimSpace(key)
					parseKeyWord(newKey, substr, keyword)
					data = strings.Replace(data, key, "", 1)

				}
				data = strings.ReplaceAll(data, sub, "")
			}

		}
	}
	data = strings.TrimSpace(data)
	if len(data) > 0 {
		if strings.EqualFold(data, "LIVE版") {
			data = "LIVE"
		}
		keyword[data] = 1
	}
}

type ByLenSort []string

func (a ByLenSort) Len() int {
	return len(a)
}

func (a ByLenSort) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}

func (a ByLenSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func ParseSongNameKeyWord(data string) []string {
	var keyword = make(map[string]int)
	if len(data) > 0 {
		data = width.Narrow.String(strings.ToUpper(data))
		substr := []string{"(", "[", "{", "<", "《", "「", "【", "『", "+", "/", ":", ",", "｡", " "}
		for index, sub := range substr {
			substr[index] = width.Narrow.String(sub)
		}
		parseKeyWord(data, substr, keyword)
	}
	keys := make([]string, 0, len(keyword))
	for k, _ := range keyword {
		keys = append(keys, k)
	}
	sort.Sort(ByLenSort(keys))
	return keys
}
func ParseSingerKeyWord(data string) []string {
	var keyword = make(map[string]int)
	if len(data) > 0 {
		data = strings.TrimSpace(strings.ToUpper(data))
		substr := []string{"、", ",", " ", "､"}
		parseKeyWord(data, substr, keyword)

	}
	keys := make([]string, 0, len(keyword))
	for k, _ := range keyword {
		keys = append(keys, k)
	}
	sort.Sort(ByLenSort(keys))
	return keys
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func ParseCookies(file string) []*http.Cookie {
	fl, err := os.Open(file)
	if err != nil {
		fmt.Println(file, err)
		return nil
	}
	defer fl.Close()
	r := bufio.NewReader(fl)
	cl, err := cookiestxt.Parse(r)
	if err != nil {
		return nil
	}
	return cl
}

func Combination(WordList []string) (result []string) {
	if WordList == nil || len(WordList) == 0 {
		return []string{
			"",
		}
	}
	if len(WordList) == 1 {
		return []string{
			"",
			WordList[0],
		}
	}
	tmp := Combination(WordList[1:])
	result = append(result, tmp...)
	for _, v := range tmp {
		result = append(result, WordList[0]+" "+v)
	}
	return result
}
