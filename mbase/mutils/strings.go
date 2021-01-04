/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:06:55
 * @LastEditTime: 2020-12-17 09:06:56
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"math/rand"
	"strings"
	"time"
)

/*分隔符是seps中指定的任意一个字符*/
func Split(s string, seps string) []string {
	ws := []string{}

	w := ""
	sLen := len(s)
	for i := 0; i < sLen; i++ {
		if strings.IndexByte(seps, s[i]) >= 0 {
			if len(w) > 0 {
				ws = append(ws, w)
				w = ""
			}
		} else {
			w = w + s[i:i+1]
		}
	}
	if len(w) > 0 {
		ws = append(ws, w)
		w = ""
	}

	return ws
}

/*或略大小写的查找*/
func IndexFold(s string, substr string) int {
	lowS := strings.ToLower(s)
	lowSubstr := strings.ToLower(substr)
	return strings.Index(lowS, lowSubstr)
}

func RandomString(size int32, digit bool, lower bool, upper bool) string {
	if size == 0 || !(digit || lower || upper) {
		return ""
	}

	str := ""
	if digit {
		str += "0123456789"
	}

	if lower {
		str += "abcdefghijklmnopqrstuvwxyz"
	}

	if upper {
		str += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}

	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; int32(i) < size; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
