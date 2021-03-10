package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func SplitStr(s string, sep string) []string {
	return strings.Split(s, sep)
}

func StrSplit(s string, sep string) []string {
	return strings.Split(s, sep)
}

func SplitStrN(s string, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func StrSplitN(s string, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func StrFind(s string, f string) int {
	return strings.Index(s, f)
}

func FindStr(s string, f string) int {
	return strings.Index(s, f)
}

func ReplaceStr(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func StrReplace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func TrimStr(s string) string {
	return strings.TrimSpace(s)
}

func StrTrim(s string) string {
	return strings.TrimSpace(s)
}

func StrContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func ContainsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

func JoinStr(a []string, sep string) string {
	return strings.Join(a, sep)
}

func StrJoin(a []string, sep string) string {
	return strings.Join(a, sep)
}

func StrToLower(s string) string {
	return strings.ToLower(s)
}

func ToLowerStr(s string) string {
	return strings.ToLower(s)
}

func StrToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToUpperStr(s string) string {
	return strings.ToUpper(s)
}

func StrTrimRight(s, cutset string) string {
	return strings.TrimRight(s, cutset)
}

func TrimRightStr(s, cutset string) string {
	return strings.TrimRight(s, cutset)
}

func Print(a ...interface{}) (int, error) {
	return fmt.Print(a...)
}

func Println(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}

func Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}

func Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func Atoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func GetInt2(str string, sep string) (int, int, error) {
	split := SplitStr(str, sep)
	if len(split) != 2 {
		return 0, 0, fmt.Errorf("GetInt2 error: %v", str)
	}

	num1, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("GetInt2 error: %v", str)
	}

	num2, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("GetInt2 error: %v", str)
	}
	return int(num1), int(num2), nil
}

func GetIntArr(str string, sep string) ([]int32, error) {
	split := SplitStr(str, sep)
	if len(split) <= 0 {
		return nil, fmt.Errorf("GetInts error: %v", str)
	}

	arr := make([]int32, 0)
	for index, str := range split {
		num, err := strconv.ParseInt(split[index], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("GetInts error: %v", str)
		}
		arr = append(arr, int32(num))
	}

	return arr, nil
}

func MD5Bytes(str []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(str)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func MD5Str(str string) string {
	return MD5Bytes([]byte(str))
}

// 过滤大于3字节的字符(比如说emoji)
func FilterEmoji(content string) string {
	new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new_content += string(value)
		}
	}
	return new_content
}
